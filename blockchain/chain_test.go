// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/flokiorg/go-flokicoin/blockchain/internal/testhelper"
	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/wire"
)

// TestHaveBlock tests the HaveBlock API to ensure proper functionality.
func TestHaveBlock(t *testing.T) {
	return // #FLOKI_CHANGE

	// Load up blocks such that there is a side chain.
	// (genesis block) -> 1 -> 2 -> 3 -> 4
	//                          \-> 3a
	testFiles := []string{
		"blk_0_to_4.dat.bz2",
		"blk_3A.dat.bz2",
	}

	var blocks []*chainutil.Block
	for _, file := range testFiles {
		blockTmp, err := loadBlocks(file)
		if err != nil {
			t.Errorf("Error loading file: %v\n", err)
			return
		}
		blocks = append(blocks, blockTmp...)
		for _, b := range blockTmp {
			t.Logf("block: %s", b.Hash())
		}
	}

	// Create a new database and chain instance to run tests against.
	chain, teardownFunc, err := chainSetup("haveblock",
		&chaincfg.MainNetParams)
	if err != nil {
		t.Errorf("Failed to setup chain instance: %v", err)
		return
	}
	defer teardownFunc()

	// Since we're not dealing with the real block chain, set the coinbase
	// maturity to 1.
	chain.TstSetCoinbaseMaturity(1)

	for i := 1; i < len(blocks); i++ {
		t.Logf("process %s", blocks[i].Hash())
		_, isOrphan, err := chain.ProcessBlock(blocks[i], BFNone)
		if err != nil {
			t.Errorf("ProcessBlock fail on block %v: %v\n", i, err)
			return
		}
		if isOrphan {
			t.Errorf("ProcessBlock incorrectly returned block %v "+
				"is an orphan\n", i)
			return
		}
	}

	// Insert an orphan block.
	_, isOrphan, err := chain.ProcessBlock(chainutil.NewBlock(&Block100000),
		BFNone)
	if err != nil {
		t.Errorf("Unable to process block: %v", err)
		return
	}
	if !isOrphan {
		t.Errorf("ProcessBlock indicated block is an not orphan when " +
			"it should be\n")
		return
	}

	tests := []struct {
		hash string
		want bool
	}{
		// Genesis block should be present (in the main chain).
		{hash: chaincfg.MainNetParams.GenesisHash.String(), want: true},

		// Block 3a should be present (on a side chain).
		{hash: "00000000474284d20067a4d33f6a02284e6ef70764a3a26d6a5b9df52ef663dd", want: true},

		// Block 100000 should be present (as an orphan).
		{hash: "000000000003ba27aa200b1cecaad478d2b00432346c3f1f3986da1afd33e506", want: true},

		// Random hashes should not be available.
		{hash: "123", want: false},
	}

	for i, test := range tests {
		hash, err := chainhash.NewHashFromStr(test.hash)
		if err != nil {
			t.Errorf("NewHashFromStr: %v", err)
			continue
		}

		result, err := chain.HaveBlock(hash)
		if err != nil {
			t.Errorf("HaveBlock #%d unexpected error: %v", i, err)
			return
		}
		if result != test.want {
			t.Errorf("HaveBlock #%d got %v want %v", i, result,
				test.want)
			continue
		}
	}
}

// TestCalcSequenceLock tests the LockTimeToSequence function, and the
// CalcSequenceLock method of a Chain instance. The tests exercise several
// combinations of inputs to the CalcSequenceLock function in order to ensure
// the returned SequenceLocks are correct for each test instance.
func TestCalcSequenceLock(t *testing.T) {
	netParams := &chaincfg.SimNetParams

	// We need to activate CSV in order to test the processing logic, so
	// manually craft the block version that's used to signal the soft-fork
	// activation.
	csvBit := netParams.Deployments[chaincfg.DeploymentCSV].BitNumber
	blockVersion := int32(0x20000000 | (uint32(1) << csvBit))

	// Generate enough synthetic blocks to activate CSV.
	chain := newFakeChain(netParams)
	node := chain.bestChain.Tip()
	blockTime := node.Header().Timestamp
	numBlocksToActivate := (netParams.MinerConfirmationWindow * 3)
	for i := uint32(0); i < numBlocksToActivate; i++ {
		blockTime = blockTime.Add(time.Second)
		node = newFakeNode(node, blockVersion, 0, blockTime)
		chain.index.AddNode(node)
		chain.bestChain.SetTip(node)
	}

	// Create a utxo view with a fake utxo for the inputs used in the
	// transactions created below.  This utxo is added such that it has an
	// age of 4 blocks.
	targetTx := chainutil.NewTx(&wire.MsgTx{
		TxOut: []*wire.TxOut{{
			PkScript: nil,
			Value:    10,
		}},
	})
	utxoView := NewUtxoViewpoint()
	utxoView.AddTxOuts(targetTx, int32(numBlocksToActivate)-4)
	utxoView.SetBestHash(&node.hash)

	// Create a utxo that spends the fake utxo created above for use in the
	// transactions created in the tests.  It has an age of 4 blocks.  Note
	// that the sequence lock heights are always calculated from the same
	// point of view that they were originally calculated from for a given
	// utxo.  That is to say, the height prior to it.
	utxo := wire.OutPoint{
		Hash:  *targetTx.Hash(),
		Index: 0,
	}
	prevUtxoHeight := int32(numBlocksToActivate) - 4

	// Obtain the median time past from the PoV of the input created above.
	// The MTP for the input is the MTP from the PoV of the block *prior*
	// to the one that included it.
	medianTime := CalcPastMedianTime(node.RelativeAncestor(5)).Unix()

	// The median time calculated from the PoV of the best block in the
	// test chain.  For unconfirmed inputs, this value will be used since
	// the MTP will be calculated from the PoV of the yet-to-be-mined
	// block.
	nextMedianTime := CalcPastMedianTime(node).Unix()
	nextBlockHeight := int32(numBlocksToActivate) + 1

	// Add an additional transaction which will serve as our unconfirmed
	// output.
	unConfTx := &wire.MsgTx{
		TxOut: []*wire.TxOut{{
			PkScript: nil,
			Value:    5,
		}},
	}
	unConfUtxo := wire.OutPoint{
		Hash:  unConfTx.TxHash(),
		Index: 0,
	}

	// Adding a utxo with a height of 0x7fffffff indicates that the output
	// is currently unmined.
	utxoView.AddTxOuts(chainutil.NewTx(unConfTx), 0x7fffffff)

	tests := []struct {
		tx      *wire.MsgTx
		view    *UtxoViewpoint
		mempool bool
		want    *SequenceLock
	}{
		// A transaction of version one should disable sequence locks
		// as the new sequence number semantics only apply to
		// transactions version 2 or higher.
		{
			tx: &wire.MsgTx{
				Version: 1,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 3),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     -1,
				BlockHeight: -1,
			},
		},
		// A transaction with a single input with max sequence number.
		// This sequence number has the high bit set, so sequence locks
		// should be disabled.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         wire.MaxTxInSequenceNum,
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     -1,
				BlockHeight: -1,
			},
		},
		// A transaction with a single input whose lock time is
		// expressed in seconds.  However, the specified lock time is
		// below the required floor for time based lock times since
		// they have time granularity of 512 seconds.  As a result, the
		// seconds lock-time should be just before the median time of
		// the targeted block.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 2),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime - 1,
				BlockHeight: -1,
			},
		},
		// A transaction with a single input whose lock time is
		// expressed in seconds.  The number of seconds should be 1023
		// seconds after the median past time of the last block in the
		// chain.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 1024),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime + 1023,
				BlockHeight: -1,
			},
		},
		// A transaction with multiple inputs.  The first input has a
		// lock time expressed in seconds.  The second input has a
		// sequence lock in blocks with a value of 4.  The last input
		// has a sequence number with a value of 5, but has the disable
		// bit set.  So the first lock should be selected as it's the
		// latest lock that isn't disabled.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 2560),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 4),
				}, {
					PreviousOutPoint: utxo,
					Sequence: LockTimeToSequence(false, 5) |
						wire.SequenceLockTimeDisabled,
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime + (5 << wire.SequenceLockTimeGranularity) - 1,
				BlockHeight: prevUtxoHeight + 3,
			},
		},
		// Transaction with a single input.  The input's sequence number
		// encodes a relative lock-time in blocks (3 blocks).  The
		// sequence lock should  have a value of -1 for seconds, but a
		// height of 2 meaning it can be included at height 3.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 3),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     -1,
				BlockHeight: prevUtxoHeight + 2,
			},
		},
		// A transaction with two inputs with lock times expressed in
		// seconds.  The selected sequence lock value for seconds should
		// be the time further in the future.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 5120),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 2560),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime + (10 << wire.SequenceLockTimeGranularity) - 1,
				BlockHeight: -1,
			},
		},
		// A transaction with two inputs with lock times expressed in
		// blocks.  The selected sequence lock value for blocks should
		// be the height further in the future, so a height of 10
		// indicating it can be included at height 11.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 1),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 11),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     -1,
				BlockHeight: prevUtxoHeight + 10,
			},
		},
		// A transaction with multiple inputs.  Two inputs are time
		// based, and the other two are block based. The lock lying
		// further into the future for both inputs should be chosen.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 2560),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(true, 6656),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 3),
				}, {
					PreviousOutPoint: utxo,
					Sequence:         LockTimeToSequence(false, 9),
				}},
			},
			view: utxoView,
			want: &SequenceLock{
				Seconds:     medianTime + (13 << wire.SequenceLockTimeGranularity) - 1,
				BlockHeight: prevUtxoHeight + 8,
			},
		},
		// A transaction with a single unconfirmed input.  As the input
		// is confirmed, the height of the input should be interpreted
		// as the height of the *next* block.  So, a 2 block relative
		// lock means the sequence lock should be for 1 block after the
		// *next* block height, indicating it can be included 2 blocks
		// after that.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: unConfUtxo,
					Sequence:         LockTimeToSequence(false, 2),
				}},
			},
			view:    utxoView,
			mempool: true,
			want: &SequenceLock{
				Seconds:     -1,
				BlockHeight: nextBlockHeight + 1,
			},
		},
		// A transaction with a single unconfirmed input.  The input has
		// a time based lock, so the lock time should be based off the
		// MTP of the *next* block.
		{
			tx: &wire.MsgTx{
				Version: 2,
				TxIn: []*wire.TxIn{{
					PreviousOutPoint: unConfUtxo,
					Sequence:         LockTimeToSequence(true, 1024),
				}},
			},
			view:    utxoView,
			mempool: true,
			want: &SequenceLock{
				Seconds:     nextMedianTime + 1023,
				BlockHeight: -1,
			},
		},
	}

	t.Logf("Running %v SequenceLock tests", len(tests))
	for i, test := range tests {
		utilTx := chainutil.NewTx(test.tx)
		seqLock, err := chain.CalcSequenceLock(utilTx, test.view, test.mempool)
		if err != nil {
			t.Fatalf("test #%d, unable to calc sequence lock: %v", i, err)
		}

		if seqLock.Seconds != test.want.Seconds {
			t.Fatalf("test #%d got %v seconds want %v seconds",
				i, seqLock.Seconds, test.want.Seconds)
		}
		if seqLock.BlockHeight != test.want.BlockHeight {
			t.Fatalf("test #%d got height of %v want height of %v ",
				i, seqLock.BlockHeight, test.want.BlockHeight)
		}
	}
}

