// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/chainjson"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/wire"
)

const (
	// defaultMaxFeeRate is the default maximum fee rate in FLC/kvB enforced
	// by flokicoind v0.19.0 or after for transaction broadcast.
	defaultMaxFeeRate chainjson.FLCPerkvB = 0.1
)

// SigHashType enumerates the available signature hashing types that the
// SignRawTransaction function accepts.
type SigHashType string

// Constants used to indicate the signature hash type for SignRawTransaction.
const (
	// SigHashAll indicates ALL of the outputs should be signed.
	SigHashAll SigHashType = "ALL"

	// SigHashNone indicates NONE of the outputs should be signed.  This
	// can be thought of as specifying the signer does not care where the
	// flokicoins go.
	SigHashNone SigHashType = "NONE"

	// SigHashSingle indicates that a SINGLE output should be signed.  This
	// can be thought of specifying the signer only cares about where ONE of
	// the outputs goes, but not any of the others.
	SigHashSingle SigHashType = "SINGLE"

	// SigHashAllAnyoneCanPay indicates that signer does not care where the
	// other inputs to the transaction come from, so it allows other people
	// to add inputs.  In addition, it uses the SigHashAll signing method
	// for outputs.
	SigHashAllAnyoneCanPay SigHashType = "ALL|ANYONECANPAY"

	// SigHashNoneAnyoneCanPay indicates that signer does not care where the
	// other inputs to the transaction come from, so it allows other people
	// to add inputs.  In addition, it uses the SigHashNone signing method
	// for outputs.
	SigHashNoneAnyoneCanPay SigHashType = "NONE|ANYONECANPAY"

	// SigHashSingleAnyoneCanPay indicates that signer does not care where
	// the other inputs to the transaction come from, so it allows other
	// people to add inputs.  In addition, it uses the SigHashSingle signing
	// method for outputs.
	SigHashSingleAnyoneCanPay SigHashType = "SINGLE|ANYONECANPAY"
)

// String returns the SighHashType in human-readable form.
func (s SigHashType) String() string {
	return string(s)
}

// FutureGetRawTransactionResult is a future promise to deliver the result of a
// GetRawTransactionAsync RPC invocation (or an applicable error).
type FutureGetRawTransactionResult chan *Response

// Receive waits for the Response promised by the future and returns a
// transaction given its hash.
func (r FutureGetRawTransactionResult) Receive() (*chainutil.Tx, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var txHex string
	err = json.Unmarshal(res, &txHex)
	if err != nil {
		return nil, err
	}

	// Decode the serialized transaction hex to raw bytes.
	serializedTx, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}

	// Deserialize the transaction and return it.
	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(serializedTx)); err != nil {
		return nil, err
	}
	return chainutil.NewTx(&msgTx), nil
}

// GetRawTransactionAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See GetRawTransaction for the blocking version and more details.
func (c *Client) GetRawTransactionAsync(txHash *chainhash.Hash) FutureGetRawTransactionResult {
	hash := ""
	if txHash != nil {
		hash = txHash.String()
	}

	cmd := chainjson.NewGetRawTransactionCmd(hash, chainjson.Int(0))
	return c.SendCmd(cmd)
}

// GetRawTransaction returns a transaction given its hash.
//
// See GetRawTransactionVerbose to obtain additional information about the
// transaction.
func (c *Client) GetRawTransaction(txHash *chainhash.Hash) (*chainutil.Tx, error) {
	return c.GetRawTransactionAsync(txHash).Receive()
}

// FutureGetRawTransactionVerboseResult is a future promise to deliver the
// result of a GetRawTransactionVerboseAsync RPC invocation (or an applicable
// error).
type FutureGetRawTransactionVerboseResult chan *Response

// Receive waits for the Response promised by the future and returns information
// about a transaction given its hash.
func (r FutureGetRawTransactionVerboseResult) Receive() (*chainjson.TxRawResult, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a gettrawtransaction result object.
	var rawTxResult chainjson.TxRawResult
	err = json.Unmarshal(res, &rawTxResult)
	if err != nil {
		return nil, err
	}

	return &rawTxResult, nil
}

// GetRawTransactionVerboseAsync returns an instance of a type that can be used
// to get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See GetRawTransactionVerbose for the blocking version and more details.
func (c *Client) GetRawTransactionVerboseAsync(txHash *chainhash.Hash) FutureGetRawTransactionVerboseResult {
	hash := ""
	if txHash != nil {
		hash = txHash.String()
	}

	cmd := chainjson.NewGetRawTransactionCmd(hash, chainjson.Int(1))
	return c.SendCmd(cmd)
}

