package wire

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
)

const MaxCoinbaseTxSize = 100000
const MaxBranchHashes = 30
const MaxBranchSize = 4 + MaxBranchHashes*chainhash.HashSize
const MaxAuxPowSize = MaxCoinbaseTxSize + chainhash.HashSize + MaxBranchSize*2 + MaxBlockHeaderPayload

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

	if n > 0x02000000 {
		return fmt.Errorf("size too large")
	}

	mb.Hashes = make([]chainhash.Hash, n)

	for i := uint64(0); i < n; i++ {
		err = readElement(r, &mb.Hashes[i])
		if err != nil {
			return err
		}
	}

	err = readElement(r, &mb.SideMask)
	if err != nil {
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
func (mb *MerkleBranch) DetermineRoot(component *chainhash.Hash) (h *chainhash.Hash, err error) {
	//log.Printf("MerkleBranch: DetermineRoot (component=%s)", component.String())
	//log.Printf("MerkleBranch contains %d hashes (0x%08x):", len(mb.Hashes), mb.SideMask)

	m := mb.SideMask
	h = component
	hbuf := make([]byte, chainhash.HashSize*2)

	if component == nil {
		panic("component must be specified")
	}

	for i := range mb.Hashes {
		//log.Printf("  %s", mb.Hashes[i].String())

		if (m & 1) != 0 {
			copy(hbuf[0:chainhash.HashSize], mb.Hashes[i][:])
			copy(hbuf[chainhash.HashSize:chainhash.HashSize*2], h[:])
		} else {
			copy(hbuf[0:chainhash.HashSize], h[:])
			copy(hbuf[chainhash.HashSize:chainhash.HashSize*2], mb.Hashes[i][:])
		}

		dh := chainhash.DoubleHashH(hbuf)
		h = &dh
		m = m >> 1
	}

	return
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

func (h *ParentAuxPowHeader) GetChainID() int32 {
	return h.Version >> 16
}

// BlockHash computes the block identifier hash for the given block header.
func (h *ParentAuxPowHeader) BlockPoWHash() chainhash.Hash {
	return chainhash.ScryptRaw(func(w io.Writer) error {
		return writeParentAuxBlockHeader(w, 0, h)
	})
}

type AuxPowHeader struct {
	CoinbaseTx        MsgTx
	ParentBlockHash   chainhash.Hash
	CoinbaseBranch    MerkleBranch
	BlockChainBranch  MerkleBranch
	ParentBlockHeader ParentAuxPowHeader
}

// CAuxPow
//   CMerkleTx
//     CTransaction
//       nVersion       int
//       vin            vector<CTxIn>
//       vout           vector<CTxOut>
//       nLockTime      unsigned int
//     hashBlock        uint256
//     vMerkleBranch    vector<uint256>
//     nIndex           int
//   vChainMerkleBranch vector<uint256>  // } These are the Merkle branch?
//   nChainIndex int                     // }
//   parentBlock CBlock (header only)

func (aph *AuxPowHeader) FlcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	err := aph.CoinbaseTx.FlcEncode(w, pver, enc)
	if err != nil {
		return err
	}

	err = writeElement(w, &aph.ParentBlockHash)
	if err != nil {
		return err
	}

	err = aph.CoinbaseBranch.FlcEncode(w, pver)
	if err != nil {
		return err
	}

	err = aph.BlockChainBranch.FlcEncode(w, pver)
	if err != nil {
		return err
	}

	err = aph.ParentBlockHeader.Serialize(w)
	if err != nil {
		return err
	}

	return nil
}

func (aph *AuxPowHeader) FlcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	err := aph.CoinbaseTx.FlcDecode(r, pver, enc)
	if err != nil {
		return err
	}

	err = readElement(r, &aph.ParentBlockHash)
	if err != nil {
		return err
	}

	err = aph.CoinbaseBranch.FlcDecode(r, pver)
	if err != nil {
		return err
	}

	err = aph.BlockChainBranch.FlcDecode(r, pver)
	if err != nil {
		return err
	}

	err = aph.ParentBlockHeader.Deserialize(r)
	if err != nil {
		return err
	}

	return nil
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

	if aph.BlockChainBranch.Size() > MaxBranchHashes {
		return fmt.Errorf("Aux POW chain merkle branch too long")
	}

	// Step 1. Determine the root hash of the block chain merkle tree.
	rootHash, err := aph.BlockChainBranch.DetermineRoot(&auxBlockHash) // rootHash
	if err != nil {
		return err
	}
	revRootHash := reverseHash(*rootHash)

	// Step 3. Determine the hash of the coinbase transaction.
	coinbaseTxHash := aph.CoinbaseTx.TxHash()

	// Step 4. Ensure that the coinbase transaction is included in the parent
	// block.
	if !aph.CoinbaseBranch.HasRoot(&coinbaseTxHash, &aph.ParentBlockHeader.MerkleRoot) {
		return fmt.Errorf("Auxpow parent block's merkle tree does not include auxpow coinbase")
	}

	if len(aph.CoinbaseTx.TxIn) == 0 {
		return fmt.Errorf("Aux POW coinbase has no inputs")
	}

	txIn := aph.CoinbaseTx.TxIn[0]
	script := txIn.SignatureScript

	hashPos := bytes.Index(script, revRootHash[:])
	if hashPos < 0 {
		// Auxillary block hash not found in coinbase input, so this coinbase input does
		// not nominate this auxillary block and validation fails.
		fmt.Printf("  Script:  %x\n", script)
		str := fmt.Sprintf("Auxpow block hash %s not found in parent block's coinbase input (%x) (%x)",
			rootHash.String(), rootHash[:], revRootHash[:])
		return messageError("ErrAuxpowCoinbaseHashNotFound", str)
	}

	headerPos := bytes.Index(script, PchMergedMiningHeader)
	if headerPos >= 0 {
		// AuxPOW header was found.

		// Namecoin: "Enforce only one chain merkle root by checking that a single instance
		// instance of the merged mining header exists just before."
		//
		// The code then proceeds to search the string beginning one byte after headerPos
		// for the auxPowHeader.
		//
		// This excludes any auxpow block the header of which contains 0xFA,0xBE,'m','m'.
		// But since bug-for-bug compatibility is the order of the day...
		headerPosA := bytes.Index(script[headerPos+1:], PchMergedMiningHeader)
		if headerPosA >= 0 {
			// Multiple merged mining headers in coinbase
			return messageError("ErrAuxpowMultipleHeaders",
				"Multiple auxpow headers found in parent block's coinbase input")
		}

		if (headerPos + len(PchMergedMiningHeader)) != hashPos {
			// Merged mining header is not just before chain merkle root
			return messageError("ErrAuxpowBadHashPosition",
				"Auxpow coinbase's input has hash at wrong position")
		}
	} else {
		// AuxPOW header was not found.
		// For backward compatibility.
		if hashPos > 20 {
			// AuxPOW merkle chain must start in the first 20 bytes of the parent coinbase.
			return messageError("ErrAuxpowNoHeader",
				"Auxpow coinbase's input must have header or hash starting within first 20 bytes")
		}
	}

	paramsPos := hashPos + chainhash.HashSize
	if (len(script) - paramsPos) < 8 {
		// Malformed AuxPOW structure in parent coinbase.
		return messageError("ErrAuxpowMalformedCoinbase",
			"Auxpow coinbase does not contain room for params")
	}

	// "Ensure we are at a deterministic point in the merkle leaves by
	//  hashing a nonce and our chain ID and comparing to the index."
	mSize := binary.LittleEndian.Uint32(script[paramsPos : paramsPos+4])
	if mSize != (1 << aph.BlockChainBranch.Size()) {
		// AuxPOW coinbase merkle branch size does not match parent coinbase.
		return messageError("ErrAuxpowWrongSize",
			"Auxpow coinbase does not specify correct merkle branch size")
	}

	// "Choose a psuedo-random slot in the chain merkle tree but have it
	//  be fixed for a size/nonce/chain combination.
	//
	//  This prevents the same work from being used twice for the same
	//  chain while reducing the chance that two chains clash for the
	//  same slot."
	mNonce := binary.LittleEndian.Uint32(script[paramsPos+4 : paramsPos+8])

	// r := mNonce
	// r = r*1103515245 + 12345
	// r += chainID
	// r = r*1103515245 + 12345
	// if h.BlockChainBranch.SideMask != (r % mSize) {

	if aph.BlockChainBranch.SideMask != getExpectedIndex(mNonce, uint32(chainID), mSize) {
		// AuxPOW wrong index.
		return messageError("ErrAuxpowWrongIndex",
			"Auxpow coinbase does not specify correct index")
	}

	return nil
}

func getExpectedIndex(nonce, chainID, h uint32) uint32 {
	rand := nonce
	rand = rand*1103515245 + 12345
	rand += uint32(chainID)
	rand = rand*1103515245 + 12345

	return rand % (1 << uint32(h))
}

func reverseHash(h chainhash.Hash) (r [chainhash.HashSize]byte) {
	b := make([]byte, chainhash.HashSize)
	copy(b, h[0:chainhash.HashSize])
	reverseBytes(b)
	copy(r[:], b[0:chainhash.HashSize])
	return
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
