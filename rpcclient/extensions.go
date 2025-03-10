// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/chainjson"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/wire"
)

// FutureDebugLevelResult is a future promise to deliver the result of a
// DebugLevelAsync RPC invocation (or an applicable error).
type FutureDebugLevelResult chan *Response

// Receive waits for the Response promised by the future and returns the result
// of setting the debug logging level to the passed level specification or the
// list of of the available subsystems for the special keyword 'show'.
func (r FutureDebugLevelResult) Receive() (string, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return "", err
	}

	// Unmashal the result as a string.
	var result string
	err = json.Unmarshal(res, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

// DebugLevelAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See DebugLevel for the blocking version and more details.
//
// NOTE: This is a flokicoind extension.
func (c *Client) DebugLevelAsync(levelSpec string) FutureDebugLevelResult {
	cmd := chainjson.NewDebugLevelCmd(levelSpec)
	return c.SendCmd(cmd)
}

// DebugLevel dynamically sets the debug logging level to the passed level
// specification.
//
// The levelspec can be either a debug level or of the form:
//
//	<subsystem>=<level>,<subsystem2>=<level2>,...
//
// Additionally, the special keyword 'show' can be used to get a list of the
// available subsystems.
//
// NOTE: This is a flokicoind extension.
func (c *Client) DebugLevel(levelSpec string) (string, error) {
	return c.DebugLevelAsync(levelSpec).Receive()
}

// FutureCreateEncryptedWalletResult is a future promise to deliver the error
// result of a CreateEncryptedWalletAsync RPC invocation.
type FutureCreateEncryptedWalletResult chan *Response

// Receive waits for and returns the error Response promised by the future.
func (r FutureCreateEncryptedWalletResult) Receive() error {
	_, err := ReceiveFuture(r)
	return err
}

// CreateEncryptedWalletAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See CreateEncryptedWallet for the blocking version and more details.
//
// NOTE: This is a walletd extension.
func (c *Client) CreateEncryptedWalletAsync(passphrase string) FutureCreateEncryptedWalletResult {
	cmd := chainjson.NewCreateEncryptedWalletCmd(passphrase)
	return c.SendCmd(cmd)
}

// CreateEncryptedWallet requests the creation of an encrypted wallet.  Wallets
// managed by walletd are only written to disk with encrypted private keys,
// and generating wallets on the fly is impossible as it requires user input for
// the encryption passphrase.  This RPC specifies the passphrase and instructs
// the wallet creation.  This may error if a wallet is already opened, or the
// new wallet cannot be written to disk.
//
// NOTE: This is a walletd extension.
func (c *Client) CreateEncryptedWallet(passphrase string) error {
	return c.CreateEncryptedWalletAsync(passphrase).Receive()
}

// FutureListAddressTransactionsResult is a future promise to deliver the result
// of a ListAddressTransactionsAsync RPC invocation (or an applicable error).
type FutureListAddressTransactionsResult chan *Response

// Receive waits for the Response promised by the future and returns information
// about all transactions associated with the provided addresses.
func (r FutureListAddressTransactionsResult) Receive() ([]chainjson.ListTransactionsResult, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal the result as an array of listtransactions objects.
	var transactions []chainjson.ListTransactionsResult
	err = json.Unmarshal(res, &transactions)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// ListAddressTransactionsAsync returns an instance of a type that can be used
// to get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See ListAddressTransactions for the blocking version and more details.
//
// NOTE: This is a flokicoind extension.
func (c *Client) ListAddressTransactionsAsync(addresses []chainutil.Address, account string) FutureListAddressTransactionsResult {
	// Convert addresses to strings.
	addrs := make([]string, 0, len(addresses))
	for _, addr := range addresses {
		addrs = append(addrs, addr.EncodeAddress())
	}
	cmd := chainjson.NewListAddressTransactionsCmd(addrs, &account)
	return c.SendCmd(cmd)
}

// ListAddressTransactions returns information about all transactions associated
// with the provided addresses.
//
// NOTE: This is a walletd extension.
func (c *Client) ListAddressTransactions(addresses []chainutil.Address, account string) ([]chainjson.ListTransactionsResult, error) {
	return c.ListAddressTransactionsAsync(addresses, account).Receive()
}

// FutureGetBestBlockResult is a future promise to deliver the result of a
// GetBestBlockAsync RPC invocation (or an applicable error).
type FutureGetBestBlockResult chan *Response

// Receive waits for the Response promised by the future and returns the hash
// and height of the block in the longest (best) chain.
func (r FutureGetBestBlockResult) Receive() (*chainhash.Hash, int32, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, 0, err
	}

	// Unmarshal result as a getbestblock result object.
	var bestBlock chainjson.GetBestBlockResult
	err = json.Unmarshal(res, &bestBlock)
	if err != nil {
		return nil, 0, err
	}

	// Convert to hash from string.
	hash, err := chainhash.NewHashFromStr(bestBlock.Hash)
	if err != nil {
		return nil, 0, err
	}

	return hash, bestBlock.Height, nil
}

// GetBestBlockAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetBestBlock for the blocking version and more details.
//
// NOTE: This is a flokicoind extension.
func (c *Client) GetBestBlockAsync() FutureGetBestBlockResult {
	cmd := chainjson.NewGetBestBlockCmd()
	return c.SendCmd(cmd)
}

// GetBestBlock returns the hash and height of the block in the longest (best)
// chain.
//
// NOTE: This is a flokicoind extension.
func (c *Client) GetBestBlock() (*chainhash.Hash, int32, error) {
	return c.GetBestBlockAsync().Receive()
}

// FutureGetCurrentNetResult is a future promise to deliver the result of a
// GetCurrentNetAsync RPC invocation (or an applicable error).
type FutureGetCurrentNetResult chan *Response

// Receive waits for the Response promised by the future and returns the network
// the server is running on.
func (r FutureGetCurrentNetResult) Receive() (wire.FlokicoinNet, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return 0, err
	}

	// Unmarshal result as an int64.
	var net int64
	err = json.Unmarshal(res, &net)
	if err != nil {
		return 0, err
	}

	return wire.FlokicoinNet(net), nil
}

