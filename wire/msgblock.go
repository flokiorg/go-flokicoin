// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"fmt"
	"io"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
)

// defaultTransactionAlloc is the default size used for the backing array
// for transactions.  The transaction array will dynamically grow as needed, but
// this figure is intended to provide enough space for the number of
// transactions in the vast majority of blocks without needing to grow the
// backing array multiple times.
const defaultTransactionAlloc = 2048

// MaxBlocksPerMsg is the maximum number of blocks allowed per message.
const MaxBlocksPerMsg = 500

// MaxBlockPayload is the maximum bytes a block message can be in bytes.
// After Segregated Witness, the max block payload has been raised to 4MB.
const MaxBlockPayload = 4000000

// maxTxPerBlock is the maximum number of transactions that could
// possibly fit into a block.
const maxTxPerBlock = (MaxBlockPayload / minTxPayload) + 1

// TxLoc holds locator data for the offset and length of where a transaction is
// located within a MsgBlock data buffer.
type TxLoc struct {
	TxStart int
	TxLen   int
}

// MsgBlock implements the Message interface and represents a flokicoin
// block message.  It is used to deliver block and transaction information in
// response to a getdata message (MsgGetData) for a given block hash.
type MsgBlock struct {
	Header       BlockHeader
	Transactions []*MsgTx
}

// Copy creates a deep copy of MsgBlock.
func (msg *MsgBlock) Copy() *MsgBlock {
	block := &MsgBlock{
		Header:       msg.Header,
		Transactions: make([]*MsgTx, len(msg.Transactions)),
	}

	for i, tx := range msg.Transactions {
		block.Transactions[i] = tx.Copy()
	}

	return block
}

// AddTransaction adds a transaction to the message.
func (msg *MsgBlock) AddTransaction(tx *MsgTx) error {
	msg.Transactions = append(msg.Transactions, tx)
	return nil

}

// ClearTransactions removes all transactions from the message.
func (msg *MsgBlock) ClearTransactions() {
	msg.Transactions = make([]*MsgTx, 0, defaultTransactionAlloc)
}

// FlcDecode decodes r using the flokicoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
// See Deserialize for decoding blocks stored to disk, such as in a database, as
// opposed to decoding blocks from the wire.
func (msg *MsgBlock) FlcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	// Decode header (wire format)
	if err := msg.Header.FlcDecode(r, pver, enc); err != nil {
		return err
	}

	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)

	txCount, err := ReadVarIntBuf(r, pver, buf)
	if err != nil {
		return err
	}

	// Prevent more transactions than could possibly fit into a block.
	// It would be possible to cause memory exhaustion and panics without
	// a sane upper bound on this count.
	if txCount > maxTxPerBlock {
		str := fmt.Sprintf("too many transactions to fit into a block "+
			"[count %d, max %d]", txCount, maxTxPerBlock)
		return messageError("MsgBlock.FlcDecode", str)
	}

	scriptBuf := scriptPool.Borrow()
	defer scriptPool.Return(scriptBuf)

	msg.Transactions = make([]*MsgTx, 0, txCount)
	for i := uint64(0); i < txCount; i++ {
		tx := MsgTx{}
		err := tx.flcDecode(r, pver, enc, buf, scriptBuf[:])
		if err != nil {
			return err
		}
		msg.Transactions = append(msg.Transactions, &tx)
	}

	return nil
}

// Deserialize decodes a block from r into the receiver using a format that is
// suitable for long-term storage such as a database while respecting the
// Version field in the block.  This function differs from FlcDecode in that
// FlcDecode decodes from the flokicoin wire protocol as it was sent across the
// network.  The wire encoding can technically differ depending on the protocol
// version and doesn't even really need to match the format of a stored block at
// all.  As of the time this comment was written, the encoded block is the same
// in both instances, but there is a distinct difference and separating the two
// allows the API to be flexible enough to deal with changes.
func (msg *MsgBlock) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of FlcDecode.
	//
	// Passing an encoding type of WitnessEncoding to FlcEncode for the
	// MessageEncoding parameter indicates that the transactions within the
	// block are expected to be serialized according to the new
	// serialization structure defined in BIP0141.
	return msg.FlcDecode(r, 0, WitnessEncoding)
}