// GetRawTransactionVerbose returns information about a transaction given
// its hash.
//
// See GetRawTransaction to obtain only the transaction already deserialized.
func (c *Client) GetRawTransactionVerbose(txHash *chainhash.Hash) (*chainjson.TxRawResult, error) {
	return c.GetRawTransactionVerboseAsync(txHash).Receive()
}

// FutureDecodeRawTransactionResult is a future promise to deliver the result
// of a DecodeRawTransactionAsync RPC invocation (or an applicable error).
type FutureDecodeRawTransactionResult chan *Response

// Receive waits for the Response promised by the future and returns information
// about a transaction given its serialized bytes.
func (r FutureDecodeRawTransactionResult) Receive() (*chainjson.TxRawResult, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a decoderawtransaction result object.
	var rawTxResult chainjson.TxRawResult
	err = json.Unmarshal(res, &rawTxResult)
	if err != nil {
		return nil, err
	}

	return &rawTxResult, nil
}

// DecodeRawTransactionAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See DecodeRawTransaction for the blocking version and more details.
func (c *Client) DecodeRawTransactionAsync(serializedTx []byte) FutureDecodeRawTransactionResult {
	txHex := hex.EncodeToString(serializedTx)
	cmd := chainjson.NewDecodeRawTransactionCmd(txHex)
	return c.SendCmd(cmd)
}

// DecodeRawTransaction returns information about a transaction given its
// serialized bytes.
func (c *Client) DecodeRawTransaction(serializedTx []byte) (*chainjson.TxRawResult, error) {
	return c.DecodeRawTransactionAsync(serializedTx).Receive()
}

// FutureFundRawTransactionResult is a future promise to deliver the result
// of a FutureFundRawTransactionAsync RPC invocation (or an applicable error).
type FutureFundRawTransactionResult chan *Response

// Receive waits for the Response promised by the future and returns information
// about a funding attempt
func (r FutureFundRawTransactionResult) Receive() (*chainjson.FundRawTransactionResult, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	var marshalled chainjson.FundRawTransactionResult
	if err := json.Unmarshal(res, &marshalled); err != nil {
		return nil, err
	}

	return &marshalled, nil
}

// FundRawTransactionAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See FundRawTransaction for the blocking version and more details.
func (c *Client) FundRawTransactionAsync(tx *wire.MsgTx, opts chainjson.FundRawTransactionOpts, isWitness *bool) FutureFundRawTransactionResult {
	var txBuf bytes.Buffer
	if err := tx.Serialize(&txBuf); err != nil {
		return newFutureError(err)
	}

	cmd := chainjson.NewFundRawTransactionCmd(txBuf.Bytes(), opts, isWitness)
	return c.SendCmd(cmd)
}

// FundRawTransaction returns the result of trying to fund the given transaction with
// funds from the node wallet
func (c *Client) FundRawTransaction(tx *wire.MsgTx, opts chainjson.FundRawTransactionOpts, isWitness *bool) (*chainjson.FundRawTransactionResult, error) {
	return c.FundRawTransactionAsync(tx, opts, isWitness).Receive()
}

// FutureCreateRawTransactionResult is a future promise to deliver the result
// of a CreateRawTransactionAsync RPC invocation (or an applicable error).
type FutureCreateRawTransactionResult chan *Response

// Receive waits for the Response promised by the future and returns a new
// transaction spending the provided inputs and sending to the provided
// addresses.
func (r FutureCreateRawTransactionResult) Receive() (*wire.MsgTx, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var txHex string
	err = json.Unmarshal(res, &txHex)
	if err != nil {
		return nil, err
	}

	// Decode the serialized transaction hex to raw bytes.
	serializedTx, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}

	// Deserialize the transaction and return it.
	var msgTx wire.MsgTx
	// we try both the new and old encoding format
	witnessErr := msgTx.Deserialize(bytes.NewReader(serializedTx))
	if witnessErr != nil {
		legacyErr := msgTx.DeserializeNoWitness(bytes.NewReader(serializedTx))
		if legacyErr != nil {
			return nil, legacyErr
		}
	}
	return &msgTx, nil
}

// CreateRawTransactionAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See CreateRawTransaction for the blocking version and more details.
func (c *Client) CreateRawTransactionAsync(inputs []chainjson.TransactionInput,
	amounts map[chainutil.Address]chainutil.Amount, lockTime *int64) FutureCreateRawTransactionResult {

	convertedAmts := make(map[string]float64, len(amounts))
	for addr, amount := range amounts {
		convertedAmts[addr.String()] = amount.ToFLC()
	}
	cmd := chainjson.NewCreateRawTransactionCmd(inputs, convertedAmts, lockTime)
	return c.SendCmd(cmd)
}

// CreateRawTransaction returns a new transaction spending the provided inputs
// and sending to the provided addresses. If the inputs are either nil or an
// empty slice, it is interpreted as an empty slice.
func (c *Client) CreateRawTransaction(inputs []chainjson.TransactionInput, amounts map[chainutil.Address]chainutil.Amount, lockTime *int64) (*wire.MsgTx, error) {

	return c.CreateRawTransactionAsync(inputs, amounts, lockTime).Receive()
}

// FutureSendRawTransactionResult is a future promise to deliver the result
// of a SendRawTransactionAsync RPC invocation (or an applicable error).
type FutureSendRawTransactionResult chan *Response

// Receive waits for the Response promised by the future and returns the result
// of submitting the encoded transaction to the server which then relays it to
// the network.
func (r FutureSendRawTransactionResult) Receive() (*chainhash.Hash, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a string.
	var txHashStr string
	err = json.Unmarshal(res, &txHashStr)
	if err != nil {
		return nil, err
	}

	return chainhash.NewHashFromStr(txHashStr)
}

// SendRawTransactionAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See SendRawTransaction for the blocking version and more details.
func (c *Client) SendRawTransactionAsync(tx *wire.MsgTx, allowHighFees bool) FutureSendRawTransactionResult {
	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return newFutureError(err)
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	// Due to differences in the sendrawtransaction API for different
	// backends, we'll need to inspect our version and construct the
	// appropriate request.
	version, err := c.BackendVersion()
	if err != nil {
		return newFutureError(err)
	}

	var cmd *chainjson.SendRawTransactionCmd
	// Starting from flokicoind v0.19.0, the MaxFeeRate field should be used.
	//
	// When unified softforks format is supported, it's 0.19 and above.
	if version.SupportUnifiedSoftForks() {
		// Using a 0 MaxFeeRate is interpreted as a maximum fee rate not
		// being enforced by flokicoind.
		var maxFeeRate chainjson.FLCPerkvB
		if !allowHighFees {
			maxFeeRate = defaultMaxFeeRate
		}
		cmd = chainjson.NewFlokicoindSendRawTransactionCmd(txHex, maxFeeRate)
	} else {
		// Otherwise, use the AllowHighFees field.
		cmd = chainjson.NewSendRawTransactionCmd(txHex, &allowHighFees)
	}

	return c.SendCmd(cmd)
}

// SendRawTransaction submits the encoded transaction to the server which will
// then relay it to the network.
func (c *Client) SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error) {
	return c.SendRawTransactionAsync(tx, allowHighFees).Receive()
}

// FutureSignRawTransactionResult is a future promise to deliver the result
// of one of the SignRawTransactionAsync family of RPC invocations (or an
// applicable error).
type FutureSignRawTransactionResult chan *Response

// Receive waits for the Response promised by the future and returns the
// signed transaction as well as whether or not all inputs are now signed.
func (r FutureSignRawTransactionResult) Receive() (*wire.MsgTx, bool, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, false, err
	}

	// Unmarshal as a signrawtransaction result.
	var signRawTxResult chainjson.SignRawTransactionResult
	err = json.Unmarshal(res, &signRawTxResult)
	if err != nil {
		return nil, false, err
	}

	// Decode the serialized transaction hex to raw bytes.
	serializedTx, err := hex.DecodeString(signRawTxResult.Hex)
	if err != nil {
		return nil, false, err
	}

	// Deserialize the transaction and return it.
	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(serializedTx)); err != nil {
		return nil, false, err
	}

	return &msgTx, signRawTxResult.Complete, nil
}

// SignRawTransactionAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function on
// the returned instance.
//
// See SignRawTransaction for the blocking version and more details.
func (c *Client) SignRawTransactionAsync(tx *wire.MsgTx) FutureSignRawTransactionResult {
	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return newFutureError(err)
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	cmd := chainjson.NewSignRawTransactionCmd(txHex, nil, nil, nil)
	return c.SendCmd(cmd)
}