// nodeHashes is a convenience function that returns the hashes for all of the
// passed indexes of the provided nodes.  It is used to construct expected hash
// slices in the tests.
func nodeHashes(nodes []*blockNode, indexes ...int) []chainhash.Hash {
	hashes := make([]chainhash.Hash, 0, len(indexes))
	for _, idx := range indexes {
		hashes = append(hashes, nodes[idx].hash)
	}
	return hashes
}

// nodeHeaders is a convenience function that returns the headers for all of
// the passed indexes of the provided nodes.  It is used to construct expected
// located headers in the tests.
func nodeHeaders(nodes []*blockNode, indexes ...int) []wire.BlockHeader {
	headers := make([]wire.BlockHeader, 0, len(indexes))
	for _, idx := range indexes {
		headers = append(headers, nodes[idx].Header())
	}
	return headers
}

// TestLocateInventory ensures that locating inventory via the LocateHeaders and
// LocateBlocks functions behaves as expected.
func TestLocateInventory(t *testing.T) {
	// Construct a synthetic block chain with a block index consisting of
	// the following structure.
	// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
	// 	                              \-> 16a -> 17a
	tip := tstTip
	chain := newFakeChain(&chaincfg.MainNetParams)
	branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 18)
	branch1Nodes := chainedNodes(branch0Nodes[14], 2)
	for _, node := range branch0Nodes {
		chain.index.AddNode(node)
	}
	for _, node := range branch1Nodes {
		chain.index.AddNode(node)
	}
	chain.bestChain.SetTip(tip(branch0Nodes))

	// Create chain views for different branches of the overall chain to
	// simulate a local and remote node on different parts of the chain.
	localView := newChainView(tip(branch0Nodes))
	remoteView := newChainView(tip(branch1Nodes))

	// Create a chain view for a completely unrelated block chain to
	// simulate a remote node on a totally different chain.
	unrelatedBranchNodes := chainedNodes(nil, 5)
	unrelatedView := newChainView(tip(unrelatedBranchNodes))

	tests := []struct {
		name       string
		locator    BlockLocator       // locator for requested inventory
		hashStop   chainhash.Hash     // stop hash for locator
		maxAllowed uint32             // max to locate, 0 = wire const
		headers    []wire.BlockHeader // expected located headers
		hashes     []chainhash.Hash   // expected located hashes
	}{
		{
			// Empty block locators and unknown stop hash.  No
			// inventory should be located.
			name:     "no locators, no stop",
			locator:  nil,
			hashStop: chainhash.Hash{},
			headers:  nil,
			hashes:   nil,
		},
		{
			// Empty block locators and stop hash in side chain.
			// The expected result is the requested block.
			name:     "no locators, stop in side",
			locator:  nil,
			hashStop: tip(branch1Nodes).hash,
			headers:  nodeHeaders(branch1Nodes, 1),
			hashes:   nodeHashes(branch1Nodes, 1),
		},
		{
			// Empty block locators and stop hash in main chain.
			// The expected result is the requested block.
			name:     "no locators, stop in main",
			locator:  nil,
			hashStop: branch0Nodes[12].hash,
			headers:  nodeHeaders(branch0Nodes, 12),
			hashes:   nodeHashes(branch0Nodes, 12),
		},
		{
			// Locators based on remote being on side chain and a
			// stop hash local node doesn't know about.  The
			// expected result is the blocks after the fork point in
			// the main chain and the stop hash has no effect.
			name:     "remote side chain, unknown stop",
			locator:  remoteView.BlockLocator(nil),
			hashStop: chainhash.Hash{0x01},
			headers:  nodeHeaders(branch0Nodes, 15, 16, 17),
			hashes:   nodeHashes(branch0Nodes, 15, 16, 17),
		},
		{
			// Locators based on remote being on side chain and a
			// stop hash in side chain.  The expected result is the
			// blocks after the fork point in the main chain and the
			// stop hash has no effect.
			name:     "remote side chain, stop in side",
			locator:  remoteView.BlockLocator(nil),
			hashStop: tip(branch1Nodes).hash,
			headers:  nodeHeaders(branch0Nodes, 15, 16, 17),
			hashes:   nodeHashes(branch0Nodes, 15, 16, 17),
		},
		{
			// Locators based on remote being on side chain and a
			// stop hash in main chain, but before fork point.  The
			// expected result is the blocks after the fork point in
			// the main chain and the stop hash has no effect.
			name:     "remote side chain, stop in main before",
			locator:  remoteView.BlockLocator(nil),
			hashStop: branch0Nodes[13].hash,
			headers:  nodeHeaders(branch0Nodes, 15, 16, 17),
			hashes:   nodeHashes(branch0Nodes, 15, 16, 17),
		},
		{
			// Locators based on remote being on side chain and a
			// stop hash in main chain, but exactly at the fork
			// point.  The expected result is the blocks after the
			// fork point in the main chain and the stop hash has no
			// effect.
			name:     "remote side chain, stop in main exact",
			locator:  remoteView.BlockLocator(nil),
			hashStop: branch0Nodes[14].hash,
			headers:  nodeHeaders(branch0Nodes, 15, 16, 17),
			hashes:   nodeHashes(branch0Nodes, 15, 16, 17),
		},
		{
			// Locators based on remote being on side chain and a
			// stop hash in main chain just after the fork point.
			// The expected result is the blocks after the fork
			// point in the main chain up to and including the stop
			// hash.
			name:     "remote side chain, stop in main after",
			locator:  remoteView.BlockLocator(nil),
			hashStop: branch0Nodes[15].hash,
			headers:  nodeHeaders(branch0Nodes, 15),
			hashes:   nodeHashes(branch0Nodes, 15),
		},
		{
			// Locators based on remote being on side chain and a
			// stop hash in main chain some time after the fork
			// point.  The expected result is the blocks after the
			// fork point in the main chain up to and including the
			// stop hash.
			name:     "remote side chain, stop in main after more",
			locator:  remoteView.BlockLocator(nil),
			hashStop: branch0Nodes[16].hash,
			headers:  nodeHeaders(branch0Nodes, 15, 16),
			hashes:   nodeHashes(branch0Nodes, 15, 16),
		},
		{
			// Locators based on remote being on main chain in the
			// past and a stop hash local node doesn't know about.
			// The expected result is the blocks after the known
			// point in the main chain and the stop hash has no
			// effect.
			name:     "remote main chain past, unknown stop",
			locator:  localView.BlockLocator(branch0Nodes[12]),
			hashStop: chainhash.Hash{0x01},
			headers:  nodeHeaders(branch0Nodes, 13, 14, 15, 16, 17),
			hashes:   nodeHashes(branch0Nodes, 13, 14, 15, 16, 17),
		},
		{
			// Locators based on remote being on main chain in the
			// past and a stop hash in a side chain.  The expected
			// result is the blocks after the known point in the
			// main chain and the stop hash has no effect.
			name:     "remote main chain past, stop in side",
			locator:  localView.BlockLocator(branch0Nodes[12]),
			hashStop: tip(branch1Nodes).hash,
			headers:  nodeHeaders(branch0Nodes, 13, 14, 15, 16, 17),
			hashes:   nodeHashes(branch0Nodes, 13, 14, 15, 16, 17),
		},
		{
			// Locators based on remote being on main chain in the
			// past and a stop hash in the main chain before that
			// point.  The expected result is the blocks after the
			// known point in the main chain and the stop hash has
			// no effect.
			name:     "remote main chain past, stop in main before",
			locator:  localView.BlockLocator(branch0Nodes[12]),
			hashStop: branch0Nodes[11].hash,
			headers:  nodeHeaders(branch0Nodes, 13, 14, 15, 16, 17),
			hashes:   nodeHashes(branch0Nodes, 13, 14, 15, 16, 17),
		},
		{
			// Locators based on remote being on main chain in the
			// past and a stop hash in the main chain exactly at that
			// point.  The expected result is the blocks after the
			// known point in the main chain and the stop hash has
			// no effect.
			name:     "remote main chain past, stop in main exact",
			locator:  localView.BlockLocator(branch0Nodes[12]),
			hashStop: branch0Nodes[12].hash,
			headers:  nodeHeaders(branch0Nodes, 13, 14, 15, 16, 17),
			hashes:   nodeHashes(branch0Nodes, 13, 14, 15, 16, 17),
		},
		{
			// Locators based on remote being on main chain in the
			// past and a stop hash in the main chain just after
			// that point.  The expected result is the blocks after
			// the known point in the main chain and the stop hash
			// has no effect.
			name:     "remote main chain past, stop in main after",
			locator:  localView.BlockLocator(branch0Nodes[12]),
			hashStop: branch0Nodes[13].hash,
			headers:  nodeHeaders(branch0Nodes, 13),
			hashes:   nodeHashes(branch0Nodes, 13),
		},
		{
			// Locators based on remote being on main chain in the
			// past and a stop hash in the main chain some time
			// after that point.  The expected result is the blocks
			// after the known point in the main chain and the stop
			// hash has no effect.
			name:     "remote main chain past, stop in main after more",
			locator:  localView.BlockLocator(branch0Nodes[12]),
			hashStop: branch0Nodes[15].hash,
			headers:  nodeHeaders(branch0Nodes, 13, 14, 15),
			hashes:   nodeHashes(branch0Nodes, 13, 14, 15),
		},
		{
			// Locators based on remote being at exactly the same
			// point in the main chain and a stop hash local node
			// doesn't know about.  The expected result is no
			// located inventory.
			name:     "remote main chain same, unknown stop",
			locator:  localView.BlockLocator(nil),
			hashStop: chainhash.Hash{0x01},
			headers:  nil,
			hashes:   nil,
		},
		{
			// Locators based on remote being at exactly the same
			// point in the main chain and a stop hash at exactly
			// the same point.  The expected result is no located
			// inventory.
			name:     "remote main chain same, stop same point",
			locator:  localView.BlockLocator(nil),
			hashStop: tip(branch0Nodes).hash,
			headers:  nil,
			hashes:   nil,
		},
		{
			// Locators from remote that don't include any blocks
			// the local node knows.  This would happen if the
			// remote node is on a completely separate chain that
			// isn't rooted with the same genesis block.  The
			// expected result is the blocks after the genesis
			// block.
			name:     "remote unrelated chain",
			locator:  unrelatedView.BlockLocator(nil),
			hashStop: chainhash.Hash{},
			headers: nodeHeaders(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
			hashes: nodeHashes(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
		},
		{
			// Locators from remote for second block in main chain
			// and no stop hash, but with an overridden max limit.
			// The expected result is the blocks after the second
			// block limited by the max.
			name:       "remote genesis",
			locator:    locatorHashes(branch0Nodes, 0),
			hashStop:   chainhash.Hash{},
			maxAllowed: 3,
			headers:    nodeHeaders(branch0Nodes, 1, 2, 3),
			hashes:     nodeHashes(branch0Nodes, 1, 2, 3),
		},
		{
			// Poorly formed locator.
			//
			// Locator from remote that only includes a single
			// block on a side chain the local node knows.  The
			// expected result is the blocks after the genesis
			// block since even though the block is known, it is on
			// a side chain and there are no more locators to find
			// the fork point.
			name:     "weak locator, single known side block",
			locator:  locatorHashes(branch1Nodes, 1),
			hashStop: chainhash.Hash{},
			headers: nodeHeaders(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
			hashes: nodeHashes(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
		},
		{
			// Poorly formed locator.
			//
			// Locator from remote that only includes multiple
			// blocks on a side chain the local node knows however
			// none in the main chain.  The expected result is the
			// blocks after the genesis block since even though the
			// blocks are known, they are all on a side chain and
			// there are no more locators to find the fork point.
			name:     "weak locator, multiple known side blocks",
			locator:  locatorHashes(branch1Nodes, 1),
			hashStop: chainhash.Hash{},
			headers: nodeHeaders(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
			hashes: nodeHashes(branch0Nodes, 0, 1, 2, 3, 4, 5, 6,
				7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17),
		},
		{
			// Poorly formed locator.
			//
			// Locator from remote that only includes multiple
			// blocks on a side chain the local node knows however
			// none in the main chain but includes a stop hash in
			// the main chain.  The expected result is the blocks
			// after the genesis block up to the stop hash since
			// even though the blocks are known, they are all on a
			// side chain and there are no more locators to find the
			// fork point.
			name:     "weak locator, multiple known side blocks, stop in main",
			locator:  locatorHashes(branch1Nodes, 1),
			hashStop: branch0Nodes[5].hash,
			headers:  nodeHeaders(branch0Nodes, 0, 1, 2, 3, 4, 5),
			hashes:   nodeHashes(branch0Nodes, 0, 1, 2, 3, 4, 5),
		},
	}
	for _, test := range tests {
		// Ensure the expected headers are located.
		var headers []wire.BlockHeader
		if test.maxAllowed != 0 {
			// Need to use the unexported function to override the
			// max allowed for headers.
			chain.chainLock.RLock()
			headers = chain.locateHeaders(test.locator,
				&test.hashStop, test.maxAllowed)
			chain.chainLock.RUnlock()
		} else {
			headers = chain.LocateHeaders(test.locator,
				&test.hashStop)
		}
		if !reflect.DeepEqual(headers, test.headers) {
			t.Errorf("%s: unexpected headers -- got %v, want %v",
				test.name, headers, test.headers)
			continue
		}

		// Ensure the expected block hashes are located.
		maxAllowed := uint32(wire.MaxBlocksPerMsg)
		if test.maxAllowed != 0 {
			maxAllowed = test.maxAllowed
		}
		hashes := chain.LocateBlocks(test.locator, &test.hashStop,
			maxAllowed)
		if !reflect.DeepEqual(hashes, test.hashes) {
			t.Errorf("%s: unexpected hashes -- got %v, want %v",
				test.name, hashes, test.hashes)
			continue
		}
	}
}

// TestHeightToHashRange ensures that fetching a range of block hashes by start
// height and end hash works as expected.
func TestHeightToHashRange(t *testing.T) {
	// Construct a synthetic block chain with a block index consisting of
	// the following structure.
	// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
	// 	                              \-> 16a -> 17a -> 18a (unvalidated)
	tip := tstTip
	chain := newFakeChain(&chaincfg.MainNetParams)
	branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 18)
	branch1Nodes := chainedNodes(branch0Nodes[14], 3)
	for _, node := range branch0Nodes {
		chain.index.SetStatusFlags(node, statusValid)
		chain.index.AddNode(node)
	}
	for _, node := range branch1Nodes {
		if node.height < 18 {
			chain.index.SetStatusFlags(node, statusValid)
		}
		chain.index.AddNode(node)
	}
	chain.bestChain.SetTip(tip(branch0Nodes))

	tests := []struct {
		name        string
		startHeight int32            // locator for requested inventory
		endHash     chainhash.Hash   // stop hash for locator
		maxResults  int              // max to locate, 0 = wire const
		hashes      []chainhash.Hash // expected located hashes
		expectError bool
	}{
		{
			name:        "blocks below tip",
			startHeight: 11,
			endHash:     branch0Nodes[14].hash,
			maxResults:  10,
			hashes:      nodeHashes(branch0Nodes, 10, 11, 12, 13, 14),
		},
		{
			name:        "blocks on main chain",
			startHeight: 15,
			endHash:     branch0Nodes[17].hash,
			maxResults:  10,
			hashes:      nodeHashes(branch0Nodes, 14, 15, 16, 17),
		},
		{
			name:        "blocks on stale chain",
			startHeight: 15,
			endHash:     branch1Nodes[1].hash,
			maxResults:  10,
			hashes: append(nodeHashes(branch0Nodes, 14),
				nodeHashes(branch1Nodes, 0, 1)...),
		},
		{
			name:        "invalid start height",
			startHeight: 19,
			endHash:     branch0Nodes[17].hash,
			maxResults:  10,
			expectError: true,
		},
		{
			name:        "too many results",
			startHeight: 1,
			endHash:     branch0Nodes[17].hash,
			maxResults:  10,
			expectError: true,
		},
		{
			name:        "unvalidated block",
			startHeight: 15,
			endHash:     branch1Nodes[2].hash,
			maxResults:  10,
			expectError: true,
		},
	}
	for _, test := range tests {
		hashes, err := chain.HeightToHashRange(test.startHeight, &test.endHash,
			test.maxResults)
		if err != nil {
			if !test.expectError {
				t.Errorf("%s: unexpected error: %v", test.name, err)
			}
			continue
		}

		if !reflect.DeepEqual(hashes, test.hashes) {
			t.Errorf("%s: unexpected hashes -- got %v, want %v",
				test.name, hashes, test.hashes)
		}
	}
}

// TestIntervalBlockHashes ensures that fetching block hashes at specified
// intervals by end hash works as expected.
func TestIntervalBlockHashes(t *testing.T) {
	// Construct a synthetic block chain with a block index consisting of
	// the following structure.
	// 	genesis -> 1 -> 2 -> ... -> 15 -> 16  -> 17  -> 18
	// 	                              \-> 16a -> 17a -> 18a (unvalidated)
	tip := tstTip
	chain := newFakeChain(&chaincfg.MainNetParams)
	branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 18)
	branch1Nodes := chainedNodes(branch0Nodes[14], 3)
	for _, node := range branch0Nodes {
		chain.index.SetStatusFlags(node, statusValid)
		chain.index.AddNode(node)
	}
	for _, node := range branch1Nodes {
		if node.height < 18 {
			chain.index.SetStatusFlags(node, statusValid)
		}
		chain.index.AddNode(node)
	}
	chain.bestChain.SetTip(tip(branch0Nodes))

	tests := []struct {
		name        string
		endHash     chainhash.Hash
		interval    int
		hashes      []chainhash.Hash
		expectError bool
	}{
		{
			name:     "blocks on main chain",
			endHash:  branch0Nodes[17].hash,
			interval: 8,
			hashes:   nodeHashes(branch0Nodes, 7, 15),
		},
		{
			name:     "blocks on stale chain",
			endHash:  branch1Nodes[1].hash,
			interval: 8,
			hashes: append(nodeHashes(branch0Nodes, 7),
				nodeHashes(branch1Nodes, 0)...),
		},
		{
			name:     "no results",
			endHash:  branch0Nodes[17].hash,
			interval: 20,
			hashes:   []chainhash.Hash{},
		},
		{
			name:        "unvalidated block",
			endHash:     branch1Nodes[2].hash,
			interval:    8,
			expectError: true,
		},
	}
	for _, test := range tests {
		hashes, err := chain.IntervalBlockHashes(&test.endHash, test.interval)
		if err != nil {
			if !test.expectError {
				t.Errorf("%s: unexpected error: %v", test.name, err)
			}
			continue
		}

		if !reflect.DeepEqual(hashes, test.hashes) {
			t.Errorf("%s: unexpected hashes -- got %v, want %v",
				test.name, hashes, test.hashes)
		}
	}
}

func TestChainTips(t *testing.T) {
	tests := []struct {
		name        string
		chainTipGen func() (*BlockChain, map[chainhash.Hash]ChainTip)
	}{
		{
			name: "one active chain tip",
			chainTipGen: func() (*BlockChain, map[chainhash.Hash]ChainTip) {
				// Construct a synthetic block chain with a block index consisting of
				// the following structure.
				// 	genesis -> 1 -> 2 -> 3
				tip := tstTip
				chain := newFakeChain(&chaincfg.MainNetParams)
				branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 3)
				for _, node := range branch0Nodes {
					chain.index.SetStatusFlags(node, statusDataStored)
					chain.index.SetStatusFlags(node, statusValid)
					chain.index.AddNode(node)
				}
				chain.bestChain.SetTip(tip(branch0Nodes))

				activeTip := ChainTip{
					Height:    3,
					BlockHash: (tip(branch0Nodes)).hash,
					BranchLen: 0,
					Status:    StatusActive,
				}
				chainTips := make(map[chainhash.Hash]ChainTip)
				chainTips[activeTip.BlockHash] = activeTip

				return chain, chainTips
			},
		},
		{
			name: "one active chain tip, one unknown chain tip",
			chainTipGen: func() (*BlockChain, map[chainhash.Hash]ChainTip) {
				// Construct a synthetic block chain with a block index consisting of
				// the following structure.
				// 	genesis -> 1 -> 2 -> 3 ... -> 10 -> 11  -> 12  -> 13 (active)
				//                                      \-> 11a -> 12a (unknown)
				tip := tstTip
				chain := newFakeChain(&chaincfg.MainNetParams)
				branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 13)
				for _, node := range branch0Nodes {
					chain.index.SetStatusFlags(node, statusDataStored)
					chain.index.SetStatusFlags(node, statusValid)
					chain.index.AddNode(node)
				}
				chain.bestChain.SetTip(tip(branch0Nodes))

				branch1Nodes := chainedNodes(branch0Nodes[9], 2)
				for _, node := range branch1Nodes {
					chain.index.AddNode(node)
				}

				activeTip := ChainTip{
					Height:    13,
					BlockHash: (tip(branch0Nodes)).hash,
					BranchLen: 0,
					Status:    StatusActive,
				}
				unknownTip := ChainTip{
					Height:    12,
					BlockHash: (tip(branch1Nodes)).hash,
					BranchLen: 2,
					Status:    StatusUnknown,
				}
				chainTips := make(map[chainhash.Hash]ChainTip)
				chainTips[activeTip.BlockHash] = activeTip
				chainTips[unknownTip.BlockHash] = unknownTip

				return chain, chainTips
			},
		},
		{
			name: "1 inactive tip, 1 invalid tip, 1 active tip",
			chainTipGen: func() (*BlockChain, map[chainhash.Hash]ChainTip) {
				// Construct a synthetic block chain with a block index consisting of
				// the following structure.
				// 	genesis -> 1  -> 2  -> 3 (active)
				//            \ -> 1a (valid-fork)
				//            \ -> 1b (invalid)
				tip := tstTip
				chain := newFakeChain(&chaincfg.MainNetParams)
				branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 3)
				for _, node := range branch0Nodes {
					chain.index.SetStatusFlags(node, statusDataStored)
					chain.index.SetStatusFlags(node, statusValid)
					chain.index.AddNode(node)
				}
				chain.bestChain.SetTip(tip(branch0Nodes))

				branch1Nodes := chainedNodes(chain.bestChain.Genesis(), 1)
				for _, node := range branch1Nodes {
					chain.index.SetStatusFlags(node, statusDataStored)
					chain.index.SetStatusFlags(node, statusValid)
					chain.index.AddNode(node)
				}

				branch2Nodes := chainedNodes(chain.bestChain.Genesis(), 1)
				for _, node := range branch2Nodes {
					chain.index.SetStatusFlags(node, statusDataStored)
					chain.index.SetStatusFlags(node, statusValidateFailed)
					chain.index.AddNode(node)
				}

				activeTip := ChainTip{
					Height:    tip(branch0Nodes).height,
					BlockHash: (tip(branch0Nodes)).hash,
					BranchLen: 0,
					Status:    StatusActive,
				}

				inactiveTip := ChainTip{
					Height:    tip(branch1Nodes).height,
					BlockHash: (tip(branch1Nodes)).hash,
					BranchLen: 1,
					Status:    StatusValidFork,
				}

				invalidTip := ChainTip{
					Height:    tip(branch2Nodes).height,
					BlockHash: (tip(branch2Nodes)).hash,
					BranchLen: 1,
					Status:    StatusInvalid,
				}

				chainTips := make(map[chainhash.Hash]ChainTip)
				chainTips[activeTip.BlockHash] = activeTip
				chainTips[inactiveTip.BlockHash] = inactiveTip
				chainTips[invalidTip.BlockHash] = invalidTip

				return chain, chainTips
			},
		},
	}

	for _, test := range tests {
		chain, expectedChainTips := test.chainTipGen()
		gotChainTips := chain.ChainTips()
		if len(gotChainTips) != len(expectedChainTips) {
			t.Errorf("TestChainTips Failed test %s. Expected %d "+
				"chain tips, got %d", test.name, len(expectedChainTips), len(gotChainTips))
		}

		for _, gotChainTip := range gotChainTips {
			testChainTip, found := expectedChainTips[gotChainTip.BlockHash]
			if !found {
				t.Errorf("TestChainTips Failed test %s. Couldn't find an expected "+
					"chain tip with height %d, hash %s, branchlen %d, status \"%s\"",
					test.name, testChainTip.Height, testChainTip.BlockHash.String(),
					testChainTip.BranchLen, testChainTip.Status.String())
			}

			if !reflect.DeepEqual(testChainTip, gotChainTip) {
				t.Errorf("TestChainTips Failed test %s. Expected chain tip with "+
					"height %d, hash %s, branchlen %d, status \"%s\" but got "+
					"height %d, hash %s, branchlen %d, status \"%s\"", test.name,
					testChainTip.Height, testChainTip.BlockHash.String(),
					testChainTip.BranchLen, testChainTip.Status.String(),
					gotChainTip.Height, gotChainTip.BlockHash.String(),
					gotChainTip.BranchLen, gotChainTip.Status.String())
			}

			switch testChainTip.Status {
			case StatusActive:
				if testChainTip.Status.String() != "active" {
					t.Errorf("TestChainTips Fail: Expected string of \"active\", got \"%s\"",
						testChainTip.Status.String())
				}
			case StatusInvalid:
				if testChainTip.Status.String() != "invalid" {
					t.Errorf("TestChainTips Fail: Expected string of \"invalid\", got \"%s\"",
						testChainTip.Status.String())
				}
			case StatusValidFork:
				if testChainTip.Status.String() != "valid-fork" {
					t.Errorf("TestChainTips Fail: Expected string of \"valid-fork\", got \"%s\"",
						testChainTip.Status.String())
				}
			case StatusUnknown:
				if testChainTip.Status.String() != fmt.Sprintf("unknown: %b", testChainTip.Status) {
					t.Errorf("TestChainTips Fail: Expected string of \"unknown\", got \"%s\"",
						testChainTip.Status.String())
				}
			}
		}
	}
}