// GetCurrentNetAsync returns an instance of a type that can be used to get the
// result of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See GetCurrentNet for the blocking version and more details.
//
// NOTE: This is a flokicoind extension.
func (c *Client) GetCurrentNetAsync() FutureGetCurrentNetResult {
	cmd := chainjson.NewGetCurrentNetCmd()
	return c.SendCmd(cmd)
}

// GetCurrentNet returns the network the server is running on.
//
// NOTE: This is a flokicoind extension.
func (c *Client) GetCurrentNet() (wire.FlokicoinNet, error) {
	return c.GetCurrentNetAsync().Receive()
}

// FutureGetHeadersResult is a future promise to deliver the result of a
// getheaders RPC invocation (or an applicable error).
type FutureGetHeadersResult chan *Response

// Receive waits for the Response promised by the future and returns the
// getheaders result.
func (r FutureGetHeadersResult) Receive() ([]wire.BlockHeader, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a slice of strings.
	var result []string
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	// Deserialize the []string into []wire.BlockHeader.
	headers := make([]wire.BlockHeader, len(result))
	for i, headerHex := range result {
		serialized, err := hex.DecodeString(headerHex)
		if err != nil {
			return nil, err
		}
		err = headers[i].Deserialize(bytes.NewReader(serialized))
		if err != nil {
			return nil, err
		}
	}
	return headers, nil
}

// GetHeadersAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the returned instance.
//
// See GetHeaders for the blocking version and more details.
func (c *Client) GetHeadersAsync(blockLocators []chainhash.Hash, hashStop *chainhash.Hash) FutureGetHeadersResult {
	locators := make([]string, len(blockLocators))
	for i := range blockLocators {
		locators[i] = blockLocators[i].String()
	}
	hash := ""
	if hashStop != nil {
		hash = hashStop.String()
	}
	cmd := chainjson.NewGetHeadersCmd(locators, hash)
	return c.SendCmd(cmd)
}

// GetHeaders mimics the wire protocol getheaders and headers messages by
// returning all headers on the main chain after the first known block in the
// locators, up until a block hash matches hashStop.
func (c *Client) GetHeaders(blockLocators []chainhash.Hash, hashStop *chainhash.Hash) ([]wire.BlockHeader, error) {
	return c.GetHeadersAsync(blockLocators, hashStop).Receive()
}