// SignRawTransaction signs inputs for the passed transaction and returns the
// signed transaction as well as whether or not all inputs are now signed.
//
// This function assumes the RPC server already knows the input transactions and
// private keys for the passed transaction which needs to be signed and uses the
// default signature hash type.  Use one of the SignRawTransaction# variants to
// specify that information if needed.
func (c *Client) SignRawTransaction(tx *wire.MsgTx) (*wire.MsgTx, bool, error) {
	return c.SignRawTransactionAsync(tx).Receive()
}

// SignRawTransaction2Async returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See SignRawTransaction2 for the blocking version and more details.
func (c *Client) SignRawTransaction2Async(tx *wire.MsgTx, inputs []chainjson.RawTxInput) FutureSignRawTransactionResult {
	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return newFutureError(err)
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	cmd := chainjson.NewSignRawTransactionCmd(txHex, &inputs, nil, nil)
	return c.SendCmd(cmd)
}

// SignRawTransaction2 signs inputs for the passed transaction given the list
// of information about the input transactions needed to perform the signing
// process.
//
// This only input transactions that need to be specified are ones the
// RPC server does not already know.  Already known input transactions will be
// merged with the specified transactions.
//
// See SignRawTransaction if the RPC server already knows the input
// transactions.
func (c *Client) SignRawTransaction2(tx *wire.MsgTx, inputs []chainjson.RawTxInput) (*wire.MsgTx, bool, error) {
	return c.SignRawTransaction2Async(tx, inputs).Receive()
}

// SignRawTransaction3Async returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See SignRawTransaction3 for the blocking version and more details.
func (c *Client) SignRawTransaction3Async(tx *wire.MsgTx,
	inputs []chainjson.RawTxInput,
	privKeysWIF []string) FutureSignRawTransactionResult {

	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return newFutureError(err)
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	cmd := chainjson.NewSignRawTransactionCmd(txHex, &inputs, &privKeysWIF,
		nil)
	return c.SendCmd(cmd)
}

// SignRawTransaction3 signs inputs for the passed transaction given the list
// of information about extra input transactions and a list of private keys
// needed to perform the signing process.  The private keys must be in wallet
// import format (WIF).
//
// This only input transactions that need to be specified are ones the
// RPC server does not already know.  Already known input transactions will be
// merged with the specified transactions.  This means the list of transaction
// inputs can be nil if the RPC server already knows them all.
//
// NOTE: Unlike the merging functionality of the input transactions, ONLY the
// specified private keys will be used, so even if the server already knows some
// of the private keys, they will NOT be used.
//
// See SignRawTransaction if the RPC server already knows the input
// transactions and private keys or SignRawTransaction2 if it already knows the
// private keys.
func (c *Client) SignRawTransaction3(tx *wire.MsgTx,
	inputs []chainjson.RawTxInput,
	privKeysWIF []string) (*wire.MsgTx, bool, error) {

	return c.SignRawTransaction3Async(tx, inputs, privKeysWIF).Receive()
}

// SignRawTransaction4Async returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See SignRawTransaction4 for the blocking version and more details.
func (c *Client) SignRawTransaction4Async(tx *wire.MsgTx,
	inputs []chainjson.RawTxInput, privKeysWIF []string,
	hashType SigHashType) FutureSignRawTransactionResult {

	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return newFutureError(err)
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	cmd := chainjson.NewSignRawTransactionCmd(txHex, &inputs, &privKeysWIF,
		chainjson.String(string(hashType)))
	return c.SendCmd(cmd)
}

// SignRawTransaction4 signs inputs for the passed transaction using
// the specified signature hash type given the list of information about extra
// input transactions and a potential list of private keys needed to perform
// the signing process.  The private keys, if specified, must be in wallet
// import format (WIF).
//
// The only input transactions that need to be specified are ones the RPC server
// does not already know.  This means the list of transaction inputs can be nil
// if the RPC server already knows them all.
//
// NOTE: Unlike the merging functionality of the input transactions, ONLY the
// specified private keys will be used, so even if the server already knows some
// of the private keys, they will NOT be used.  The list of private keys can be
// nil in which case any private keys the RPC server knows will be used.
//
// This function should only used if a non-default signature hash type is
// desired.  Otherwise, see SignRawTransaction if the RPC server already knows
// the input transactions and private keys, SignRawTransaction2 if it already
// knows the private keys, or SignRawTransaction3 if it does not know both.
func (c *Client) SignRawTransaction4(tx *wire.MsgTx,
	inputs []chainjson.RawTxInput, privKeysWIF []string,
	hashType SigHashType) (*wire.MsgTx, bool, error) {

	return c.SignRawTransaction4Async(tx, inputs, privKeysWIF,
		hashType).Receive()
}