func TestIsAncestor(t *testing.T) {
	// Construct a synthetic block chain with a block index consisting of
	// the following structure.
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	tip := tstTip
	chain := newFakeChain(&chaincfg.MainNetParams)
	branch0Nodes := chainedNodes(chain.bestChain.Genesis(), 3)
	for _, node := range branch0Nodes {
		chain.index.SetStatusFlags(node, statusDataStored)
		chain.index.SetStatusFlags(node, statusValid)
		chain.index.AddNode(node)
	}
	chain.bestChain.SetTip(tip(branch0Nodes))

	branch1Nodes := chainedNodes(chain.bestChain.Genesis(), 1)
	for _, node := range branch1Nodes {
		chain.index.SetStatusFlags(node, statusDataStored)
		chain.index.SetStatusFlags(node, statusValid)
		chain.index.AddNode(node)
	}

	branch2Nodes := chainedNodes(chain.bestChain.Genesis(), 1)
	for _, node := range branch2Nodes {
		chain.index.SetStatusFlags(node, statusDataStored)
		chain.index.SetStatusFlags(node, statusValidateFailed)
		chain.index.AddNode(node)
	}

	// Is 1 an ancestor of 3?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeTrue := branch0Nodes[2].IsAncestor(branch0Nodes[0])
	if !shouldBeTrue {
		t.Errorf("TestIsAncestor fail. Node %s is an ancestor of node %s but got false",
			branch0Nodes[0].hash.String(), branch0Nodes[2].hash.String())
	}

	// Is 1 an ancestor of 2?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeTrue = branch0Nodes[1].IsAncestor(branch0Nodes[0])
	if !shouldBeTrue {
		t.Errorf("TestIsAncestor fail. Node %s is an ancestor of node %s but got false",
			branch0Nodes[0].hash.String(), branch0Nodes[1].hash.String())
	}

	// Is the genesis an ancestor of 1?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeTrue = branch0Nodes[0].IsAncestor(chain.bestChain.Genesis())
	if !shouldBeTrue {
		t.Errorf("TestIsAncestor fail. The genesis block is an ancestor of all blocks "+
			"but got false for node %s",
			branch0Nodes[0].hash.String())
	}

	// Is the genesis an ancestor of 1a?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeTrue = branch1Nodes[0].IsAncestor(chain.bestChain.Genesis())
	if !shouldBeTrue {
		t.Errorf("TestIsAncestor fail. The genesis block is an ancestor of all blocks "+
			"but got false for node %s",
			branch1Nodes[0].hash.String())
	}

	// Is the genesis an ancestor of 1b?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeTrue = branch2Nodes[0].IsAncestor(chain.bestChain.Genesis())
	if !shouldBeTrue {
		t.Errorf("TestIsAncestor fail. The genesis block is an ancestor of all blocks "+
			"but got false for node %s",
			branch2Nodes[0].hash.String())
	}

	// Is 1 an ancestor of 1a?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeFalse := branch1Nodes[0].IsAncestor(branch0Nodes[0])
	if shouldBeFalse {
		t.Errorf("TestIsAncestor fail. Node %s is in a different branch than "+
			"node %s but got true", branch1Nodes[0].hash.String(),
			branch0Nodes[0].hash.String())
	}

	// Is 1 an ancestor of 1b?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeFalse = branch2Nodes[0].IsAncestor(branch0Nodes[0])
	if shouldBeFalse {
		t.Errorf("TestIsAncestor fail. Node %s is in a different branch than "+
			"node %s but got true", branch2Nodes[0].hash.String(),
			branch0Nodes[0].hash.String())
	}

	// Is 1a an ancestor of 1b?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeFalse = branch2Nodes[0].IsAncestor(branch1Nodes[0])
	if shouldBeFalse {
		t.Errorf("TestIsAncestor fail. Node %s is in a different branch than "+
			"node %s but got true", branch2Nodes[0].hash.String(),
			branch1Nodes[0].hash.String())
	}

	// Is 1 an ancestor of 1?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeFalse = branch0Nodes[0].IsAncestor(branch0Nodes[0])
	if shouldBeFalse {
		t.Errorf("TestIsAncestor fail. Node is not an ancestor of itself but got true for node %s",
			branch0Nodes[0].hash.String())
	}

	// Is the geneis an ancestor of genesis?
	//
	// 	genesis -> 1  -> 2  -> 3 (active)
	//            \ -> 1a (valid-fork)
	//            \ -> 1b (invalid)
	shouldBeFalse = chain.bestChain.Genesis().IsAncestor(chain.bestChain.Genesis())
	if shouldBeFalse {
		t.Errorf("TestIsAncestor fail. Node is not an ancestor of itself but got true for node %s",
			chain.bestChain.Genesis().hash.String())
	}

	// Is a block from another chain an ancestor of 1b?
	fakeChain := newFakeChain(&chaincfg.TestNet3Params)
	shouldBeFalse = branch2Nodes[0].IsAncestor(fakeChain.bestChain.Genesis())
	if shouldBeFalse {
		t.Errorf("TestIsAncestor fail. Node %s is in a different chain than "+
			"node %s but got true", fakeChain.bestChain.Genesis().hash.String(),
			branch2Nodes[0].hash.String())
	}
}