// FromBytes deserializes a transaction byte slice.
func (msg *MsgBlock) FromBytes(b []byte) error {
	r := bytes.NewReader(b)
	return msg.Deserialize(r)
}

// DeserializeNoWitness decodes a block from r into the receiver similar to
// Deserialize, however DeserializeWitness strips all (if any) witness data
// from the transactions within the block before encoding them.
func (msg *MsgBlock) DeserializeNoWitness(r io.Reader) error {
	return msg.FlcDecode(r, 0, BaseEncoding)
}

// decodeTxs is a shared helper that reads the transaction count and decodes
// all transactions. If trackLocs is true, it also returns per-transaction
// locations (requires r to be *bytes.Buffer).
func (msg *MsgBlock) DecodeTransactions(r io.Reader, pver uint32, enc MessageEncoding) error {

	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)

	txCount, err := ReadVarIntBuf(r, pver, buf)
	if err != nil {
		return err
	}

	// Prevent more transactions than could possibly fit into a block.
	// It would be possible to cause memory exhaustion and panics without
	// a sane upper bound on this count.
	if txCount > maxTxPerBlock {
		str := fmt.Sprintf("too many transactions to fit into a block "+
			"[count %d, max %d]", txCount, maxTxPerBlock)
		return messageError("MsgBlock.FlcDecode", str)
	}

	scriptBuf := scriptPool.Borrow()
	defer scriptPool.Return(scriptBuf)

	msg.Transactions = make([]*MsgTx, 0, txCount)
	for i := uint64(0); i < txCount; i++ {
		tx := MsgTx{}
		err := tx.flcDecode(r, pver, enc, buf, scriptBuf[:])
		if err != nil {
			return err
		}
		msg.Transactions = append(msg.Transactions, &tx)
	}

	return nil
}

// DeserializeTxLoc decodes r in the same manner Deserialize does, but it takes
// a byte buffer instead of a generic reader and returns a slice containing the
// start and length of each transaction within the raw data that is being
// deserialized.
func (msg *MsgBlock) DeserializeTxLoc(r *bytes.Buffer) ([]TxLoc, error) {
	fullLen := r.Len()

	// Decode header (disk format)
	if err := msg.Header.Deserialize(r); err != nil {
		return nil, err
	}

	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)

	txCount, err := ReadVarIntBuf(r, 0, buf)
	if err != nil {
		return nil, err
	}

	// Prevent more transactions than could possibly fit into a block.
	// It would be possible to cause memory exhaustion and panics without
	// a sane upper bound on this count.
	if txCount > maxTxPerBlock {
		str := fmt.Sprintf("too many transactions to fit into a block "+
			"[count %d, max %d]", txCount, maxTxPerBlock)
		return nil, messageError("MsgBlock.DeserializeTxLoc", str)
	}

	scriptBuf := scriptPool.Borrow()
	defer scriptPool.Return(scriptBuf)

	// Deserialize each transaction while keeping track of its location
	// within the byte stream.
	msg.Transactions = make([]*MsgTx, 0, txCount)
	txLocs := make([]TxLoc, txCount)
	for i := uint64(0); i < txCount; i++ {
		txLocs[i].TxStart = fullLen - r.Len()
		tx := MsgTx{}
		err := tx.flcDecode(r, 0, WitnessEncoding, buf, scriptBuf[:])
		if err != nil {
			return nil, err
		}
		msg.Transactions = append(msg.Transactions, &tx)
		txLocs[i].TxLen = (fullLen - r.Len()) - txLocs[i].TxStart
	}

	return txLocs, nil
}

// FlcEncode encodes the receiver to w using the flokicoin protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding blocks to be stored to disk, such as in a
// database, as opposed to encoding blocks for the wire.
func (msg *MsgBlock) FlcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	// Encode full header including AuxPoW payload (if any).
	if err := msg.Header.FlcEncode(w, pver, enc); err != nil {
		return err
	}
	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)
	err := WriteVarIntBuf(w, pver, uint64(len(msg.Transactions)), buf)
	if err != nil {
		return err
	}

	for _, tx := range msg.Transactions {
		err = tx.flcEncode(w, pver, enc, buf)
		if err != nil {
			return err
		}
	}

	return nil
}

