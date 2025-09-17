package wire

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/bits"
	"testing"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
)

func TestAuxPowCheck(t *testing.T) {
	// Child (aux) block hash (random but fixed)
	childHex := "fd5874864752d756e01849c5d4d5a35fedac5b61b8723925f8ec7b594eef20ff"
	child := mustHashFromHex(t, childHex)
	var chainID int32 = 0x21

	tests := []struct {
		name string
		opts auxPowOpts
		ok   bool
	}{
		{
			name: "valid-minimal (header present, height=0, mSize=1, correct index)",
			opts: auxPowOpts{withHeader: true},
			ok:   true,
		},
		{
			name: "missing-header-but-early (allowed: no header, hash within first 20 bytes)",
			opts: auxPowOpts{withHeader: false, forceOffset: 0},
			ok:   true,
		},
		{
			name: "missing-header-too-late (hash appears after first 20 bytes) => ErrAuxpowNoHeader",
			opts: auxPowOpts{withHeader: false, forceOffset: 64},
			ok:   false,
		},
		{
			name: "multiple-headers => ErrAuxpowMultipleHeaders",
			opts: auxPowOpts{withHeader: true, dupHeader: true},
			ok:   false,
		},
		{
			name: "wrong-index (sideMask != expected)",
			opts: func() auxPowOpts {
				o := auxPowOpts{withHeader: true}
				sm := uint32(2) // force a different mask for mSize=1 (rand%2 -> 0 or 1; our helper picks 0 for nonce 42 often; we override)
				o.overrideSideMask = &sm
				return o
			}(),
			ok: false,
		},
		{
			name: "wrong-size-in-coinbase (mSize field mismatch)",
			opts: func() auxPowOpts {
				o := auxPowOpts{withHeader: true}
				wrong := uint32(2) // mSize in coinbase != (1<<height)=1
				o.wrongCoinbaseMSz = &wrong
				return o
			}(),
			ok: false,
		},
		{
			name: "coinbase-not-in-parent-merkle",
			opts: auxPowOpts{withHeader: true, breakCoinbaseInclusion: true},
			ok:   false,
		},
		{
			name: "parent-hash-mismatch (endianness-checked) => fail",
			opts: auxPowOpts{withHeader: true, mismatchParentHash: true},
			ok:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			aph := mkAuxPow(t, child, chainID, tc.opts)
			err := aph.Check(child, chainID)
			if tc.ok && err != nil {
				t.Fatalf("expected ok, got error: %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected error, got ok")
			}
		})
	}
}

// ---------------------------
// Scenario tests requested
// ---------------------------

// TestAuxPow_SingleTx_NoBranches verifies the minimal case:
// parent has only coinbase tx so MerkleRoot == coinbase txid,
// and both coinbase branch and blockchain branch are empty.
func TestAuxPow_SingleTx_NoBranches(t *testing.T) {
	child := mustHashFromHex(t, "fd5874864752d756e01849c5d4d5a35fedac5b61b8723925f8ec7b594eef20ff")
	var chainID int32 = 0x21

	aph := mkAuxPow(t, child, chainID, auxPowOpts{withHeader: true})

	if err := aph.Check(child, chainID); err != nil {
		t.Fatalf("single-tx no-branches should pass: %v", err)
	}
}

