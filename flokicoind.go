// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"runtime/trace"

	"github.com/flokiorg/go-flokicoin/blockchain/indexers"
	"github.com/flokiorg/go-flokicoin/database"
	"github.com/flokiorg/go-flokicoin/limits"
	"github.com/flokiorg/go-flokicoin/ossec"
)

const (
	// blockDbNamePrefix is the prefix for the block database name.  The
	// database type is appended to this value to form the full block
	// database name.
	blockDbNamePrefix = "blocks"
)

var (
	cfg *config
)

// winServiceMain is only invoked on Windows.  It detects when flokicoind is running
// as a service and reacts accordingly.
var winServiceMain func() (bool, error)

// fmain is the real main function for flokicoind.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func fmain(serverChan chan<- *server) error {

	fmt.Println(Orange + Bright + bonerArt + Reset)

	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg
	defer func() {
		if logRotator != nil {
			logRotator.Close()
		}
	}()

	// Get a channel that will be closed when a shutdown signal has been
	// triggered either from an OS signal such as SIGINT (Ctrl+C) or from
	// another subsystem such as the RPC server.
	interrupt := interruptListener()
	defer flcdLog.Info("Shutdown complete")

	// Show version at startup.
	flcdLog.Infof("Version %s", version())

	// Enable http profiling server if requested.
	if cfg.Profile != "" {
		go func() {
			listenAddr := net.JoinHostPort("", cfg.Profile)
			flcdLog.Infof("Profile server listening on %s", listenAddr)
			profileRedirect := http.RedirectHandler("/debug/pprof",
				http.StatusSeeOther)
			http.Handle("/", profileRedirect)
			flcdLog.Errorf("%v", http.ListenAndServe(listenAddr, nil))
		}()
	}

	// Write cpu profile if requested.
	if cfg.CPUProfile != "" {
		f, err := os.Create(cfg.CPUProfile)
		if err != nil {
			flcdLog.Errorf("Unable to create cpu profile: %v", err)
			return err
		}
		pprof.StartCPUProfile(f)
		defer f.Close()
		defer pprof.StopCPUProfile()
	}

	// Write mem profile if requested.
	if cfg.MemoryProfile != "" {
		f, err := os.Create(cfg.MemoryProfile)
		if err != nil {
			flcdLog.Errorf("Unable to create memory profile: %v", err)
			return err
		}
		defer f.Close()
		defer pprof.WriteHeapProfile(f)
		defer runtime.GC()
	}

	// Write execution trace if requested.
	if cfg.TraceProfile != "" {
		f, err := os.Create(cfg.TraceProfile)
		if err != nil {
			flcdLog.Errorf("Unable to create execution trace: %v", err)
			return err
		}
		trace.Start(f)
		defer f.Close()
		defer trace.Stop()
	}

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}

	// Load the block database.
	db, err := loadBlockDB()
	if err != nil {
		flcdLog.Errorf("%v", err)
		return err
	}
	defer func() {
		// Ensure the database is sync'd and closed on shutdown.
		flcdLog.Infof("Gracefully shutting down the database...")
		db.Close()
	}()

	// Return now if an interrupt signal was triggered.
	if interruptRequested(interrupt) {
		return nil
	}

	// Drop indexes and exit if requested.
	//
	// NOTE: The order is important here because dropping the tx index also
	// drops the address index since it relies on it.
	if cfg.DropAddrIndex {
		if err := indexers.DropAddrIndex(db, interrupt); err != nil {
			flcdLog.Errorf("%v", err)
			return err
		}

		return nil
	}
	if cfg.DropTxIndex {
		if err := indexers.DropTxIndex(db, interrupt); err != nil {
			flcdLog.Errorf("%v", err)
			return err
		}

		return nil
	}
	if cfg.DropCfIndex {
		if err := indexers.DropCfIndex(db, interrupt); err != nil {
			flcdLog.Errorf("%v", err)
			return err
		}

		return nil
	}

	// Check if the database had previously been pruned.  If it had been, it's
	// not possible to newly generate the tx index and addr index.
	var beenPruned bool
	db.View(func(dbTx database.Tx) error {
		beenPruned, err = dbTx.BeenPruned()
		return err
	})
	if err != nil {
		flcdLog.Errorf("%v", err)
		return err
	}
	if beenPruned && cfg.Prune == 0 {
		err = fmt.Errorf("--prune cannot be disabled as the node has been "+
			"previously pruned. You must delete the files in the datadir: \"%s\" "+
			"and sync from the beginning to disable pruning", cfg.DataDir)
		flcdLog.Errorf("%v", err)
		return err
	}
	if beenPruned && cfg.TxIndex {
		err = fmt.Errorf("--txindex cannot be enabled as the node has been "+
			"previously pruned. You must delete the files in the datadir: \"%s\" "+
			"and sync from the beginning to enable the desired index", cfg.DataDir)
		flcdLog.Errorf("%v", err)
		return err
	}
	if beenPruned && cfg.AddrIndex {
		err = fmt.Errorf("--addrindex cannot be enabled as the node has been "+
			"previously pruned. You must delete the files in the datadir: \"%s\" "+
			"and sync from the beginning to enable the desired index", cfg.DataDir)
		flcdLog.Errorf("%v", err)
		return err
	}
	// If we've previously been pruned and the cfindex isn't present, it means that the
	// user wants to enable the cfindex after the node has already synced up and been
	// pruned.
	if beenPruned && !indexers.CfIndexInitialized(db) && !cfg.NoCFilters {
		err = fmt.Errorf("compact filters cannot be enabled as the node has been "+
			"previously pruned. You must delete the files in the datadir: \"%s\" "+
			"and sync from the beginning to enable the desired index. You may "+
			"use the --nocfilters flag to start the node up without the compact "+
			"filters", cfg.DataDir)
		flcdLog.Errorf("%v", err)
		return err
	}
	// If the user wants to disable the cfindex and is pruned or has enabled pruning, force
	// the user to either drop the cfindex manually or restart the node without the --nocfilters
	// flag.
	if (beenPruned || cfg.Prune != 0) && indexers.CfIndexInitialized(db) && cfg.NoCFilters {
		err = fmt.Errorf("--nocfilters flag was given but the compact filters have " +
			"previously been enabled on this node and the index data currently " +
			"exists in the database. The node has also been previously pruned and " +
			"the database would be left in an inconsistent state if the compact " +
			"filters don't get indexed now. To disable compact filters, please drop the " +
			"index completely with the --dropcfindex flag and restart the node. " +
			"To keep the compact filters, restart the node without the --nocfilters " +
			"flag")
		flcdLog.Errorf("%v", err)
		return err
	}

	// Enforce removal of txindex and addrindex if user requested pruning.
	// This is to require explicit action from the user before removing
	// indexes that won't be useful when block files are pruned.
	//
	// NOTE: The order is important here because dropping the tx index also
	// drops the address index since it relies on it.  We explicitly make the
	// user drop both indexes if --addrindex was enabled previously.
	if cfg.Prune != 0 && indexers.AddrIndexInitialized(db) {
		err = fmt.Errorf("--prune flag may not be given when the address index " +
			"has been initialized. Please drop the address index with the " +
			"--dropaddrindex flag before enabling pruning")
		flcdLog.Errorf("%v", err)
		return err
	}
	if cfg.Prune != 0 && indexers.TxIndexInitialized(db) {
		err = fmt.Errorf("--prune flag may not be given when the transaction index " +
			"has been initialized. Please drop the transaction index with the " +
			"--droptxindex flag before enabling pruning")
		flcdLog.Errorf("%v", err)
		return err
	}

	// The config file is already created if it did not exist and the log
	// file has already been opened by now so we only need to allow
	// creating rpc cert and key files if they don't exist.
	unveilx(cfg.RPCKey, "rwc")
	unveilx(cfg.RPCCert, "rwc")
	unveilx(cfg.DataDir, "rwc")

	// drop unveil and tty
	pledgex("stdio rpath wpath cpath flock dns inet")

	// Create server and start it.
	server, err := newServer(cfg.Listeners, cfg.AgentBlacklist,
		cfg.AgentWhitelist, db, activeNetParams.Params, interrupt)
	if err != nil {
		// TODO: this logging could do with some beautifying.
		flcdLog.Errorf("Unable to start server on %v: %v",
			cfg.Listeners, err)
		return err
	}
	defer func() {
		flcdLog.Infof("Gracefully shutting down the server...")
		server.Stop()
		server.WaitForShutdown()
		srvrLog.Infof("Server shutdown complete")
	}()
	server.Start()
	if serverChan != nil {
		serverChan <- server
	}

	// Wait until the interrupt signal is received from an OS signal or
	// shutdown is requested through one of the subsystems such as the RPC
	// server.
	<-interrupt
	return nil
}