// Serialize encodes the block to w using a format that suitable for long-term
// storage such as a database while respecting the Version field in the block.
// This function differs from FlcEncode in that FlcEncode encodes the block to
// the flokicoin wire protocol in order to be sent across the network.  The wire
// encoding can technically differ depending on the protocol version and doesn't
// even really need to match the format of a stored block at all.  As of the
// time this comment was written, the encoded block is the same in both
// instances, but there is a distinct difference and separating the two allows
// the API to be flexible enough to deal with changes.
func (msg *MsgBlock) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of FlcEncode.
	//
	// Passing WitnessEncoding as the encoding type here indicates that
	// each of the transactions should be serialized using the witness
	// serialization structure defined in BIP0141.
	return msg.FlcEncode(w, 0, WitnessEncoding)
}

// Bytes returns the serialized form of the block in bytes.
func (msg *MsgBlock) Bytes() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, msg.SerializeSize()))
	err := msg.Serialize(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SerializeNoWitness encodes a block to w using an identical format to
// Serialize, with all (if any) witness data stripped from all transactions.
// This method is provided in addition to the regular Serialize, in order to
// allow one to selectively encode transaction witness data to non-upgraded
// peers which are unaware of the new encoding.
func (msg *MsgBlock) SerializeNoWitness(w io.Writer) error {
	return msg.FlcEncode(w, 0, BaseEncoding)
}

// SerializeSize returns the number of bytes it would take to serialize the
// block, factoring in any witness data within transaction.
func (msg *MsgBlock) SerializeSize() int {
	// Header bytes (including AuxPoW if present) + varint count + txs.
	headerSize := BlockHeaderLen
	if msg.Header.AuxPow() {
		// Compute serialized size dynamically when AuxPoW is present.
		var hb bytes.Buffer
		_ = msg.Header.Serialize(&hb)
		headerSize = hb.Len()
	}
	n := headerSize + VarIntSerializeSize(uint64(len(msg.Transactions)))

	for _, tx := range msg.Transactions {
		n += tx.SerializeSize()
	}

	return n
}

// SerializeSizeStripped returns the number of bytes it would take to serialize
// the block, excluding any witness data (if any).
func (msg *MsgBlock) SerializeSizeStripped() int {
	headerSize := BlockHeaderLen
	if msg.Header.AuxPow() {
		var hb bytes.Buffer
		_ = msg.Header.Serialize(&hb)
		headerSize = hb.Len()
	}
	n := headerSize + VarIntSerializeSize(uint64(len(msg.Transactions)))

	for _, tx := range msg.Transactions {
		n += tx.SerializeSizeStripped()
	}

	return n
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgBlock) Command() string {
	return CmdBlock
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgBlock) MaxPayloadLength(pver uint32) uint32 {
	// Block header at 80 bytes + transaction count + max transactions
	// which can vary up to the MaxBlockPayload (including the block header
	// and transaction count).
	return MaxBlockPayload
}

// BlockHash computes the block identifier hash for this block.
func (msg *MsgBlock) BlockHash() chainhash.Hash {
	return msg.Header.BlockHash()
}

// TxHashes returns a slice of hashes of all of transactions in this block.
func (msg *MsgBlock) TxHashes() ([]chainhash.Hash, error) {
	hashList := make([]chainhash.Hash, 0, len(msg.Transactions))
	for _, tx := range msg.Transactions {
		hashList = append(hashList, tx.TxHash())
	}
	return hashList, nil
}

// NewMsgBlock returns a new flokicoin block message that conforms to the
// Message interface.  See MsgBlock for details.
func NewMsgBlock(blockHeader *BlockHeader) *MsgBlock {
	return &MsgBlock{
		Header:       *blockHeader,
		Transactions: make([]*MsgTx, 0, defaultTransactionAlloc),
	}
}