// TestAuxPow_MultiTx_CoinbaseBranchOnly verifies a parent block with multiple
// transactions where coinbase is included via a non-empty coinbase merkle
// branch, and the aux chain blockchain branch has height 0 (empty).
func TestAuxPow_MultiTx_CoinbaseBranchOnly(t *testing.T) {
	var chainID int32 = 0x21
	child := mustHashFromHex(t, "fd5874864752d756e01849c5d4d5a35fedac5b61b8723925f8ec7b594eef20ff")

	// Aux chain branch height 0 => auxRoot == child
	mSize := uint32(1)
	mNonce := uint32(42)
	auxRoot := child

	// Build coinbase script with header and auxRoot (BE) and params.
	revAux := rev(auxRoot)
	script := makeMMCoinbaseScript(revAux[:], mSize, mNonce, true, 0, false)
	cbTx := mkCoinbaseTx(script)
	cbHash := cbTx.TxHash()

	// Create additional txids to form a multi-tx parent block.
	txids := []chainhash.Hash{cbHash}
	for i := 1; i < 5; i++ { // 5 total txs
		txids = append(txids, hashFromInt(i))
	}
	root, cbb := merkleRootAndBranch(txids, 0)

	// Parent header (chain header separate from aux chain header)
	parentVersion := int32(chainID<<16) | 1
	var zeroPrev chainhash.Hash
	pHdr := mkParentHeaderWithRoot(parentVersion, zeroPrev, root, 0x1d00ffff, 1337, time.Unix(1700000000, 0))

	// Compute expectedIndex using branch height (log2(mSize)).
	expectedIndex := getExpectedIndex(mNonce, uint32(chainID), uint32(bits.TrailingZeros32(mSize)))

	aph := &AuxPowHeader{
		CoinbaseTx:        cbTx,
		CoinbaseBranch:    cbb,
		BlockChainBranch:  MerkleBranch{Hashes: nil, SideMask: expectedIndex},
		ParentBlockHeader: pHdr,
	}

	if err := aph.Check(child, chainID); err != nil {
		t.Fatalf("multi-tx with coinbase branch only should pass: %v", err)
	}
}

// TestAuxPow_MultiTx_CoinbaseAndBlockchainBranches verifies a parent block
// with multiple transactions and a non-empty aux chain blockchain branch.
func TestAuxPow_MultiTx_CoinbaseAndBlockchainBranches(t *testing.T) {
	var chainID int32 = 0x21
	child := mustHashFromHex(t, "fd5874864752d756e01849c5d4d5a35fedac5b61b8723925f8ec7b594eef20ff")

	// Aux chain branch height 1 (two-leaf tree): compute auxRoot from child and a sibling.
	mSize := uint32(2) // 1<<1
	mNonce := uint32(42)
	expectedIndex := getExpectedIndex(mNonce, uint32(chainID), uint32(bits.TrailingZeros32(mSize)))

	sibling := hashFromInt(99)
	var auxRoot chainhash.Hash
	if expectedIndex&1 == 1 {
		// Component is right child => sibling || component
		auxRoot = chainhash.DoubleHashH(append(sibling[:], child[:]...))
	} else {
		// Component is left child => component || sibling
		auxRoot = chainhash.DoubleHashH(append(child[:], sibling[:]...))
	}

	// Build coinbase script with header and auxRoot (BE) and params.
	revAux := rev(auxRoot)
	script := makeMMCoinbaseScript(revAux[:], mSize, mNonce, true, 0, false)
	cbTx := mkCoinbaseTx(script)
	cbHash := cbTx.TxHash()

	// Multi-tx parent: build root and branch for coinbase index 0.
	txids := []chainhash.Hash{cbHash}
	for i := 1; i < 4; i++ {
		txids = append(txids, hashFromInt(i))
	}
	root, cbb := merkleRootAndBranch(txids, 0)

	parentVersion := int32(chainID<<16) | 1
	var zeroPrev chainhash.Hash
	pHdr := mkParentHeaderWithRoot(parentVersion, zeroPrev, root, 0x1d00ffff, 4242, time.Unix(1700000100, 0))

	aph := &AuxPowHeader{
		CoinbaseTx:     cbTx,
		CoinbaseBranch: cbb,
		BlockChainBranch: MerkleBranch{
			Hashes:   []chainhash.Hash{sibling},
			SideMask: expectedIndex,
		},
		ParentBlockHeader: pHdr,
	}

	if err := aph.Check(child, chainID); err != nil {
		t.Fatalf("multi-tx with both branches should pass: %v", err)
	}
}