// FutureSignRawTransactionWithWalletResult is a future promise to deliver
// the result of the SignRawTransactionWithWalletAsync RPC invocation (or
// an applicable error).
type FutureSignRawTransactionWithWalletResult chan *Response

// Receive waits for the Response promised by the future and returns the
// signed transaction as well as whether or not all inputs are now signed.
func (r FutureSignRawTransactionWithWalletResult) Receive() (*wire.MsgTx, bool, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, false, err
	}

	// Unmarshal as a signtransactionwithwallet result.
	var signRawTxWithWalletResult chainjson.SignRawTransactionWithWalletResult
	err = json.Unmarshal(res, &signRawTxWithWalletResult)
	if err != nil {
		return nil, false, err
	}

	// Decode the serialized transaction hex to raw bytes.
	serializedTx, err := hex.DecodeString(signRawTxWithWalletResult.Hex)
	if err != nil {
		return nil, false, err
	}

	// Deserialize the transaction and return it.
	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(serializedTx)); err != nil {
		return nil, false, err
	}

	return &msgTx, signRawTxWithWalletResult.Complete, nil
}

// SignRawTransactionWithWalletAsync returns an instance of a type that can be used
// to get the result of the RPC at some future time by invoking the Receive function
// on the returned instance.
//
// See SignRawTransactionWithWallet for the blocking version and more details.
func (c *Client) SignRawTransactionWithWalletAsync(tx *wire.MsgTx) FutureSignRawTransactionWithWalletResult {
	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return newFutureError(err)
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	cmd := chainjson.NewSignRawTransactionWithWalletCmd(txHex, nil, nil)
	return c.SendCmd(cmd)
}

// SignRawTransactionWithWallet signs inputs for the passed transaction and returns
// the signed transaction as well as whether or not all inputs are now signed.
//
// This function assumes the RPC server already knows the input transactions for the
// passed transaction which needs to be signed and uses the default signature hash
// type.  Use one of the SignRawTransactionWithWallet# variants to specify that
// information if needed.
func (c *Client) SignRawTransactionWithWallet(tx *wire.MsgTx) (*wire.MsgTx, bool, error) {
	return c.SignRawTransactionWithWalletAsync(tx).Receive()
}

// SignRawTransactionWithWallet2Async returns an instance of a type that can be
// used to get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See SignRawTransactionWithWallet2 for the blocking version and more details.
func (c *Client) SignRawTransactionWithWallet2Async(tx *wire.MsgTx,
	inputs []chainjson.RawTxWitnessInput) FutureSignRawTransactionWithWalletResult {

	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return newFutureError(err)
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	cmd := chainjson.NewSignRawTransactionWithWalletCmd(txHex, &inputs, nil)
	return c.SendCmd(cmd)
}

// SignRawTransactionWithWallet2 signs inputs for the passed transaction given the
// list of information about the input transactions needed to perform the signing
// process.
//
// This only input transactions that need to be specified are ones the
// RPC server does not already know.  Already known input transactions will be
// merged with the specified transactions.
//
// See SignRawTransactionWithWallet if the RPC server already knows the input
// transactions.
func (c *Client) SignRawTransactionWithWallet2(tx *wire.MsgTx,
	inputs []chainjson.RawTxWitnessInput) (*wire.MsgTx, bool, error) {

	return c.SignRawTransactionWithWallet2Async(tx, inputs).Receive()
}

// SignRawTransactionWithWallet3Async returns an instance of a type that can
// be used to get the result of the RPC at some future time by invoking the
// Receive function on the returned instance.
//
// See SignRawTransactionWithWallet3 for the blocking version and more details.
func (c *Client) SignRawTransactionWithWallet3Async(tx *wire.MsgTx,
	inputs []chainjson.RawTxWitnessInput, hashType SigHashType) FutureSignRawTransactionWithWalletResult {

	txHex := ""
	if tx != nil {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(buf); err != nil {
			return newFutureError(err)
		}
		txHex = hex.EncodeToString(buf.Bytes())
	}

	cmd := chainjson.NewSignRawTransactionWithWalletCmd(txHex, &inputs, chainjson.String(string(hashType)))
	return c.SendCmd(cmd)
}