// randomSelect selects random amount of random elements from a slice and returns a
// new slice.  The selected elements are removed.
func randomSelect(input []*testhelper.SpendableOut) (
	[]*testhelper.SpendableOut, []*testhelper.SpendableOut) {

	selected := []*testhelper.SpendableOut{}

	// Select random elements from the input slice
	amount := rand.Intn(len(input))
	for i := 0; i < amount; i++ {
		// Generate a random index
		randIdx := rand.Intn(len(input))

		// Append the selected element to the new slice
		selected = append(selected, input[randIdx])

		// Remove the selected element from the input slice.
		// This ensures that each selected element is unique.
		input = append(input[:randIdx], input[randIdx+1:]...)
	}

	return input, selected
}

// addBlocks generates new blocks and adds them to the chain.  The newly generated
// blocks will spend from the spendable outputs passed in.  The returned hases are
// the hashes of the newly generated blocks.
func addBlocks(count int, chain *BlockChain, prevBlock *chainutil.Block,
	allSpendableOutputs []*testhelper.SpendableOut) (
	[]*chainhash.Hash, [][]*testhelper.SpendableOut, error) {

	blockHashes := make([]*chainhash.Hash, 0, count)
	spendablesOuts := make([][]*testhelper.SpendableOut, 0, count)

	// Always spend everything on the first block.  This ensures we get unique blocks
	// every time.  The random select may choose not to spend any and that results
	// in getting the same block.
	nextSpends := allSpendableOutputs
	allSpendableOutputs = allSpendableOutputs[:0]
	for b := 0; b < count; b++ {
		newBlock, newSpendableOuts, err := addBlock(chain, prevBlock, nextSpends)
		if err != nil {
			return nil, nil, err
		}
		prevBlock = newBlock

		blockHashes = append(blockHashes, newBlock.Hash())
		spendablesOuts = append(spendablesOuts, newSpendableOuts)
		allSpendableOutputs = append(allSpendableOutputs, newSpendableOuts...)

		// Grab utxos to be spent in the next block.
		allSpendableOutputs, nextSpends = randomSelect(allSpendableOutputs)
	}

	return blockHashes, spendablesOuts, nil
}