// TestAuxPow_Invalid_ZeroTxParent synthesizes an invalid parent with a merkle
// root that cannot include coinbase (simulate zero-tx parent) and expects a
// failure from Check.
func TestAuxPow_Invalid_ZeroTxParent(t *testing.T) {
	var chainID int32 = 0x21
	child := mustHashFromHex(t, "fd5874864752d756e01849c5d4d5a35fedac5b61b8723925f8ec7b594eef20ff")

	// Height 0 aux branch in coinbase.
	mSize := uint32(1)
	mNonce := uint32(42)
	script := makeMMCoinbaseScript(child[:], mSize, mNonce, true, 0, false)
	cbTx := mkCoinbaseTx(script)

	// Bogus parent merkle root (simulate invalid/zero-tx tree).
	var bogusRoot chainhash.Hash
	for i := range bogusRoot {
		bogusRoot[i] = 0
	}

	parentVersion := int32(chainID<<16) | 1
	var zeroPrev chainhash.Hash
	pHdr := mkParentHeaderWithRoot(parentVersion, zeroPrev, bogusRoot, 0x1d00ffff, 7, time.Unix(1700000200, 0))

	aph := &AuxPowHeader{
		CoinbaseTx:        cbTx,
		CoinbaseBranch:    MerkleBranch{Hashes: nil, SideMask: 0},
		BlockChainBranch:  MerkleBranch{Hashes: nil, SideMask: 0},
		ParentBlockHeader: pHdr,
	}

	if err := aph.Check(child, chainID); err == nil {
		t.Fatalf("invalid zero-tx-like parent should fail")
	}
}

func TestAuxPowSignatureScript(t *testing.T) {
	blockhashHex := "f2e24410411b34c12b85a6998afa4594c8e7583bc7f22f08c769ba7eb15525e6"
	scriptHex := "0000000000fabe6d6df2e24410411b34c12b85a6998afa4594c8e7583bc7f22f08c769ba7eb15525e6010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	script, _ := hex.DecodeString(scriptHex)
	blockhash := mustHashFromHex(t, blockhashHex)

	t.Logf("script: %x blockhash: %s", script, blockhash)
	t.Logf("found: %v", bytes.Index(script, blockhash[:]))
}

// TestAuxPow_IndexComputation_Table focuses specifically on the index
// computation from (mSize, mNonce, chainID) and the BlockChainBranch.SideMask.
// It uses a table of scenarios and asserts that:
// - When SideMask == getExpectedIndex(mNonce, chainID, height) the check passes
// - When SideMask != expected, the check fails with the index error
func TestAuxPow_IndexComputation_Table(t *testing.T) {
	child := mustHashFromHex(t, "fd5874864752d756e01849c5d4d5a35fedac5b61b8723925f8ec7b594eef20ff")
	var chainID int32 = 0x21

	type tc struct {
		name   string
		mSize  uint32
		mNonce uint32
		match  bool // whether to use matching SideMask or intentionally mismatch
	}

	cases := []tc{
		{name: "size1_nonce0_match", mSize: 1, mNonce: 0, match: true},
		{name: "size1_nonce1_match", mSize: 1, mNonce: 1, match: true},
		{name: "size1_nonce1_mismatch", mSize: 1, mNonce: 1, match: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Compute expected index using height = log2(mSize)
			expected := getExpectedIndex(c.mNonce, uint32(chainID), uint32(bits.TrailingZeros32(c.mSize)))
			side := expected
			if !c.match {
				// Force a different value even if mSize==1 (expected is 0), by setting to 1.
				// Check() compares equality with expectedIndex, it does not bound SideMask here.
				if expected == 0 {
					side = 1
				} else {
					side = expected - 1
				}
			}

			// Build opts
			opts := auxPowOpts{withHeader: true}
			opts.overrideMSz = &c.mSize
			opts.overrideSideMask = &side
			opts.overrideNonce = &c.mNonce

			// Coinbase mSize matches overrideMSz above.

			aph := mkAuxPow(t, child, chainID, opts)

			err := aph.Check(child, chainID)
			if c.match {
				if err != nil {
					t.Fatalf("expected pass, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected failure on wrong index, got ok")
				}
			}
		})
	}
}

func leU32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

func mustHashFromHex(t *testing.T, h string) chainhash.Hash {
	b, err := hex.DecodeString(h)
	if err != nil {
		t.Fatalf("bad hex: %v", err)
	}
	if len(b) != chainhash.HashSize {
		t.Fatalf("want 32 bytes, got %d", len(b))
	}
	var out chainhash.Hash
	copy(out[:], b)
	return out
}

func rev(h chainhash.Hash) [32]byte { return reverseHash(h) }

// ---------------------------
// Coinbase / header builders
// ---------------------------