// SignRawTransactionWithWallet3 signs inputs for the passed transaction using
// the specified signature hash type given the list of information about extra
// input transactions.
//
// The only input transactions that need to be specified are ones the RPC server
// does not already know.  This means the list of transaction inputs can be nil
// if the RPC server already knows them all.
//
// This function should only used if a non-default signature hash type is
// desired.  Otherwise, see SignRawTransactionWithWallet if the RPC server already
// knows the input transactions, or SignRawTransactionWithWallet2 if it does not.
func (c *Client) SignRawTransactionWithWallet3(tx *wire.MsgTx,
	inputs []chainjson.RawTxWitnessInput, hashType SigHashType) (*wire.MsgTx, bool, error) {

	return c.SignRawTransactionWithWallet3Async(tx, inputs, hashType).Receive()
}

// FutureSearchRawTransactionsResult is a future promise to deliver the result
// of the SearchRawTransactionsAsync RPC invocation (or an applicable error).
type FutureSearchRawTransactionsResult chan *Response

// Receive waits for the Response promised by the future and returns the
// found raw transactions.
func (r FutureSearchRawTransactionsResult) Receive() ([]*wire.MsgTx, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an array of strings.
	var searchRawTxnsResult []string
	err = json.Unmarshal(res, &searchRawTxnsResult)
	if err != nil {
		return nil, err
	}

	// Decode and deserialize each transaction.
	msgTxns := make([]*wire.MsgTx, 0, len(searchRawTxnsResult))
	for _, hexTx := range searchRawTxnsResult {
		// Decode the serialized transaction hex to raw bytes.
		serializedTx, err := hex.DecodeString(hexTx)
		if err != nil {
			return nil, err
		}

		// Deserialize the transaction and add it to the result slice.
		var msgTx wire.MsgTx
		err = msgTx.Deserialize(bytes.NewReader(serializedTx))
		if err != nil {
			return nil, err
		}
		msgTxns = append(msgTxns, &msgTx)
	}

	return msgTxns, nil
}

// SearchRawTransactionsAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See SearchRawTransactions for the blocking version and more details.
func (c *Client) SearchRawTransactionsAsync(address chainutil.Address, skip, count int, reverse bool, filterAddrs []string) FutureSearchRawTransactionsResult {
	addr := address.EncodeAddress()
	verbose := chainjson.Int(0)
	cmd := chainjson.NewSearchRawTransactionsCmd(addr, verbose, &skip, &count,
		nil, &reverse, &filterAddrs)
	return c.SendCmd(cmd)
}

// SearchRawTransactions returns transactions that involve the passed address.
//
// NOTE: Chain servers do not typically provide this capability unless it has
// specifically been enabled.
//
// See SearchRawTransactionsVerbose to retrieve a list of data structures with
// information about the transactions instead of the transactions themselves.
func (c *Client) SearchRawTransactions(address chainutil.Address, skip, count int, reverse bool, filterAddrs []string) ([]*wire.MsgTx, error) {
	return c.SearchRawTransactionsAsync(address, skip, count, reverse, filterAddrs).Receive()
}

// FutureSearchRawTransactionsVerboseResult is a future promise to deliver the
// result of the SearchRawTransactionsVerboseAsync RPC invocation (or an
// applicable error).
type FutureSearchRawTransactionsVerboseResult chan *Response

