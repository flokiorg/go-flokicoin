// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
)

// MaxBlockHeaderPayload is the maximum number of bytes a block header can be.
// Version 4 bytes + Timestamp 4 bytes + Bits 4 bytes + Nonce 4 bytes +
// PrevBlock and MerkleRoot hashes.
const (
	MaxBlockHeaderPayload = 16 + (chainhash.HashSize * 2)

	VersionAuxPow int32 = (1 << 8)

	// blockHeaderLen is a constant that represents the number of bytes for a block header.
	BlockHeaderLen = 80
	// ChainIDMask covers bits [16..21] (6 bits) used to store the chain ID.
	ChainIDMask int32 = 0x003F0000
)

// BlockHeader defines information about a block and is used in the flokicoin
// block (MsgBlock) and headers (MsgHeaders) messages.
type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int32

	// Hash of the previous block header in the block chain.
	PrevBlock chainhash.Hash

	// Merkle tree reference to hash of all transactions for the block.
	MerkleRoot chainhash.Hash

	// Time the block was created.  This is, unfortunately, encoded as a
	// uint32 on the wire and therefore is limited to 2106.
	Timestamp time.Time

	// Difficulty target for the block.
	Bits uint32

	// Nonce used to generate the block.
	Nonce uint32

	// Aux data
	AuxPowHeader *AuxPowHeader
}

func (h *BlockHeader) AuxPow() bool {
	return (h.Version & VersionAuxPow) != 0
}

func (h *BlockHeader) GetChainID() int32 {
	return (h.Version & ChainIDMask) >> 16
}

func (h *BlockHeader) SetAuxPow(auxpow bool) {
	if auxpow {
		h.Version |= VersionAuxPow
	} else {
		h.Version &= ^VersionAuxPow
	}
}

func (h *BlockHeader) SetChainID(chainID int32) {
	// Preserve all bits except the chain-id field, then set masked chain-id.
	h.Version &= ^ChainIDMask
	h.Version |= (chainID << 16) & ChainIDMask
}

func (h *BlockHeader) IsLegacy() bool {
	return h.Version == 1 || h.Version == 2 || h.Version == 0x20000000
}

// BlockHash computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockPoWHash() chainhash.Hash {
	return chainhash.ScryptRaw(func(w io.Writer) error {
		return writeBlockHeader(w, 0, h)
	})
}

// BlockHash computes the block identifier hash for the given block header.
func (h *BlockHeader) BlockHash() chainhash.Hash {
	return chainhash.DoubleHashRaw(func(w io.Writer) error {
		return writeBlockHeader(w, 0, h)
	})
}

// FlcDecode decodes r using the flokicoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
// See Deserialize for decoding block headers stored to disk, such as in a
// database, as opposed to decoding block headers from the wire.
func (h *BlockHeader) FlcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	if err := readBlockHeader(r, pver, h); err != nil {
		return err
	}

	if h.AuxPow() {
		if h.AuxPowHeader == nil {
			h.AuxPowHeader = &AuxPowHeader{}
		}
		if err := h.AuxPowHeader.FlcDecode(r, pver, enc); err != nil {
			return err
		}
	}

	return nil
}

// FlcEncode encodes the receiver to w using the flokicoin protocol encoding.
// This is part of the Message interface implementation.
// See Serialize for encoding block headers to be stored to disk, such as in a
// database, as opposed to encoding block headers for the wire.
func (h *BlockHeader) FlcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	if err := writeBlockHeader(w, pver, h); err != nil {
		return err
	}

	if h.AuxPow() {
		if h.AuxPowHeader == nil {
			return fmt.Errorf("auxpow header is nil for auxpow block header: hash: %s (version: %x)", h.BlockHash(), h.Version)
		}
		if err := h.AuxPowHeader.FlcEncode(w, pver, enc); err != nil {
			return err
		}
	}

	return nil
}

// Deserialize decodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of readBlockHeader.
	if err := readBlockHeader(r, 0, h); err != nil {
		return err
	}

	if h.AuxPow() {
		if h.AuxPowHeader == nil {
			h.AuxPowHeader = &AuxPowHeader{}
		}
		if err := h.AuxPowHeader.Deserialize(r); err != nil {
			return err
		}
	}

	return nil
}

// FromBytes deserializes a block header byte slice.
func (h *BlockHeader) FromBytes(b []byte) error {
	r := bytes.NewReader(b)
	return h.Deserialize(r)
}

func (h *BlockHeader) Bytes() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, MaxBlockHeaderPayload))
	err := h.Serialize(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Serialize encodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *BlockHeader) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of writeBlockHeader.
	if err := writeBlockHeader(w, 0, h); err != nil {
		return err
	}

	if h.AuxPow() {
		if h.AuxPowHeader == nil {
			return fmt.Errorf("auxpow header is nil for auxpow block header: hash: %s (version: %x)", h.BlockHash(), h.Version)
		}
		if err := h.AuxPowHeader.Serialize(w); err != nil {
			return err
		}
	}

	return nil
}