// makeMMCoinbaseScript builds a coinbase sigScript that contains either:
//
//	header (optional) || auxRootBE || mSize(LE u32) || mNonce(LE u32)
//
// header must be immediately before auxRoot if withHeader==true,
// and the whole “start“ must appear within the first 20 bytes (per spec),
// unless forceOffset>20 is requested to make it invalid.
func makeMMCoinbaseScript(auxRootBE []byte, mSize, mNonce uint32, withHeader bool, forceOffset int, dupHeader bool) []byte {
	var pre bytes.Buffer
	// pad up to forceOffset bytes if requested
	if forceOffset > 0 {
		if forceOffset < 20 {
			// ensure it's still within 20 for valid cases
			forceOffset = 20 - forceOffset
		}
		pre.Write(bytes.Repeat([]byte{0}, forceOffset))
	} else {
		// A few prefix bytes (kept small so offset < 20).
		pre.Write([]byte{0, 0, 0, 0, 0})
	}

	var body bytes.Buffer
	if withHeader {
		body.Write(PchMergedMiningHeader) // 0xFA 0xBE 'm' 'm'
	}
	body.Write(auxRootBE)     // aux root (BE) must follow header immediately when present
	body.Write(leU32(mSize))  // branch size
	body.Write(leU32(mNonce)) // nonce

	if dupHeader {
		// Intentionally add a second header further in to trigger "multiple headers" error.
		body.Write(PchMergedMiningHeader)
	}

	// Return a sigScript-like blob (not a full script program; MsgTx just carries bytes).
	return append(pre.Bytes(), body.Bytes()...)
}

// mkCoinbaseTx wraps the script into a minimal coinbase tx with 1 input and 1 zero output.
func mkCoinbaseTx(sigScript []byte) MsgTx {
	var tx MsgTx
	tx.Version = 1
	tx.TxIn = []*TxIn{{
		PreviousOutPoint: OutPoint{}, // coinbase prevout
		SignatureScript:  sigScript,
		Sequence:         0xffffffff,
	}}
	tx.TxOut = []*TxOut{{
		Value:    0,
		PkScript: []byte{0x6a}, // OP_RETURN (placeholder)
	}}
	return tx
}

// mkParentHeader builds a parent header where MerkleRoot == coinbase txid (only coinbase in block).
func mkParentHeader(version int32, prev chainhash.Hash, coinbaseTx MsgTx, bits uint32, nonce uint32, ts time.Time) ParentAuxPowHeader {
	cbHash := coinbaseTx.TxHash() // txid is LE in memory; write function handles endianness
	return ParentAuxPowHeader{
		Version:    version,
		PrevBlock:  prev,
		MerkleRoot: cbHash,
		Timestamp:  ts,
		Bits:       bits,
		Nonce:      nonce,
	}
}

// mkParentHeaderWithRoot is like mkParentHeader but allows explicitly setting
// the parent MerkleRoot (used for multi-tx and error scenarios).
func mkParentHeaderWithRoot(version int32, prev chainhash.Hash, merkleRoot chainhash.Hash, bits uint32, nonce uint32, ts time.Time) ParentAuxPowHeader {
	return ParentAuxPowHeader{
		Version:    version,
		PrevBlock:  prev,
		MerkleRoot: merkleRoot,
		Timestamp:  ts,
		Bits:       bits,
		Nonce:      nonce,
	}
}

// mkAuxPow constructs a *valid* AuxPowHeader for a given child block hash and chainID.
// You can tweak flags to fabricate invalid cases.
type auxPowOpts struct {
	withHeader       bool // include merged-mining header bytes
	forceOffset      int  // if >20, triggers "must start within first 20 bytes" path
	dupHeader        bool // insert a second header to trigger "multiple headers"
	overrideMSz      *uint32
	overrideNonce    *uint32
	overrideSideMask *uint32
	// Mismatch parent hash vs header
	mismatchParentHash bool
	// Wrong branch size in coinbase payload (mSize field) to trigger size mismatch
	wrongCoinbaseMSz *uint32
	// Wrong coinbase merkle (parent merkle root does not include coinbase)
	breakCoinbaseInclusion bool
}

