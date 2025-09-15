package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/flokiorg/go-flokicoin/blockchain"
	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/database"
	_ "github.com/flokiorg/go-flokicoin/database/ffldb" // registers ffldb driver
	"github.com/flokiorg/go-flokicoin/wire"

	"github.com/dsnet/compress/bzip2"

	flags "github.com/jessevdk/go-flags"
)

// Command: testexport
// Usage examples:
//   testexport --db "/home/user/.flcd/data/mainnet/blocks_ffldb" \
//             --network mainnet \
//             --out ./export \
//             --range 0-10 --range 18-90 --range 250000

type Options struct {
	DBPath  string   `short:"d" long:"db"      description:"Path to ffldb directory (…/blocks_ffldb)" required:"true"`
	Network string   `short:"n" long:"network" description:"Network: mainnet | testnet3 | simnet | regtest" default:"mainnet"`
	OutDir  string   `short:"o" long:"outdir"  description:"Output directory" required:"true"`
	OutFile string   `short:"f" long:"outfile" description:"Output filename (e.g., blk_3A.dat.bz2)" required:"true"`
	Ranges  []string `short:"r" long:"range"   description:"Height range N-M or single N (repeatable)"`
	Verbose bool     `short:"v" long:"verbose" description:"Verbose logging"`
	NoBZip2 bool     `long:"no-bz2"            description:"Write uncompressed (for debug); file extension should not be .bz2"`
}

func main() {
	var opts Options
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	netID, params := mustNetwork(opts.Network)

	db, err := database.Open("ffldb", opts.DBPath, netID)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Chain view for height-based fetches
	chain, err := blockchain.New(&blockchain.Config{
		DB:          db,
		ChainParams: params,
		TimeSource:  blockchain.NewMedianTime(),
	})
	if err != nil {
		log.Fatalf("init blockchain: %v", err)
	}

	// Parse & dedupe heights
	best := chain.BestSnapshot().Height
	heights, err := parseHeights(opts.Ranges, best)
	if err != nil {
		log.Fatalf("ranges: %v", err)
	}
	if len(heights) == 0 {
		log.Fatalf("no heights to export (add --range)")
	}

	// Prepare output stream
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", opts.OutDir, err)
	}
	outPath := filepath.Join(opts.OutDir, opts.OutFile)
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("create %s: %v", outPath, err)
	}
	defer outFile.Close()

	var writer = any(outFile)
	var bz2w *bzip2.Writer
	if !opts.NoBZip2 {
		bz2w, err = bzip2.NewWriter(outFile, &bzip2.WriterConfig{Level: bzip2.BestCompression})
		if err != nil {
			log.Fatalf("bzip2 writer: %v", err)
		}
		defer bz2w.Close()
		writer = bz2w
	}

	network := uint32(netID) // this is what your loader checks against

	start := time.Now()
	exported := 0

	for _, h := range heights {
		blk, err := chain.BlockByHeight(h)
		if err != nil {
			log.Fatalf("fetch height %d: %v", h, err)
		}

		// Serialize block to raw wire bytes
		var buf bytes.Buffer
		if err := blk.MsgBlock().Serialize(&buf); err != nil {
			log.Fatalf("serialize height %d: %v", h, err)
		}

		raw := buf.Bytes()

		// Write [uint32 network][uint32 blocklen][raw]
		// Note: your loader uses LittleEndian.
		if err := binary.Write(writer.(interface{ Write([]byte) (int, error) }), binary.LittleEndian, network); err != nil {
			log.Fatalf("write network: %v", err)
		}
		blockLen := uint32(len(raw))
		lenLE := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenLE, blockLen)
		if _, err := writer.(interface{ Write([]byte) (int, error) }).Write(lenLE); err != nil {
			log.Fatalf("write length: %v", err)
		}
		if _, err := writer.(interface{ Write([]byte) (int, error) }).Write(raw); err != nil {
			log.Fatalf("write block: %v", err)
		}

		exported++
		if opts.Verbose {
			fmt.Printf("wrote height=%d bits=%d hash=%s bytes=%d\n", h, blk.MsgBlock().Header.Bits, blk.Hash().String(), len(raw))
		}
	}

	if bz2w != nil {
		if err := bz2w.Close(); err != nil {
			log.Fatalf("finalize bz2: %v", err)
		}
	}
	if err := outFile.Sync(); err != nil {
		log.Fatalf("sync: %v", err)
	}

	fmt.Printf("Done. %d blocks → %s in %s\n", exported, outPath, time.Since(start).Round(time.Millisecond))
}

// parseHeights: accepts N or N-M, inclusive, multiple times, deduped & sorted.
func parseHeights(ranges []string, tip int32) ([]int32, error) {
	if len(ranges) == 0 {
		return nil, nil
	}
	re := regexp.MustCompile(`^\s*(\d+)\s*(?:-\s*(\d+)\s*)?$`)
	set := make(map[int32]struct{})
	for _, s := range ranges {
		m := re.FindStringSubmatch(s)
		if m == nil {
			return nil, fmt.Errorf("invalid --range %q (use N or N-M)", s)
		}
		a64, _ := strconv.ParseInt(m[1], 10, 32)
		var b64 int64
		if m[2] != "" {
			b64, _ = strconv.ParseInt(m[2], 10, 32)
		} else {
			b64 = a64
		}
		a, b := int32(a64), int32(b64)
		if a > b {
			a, b = b, a
		}
		if a < 0 || b > tip {
			return nil, fmt.Errorf("range %d-%d out of bounds (tip=%d)", a, b, tip)
		}
		for h := a; h <= b; h++ {
			set[h] = struct{}{}
		}
	}
	out := make([]int32, 0, len(set))
	for h := range set {
		out = append(out, h)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

func mustNetwork(s string) (wire.FlokicoinNet, *chaincfg.Params) {
	switch strings.ToLower(s) {
	case "mainnet", "main":
		return wire.MainNet, &chaincfg.MainNetParams
	case "testnet3", "testnet":
		return wire.TestNet3, &chaincfg.TestNet3Params
	case "simnet":
		return wire.SimNet, &chaincfg.SimNetParams
	default:
		log.Fatalf("unknown network %q (use mainnet|testnet3|simnet|regtest)", s)
		return 0, nil
	}
}
