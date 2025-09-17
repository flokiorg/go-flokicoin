package wire

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
)

const MaxChainBranchHashes = 30
const MaxCoinbaseTxSize = 100000

var (
	PchMergedMiningHeader []byte = []byte{0xFA, 0xBE, 'm', 'm'}
)

type MerkleBranch struct {
	Hashes   []chainhash.Hash
	SideMask uint32
}

func (mb *MerkleBranch) Size() uint {
	return uint(len(mb.Hashes))
}

func (mb *MerkleBranch) FlcEncode(w io.Writer, pver uint32) error {
	var err error

	err = WriteVarInt(w, pver, uint64(len(mb.Hashes)))
	if err != nil {
		return err
	}

	for i := range mb.Hashes {
		err = writeElement(w, &mb.Hashes[i])
		if err != nil {
			return err
		}
	}

	err = writeElement(w, mb.SideMask)
	if err != nil {
		return err
	}

	return nil
}

func (mb *MerkleBranch) Serialize(w io.Writer) error {
	return mb.FlcEncode(w, 0)
}

func (mb *MerkleBranch) FlcDecode(r io.Reader, pver uint32) error {
	n, err := ReadVarInt(r, pver)
	if err != nil {
		return err
	}

	if n > uint64(MaxChainBranchHashes) {
		return fmt.Errorf("merkle branch too large: %d > %d", n, MaxChainBranchHashes)
	}
	mb.Hashes = make([]chainhash.Hash, n)
	for i := uint64(0); i < n; i++ {
		if err := readElement(r, &mb.Hashes[i]); err != nil {
			return err
		}
	}
	if err := readElement(r, &mb.SideMask); err != nil {
		return err
	}
	return nil
}

func (mb *MerkleBranch) Deserialize(r io.Reader) error {
	return mb.FlcDecode(r, 0)
}

func (mb *MerkleBranch) SerializeSize() int {
	n := VarIntSerializeSize(uint64(len(mb.Hashes))) + chainhash.HashSize*len(mb.Hashes) + 4
	return n
}

// Determine the root hash for the Merkle tree formed from the Merkle branch
// and the component hash specified.
//
// Note: returns a value (not a pointer) to avoid any potential aliasing with
// the input component hash.
func (mb *MerkleBranch) DetermineRoot(component *chainhash.Hash) (chainhash.Hash, error) {
	m := mb.SideMask
	if component == nil {
		return chainhash.Hash{}, fmt.Errorf("component is nil")
	}

	// Work with a local copy of the component to avoid aliasing.
	h := *component
	hbuf := make([]byte, chainhash.HashSize*2)

	for i := range mb.Hashes {
		if (m & 1) != 0 {
			copy(hbuf[0:chainhash.HashSize], mb.Hashes[i][:])
			copy(hbuf[chainhash.HashSize:chainhash.HashSize*2], h[:])
		} else {
			copy(hbuf[0:chainhash.HashSize], h[:])
			copy(hbuf[chainhash.HashSize:chainhash.HashSize*2], mb.Hashes[i][:])
		}

		dh := chainhash.DoubleHashH(hbuf)
		h = dh
		m = m >> 1
	}

	return h, nil
}

func (mb *MerkleBranch) HasRoot(component *chainhash.Hash, root *chainhash.Hash) bool {
	r, err := mb.DetermineRoot(component)
	if err != nil {
		return false
	}
	return r.IsEqual(root)
}

// BlockHeader defines information about a block and is used in the flokicoin
// block (MsgBlock) and headers (MsgHeaders) messages.
type ParentAuxPowHeader struct {
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
}

// Deserialize decodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *ParentAuxPowHeader) Deserialize(r io.Reader) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of readBlockHeader.
	if err := readParentAuxBlockHeader(r, 0, h); err != nil {
		return err
	}

	return nil
}

// Serialize encodes a block header from r into the receiver using a format
// that is suitable for long-term storage such as a database while respecting
// the Version field.
func (h *ParentAuxPowHeader) Serialize(w io.Writer) error {
	// At the current time, there is no difference between the wire encoding
	// at protocol version 0 and the stable long-term storage format.  As
	// a result, make use of writeBlockHeader.
	if err := writeParentAuxBlockHeader(w, 0, h); err != nil {
		return err
	}

	return nil
}

func (h *ParentAuxPowHeader) GetChainID() int32 { return (h.Version & ChainIDMask) >> 16 }

// BlockHash computes the block identifier hash for the given block header.
func (h *ParentAuxPowHeader) BlockPoWHash() chainhash.Hash {
	return chainhash.ScryptRaw(func(w io.Writer) error {
		return writeParentAuxBlockHeader(w, 0, h)
	})
}

// BlockHash computes the block identifier hash for the given block header.
func (h *ParentAuxPowHeader) BlockHash() chainhash.Hash {
	return chainhash.DoubleHashRaw(func(w io.Writer) error {
		return writeParentAuxBlockHeader(w, 0, h)
	})
}