// removeRegressionDB removes the existing regression test database if running
// in regression test mode and it already exists.
func removeRegressionDB(dbPath string) error {
	// Don't do anything if not in regression test mode.
	if !cfg.RegressionTest {
		return nil
	}

	// Remove the old regression test database if it already exists.
	fi, err := os.Stat(dbPath)
	if err == nil {
		flcdLog.Infof("Removing regression test database from '%s'", dbPath)
		if fi.IsDir() {
			err := os.RemoveAll(dbPath)
			if err != nil {
				return err
			}
		} else {
			err := os.Remove(dbPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// dbPath returns the path to the block database given a database type.
func blockDbPath(dbType string) string {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + dbType
	if dbType == "sqlite" {
		dbName = dbName + ".db"
	}
	dbPath := filepath.Join(cfg.DataDir, dbName)
	return dbPath
}

// warnMultipleDBs shows a warning if multiple block database types are detected.
// This is not a situation most users want.  It is handy for development however
// to support multiple side-by-side databases.
func warnMultipleDBs() {
	// This is intentionally not using the known db types which depend
	// on the database types compiled into the binary since we want to
	// detect legacy db types as well.
	dbTypes := []string{"ffldb", "leveldb", "sqlite"}
	duplicateDbPaths := make([]string, 0, len(dbTypes)-1)
	for _, dbType := range dbTypes {
		if dbType == cfg.DbType {
			continue
		}

		// Store db path as a duplicate db if it exists.
		dbPath := blockDbPath(dbType)
		if fileExists(dbPath) {
			duplicateDbPaths = append(duplicateDbPaths, dbPath)
		}
	}

	// Warn if there are extra databases.
	if len(duplicateDbPaths) > 0 {
		selectedDbPath := blockDbPath(cfg.DbType)
		flcdLog.Warnf("WARNING: There are multiple block chain databases "+
			"using different database types.\nYou probably don't "+
			"want to waste disk space by having more than one.\n"+
			"Your current database is located at [%v].\nThe "+
			"additional database is located at %v", selectedDbPath,
			duplicateDbPaths)
	}
}

// loadBlockDB loads (or creates when needed) the block database taking into
// account the selected database backend and returns a handle to it.  It also
// contains additional logic such warning the user if there are multiple
// databases which consume space on the file system and ensuring the regression
// test database is clean when in regression test mode.
func loadBlockDB() (database.DB, error) {
	// The memdb backend does not have a file path associated with it, so
	// handle it uniquely.  We also don't want to worry about the multiple
	// database type warnings when running with the memory database.
	if cfg.DbType == "memdb" {
		flcdLog.Infof("Creating block database in memory.")
		db, err := database.Create(cfg.DbType)
		if err != nil {
			return nil, err
		}
		return db, nil
	}

	warnMultipleDBs()

	// The database name is based on the database type.
	dbPath := blockDbPath(cfg.DbType)

	// The regression test is special in that it needs a clean database for
	// each run, so remove it now if it already exists.
	removeRegressionDB(dbPath)

	flcdLog.Infof("Loading block database from '%s'", dbPath)
	db, err := database.Open(cfg.DbType, dbPath, activeNetParams.Net)
	if err != nil {
		// Return the error if it's not because the database doesn't
		// exist.
		if dbErr, ok := err.(database.Error); !ok || dbErr.ErrorCode !=
			database.ErrDbDoesNotExist {

			return nil, err
		}

		// Create the db if it does not exist.
		err = os.MkdirAll(cfg.DataDir, 0700)
		if err != nil {
			return nil, err
		}
		db, err = database.Create(cfg.DbType, dbPath, activeNetParams.Net)
		if err != nil {
			return nil, err
		}
	}

	flcdLog.Info("Block database loaded")
	return db, nil
}

func unveilx(path string, perms string) {
	err := ossec.Unveil(path, perms)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unveil failed: %v\n", err)
		os.Exit(1)
	}
}

func pledgex(promises string) {
	err := ossec.PledgePromises(promises)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pledge failed: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	pledgex("unveil stdio id rpath wpath cpath flock dns inet tty")
}

func main() {
	// If GOGC is not explicitly set, override GC percent.
	if os.Getenv("GOGC") == "" {
		// Block and transaction processing can cause bursty allocations.  This
		// limits the garbage collector from excessively overallocating during
		// bursts.  This value was arrived at with the help of profiling live
		// usage.
		debug.SetGCPercent(10)
	}

	// Up some limits.
	if err := limits.SetLimits(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set limits: %v\n", err)
		os.Exit(1)
	}

	// Call serviceMain on Windows to handle running as a service.  When
	// the return isService flag is true, exit now since we ran as a
	// service.  Otherwise, just fall through to normal operation.
	if runtime.GOOS == "windows" {
		isService, err := winServiceMain()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if isService {
			os.Exit(0)
		}
	}

	// Work around defer not working after os.Exit()
	if err := fmain(nil); err != nil {
		os.Exit(1)
	}
}