// FutureExportWatchingWalletResult is a future promise to deliver the result of
// an ExportWatchingWalletAsync RPC invocation (or an applicable error).
type FutureExportWatchingWalletResult chan *Response

// Receive waits for the Response promised by the future and returns the
// exported wallet.
func (r FutureExportWatchingWalletResult) Receive() ([]byte, []byte, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, nil, err
	}

	// Unmarshal result as a JSON object.
	var obj map[string]interface{}
	err = json.Unmarshal(res, &obj)
	if err != nil {
		return nil, nil, err
	}

	// Check for the wallet and tx string fields in the object.
	base64Wallet, ok := obj["wallet"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected response type for "+
			"exportwatchingwallet 'wallet' field: %T\n",
			obj["wallet"])
	}
	base64TxStore, ok := obj["tx"].(string)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected response type for "+
			"exportwatchingwallet 'tx' field: %T\n",
			obj["tx"])
	}

	walletBytes, err := base64.StdEncoding.DecodeString(base64Wallet)
	if err != nil {
		return nil, nil, err
	}

	txStoreBytes, err := base64.StdEncoding.DecodeString(base64TxStore)
	if err != nil {
		return nil, nil, err
	}

	return walletBytes, txStoreBytes, nil

}

// ExportWatchingWalletAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See ExportWatchingWallet for the blocking version and more details.
//
// NOTE: This is a walletd extension.
func (c *Client) ExportWatchingWalletAsync(account string) FutureExportWatchingWalletResult {
	cmd := chainjson.NewExportWatchingWalletCmd(&account, chainjson.Bool(true))
	return c.SendCmd(cmd)
}

// ExportWatchingWallet returns the raw bytes for a watching-only version of
// wallet.bin and tx.bin, respectively, for the specified account that can be
// used by walletd to enable a wallet which does not have the private keys
// necessary to spend funds.
//
// NOTE: This is a walletd extension.
func (c *Client) ExportWatchingWallet(account string) ([]byte, []byte, error) {
	return c.ExportWatchingWalletAsync(account).Receive()
}

// FutureSessionResult is a future promise to deliver the result of a
// SessionAsync RPC invocation (or an applicable error).
type FutureSessionResult chan *Response

// Receive waits for the Response promised by the future and returns the
// session result.
func (r FutureSessionResult) Receive() (*chainjson.SessionResult, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a session result object.
	var session chainjson.SessionResult
	err = json.Unmarshal(res, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// SessionAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See Session for the blocking version and more details.
func (c *Client) SessionAsync() FutureSessionResult {
	// Not supported in HTTP POST mode.
	if c.config.HTTPPostMode {
		return newFutureError(ErrWebsocketsRequired)
	}

	cmd := chainjson.NewSessionCmd()
	return c.SendCmd(cmd)
}

// Session returns details regarding a websocket client's current connection.
//
// This RPC requires the client to be running in websocket mode.
func (c *Client) Session() (*chainjson.SessionResult, error) {
	return c.SessionAsync().Receive()
}

// FutureVersionResult is a future promise to deliver the result of a version
// RPC invocation (or an applicable error).
type FutureVersionResult chan *Response

// Receive waits for the Response promised by the future and returns the version
// result.
func (r FutureVersionResult) Receive() (map[string]chainjson.VersionResult,
	error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a version result object.
	var vr map[string]chainjson.VersionResult
	err = json.Unmarshal(res, &vr)
	if err != nil {
		return nil, err
	}

	return vr, nil
}

// VersionAsync returns an instance of a type that can be used to get the result
// of the RPC at some future time by invoking the Receive function on the
// returned instance.
//
// See Version for the blocking version and more details.
func (c *Client) VersionAsync() FutureVersionResult {
	cmd := chainjson.NewVersionCmd()
	return c.SendCmd(cmd)
}

// Version returns information about the server's JSON-RPC API versions.
func (c *Client) Version() (map[string]chainjson.VersionResult, error) {
	return c.VersionAsync().Receive()
}