type AuxPowHeader struct {
	CoinbaseTx        MsgTx
	CoinbaseBranch    MerkleBranch
	BlockChainBranch  MerkleBranch
	ParentBlockHeader ParentAuxPowHeader
}

func (aph *AuxPowHeader) FlcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	if _, err := w.Write([]byte{0x00}); err != nil {
		return err
	}
	if err := aph.CoinbaseTx.FlcEncode(w, pver, enc); err != nil {
		return err
	}
	if err := aph.CoinbaseBranch.FlcEncode(w, pver); err != nil {
		return err
	}
	if err := aph.BlockChainBranch.FlcEncode(w, pver); err != nil {
		return err
	}
	return aph.ParentBlockHeader.Serialize(w)
}

func (aph *AuxPowHeader) FlcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {

	var ver [1]byte
	if _, err := io.ReadFull(r, ver[:]); err != nil {
		return err
	}
	switch ver[0] {
	case 0x00: // v0
	// ok
	default:
		return fmt.Errorf("unsupported auxpow version: %d", ver[0])
	}
	if err := aph.CoinbaseTx.FlcDecode(r, pver, enc); err != nil {
		return err
	}
	if err := aph.CoinbaseBranch.FlcDecode(r, pver); err != nil {
		return err
	}
	if err := aph.BlockChainBranch.FlcDecode(r, pver); err != nil {
		return err
	}
	return aph.ParentBlockHeader.Deserialize(r)
}

func (aph *AuxPowHeader) Deserialize(r io.Reader) error {
	return aph.FlcDecode(r, 0, BaseEncoding)
}

func (aph *AuxPowHeader) Serialize(w io.Writer) error {
	return aph.FlcEncode(w, 0, BaseEncoding)
}

func (aph *AuxPowHeader) SerializeSize() int {
	n := chainhash.HashSize + BlockHeaderLen
	n += aph.CoinbaseTx.SerializeSize()
	n += aph.CoinbaseBranch.SerializeSize()
	n += aph.BlockChainBranch.SerializeSize()
	return n
}

func (aph *AuxPowHeader) Check(auxBlockHash chainhash.Hash, chainID int32) error {

	if aph.CoinbaseTx.SerializeSize() > MaxCoinbaseTxSize {
		return fmt.Errorf("aux POW coinbase too large")
	}

	if aph.CoinbaseBranch.SideMask != 0 {
		return fmt.Errorf("auxpow is not a generate")
	}

	if aph.BlockChainBranch.Size() > MaxChainBranchHashes {
		return fmt.Errorf("auxpow chain merkle branch too long")
	}

	// Check that the chain merkle root is in the coinbase
	rootHash, err := aph.BlockChainBranch.DetermineRoot(&auxBlockHash)
	if err != nil {
		return err
	}
	revRootHash := reverseHash(rootHash)

	// Determine the hash of the coinbase transaction.
	coinbaseTxHash := aph.CoinbaseTx.TxHash()

	// Ensure that the coinbase transaction is included in the parent block.
	if !aph.CoinbaseBranch.HasRoot(&coinbaseTxHash, &aph.ParentBlockHeader.MerkleRoot) {
		return fmt.Errorf("auxpow parent block's merkle tree does not include auxpow coinbase")
	}

	if len(aph.CoinbaseTx.TxIn) == 0 {
		return fmt.Errorf("auxpow coinbase has no inputs")
	}

	script := aph.CoinbaseTx.TxIn[0].SignatureScript
	hashPos := bytes.Index(script, revRootHash[:])
	if hashPos < 0 {
		// Auxillary block hash not found in coinbase input, so this coinbase input does
		// not nominate this auxillary block and validation fails.
		return fmt.Errorf("auxpow block hash %s not found in parent block's coinbase input (%x) (%s)", auxBlockHash, script, rootHash)
	}

	headerPos := bytes.Index(script, PchMergedMiningHeader)
	if headerPos >= 0 {
		// AuxPOW header was found.
		headerPosA := bytes.Index(script[headerPos+1:], PchMergedMiningHeader)
		if headerPosA >= 0 {
			// Multiple merged mining headers in coinbase
			return fmt.Errorf("multiple auxpow headers found in parent block's coinbase input")
		}

		if (headerPos + len(PchMergedMiningHeader)) != hashPos {
			// Merged mining header is not just before chain merkle root
			return fmt.Errorf("auxpow coinbase's input has hash at wrong position")
		}
	} else {
		// AuxPOW header was not found.
		// For backward compatibility.
		if hashPos > 20 {
			// AuxPOW merkle chain must start in the first 20 bytes of the parent coinbase.
			return fmt.Errorf("auxpow coinbase's input must have header or hash starting within first 20 bytes")
		}
	}

	paramsPos := hashPos + chainhash.HashSize
	if (len(script) - paramsPos) < 8 {
		// Malformed AuxPOW structure in parent coinbase.
		return fmt.Errorf("auxpow coinbase does not contain room for params")
	}

	// "Ensure we are at a deterministic point in the merkle leaves by
	//  hashing a nonce and our chain ID and comparing to the index."
	mSize := binary.LittleEndian.Uint32(script[paramsPos : paramsPos+4])
	if mSize != (1 << aph.BlockChainBranch.Size()) {
		// AuxPOW coinbase merkle branch size does not match parent coinbase.
		return fmt.Errorf("auxpow coinbase does not specify correct merkle branch size")
	}

	// "Choose a psuedo-random slot in the chain merkle tree but have it
	//  be fixed for a size/nonce/chain combination.
	//
	//  This prevents the same work from being used twice for the same
	//  chain while reducing the chance that two chains clash for the
	//  same slot."
	mNonce := binary.LittleEndian.Uint32(script[paramsPos+4 : paramsPos+8])

	expectedIndex := getExpectedIndex(mNonce, uint32(chainID), uint32(aph.BlockChainBranch.Size()))
	if aph.BlockChainBranch.SideMask != expectedIndex {
		// AuxPOW wrong index.
		return fmt.Errorf("auxpow wrong chain index. got: %d want: %d", aph.BlockChainBranch.SideMask, expectedIndex)
	}

	return nil
}