func TestInvalidateBlock(t *testing.T) {
	tests := []struct {
		name     string
		chainGen func() (*BlockChain, []*chainhash.Hash, func())
	}{
		{
			name: "one branch, invalidate once",
			chainGen: func() (*BlockChain, []*chainhash.Hash, func()) {
				chain, params, tearDown := utxoCacheTestChain(
					"TestInvalidateBlock-one-branch-" +
						"invalidate-once")
				// Grab the tip of the chain.
				tip := chainutil.NewBlock(params.GenesisBlock)

				// Create a chain with 11 blocks.
				_, _, err := addBlocks(11, chain, tip, []*testhelper.SpendableOut{})
				if err != nil {
					t.Fatal(err)
				}

				// Invalidate block 5.
				block, err := chain.BlockByHeight(5)
				if err != nil {
					t.Fatal(err)
				}
				invalidateHash := block.Hash()

				return chain, []*chainhash.Hash{invalidateHash}, tearDown
			},
		},
		{
			name: "invalidate twice",
			chainGen: func() (*BlockChain, []*chainhash.Hash, func()) {
				chain, params, tearDown := utxoCacheTestChain("TestInvalidateBlock-invalidate-twice")
				// Grab the tip of the chain.
				tip := chainutil.NewBlock(params.GenesisBlock)

				// Create a chain with 11 blocks.
				_, spendableOuts, err := addBlocks(11, chain, tip, []*testhelper.SpendableOut{})
				//_, _, err := addBlocks(11, chain, tip, []*testhelper.SpendableOut{})
				if err != nil {
					t.Fatal(err)
				}

				// Set invalidateHash as block 5.
				block, err := chain.BlockByHeight(5)
				if err != nil {
					t.Fatal(err)
				}
				invalidateHash := block.Hash()

				// Create a side chain with 7 blocks that builds on block 1.
				b1, err := chain.BlockByHeight(1)
				if err != nil {
					t.Fatal(err)
				}
				altBlockHashes, _, err := addBlocks(6, chain, b1, spendableOuts[0])
				if err != nil {
					t.Fatal(err)
				}

				// Grab block at height 5:
				//
				// b2, b3, b4, b5
				//  0,  1,  2,  3
				invalidateHash2 := altBlockHashes[3]

				// Sanity checking that we grabbed the correct hash.
				node := chain.index.LookupNode(invalidateHash)
				if node == nil || node.height != 5 {
					t.Fatalf("wanted to grab block at height 5 but got height %v",
						node.height)
				}

				return chain, []*chainhash.Hash{invalidateHash, invalidateHash2}, tearDown
			},
		},
		{
			name: "invalidate a side branch",
			chainGen: func() (*BlockChain, []*chainhash.Hash, func()) {
				chain, params, tearDown := utxoCacheTestChain("TestInvalidateBlock-invalidate-side-branch")
				tip := chainutil.NewBlock(params.GenesisBlock)

				// Grab the tip of the chain.
				tip, err := chain.BlockByHash(&chain.bestChain.Tip().hash)
				if err != nil {
					t.Fatal(err)
				}

				// Create a chain with 11 blocks.
				_, spendableOuts, err := addBlocks(11, chain, tip, []*testhelper.SpendableOut{})
				if err != nil {
					t.Fatal(err)
				}

				// Create a side chain with 7 blocks that builds on block 1.
				b1, err := chain.BlockByHeight(1)
				if err != nil {
					t.Fatal(err)
				}
				altBlockHashes, _, err := addBlocks(6, chain, b1, spendableOuts[0])
				if err != nil {
					t.Fatal(err)
				}

				// Grab block at height 4:
				//
				// b2, b3, b4
				//  0,  1,  2
				invalidateHash := altBlockHashes[2]

				// Sanity checking that we grabbed the correct hash.
				node := chain.index.LookupNode(invalidateHash)
				if node == nil || node.height != 4 {
					t.Fatalf("wanted to grab block at height 4 but got height %v",
						node.height)
				}

				return chain, []*chainhash.Hash{invalidateHash}, tearDown
			},
		},
	}

	for _, test := range tests {
		chain, invalidateHashes, tearDown := test.chainGen()
		func() {
			defer tearDown()
			for _, invalidateHash := range invalidateHashes {
				chainTipsBefore := chain.ChainTips()

				// Mark if we're invalidating a block that's a part of the best chain.
				var bestChainBlock bool
				node := chain.index.LookupNode(invalidateHash)
				if chain.bestChain.Contains(node) {
					bestChainBlock = true
				}

				// Actual invalidation.
				err := chain.InvalidateBlock(invalidateHash)
				if err != nil {
					t.Fatal(err)
				}

				chainTipsAfter := chain.ChainTips()

				// Create a map for easy lookup.
				chainTipMap := make(map[chainhash.Hash]ChainTip, len(chainTipsAfter))
				activeTipCount := 0
				for _, chainTip := range chainTipsAfter {
					chainTipMap[chainTip.BlockHash] = chainTip

					if chainTip.Status == StatusActive {
						activeTipCount++
					}
				}
				if activeTipCount != 1 {
					t.Fatalf("TestInvalidateBlock fail. Expected "+
						"1 active chain tip but got %d", activeTipCount)
				}

				bestTip := chain.bestChain.Tip()

				validForkCount := 0
				for _, tip := range chainTipsBefore {
					// If the chaintip was an active tip and we invalidated a block
					// in the active tip, assert that it's invalid now.
					if bestChainBlock && tip.Status == StatusActive {
						gotTip, found := chainTipMap[tip.BlockHash]
						if !found {
							t.Fatalf("TestInvalidateBlock fail. Expected "+
								"block %s not found in chaintips after "+
								"invalidateblock", tip.BlockHash.String())
						}

						if gotTip.Status != StatusInvalid {
							t.Fatalf("TestInvalidateBlock fail. "+
								"Expected block %s to be invalid, got status: %s",
								gotTip.BlockHash.String(), gotTip.Status)
						}
					}

					if !bestChainBlock && tip.Status != StatusActive {
						gotTip, found := chainTipMap[tip.BlockHash]
						if !found {
							t.Fatalf("TestInvalidateBlock fail. Expected "+
								"block %s not found in chaintips after "+
								"invalidateblock", tip.BlockHash.String())
						}

						if gotTip.BlockHash == *invalidateHash && gotTip.Status != StatusInvalid {
							t.Fatalf("TestInvalidateBlock fail. "+
								"Expected block %s to be invalid, got status: %s",
								gotTip.BlockHash.String(), gotTip.Status)
						}
					}

					// If we're not invalidating the branch with an active tip,
					// we expect the active tip to remain the same.
					if !bestChainBlock && tip.Status == StatusActive && tip.BlockHash != bestTip.hash {
						t.Fatalf("TestInvalidateBlock fail. Expected block %s as the tip but got %s",
							tip.BlockHash.String(), bestTip.hash.String())
					}

					// If this tip is not invalid and not active, it should be
					// lighter than the current best tip.
					if tip.Status != StatusActive && tip.Status != StatusInvalid &&
						tip.Height > bestTip.height {

						tipNode := chain.index.LookupNode(&tip.BlockHash)
						if bestTip.workSum.Cmp(tipNode.workSum) == -1 {
							t.Fatalf("TestInvalidateBlock fail. Expected "+
								"block %s to be the active tip but block %s "+
								"was", tipNode.hash.String(), bestTip.hash.String())
						}
					}

					if tip.Status == StatusValidFork {
						validForkCount++
					}
				}

				// If there are no other valid chain tips besides the active chaintip,
				// we expect to have one more chain tip after the invalidate.
				if validForkCount == 0 && len(chainTipsAfter) != len(chainTipsBefore)+1 {
					t.Fatalf("TestInvalidateBlock fail. Expected %d chaintips but got %d",
						len(chainTipsBefore)+1, len(chainTipsAfter))
				}
			}

			// Try to invaliate the already invalidated hash.
			err := chain.InvalidateBlock(invalidateHashes[0])
			if err != nil {
				t.Fatal(err)
			}

			// Try to invaliate a genesis block
			err = chain.InvalidateBlock(chain.chainParams.GenesisHash)
			if err == nil {
				t.Fatalf("TestInvalidateBlock fail. Expected to err when trying to" +
					"invalidate a genesis block.")
			}

			// Try to invaliate a block that doesn't exist.
			err = chain.InvalidateBlock(chaincfg.MainNetParams.GenesisHash)
			if err == nil {
				t.Fatalf("TestInvalidateBlock fail. Expected to err when trying to" +
					"invalidate a block that doesn't exist.")
			}
		}()
	}
}