// Receive waits for the Response promised by the future and returns the
// found raw transactions.
func (r FutureSearchRawTransactionsVerboseResult) Receive() ([]*chainjson.SearchRawTransactionsResult, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an array of raw transaction results.
	var result []*chainjson.SearchRawTransactionsResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SearchRawTransactionsVerboseAsync returns an instance of a type that can be
// used to get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See SearchRawTransactionsVerbose for the blocking version and more details.
func (c *Client) SearchRawTransactionsVerboseAsync(address chainutil.Address, skip,
	count int, includePrevOut, reverse bool, filterAddrs *[]string) FutureSearchRawTransactionsVerboseResult {

	addr := address.EncodeAddress()
	verbose := chainjson.Int(1)
	var prevOut *int
	if includePrevOut {
		prevOut = chainjson.Int(1)
	}
	cmd := chainjson.NewSearchRawTransactionsCmd(addr, verbose, &skip, &count,
		prevOut, &reverse, filterAddrs)
	return c.SendCmd(cmd)
}

// SearchRawTransactionsVerbose returns a list of data structures that describe
// transactions which involve the passed address.
//
// NOTE: Chain servers do not typically provide this capability unless it has
// specifically been enabled.
//
// See SearchRawTransactions to retrieve a list of raw transactions instead.
func (c *Client) SearchRawTransactionsVerbose(address chainutil.Address, skip,
	count int, includePrevOut, reverse bool, filterAddrs []string) ([]*chainjson.SearchRawTransactionsResult, error) {

	return c.SearchRawTransactionsVerboseAsync(address, skip, count,
		includePrevOut, reverse, &filterAddrs).Receive()
}

// FutureDecodeScriptResult is a future promise to deliver the result
// of a DecodeScriptAsync RPC invocation (or an applicable error).
type FutureDecodeScriptResult chan *Response

// Receive waits for the Response promised by the future and returns information
// about a script given its serialized bytes.
func (r FutureDecodeScriptResult) Receive() (*chainjson.DecodeScriptResult, error) {
	res, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal result as a decodescript result object.
	var decodeScriptResult chainjson.DecodeScriptResult
	err = json.Unmarshal(res, &decodeScriptResult)
	if err != nil {
		return nil, err
	}

	return &decodeScriptResult, nil
}

// DecodeScriptAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See DecodeScript for the blocking version and more details.
func (c *Client) DecodeScriptAsync(serializedScript []byte) FutureDecodeScriptResult {
	scriptHex := hex.EncodeToString(serializedScript)
	cmd := chainjson.NewDecodeScriptCmd(scriptHex)
	return c.SendCmd(cmd)
}

// DecodeScript returns information about a script given its serialized bytes.
func (c *Client) DecodeScript(serializedScript []byte) (*chainjson.DecodeScriptResult, error) {
	return c.DecodeScriptAsync(serializedScript).Receive()
}

// FutureTestMempoolAcceptResult is a future promise to deliver the result
// of a TestMempoolAccept RPC invocation (or an applicable error).
type FutureTestMempoolAcceptResult chan *Response

// Receive waits for the Response promised by the future and returns the
// response from TestMempoolAccept.
func (r FutureTestMempoolAcceptResult) Receive() (
	[]*chainjson.TestMempoolAcceptResult, error) {

	response, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an array of TestMempoolAcceptResult items.
	var results []*chainjson.TestMempoolAcceptResult

	err = json.Unmarshal(response, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// TestMempoolAcceptAsync returns an instance of a type that can be used to get
// the result of the RPC at some future time by invoking the Receive function
// on the returned instance.
//
// See TestMempoolAccept for the blocking version and more details.
func (c *Client) TestMempoolAcceptAsync(txns []*wire.MsgTx,
	maxFeeRate chainjson.FLCPerkvB) FutureTestMempoolAcceptResult {

	// Due to differences in the testmempoolaccept API for different
	// backends, we'll need to inspect our version and construct the
	// appropriate request.
	version, err := c.BackendVersion()
	if err != nil {
		return newFutureError(err)
	}

	log.Debugf("TestMempoolAcceptAsync: backend version %s", version)

	// Exit early if the version is below 22.0.0.
	//
	// Based on the history of `testmempoolaccept` in flokicoind,
	// - introduced in 0.17.0
	// - unchanged in 0.18.0
	// - allowhighfees(bool) param is changed to maxfeerate(numeric) in
	//   0.19.0
	// - unchanged in 0.20.0
	// - added fees and vsize fields in its response in 0.21.0
	// - allow more than one txes in param rawtx and added package-error
	//   and wtxid fields in its response in 0.22.0
	// - unchanged in 0.23.0
	// - unchanged in 0.24.0
	// - added effective-feerate and effective-includes fields in its
	//   response in 0.25.0
	//
	// We decide to not support this call for versions below 22.0.0. as the
	// request/response formats are very different.
	if !version.SupportTestMempoolAccept() {
		err := fmt.Errorf("%w: %v", ErrBackendVersion, version)
		return newFutureError(err)
	}

	// The maximum number of transactions allowed is 25.
	if len(txns) > 25 {
		err := fmt.Errorf("%w: too many transactions provided",
			ErrInvalidParam)
		return newFutureError(err)
	}

	// Exit early if an empty array of transactions is provided.
	if len(txns) == 0 {
		err := fmt.Errorf("%w: no transactions provided",
			ErrInvalidParam)
		return newFutureError(err)
	}

	// Iterate all the transactions and turn them into hex strings.
	rawTxns := make([]string, 0, len(txns))
	for _, tx := range txns {
		// Serialize the transaction and convert to hex string.
		buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))

		// TODO(yy): add similar checks found in `FlcDecode` to
		// `FlcEncode` - atm it just serializes bytes without any
		// flokicoin-specific checks.
		if err := tx.Serialize(buf); err != nil {
			err = fmt.Errorf("%w: %v", ErrInvalidParam, err)
			return newFutureError(err)
		}

		rawTx := hex.EncodeToString(buf.Bytes())
		rawTxns = append(rawTxns, rawTx)

		// Sanity check the provided tx is valid, which can be removed
		// once we have similar checks added in `FlcEncode`.
		//
		// NOTE: must be performed after buf.Bytes is copied above.
		//
		// TODO(yy): remove it once the above TODO is addressed.
		if err := tx.Deserialize(buf); err != nil {
			err = fmt.Errorf("%w: %v", ErrInvalidParam, err)
			return newFutureError(err)
		}
	}

	cmd := chainjson.NewTestMempoolAcceptCmd(rawTxns, maxFeeRate)

	return c.SendCmd(cmd)
}

// TestMempoolAccept returns result of mempool acceptance tests indicating if
// raw transaction(s) would be accepted by mempool.
//
// If multiple transactions are passed in, parents must come before children
// and package policies apply: the transactions cannot conflict with any
// mempool transactions or each other.
//
// If one transaction fails, other transactions may not be fully validated (the
// 'allowed' key will be blank).
//
// The maximum number of transactions allowed is 25.
func (c *Client) TestMempoolAccept(txns []*wire.MsgTx,
	maxFeeRate chainjson.FLCPerkvB) ([]*chainjson.TestMempoolAcceptResult, error) {

	return c.TestMempoolAcceptAsync(txns, maxFeeRate).Receive()
}

// FutureGetTxSpendingPrevOut is a future promise to deliver the result of a
// GetTxSpendingPrevOut RPC invocation (or an applicable error).
type FutureGetTxSpendingPrevOut chan *Response

// Receive waits for the Response promised by the future and returns the
// response from GetTxSpendingPrevOut.
func (r FutureGetTxSpendingPrevOut) Receive() (
	[]*chainjson.GetTxSpendingPrevOutResult, error) {

	response, err := ReceiveFuture(r)
	if err != nil {
		return nil, err
	}

	// Unmarshal as an array of GetTxSpendingPrevOutResult items.
	var results []*chainjson.GetTxSpendingPrevOutResult

	err = json.Unmarshal(response, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetTxSpendingPrevOutAsync returns an instance of a type that can be used to
// get the result of the RPC at some future time by invoking the Receive
// function on the returned instance.
//
// See GetTxSpendingPrevOut for the blocking version and more details.
func (c *Client) GetTxSpendingPrevOutAsync(
	outpoints []wire.OutPoint) FutureGetTxSpendingPrevOut {

	// Due to differences in the testmempoolaccept API for different
	// backends, we'll need to inspect our version and construct the
	// appropriate request.
	version, err := c.BackendVersion()
	if err != nil {
		return newFutureError(err)
	}

	log.Debugf("GetTxSpendingPrevOutAsync: backend version %s", version)

	// Exit early if the version is below 24.0.0.
	if !version.SupportGetTxSpendingPrevOut() {
		err := fmt.Errorf("%w: %v", ErrBackendVersion, version)
		return newFutureError(err)
	}

	// Exit early if an empty array of outpoints is provided.
	if len(outpoints) == 0 {
		err := fmt.Errorf("%w: no outpoints provided", ErrInvalidParam)
		return newFutureError(err)
	}

	cmd := chainjson.NewGetTxSpendingPrevOutCmd(outpoints)

	return c.SendCmd(cmd)
}

// GetTxSpendingPrevOut returns the result from calling `gettxspendingprevout`
// RPC.
func (c *Client) GetTxSpendingPrevOut(outpoints []wire.OutPoint) (
	[]*chainjson.GetTxSpendingPrevOutResult, error) {

	return c.GetTxSpendingPrevOutAsync(outpoints).Receive()
}