// SerializeHeader encodes only the 80-byte base block header fields to w
// (Version, PrevBlock, MerkleRoot, Timestamp, Bits, Nonce) and does not
// include any AuxPoW payload regardless of the AuxPoW version bit.
func (h *BlockHeader) SerializeHeader(w io.Writer) error {
	return writeBlockHeader(w, 0, h)
}

// DeserializeHeader decodes only the 80-byte base block header fields from r
// into the receiver (Version, PrevBlock, MerkleRoot, Timestamp, Bits, Nonce)
// and ignores any AuxPoW payload regardless of the AuxPoW version bit.
func (h *BlockHeader) DeserializeHeader(r io.Reader) error {
	return readBlockHeader(r, 0, h)
}

// NewBlockHeader returns a new BlockHeader using the provided version, previous
// block hash, merkle root hash, difficulty bits, and nonce used to generate the
// block with defaults for the remaining fields.
func NewBlockHeader(version int32, prevHash, merkleRootHash *chainhash.Hash, bits uint32, nonce uint32) *BlockHeader {

	// Limit the timestamp to one second precision since the protocol
	// doesn't support better.
	return &BlockHeader{
		Version:    version,
		PrevBlock:  *prevHash,
		MerkleRoot: *merkleRootHash,
		Timestamp:  time.Unix(time.Now().Unix(), 0),
		Bits:       bits,
		Nonce:      nonce,
	}
}

// readBlockHeader reads a flokicoin block header from r.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the wire.
//
// DEPRECATED: Use readBlockHeaderBuf instead.
func readBlockHeader(r io.Reader, pver uint32, bh *BlockHeader) error {
	buf := binarySerializer.Borrow()
	err := readBlockHeaderBuf(r, pver, bh, buf)
	binarySerializer.Return(buf)
	return err
}

// readBlockHeaderBuf reads a flokicoin block header from r.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the wire.
//
// If b is non-nil, the provided buffer will be used for serializing small
// values.  Otherwise a buffer will be drawn from the binarySerializer's pool
// and return when the method finishes.
//
// NOTE: b MUST either be nil or at least an 8-byte slice.
func readBlockHeaderBuf(r io.Reader, pver uint32, bh *BlockHeader,
	buf []byte) error {

	if _, err := io.ReadFull(r, buf[:4]); err != nil {
		return err
	}
	bh.Version = int32(littleEndian.Uint32(buf[:4]))

	if _, err := io.ReadFull(r, bh.PrevBlock[:]); err != nil {
		return err
	}

	if _, err := io.ReadFull(r, bh.MerkleRoot[:]); err != nil {
		return err
	}

	if _, err := io.ReadFull(r, buf[:4]); err != nil {
		return err
	}
	bh.Timestamp = time.Unix(int64(littleEndian.Uint32(buf[:4])), 0)

	if _, err := io.ReadFull(r, buf[:4]); err != nil {
		return err
	}
	bh.Bits = littleEndian.Uint32(buf[:4])

	if _, err := io.ReadFull(r, buf[:4]); err != nil {
		return err
	}
	bh.Nonce = littleEndian.Uint32(buf[:4])

	return nil
}

// writeBlockHeader writes a flokicoin block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the wire.
//
// DEPRECATED: Use writeBlockHeaderBuf instead.
func writeBlockHeader(w io.Writer, pver uint32, bh *BlockHeader) error {
	buf := binarySerializer.Borrow()
	err := writeBlockHeaderBuf(w, pver, bh, buf)
	binarySerializer.Return(buf)
	return err
}

// writeBlockHeaderBuf writes a flokicoin block header to w.  See Serialize for
// encoding block headers to be stored to disk, such as in a database, as
// opposed to encoding for the wire.
//
// If b is non-nil, the provided buffer will be used for serializing small
// values.  Otherwise a buffer will be drawn from the binarySerializer's pool
// and return when the method finishes.
//
// NOTE: b MUST either be nil or at least an 8-byte slice.
func writeBlockHeaderBuf(w io.Writer, pver uint32, bh *BlockHeader,
	buf []byte) error {

	littleEndian.PutUint32(buf[:4], uint32(bh.Version))
	if _, err := w.Write(buf[:4]); err != nil {
		return err
	}

	if _, err := w.Write(bh.PrevBlock[:]); err != nil {
		return err
	}

	if _, err := w.Write(bh.MerkleRoot[:]); err != nil {
		return err
	}

	littleEndian.PutUint32(buf[:4], uint32(bh.Timestamp.Unix()))
	if _, err := w.Write(buf[:4]); err != nil {
		return err
	}

	littleEndian.PutUint32(buf[:4], bh.Bits)
	if _, err := w.Write(buf[:4]); err != nil {
		return err
	}

	littleEndian.PutUint32(buf[:4], bh.Nonce)
	if _, err := w.Write(buf[:4]); err != nil {
		return err
	}

	return nil
}