// In wire/auxpow.go
func (aph *AuxPowHeader) String() string {
	var b bytes.Buffer

	fmt.Fprintf(&b, "AuxPowHeader{\n")
	fmt.Fprintf(&b, "  CoinbaseTx: %s\n", aph.CoinbaseTx.TxHash())
	fmt.Fprintf(&b, "  ParentBlockHash: %x\n", aph.ParentBlockHeader.BlockHash())
	fmt.Fprintf(&b, "  CoinbaseBranch: size=%d, sideMask=%#x\n", aph.CoinbaseBranch.Size(), aph.CoinbaseBranch.SideMask)
	for i, h := range aph.CoinbaseBranch.Hashes {
		fmt.Fprintf(&b, "    [%d] %s\n", i, h)
	}
	fmt.Fprintf(&b, "  BlockChainBranch: size=%d, sideMask=%#x\n", aph.BlockChainBranch.Size(), aph.BlockChainBranch.SideMask)
	for i, h := range aph.BlockChainBranch.Hashes {
		fmt.Fprintf(&b, "    [%d] %s\n", i, h)
	}
	fmt.Fprintf(&b, "  ParentBlockHeader:\n")
	fmt.Fprintf(&b, "    Version:    %d\n", aph.ParentBlockHeader.Version)
	fmt.Fprintf(&b, "    PrevBlock:  %s\n", aph.ParentBlockHeader.PrevBlock)
	fmt.Fprintf(&b, "    MerkleRoot: %s\n", aph.ParentBlockHeader.MerkleRoot)
	fmt.Fprintf(&b, "    Timestamp:  %s\n", aph.ParentBlockHeader.Timestamp.UTC())
	fmt.Fprintf(&b, "    Bits:       %08x\n", aph.ParentBlockHeader.Bits)
	fmt.Fprintf(&b, "    Nonce:      %d\n", aph.ParentBlockHeader.Nonce)
	fmt.Fprintf(&b, "}")

	return b.String()
}

func getExpectedIndex(nonce, chainID, h uint32) uint32 {
	rand := nonce
	rand = rand*1103515245 + 12345
	rand += uint32(chainID)
	rand = rand*1103515245 + 12345

	return rand % (1 << uint32(h))
}

func reverseHash(h chainhash.Hash) (r chainhash.Hash) {
	// Convert to a byte slice, reverse in place, then copy back.
	b := make([]byte, chainhash.HashSize)
	copy(b, h[:])
	reverseBytes(b)
	copy(r[:], b)
	return r
}

func reverseBytes(b []byte) {
	L := len(b)
	for i := 0; i < L/2; i++ {
		b[i], b[L-i-1] = b[L-i-1], b[i]
	}
}

// readBlockHeader reads a flokicoin block header from r.  See Deserialize for
// decoding block headers stored to disk, such as in a database, as opposed to
// decoding from the wire.
//
// DEPRECATED: Use readBlockHeaderBuf instead.
func readParentAuxBlockHeader(r io.Reader, pver uint32, bh *ParentAuxPowHeader) error {
	buf := binarySerializer.Borrow()
	err := readParentAuxBlockHeaderBuf(r, pver, bh, buf)
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
func readParentAuxBlockHeaderBuf(r io.Reader, pver uint32, bh *ParentAuxPowHeader,
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
func writeParentAuxBlockHeader(w io.Writer, pver uint32, bh *ParentAuxPowHeader) error {
	buf := binarySerializer.Borrow()
	err := writeParentAuxBlockHeaderBuf(w, pver, bh, buf)
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
func writeParentAuxBlockHeaderBuf(w io.Writer, pver uint32, bh *ParentAuxPowHeader,
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