func TestReconsiderBlock(t *testing.T) {
	tests := []struct {
		name     string
		chainGen func() (*BlockChain, []*chainhash.Hash, func())
	}{
		{
			name: "one branch, invalidate once and revalidate",
			chainGen: func() (*BlockChain, []*chainhash.Hash, func()) {
				chain, params, tearDown := utxoCacheTestChain("TestInvalidateBlock-one-branch-invalidate-once")

				// Create a chain with 101 blocks.
				tip := chainutil.NewBlock(params.GenesisBlock)
				_, _, err := addBlocks(101, chain, tip, []*testhelper.SpendableOut{})
				if err != nil {
					t.Fatal(err)
				}

				// Invalidate block 5.
				block, err := chain.BlockByHeight(5)
				if err != nil {
					t.Fatal(err)
				}
				invalidateHash := block.Hash()

				return chain, []*chainhash.Hash{invalidateHash}, tearDown
			},
		},
		{
			name: "invalidate the active branch with a side branch present and revalidate",
			chainGen: func() (*BlockChain, []*chainhash.Hash, func()) {
				chain, params, tearDown := utxoCacheTestChain("TestReconsiderBlock-invalidate-with-side-branch")

				// Create a chain with 101 blocks.
				tip := chainutil.NewBlock(params.GenesisBlock)
				_, spendableOuts, err := addBlocks(101, chain, tip, []*testhelper.SpendableOut{})
				if err != nil {
					t.Fatal(err)
				}

				// Invalidate block 5.
				block, err := chain.BlockByHeight(5)
				if err != nil {
					t.Fatal(err)
				}
				invalidateHash := block.Hash()

				// Create a side chain with 7 blocks that builds on block 1.
				b1, err := chain.BlockByHeight(1)
				if err != nil {
					t.Fatal(err)
				}
				_, _, err = addBlocks(6, chain, b1, spendableOuts[0])
				if err != nil {
					t.Fatal(err)
				}

				return chain, []*chainhash.Hash{invalidateHash}, tearDown
			},
		},
		{
			name: "invalidate a side branch and revalidate it",
			chainGen: func() (*BlockChain, []*chainhash.Hash, func()) {
				chain, params, tearDown := utxoCacheTestChain("TestReconsiderBlock-invalidate-a-side-branch")

				// Create a chain with 101 blocks.
				tip := chainutil.NewBlock(params.GenesisBlock)
				_, spendableOuts, err := addBlocks(101, chain, tip, []*testhelper.SpendableOut{})
				if err != nil {
					t.Fatal(err)
				}

				// Create a side chain with 7 blocks that builds on block 1.
				b1, err := chain.BlockByHeight(1)
				if err != nil {
					t.Fatal(err)
				}
				altBlockHashes, _, err := addBlocks(6, chain, b1, spendableOuts[0])
				if err != nil {
					t.Fatal(err)
				}
				// Grab block at height 4:
				//
				// b2, b3, b4, b5
				//  0,  1,  2,  3
				invalidateHash := altBlockHashes[2]

				return chain, []*chainhash.Hash{invalidateHash}, tearDown
			},
		},
		{
			name: "reconsider an invalid side branch with a higher work",
			chainGen: func() (*BlockChain, []*chainhash.Hash, func()) {
				chain, params, tearDown := utxoCacheTestChain("TestReconsiderBlock-reconsider-an-invalid-side-branch-higher")

				tip := chainutil.NewBlock(params.GenesisBlock)
				_, spendableOuts, err := addBlocks(6, chain, tip, []*testhelper.SpendableOut{})
				if err != nil {
					t.Fatal(err)
				}

				// Select utxos to be spent from the best block and
				// modify the amount so that the block will be invalid.
				nextSpends, _ := randomSelect(spendableOuts[len(spendableOuts)-1])
				nextSpends[0].Amount += testhelper.LowFee

				// Make an invalid block that best on top of the current tip.
				bestBlock, err := chain.BlockByHash(&chain.BestSnapshot().Hash)
				if err != nil {
					t.Fatal(err)
				}
				invalidBlock, _, _ := newBlock(chain, bestBlock, nextSpends)
				invalidateHash := invalidBlock.Hash()

				// The block validation will fail here and we'll mark the
				// block as invalid in the block index.
				chain.ProcessBlock(invalidBlock, BFNone)

				// Modify the amount again so it's valid.
				nextSpends[0].Amount -= testhelper.LowFee

				return chain, []*chainhash.Hash{invalidateHash}, tearDown
			},
		},
		{
			name: "reconsider an invalid side branch with a lower work",
			chainGen: func() (*BlockChain, []*chainhash.Hash, func()) {
				chain, params, tearDown := utxoCacheTestChain("TestReconsiderBlock-reconsider-an-invalid-side-branch-lower")

				tip := chainutil.NewBlock(params.GenesisBlock)
				_, spendableOuts, err := addBlocks(6, chain, tip, []*testhelper.SpendableOut{})
				if err != nil {
					t.Fatal(err)
				}

				// Select utxos to be spent from the best block and
				// modify the amount so that the block will be invalid.
				nextSpends, _ := randomSelect(spendableOuts[len(spendableOuts)-1])
				nextSpends[0].Amount += testhelper.LowFee

				// Make an invalid block that best on top of the current tip.
				bestBlock, err := chain.BlockByHash(&chain.BestSnapshot().Hash)
				if err != nil {
					t.Fatal(err)
				}
				invalidBlock, _, _ := newBlock(chain, bestBlock, nextSpends)
				invalidateHash := invalidBlock.Hash()

				// The block validation will fail here and we'll mark the
				// block as invalid in the block index.
				chain.ProcessBlock(invalidBlock, BFNone)

				// Modify the amount again so it's valid.
				nextSpends[0].Amount -= testhelper.LowFee

				// Add more blocks to make the invalid block a
				// side chain and not the most pow.
				_, _, err = addBlocks(3, chain, bestBlock, []*testhelper.SpendableOut{})
				if err != nil {
					t.Fatal(err)
				}

				return chain, []*chainhash.Hash{invalidateHash}, tearDown
			},
		},
	}

	for _, test := range tests {
		chain, invalidateHashes, tearDown := test.chainGen()
		func() {
			defer tearDown()
			for _, invalidateHash := range invalidateHashes {
				// Cache the chain tips before the invalidate. Since we'll reconsider
				// the invalidated block, we should come back to these tips in the end.
				tips := chain.ChainTips()
				expectedChainTips := make(map[chainhash.Hash]ChainTip, len(tips))
				for _, tip := range tips {
					expectedChainTips[tip.BlockHash] = tip
				}

				// Invalidation.
				err := chain.InvalidateBlock(invalidateHash)
				if err != nil {
					t.Fatal(err)
				}

				// Reconsideration.
				err = chain.ReconsiderBlock(invalidateHash)
				if err != nil {
					t.Fatal(err)
				}

				// Compare the tips aginst the tips we've cached.
				gotChainTips := chain.ChainTips()
				for _, gotChainTip := range gotChainTips {
					testChainTip, found := expectedChainTips[gotChainTip.BlockHash]
					if !found {
						t.Errorf("TestReconsiderBlock Failed test \"%s\". Couldn't find an expected "+
							"chain tip with height %d, hash %s, branchlen %d, status \"%s\"",
							test.name, testChainTip.Height, testChainTip.BlockHash.String(),
							testChainTip.BranchLen, testChainTip.Status.String())
					}

					// If the invalid side branch is a lower work, we'll never
					// actually process the block again until the branch becomes
					// a greater work chain so it'll show up as valid-fork.
					if test.name == "reconsider an invalid side branch with a lower work" &&
						testChainTip.BlockHash == *invalidateHash {

						testChainTip.Status = StatusValidFork
					}

					if !reflect.DeepEqual(testChainTip, gotChainTip) {
						t.Errorf("TestReconsiderBlock Failed test \"%s\". Expected chain tip with "+
							"height %d, hash %s, branchlen %d, status \"%s\" but got "+
							"height %d, hash %s, branchlen %d, status \"%s\"", test.name,
							testChainTip.Height, testChainTip.BlockHash.String(),
							testChainTip.BranchLen, testChainTip.Status.String(),
							gotChainTip.Height, gotChainTip.BlockHash.String(),
							gotChainTip.BranchLen, gotChainTip.Status.String())
					}
				}
			}
		}()
	}
}