func mkAuxPow(t *testing.T, childHash chainhash.Hash, chainID int32, opts auxPowOpts) *AuxPowHeader {
	t.Helper()

	// For simplicity: aux chain merkle branch height 0 => mSize = 1
	// SideMask must equal getExpectedIndex(mNonce, chainID, height).
	mSize := uint32(1) // 1 << height(0)
	mNonce := uint32(42)
	if opts.overrideMSz != nil {
		mSize = *opts.overrideMSz
	}
	if opts.overrideNonce != nil {
		mNonce = *opts.overrideNonce
	}

	expectedIndex := getExpectedIndex(mNonce, uint32(chainID), uint32(bits.TrailingZeros32(mSize)))
	sideMask := expectedIndex
	if opts.overrideSideMask != nil {
		sideMask = *opts.overrideSideMask
	}

	// Aux root used in coinbase payload must be big-endian (display) bytes of child hash (height 0 case).
	auxRootBE := rev(childHash)
	// coinbase script
	cmSize := mSize
	if opts.wrongCoinbaseMSz != nil {
		cmSize = *opts.wrongCoinbaseMSz
	}
	script := makeMMCoinbaseScript(auxRootBE[:], cmSize, mNonce, opts.withHeader, opts.forceOffset, opts.dupHeader)
	cbTx := mkCoinbaseTx(script)

	// Parent header: single-tx block -> merkle root = coinbase txid
	parentVersion := int32(chainID<<16) | 1
	var zeroPrev chainhash.Hash
	parentBits := uint32(0x1d00ffff)
	parentNonce := uint32(247178490)
	parentTime := time.Unix(1600101920, 0) // fixed for determinism
	pHdr := mkParentHeader(parentVersion, zeroPrev, cbTx, parentBits, parentNonce, parentTime)

	// Construct AuxPowHeader
	aph := &AuxPowHeader{
		CoinbaseTx:        cbTx,
		CoinbaseBranch:    MerkleBranch{Hashes: nil, SideMask: 0}, // only coinbase -> empty branch
		BlockChainBranch:  MerkleBranch{Hashes: nil, SideMask: sideMask},
		ParentBlockHeader: pHdr,
	}

	// Optionally break coinbase inclusion by altering parent header merkle root.
	if opts.breakCoinbaseInclusion {
		var bogus chainhash.Hash
		for i := 0; i < len(bogus); i++ {
			bogus[i] = 0xAA
		}
		aph.ParentBlockHeader.MerkleRoot = bogus
	}

	// Simulate a parent hash mismatch (endianness) by reversing the
	// coinbase txid and forcing it as the parent merkle root. This causes
	// the coinbase inclusion check to fail in Check() as intended by the
	// test case.
	if opts.mismatchParentHash {
		cbHash := cbTx.TxHash()
		revHash := rev(cbHash)
		aph.ParentBlockHeader.MerkleRoot = revHash
	}

	return aph

	return aph
}

// ---------------------------
// Merkle helpers (parent side)
// ---------------------------

// hashFromInt deterministically derives a hash from an int.
func hashFromInt(i int) chainhash.Hash {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(i))
	return chainhash.DoubleHashH(b[:])
}

// merkleRootAndBranch computes the Bitcoin-style Merkle root for txids and the
// merkle branch needed to connect the tx at index to the root. If the number of
// leaves on a level is odd, the last is duplicated.
func merkleRootAndBranch(txids []chainhash.Hash, index int) (chainhash.Hash, MerkleBranch) {
	if len(txids) == 0 {
		var zero chainhash.Hash
		return zero, MerkleBranch{Hashes: nil, SideMask: 0}
	}
	// Work on a copy.
	level := make([]chainhash.Hash, len(txids))
	copy(level, txids)

	var branch []chainhash.Hash
	var mask uint32
	idx := index

	for len(level) > 1 {
		// Determine sibling for current idx and add to branch.
		// If odd number of nodes, duplicate the last.
		if len(level)%2 == 1 {
			level = append(level, level[len(level)-1])
		}
		siblingIndex := idx ^ 1
		branch = append(branch, level[siblingIndex])
		// Set bit if current node is right child.
		if idx&1 == 1 {
			mask |= (1 << uint32(len(branch)-1))
		}

		// Build next level.
		next := make([]chainhash.Hash, len(level)/2)
		for i := 0; i < len(level); i += 2 {
			left := level[i]
			right := level[i+1]
			next[i/2] = chainhash.DoubleHashH(append(left[:], right[:]...))
		}
		level = next
		idx /= 2
	}

	return level[0], MerkleBranch{Hashes: branch, SideMask: mask}
}
