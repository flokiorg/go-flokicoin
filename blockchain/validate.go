// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/txscript"
	"github.com/flokiorg/go-flokicoin/wire"
)

const (
	// MaxTimeOffsetSeconds is the maximum number of seconds a block time
	// is allowed to be ahead of the current time.  This is currently 2
	// hours.
	MaxTimeOffsetSeconds = 2 * 60 * 60

	// MinCoinbaseScriptLen is the minimum length a coinbase script can be.
	MinCoinbaseScriptLen = 2

	// MaxCoinbaseScriptLen is the maximum length a coinbase script can be.
	MaxCoinbaseScriptLen = 100

	// medianTimeBlocks is the number of previous blocks which should be
	// used to calculate the median time used to validate block timestamps.
	medianTimeBlocks = 11

	// serializedHeightVersion is the block version which changed block
	// coinbases to start with the serialized block height.
	serializedHeightVersion = 2

	// baseSubsidy is the starting subsidy amount for mined blocks.  This
	// value is halved every SubsidyHalvingInterval blocks.
	baseSubsidy int64 = 1000 * chainutil.LokiPerFlokicoin // #FLOKI_CHANGE

	stableSubsidy int64 = 21 * chainutil.LokiPerFlokicoin // Minimum subsidy

	// coinbaseHeightAllocSize is the amount of bytes that the
	// ScriptBuilder will allocate when validating the coinbase height.
	coinbaseHeightAllocSize = 5

	// maxTimeWarp is a maximum number of seconds that the timestamp of the first
	// block of a difficulty adjustment period is allowed to
	// be earlier than the last block of the previous period (BIP94).
	maxTimeWarp = 600 * time.Second
)

var (
	// zeroHash is the zero value for a chainhash.Hash and is defined as
	// a package level variable to avoid the need to create a new instance
	// every time a check is needed.
	zeroHash chainhash.Hash

	// block91842Hash is one of the two nodes which violate the rules
	// set forth in BIP0030.  It is defined as a package level variable to
	// avoid the need to create a new instance every time a check is needed.
	block91842Hash = newHashFromStr("00000000000a4d0a398161ffc163c503763b1f4360639393e0e4c8e300e0caec")

	// block91880Hash is one of the two nodes which violate the rules
	// set forth in BIP0030.  It is defined as a package level variable to
	// avoid the need to create a new instance every time a check is needed.
	block91880Hash = newHashFromStr("00000000000743f190a18c5577a3c2d2a1f610ae9601ac046a38084ccb7cd721")
)

// isNullOutpoint determines whether or not a previous transaction output point
// is set.
func isNullOutpoint(outpoint *wire.OutPoint) bool {
	if outpoint.Index == math.MaxUint32 && outpoint.Hash == zeroHash {
		return true
	}
	return false
}

// ShouldHaveSerializedBlockHeight determines if a block should have a
// serialized block height embedded within the scriptSig of its
// coinbase transaction. Judgement is based on the block version in the block
// header. Blocks with version 2 and above satisfy this criteria. See BIP0034
// for further information.
func ShouldHaveSerializedBlockHeight(header *wire.BlockHeader) bool {
	return header.Version >= serializedHeightVersion
}

// IsCoinBaseTx determines whether or not a transaction is a coinbase.  A coinbase
// is a special transaction created by miners that has no inputs.  This is
// represented in the block chain by a transaction with a single input that has
// a previous output transaction index set to the maximum value along with a
// zero hash.
//
// This function only differs from IsCoinBase in that it works with a raw wire
// transaction as opposed to a higher level util transaction.
func IsCoinBaseTx(msgTx *wire.MsgTx) bool {
	// A coin base must only have one transaction input.
	if len(msgTx.TxIn) != 1 {
		return false
	}

	// The previous output of a coin base must have a max value index and
	// a zero hash.
	prevOut := &msgTx.TxIn[0].PreviousOutPoint
	if prevOut.Index != math.MaxUint32 || prevOut.Hash != zeroHash {
		return false
	}

	return true
}

// IsCoinBase determines whether or not a transaction is a coinbase.  A coinbase
// is a special transaction created by miners that has no inputs.  This is
// represented in the block chain by a transaction with a single input that has
// a previous output transaction index set to the maximum value along with a
// zero hash.
//
// This function only differs from IsCoinBaseTx in that it works with a higher
// level util transaction as opposed to a raw wire transaction.
func IsCoinBase(tx *chainutil.Tx) bool {
	return IsCoinBaseTx(tx.MsgTx())
}

// SequenceLockActive determines if a transaction's sequence locks have been
// met, meaning that all the inputs of a given transaction have reached a
// height or time sufficient for their relative lock-time maturity.
func SequenceLockActive(sequenceLock *SequenceLock, blockHeight int32,
	medianTimePast time.Time) bool {

	// If either the seconds, or height relative-lock time has not yet
	// reached, then the transaction is not yet mature according to its
	// sequence locks.
	if sequenceLock.Seconds >= medianTimePast.Unix() ||
		sequenceLock.BlockHeight >= blockHeight {
		return false
	}

	return true
}

// IsFinalizedTransaction determines whether or not a transaction is finalized.
func IsFinalizedTransaction(tx *chainutil.Tx, blockHeight int32, blockTime time.Time) bool {
	msgTx := tx.MsgTx()

	// Lock time of zero means the transaction is finalized.
	lockTime := msgTx.LockTime
	if lockTime == 0 {
		return true
	}

	// The lock time field of a transaction is either a block height at
	// which the transaction is finalized or a timestamp depending on if the
	// value is before the txscript.LockTimeThreshold.  When it is under the
	// threshold it is a block height.
	blockTimeOrHeight := int64(0)
	if lockTime < txscript.LockTimeThreshold {
		blockTimeOrHeight = int64(blockHeight)
	} else {
		blockTimeOrHeight = blockTime.Unix()
	}
	if int64(lockTime) < blockTimeOrHeight {
		return true
	}

	// At this point, the transaction's lock time hasn't occurred yet, but
	// the transaction might still be finalized if the sequence number
	// for all transaction inputs is maxed out.
	for _, txIn := range msgTx.TxIn {
		if txIn.Sequence != math.MaxUint32 {
			return false
		}
	}
	return true
}

// isBIP0030Node returns whether or not the passed node represents one of the
// two blocks that violate the BIP0030 rule which prevents transactions from
// overwriting old ones.
func isBIP0030Node(node *blockNode) bool {
	if node.height == 91842 && node.hash.IsEqual(block91842Hash) {
		return true
	}

	if node.height == 91880 && node.hash.IsEqual(block91880Hash) {
		return true
	}

	return false
}

// CalcBlockSubsidy returns the subsidy amount a block at the provided height
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
//
// The subsidy is halved every SubsidyReductionInterval blocks.  Mathematically
// this is: baseSubsidy / 2^(height/SubsidyReductionInterval)
//
// At the target block generation rate for the main network, this is
// approximately every 6 months.
func CalcBlockSubsidy(height int32, chainParams *chaincfg.Params) int64 {
	if chainParams.SubsidyReductionInterval == 0 {
		return baseSubsidy
	}

	// Equivalent to: baseSubsidy / 2^(height/subsidyHalvingInterval)
	subsidy := baseSubsidy >> uint(height/chainParams.SubsidyReductionInterval)

	if subsidy < stableSubsidy { // constant inflation
		return stableSubsidy
	}

	return subsidy
}

// CheckTransactionSanity performs some preliminary checks on a transaction to
// ensure it is sane.  These checks are context free.
func CheckTransactionSanity(tx *chainutil.Tx) error {
	// A transaction must have at least one input.
	msgTx := tx.MsgTx()
	if len(msgTx.TxIn) == 0 {
		return ruleError(ErrNoTxInputs, "transaction has no inputs")
	}

	// A transaction must have at least one output.
	if len(msgTx.TxOut) == 0 {
		return ruleError(ErrNoTxOutputs, "transaction has no outputs")
	}

	// A transaction must not exceed the maximum allowed block payload when
	// serialized.
	serializedTxSize := tx.MsgTx().SerializeSizeStripped()
	if serializedTxSize > MaxBlockBaseSize {
		str := fmt.Sprintf("serialized transaction is too big - got "+
			"%d, max %d", serializedTxSize, MaxBlockBaseSize)
		return ruleError(ErrTxTooBig, str)
	}

	// Ensure the transaction amounts are in range.  Each transaction
	// output must not be negative or more than the max allowed per
	// transaction.  Also, the total of all outputs must abide by the same
	// restrictions.  All amounts in a transaction are in a unit value known
	// as a loki.  One flokicoin is a quantity of loki as defined by the
	// LokiPerFlokicoin constant.
	var totalLoki int64
	for _, txOut := range msgTx.TxOut {
		loki := txOut.Value
		if loki < 0 {
			str := fmt.Sprintf("transaction output has negative "+
				"value of %v", loki)
			return ruleError(ErrBadTxOutValue, str)
		}
		if loki > chainutil.MaxLoki {
			str := fmt.Sprintf("transaction output value is "+
				"higher than max allowed value: %v > %v ",
				loki, chainutil.MaxLoki)
			return ruleError(ErrBadTxOutValue, str)
		}

		// Two's complement int64 overflow guarantees that any overflow
		// is detected and reported.  This is impossible for Flokicoin, but
		// perhaps possible if an alt increases the total money supply.
		totalLoki += loki
		if totalLoki < 0 {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs exceeds max allowed value of %v",
				chainutil.MaxLoki)
			return ruleError(ErrBadTxOutValue, str)
		}
		if totalLoki > chainutil.MaxLoki {
			str := fmt.Sprintf("total value of all transaction "+
				"outputs is %v which is higher than max "+
				"allowed value of %v", totalLoki,
				chainutil.MaxLoki)
			return ruleError(ErrBadTxOutValue, str)
		}
	}

	// Check for duplicate transaction inputs.
	existingTxOut := make(map[wire.OutPoint]struct{})
	for _, txIn := range msgTx.TxIn {
		if _, exists := existingTxOut[txIn.PreviousOutPoint]; exists {
			return ruleError(ErrDuplicateTxInputs, "transaction "+
				"contains duplicate inputs")
		}
		existingTxOut[txIn.PreviousOutPoint] = struct{}{}
	}

	// Coinbase script length must be between min and max length.
	if IsCoinBase(tx) {
		slen := len(msgTx.TxIn[0].SignatureScript)
		if slen < MinCoinbaseScriptLen || slen > MaxCoinbaseScriptLen {
			str := fmt.Sprintf("coinbase transaction script length "+
				"of %d is out of range (min: %d, max: %d)",
				slen, MinCoinbaseScriptLen, MaxCoinbaseScriptLen)
			return ruleError(ErrBadCoinbaseScriptLen, str)
		}
	} else {
		// Previous transaction outputs referenced by the inputs to this
		// transaction must not be null.
		for _, txIn := range msgTx.TxIn {
			if isNullOutpoint(&txIn.PreviousOutPoint) {
				return ruleError(ErrBadTxInput, "transaction "+
					"input refers to previous output that "+
					"is null")
			}
		}
	}

	return nil
}

// checkProofOfWork ensures the block header bits which indicate the target
// difficulty is in min/max range and that the block hash is less than the
// target difficulty as claimed.
//
// The flags modify the behavior of this function as follows:
//   - BFNoPoWCheck: The check to ensure the block hash is less than the target
//     difficulty is not performed.
func checkProofOfWork(header *wire.BlockHeader, powLimit *big.Int, flags BehaviorFlags) error {

	// The target difficulty must be larger than zero.
	target := CompactToBig(header.Bits)
	if target.Sign() <= 0 {
		str := fmt.Sprintf("block target difficulty of %064x is too low",
			target)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The target difficulty must be less than the maximum allowed.
	if target.Cmp(powLimit) > 0 {
		str := fmt.Sprintf("block target difficulty of %064x is "+
			"higher than max of %064x", target, powLimit)
		return ruleError(ErrUnexpectedDifficulty, str)
	}

	// The block hash must be less than the claimed target unless the flag
	// to avoid proof of work checks is set.
	if flags&BFNoPoWCheck == BFNoPoWCheck {
		return nil
	}
	// The block hash must be less than the claimed target.
	// hash := header.BlockHash() #FLOKI_CHANGE
	if header.AuxPowHeader == nil {
		if header.AuxPow() {
			return ruleError(ErrAuxpowNoVersion, "no auxpow on block with auxpow version")
		}
		hash := header.BlockPoWHash()
		hashNum := HashToBig(&hash)
		if hashNum.Cmp(target) > 0 {
			str := fmt.Sprintf("block hash of %064x is higher than "+
				"expected max of %064x", hashNum, target)
			return ruleError(ErrHighHash, str)
		}
		return nil
	}

	// auxpow case
	if !header.AuxPow() {
		return ruleError(ErrAuxpowNoVersion, "auxpow on block with non-auxpow version")
	}

	{
		hash := header.AuxPowHeader.ParentBlockHeader.BlockPoWHash()
		hashNum := HashToBig(&hash)
		if hashNum.Cmp(target) > 0 {
			str := fmt.Sprintf("auxpow failed, parent block hash of %064x is higher than "+
				"expected max of %064x", hashNum, target)
			return ruleError(ErrHighHash, str)
		}
	}

	if err := header.AuxPowHeader.Check(header.BlockHash(), header.GetChainID()); err != nil {
		return err
	}

	return nil
}

// CheckProofOfWork ensures the block header bits which indicate the target
// difficulty is in min/max range and that the block hash is less than the
// target difficulty as claimed.
func CheckProofOfWork(block *chainutil.Block, powLimit *big.Int) error {
	return checkProofOfWork(&block.MsgBlock().Header, powLimit, BFNone)
}

// CountSigOps returns the number of signature operations for all transaction
// input and output scripts in the provided transaction.  This uses the
// quicker, but imprecise, signature operation counting mechanism from
// txscript.
func CountSigOps(tx *chainutil.Tx) int {
	msgTx := tx.MsgTx()

	// Accumulate the number of signature operations in all transaction
	// inputs.
	totalSigOps := 0
	for _, txIn := range msgTx.TxIn {
		numSigOps := txscript.GetSigOpCount(txIn.SignatureScript)
		totalSigOps += numSigOps
	}

	// Accumulate the number of signature operations in all transaction
	// outputs.
	for _, txOut := range msgTx.TxOut {
		numSigOps := txscript.GetSigOpCount(txOut.PkScript)
		totalSigOps += numSigOps
	}

	return totalSigOps
}

// CountP2SHSigOps returns the number of signature operations for all input
// transactions which are of the pay-to-script-hash type.  This uses the
// precise, signature operation counting mechanism from the script engine which
// requires access to the input transaction scripts.
func CountP2SHSigOps(tx *chainutil.Tx, isCoinBaseTx bool, utxoView *UtxoViewpoint) (int, error) {
	// Coinbase transactions have no interesting inputs.
	if isCoinBaseTx {
		return 0, nil
	}

	// Accumulate the number of signature operations in all transaction
	// inputs.
	msgTx := tx.MsgTx()
	totalSigOps := 0
	for txInIndex, txIn := range msgTx.TxIn {
		// Ensure the referenced input transaction is available.
		utxo := utxoView.LookupEntry(txIn.PreviousOutPoint)
		if utxo == nil || utxo.IsSpent() {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOutPoint,
				tx.Hash(), txInIndex)
			return 0, ruleError(ErrMissingTxOut, str)
		}

		// We're only interested in pay-to-script-hash types, so skip
		// this input if it's not one.
		pkScript := utxo.PkScript()
		if !txscript.IsPayToScriptHash(pkScript) {
			continue
		}

		// Count the precise number of signature operations in the
		// referenced public key script.
		sigScript := txIn.SignatureScript
		numSigOps := txscript.GetPreciseSigOpCount(sigScript, pkScript,
			true)

		// We could potentially overflow the accumulator so check for
		// overflow.
		lastSigOps := totalSigOps
		totalSigOps += numSigOps
		if totalSigOps < lastSigOps {
			str := fmt.Sprintf("the public key script from output "+
				"%v contains too many signature operations - "+
				"overflow", txIn.PreviousOutPoint)
			return 0, ruleError(ErrTooManySigOps, str)
		}
	}

	return totalSigOps, nil
}

// CheckBlockHeaderSanity performs some preliminary checks on a block header to
// ensure it is sane before continuing with processing.  These checks are
// context free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkProofOfWork.
func CheckBlockHeaderSanity(header *wire.BlockHeader, powLimit *big.Int,
	timeSource MedianTimeSource, flags BehaviorFlags) error {

	// Ensure the proof of work bits in the block header is in min/max range
	// and the block hash is less than the target value described by the
	// bits.
	err := checkProofOfWork(header, powLimit, flags)
	if err != nil {
		return err
	}

	// A block timestamp must not have a greater precision than one second.
	// This check is necessary because Go time.Time values support
	// nanosecond precision whereas the consensus rules only apply to
	// seconds and it's much nicer to deal with standard Go time values
	// instead of converting to seconds everywhere.
	if !header.Timestamp.Equal(time.Unix(header.Timestamp.Unix(), 0)) {
		str := fmt.Sprintf("block timestamp of %v has a higher "+
			"precision than one second", header.Timestamp)
		return ruleError(ErrInvalidTime, str)
	}

	// Ensure the block time is not too far in the future.
	maxTimestamp := timeSource.AdjustedTime().Add(time.Second *
		MaxTimeOffsetSeconds)
	if header.Timestamp.After(maxTimestamp) {
		str := fmt.Sprintf("block timestamp of %v is too far in the "+
			"future", header.Timestamp)
		return ruleError(ErrTimeTooNew, str)
	}

	return nil
}

// checkBlockSanity performs some preliminary checks on a block to ensure it is
// sane before continuing with block processing.  These checks are context free.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to checkBlockHeaderSanity.
func checkBlockSanity(block *chainutil.Block, powLimit *big.Int, timeSource MedianTimeSource, flags BehaviorFlags) error {
	msgBlock := block.MsgBlock()
	header := &msgBlock.Header
	err := CheckBlockHeaderSanity(header, powLimit, timeSource, flags)
	if err != nil {
		return err
	}

	// A block must have at least one transaction.
	numTx := len(msgBlock.Transactions)
	if numTx == 0 {
		return ruleError(ErrNoTransactions, "block does not contain "+
			"any transactions")
	}

	// A block must not have more transactions than the max block payload or
	// else it is certainly over the weight limit.
	if numTx > MaxBlockBaseSize {
		str := fmt.Sprintf("block contains too many transactions - "+
			"got %d, max %d", numTx, MaxBlockBaseSize)
		return ruleError(ErrBlockTooBig, str)
	}

	// A block must not exceed the maximum allowed block payload when
	// serialized.
	serializedSize := msgBlock.SerializeSizeStripped()
	if serializedSize > MaxBlockBaseSize {
		str := fmt.Sprintf("serialized block is too big - got %d, "+
			"max %d", serializedSize, MaxBlockBaseSize)
		return ruleError(ErrBlockTooBig, str)
	}

	// The first transaction in a block must be a coinbase.
	transactions := block.Transactions()
	if !IsCoinBase(transactions[0]) {
		return ruleError(ErrFirstTxNotCoinbase, "first transaction in "+
			"block is not a coinbase")
	}

	// A block must not have more than one coinbase.
	for i, tx := range transactions[1:] {
		if IsCoinBase(tx) {
			str := fmt.Sprintf("block contains second coinbase at "+
				"index %d", i+1)
			return ruleError(ErrMultipleCoinbases, str)
		}
	}

	// Do some preliminary checks on each transaction to ensure they are
	// sane before continuing.
	for _, tx := range transactions {
		err := CheckTransactionSanity(tx)
		if err != nil {
			return err
		}
	}

	// Build merkle tree and ensure the calculated merkle root matches the
	// entry in the block header.  This also has the effect of caching all
	// of the transaction hashes in the block to speed up future hash
	// checks.  Flokicoind builds the tree here and checks the merkle root
	// after the following checks, but there is no reason not to check the
	// merkle root matches here.
	calcMerkleRoot := CalcMerkleRoot(block.Transactions(), false)
	if !header.MerkleRoot.IsEqual(&calcMerkleRoot) {
		str := fmt.Sprintf("block merkle root is invalid - block "+
			"header indicates %v, but calculated value is %v",
			header.MerkleRoot, calcMerkleRoot)
		return ruleError(ErrBadMerkleRoot, str)
	}

	// Check for duplicate transactions.  This check will be fairly quick
	// since the transaction hashes are already cached due to building the
	// merkle tree above.
	existingTxHashes := make(map[chainhash.Hash]struct{})
	for _, tx := range transactions {
		hash := tx.Hash()
		if _, exists := existingTxHashes[*hash]; exists {
			str := fmt.Sprintf("block contains duplicate "+
				"transaction %v", hash)
			return ruleError(ErrDuplicateTx, str)
		}
		existingTxHashes[*hash] = struct{}{}
	}

	// The number of signature operations must be less than the maximum
	// allowed per block.
	totalSigOps := 0
	for _, tx := range transactions {
		// We could potentially overflow the accumulator so check for
		// overflow.
		lastSigOps := totalSigOps
		totalSigOps += (CountSigOps(tx) * WitnessScaleFactor)
		if totalSigOps < lastSigOps || totalSigOps > MaxBlockSigOpsCost {
			str := fmt.Sprintf("block contains too many signature "+
				"operations - got %v, max %v", totalSigOps,
				MaxBlockSigOpsCost)
			return ruleError(ErrTooManySigOps, str)
		}
	}

	return nil
}

// CheckBlockSanity performs some preliminary checks on a block to ensure it is
// sane before continuing with block processing.  These checks are context free.
func CheckBlockSanity(block *chainutil.Block, powLimit *big.Int, timeSource MedianTimeSource) error {
	return checkBlockSanity(block, powLimit, timeSource, BFNone)
}

// ExtractCoinbaseHeight attempts to extract the height of the block from the
// scriptSig of a coinbase transaction.  Coinbase heights are only present in
// blocks of version 2 or later.  This was added as part of BIP0034.
func ExtractCoinbaseHeight(coinbaseTx *chainutil.Tx) (int32, error) {
	sigScript := coinbaseTx.MsgTx().TxIn[0].SignatureScript
	if len(sigScript) < 1 {
		str := "the coinbase signature script for blocks of " +
			"version %d or greater must start with the " +
			"length of the serialized block height"
		str = fmt.Sprintf(str, serializedHeightVersion)
		return 0, ruleError(ErrMissingCoinbaseHeight, str)
	}

	// Detect the case when the block height is a small integer encoded with
	// as single byte.
	opcode := int(sigScript[0])
	if opcode == txscript.OP_0 {
		return 0, nil
	}
	if opcode >= txscript.OP_1 && opcode <= txscript.OP_16 {
		return int32(opcode - (txscript.OP_1 - 1)), nil
	}

	// Otherwise, the opcode is the length of the following bytes which
	// encode in the block height.
	serializedLen := int(sigScript[0])
	if len(sigScript[1:]) < serializedLen {
		str := "the coinbase signature script for blocks of " +
			"version %d or greater must start with the " +
			"serialized block height"
		str = fmt.Sprintf(str, serializedLen)
		return 0, ruleError(ErrMissingCoinbaseHeight, str)
	}

	// We use 4 bytes here since it saves us allocations. We use a stack
	// allocation rather than a heap allocation here.
	var serializedHeightBytes [4]byte
	copy(serializedHeightBytes[:], sigScript[1:serializedLen+1])

	serializedHeight := int32(
		binary.LittleEndian.Uint32(serializedHeightBytes[:]),
	)

	if err := compareScript(serializedHeight, sigScript); err != nil {
		return 0, err
	}

	return serializedHeight, nil
}

// CheckSerializedHeight checks if the signature script in the passed
// transaction starts with the serialized block height of wantHeight.
func CheckSerializedHeight(coinbaseTx *chainutil.Tx, wantHeight int32) error {
	serializedHeight, err := ExtractCoinbaseHeight(coinbaseTx)
	if err != nil {
		return err
	}

	if serializedHeight != wantHeight {
		str := fmt.Sprintf("the coinbase signature script serialized "+
			"block height is %d when %d was expected",
			serializedHeight, wantHeight)
		return ruleError(ErrBadCoinbaseHeight, str)
	}
	return nil
}

func compareScript(height int32, script []byte) error {
	scriptBuilder := txscript.NewScriptBuilder(
		txscript.WithScriptAllocSize(coinbaseHeightAllocSize),
	)
	scriptHeight, err := scriptBuilder.AddInt64(
		int64(height),
	).Script()
	if err != nil {
		return err
	}

	if !bytes.HasPrefix(script, scriptHeight) {
		str := fmt.Sprintf("the coinbase signature script does not "+
			"minimally encode the height %d", height)
		return ruleError(ErrBadCoinbaseHeight, str)
	}

	return nil
}

// CheckBlockHeaderContext performs several validation checks on the block header
// which depend on its position within the block chain.
//
// The flags modify the behavior of this function as follows:
//   - BFFastAdd: All checks except those involving comparing the header against
//     the checkpoints are not performed.
//
// The skipCheckpoint boolean is used so that libraries can skip the checkpoint
// sanity checks.
//
// This function MUST be called with the chain state lock held (for writes).
// NOTE: Ignore the above lock requirement if this function is not passed a
// *Blockchain instance as the ChainCtx argument.
func CheckBlockHeaderContext(header *wire.BlockHeader, prevNode HeaderCtx,
	flags BehaviorFlags, c ChainCtx, skipCheckpoint bool) error {

	// The height of this block is one more than the referenced previous
	// block.
	blockHeight := prevNode.Height() + 1
	params := c.ChainParams()

	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		// Ensure the difficulty specified in the block header matches
		// the calculated difficulty based on the previous block and
		// difficulty retarget rules.
		expectedDifficulty, err := calcNextRequiredDifficulty(
			prevNode, header.Timestamp, c,
		)
		if err != nil {
			return err
		}
		blockDifficulty := header.Bits
		if blockDifficulty != expectedDifficulty {
			str := "block difficulty of %d is not the expected value of %d"
			str = fmt.Sprintf(str, blockDifficulty, expectedDifficulty)
			return ruleError(ErrUnexpectedDifficulty, str)
		}

		// Ensure the timestamp for the block header is after the
		// median time of the last several blocks (medianTimeBlocks).
		medianTime := CalcPastMedianTime(prevNode)
		if !header.Timestamp.After(medianTime) {
			str := "block timestamp of %v is not after expected %v"
			str = fmt.Sprintf(str, header.Timestamp, medianTime)
			return ruleError(ErrTimeTooOld, str)
		}

		// Testnet4 only: Check timestamp against prev for
		// difficulty-adjustment blocks to prevent timewarp attacks.
		if params.EnforceBIP94 {
			err := assertNoTimeWarp(
				blockHeight, c.BlocksPerRetarget(),
				header.Timestamp,
				time.Unix(prevNode.Timestamp(), 0),
			)
			if err != nil {
				return err
			}
		}
	}

	// Reject outdated block versions once a majority of the network
	// has upgraded.  These were originally voted on by BIP0034,
	// BIP0065, and BIP0066.
	if header.Version < 2 && blockHeight >= params.BIP0034Height ||
		header.Version < 3 && blockHeight >= params.BIP0066Height ||
		header.Version < 4 && blockHeight >= params.BIP0065Height {

		str := "new blocks with version %d are no longer valid"
		str = fmt.Sprintf(str, header.Version)
		return ruleError(ErrBlockVersionTooOld, str)
	}

	if skipCheckpoint {
		// If the caller wants us to skip the checkpoint checks, we'll
		// return early.
		return nil
	}

	// Ensure chain matches up to predetermined checkpoints.
	blockHash := header.BlockHash()
	if !c.VerifyCheckpoint(blockHeight, &blockHash) {
		str := fmt.Sprintf("block at height %d does not match "+
			"checkpoint hash", blockHeight)
		return ruleError(ErrBadCheckpoint, str)
	}

	// Find the previous checkpoint and prevent blocks which fork the main
	// chain before it.  This prevents storage of new, otherwise valid,
	// blocks which build off of old blocks that are likely at a much easier
	// difficulty and therefore could be used to waste cache and disk space.
	checkpointNode, err := c.FindPreviousCheckpoint()
	if err != nil {
		return err
	}
	if checkpointNode != nil && blockHeight < checkpointNode.Height() {
		str := fmt.Sprintf("block at height %d forks the main chain "+
			"before the previous checkpoint at height %d",
			blockHeight, checkpointNode.Height())
		return ruleError(ErrForkTooOld, str)
	}

	return nil
}

// assertNoTimeWarp checks the timestamp of the block against the previous
// block's timestamp for the first block of each difficulty adjustment interval
// to prevent timewarp attacks. This is defined in BIP-0094.
func assertNoTimeWarp(blockHeight, blocksPerReTarget int32, headerTimestamp,
	prevBlockTimestamp time.Time) error {

	// If this isn't the first block of the difficulty adjustment interval,
	// then we can exit early.
	if blockHeight%blocksPerReTarget != 0 {
		return nil
	}

	// Check timestamp for the first block of each difficulty adjustment
	// interval, except the genesis block.
	if headerTimestamp.Before(prevBlockTimestamp.Add(-maxTimeWarp)) {
		str := "block's timestamp %v is too early on diff adjustment " +
			"block %v"
		str = fmt.Sprintf(str, headerTimestamp, prevBlockTimestamp)
		return ruleError(ErrTimewarpAttack, str)
	}

	return nil
}

// checkBlockContext performs several validation checks on the block which depend
// on its position within the block chain.
//
// The flags modify the behavior of this function as follows:
//   - BFFastAdd: The transaction are not checked to see if they are finalized
//     and the somewhat expensive BIP0034 validation is not performed.
//
// The flags are also passed to checkBlockHeaderContext.  See its documentation
// for how the flags modify its behavior.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) checkBlockContext(block *chainutil.Block, prevNode *blockNode, flags BehaviorFlags) error {
	// Perform all block header related validation checks.
	header := &block.MsgBlock().Header
	err := CheckBlockHeaderContext(header, prevNode, flags, b, false)
	if err != nil {
		return err
	}

	fastAdd := flags&BFFastAdd == BFFastAdd
	if !fastAdd {
		// Obtain the latest state of the deployed CSV soft-fork in
		// order to properly guard the new validation behavior based on
		// the current BIP 9 version bits state.
		csvState, err := b.deploymentState(prevNode, chaincfg.DeploymentCSV)
		if err != nil {
			return err
		}

		// Once the CSV soft-fork is fully active, we'll switch to
		// using the current median time past of the past block's
		// timestamps for all lock-time based checks.
		blockTime := header.Timestamp
		if csvState == ThresholdActive {
			blockTime = CalcPastMedianTime(prevNode)
		}

		// The height of this block is one more than the referenced
		// previous block.
		blockHeight := prevNode.height + 1

		// Ensure all transactions in the block are finalized.
		for _, tx := range block.Transactions() {
			if !IsFinalizedTransaction(tx, blockHeight,
				blockTime) {

				str := fmt.Sprintf("block contains unfinalized "+
					"transaction %v", tx.Hash())
				return ruleError(ErrUnfinalizedTx, str)
			}
		}

		// Ensure coinbase starts with serialized block heights for
		// blocks whose version is the serializedHeightVersion or newer
		// once a majority of the network has upgraded.  This is part of
		// BIP0034.
		if ShouldHaveSerializedBlockHeight(header) &&
			blockHeight >= b.chainParams.BIP0034Height {

			coinbaseTx := block.Transactions()[0]
			err := CheckSerializedHeight(coinbaseTx, blockHeight)
			if err != nil {
				return err
			}
		}

		// Query for the Version Bits state for the segwit soft-fork
		// deployment. If segwit is active, we'll switch over to
		// enforcing all the new rules.
		segwitState, err := b.deploymentState(prevNode,
			chaincfg.DeploymentSegwit)
		if err != nil {
			return err
		}

		// If segwit is active, then we'll need to fully validate the
		// new witness commitment for adherence to the rules.
		if segwitState == ThresholdActive {
			// Validate the witness commitment (if any) within the
			// block.  This involves asserting that if the coinbase
			// contains the special commitment output, then this
			// merkle root matches a computed merkle root of all
			// the wtxid's of the transactions within the block. In
			// addition, various other checks against the
			// coinbase's witness stack.
			if err := ValidateWitnessCommitment(block); err != nil {
				return err
			}

			// Once the witness commitment, witness nonce, and sig
			// op cost have been validated, we can finally assert
			// that the block's weight doesn't exceed the current
			// consensus parameter.
			blockWeight := GetBlockWeight(block)
			if blockWeight > MaxBlockWeight {
				str := fmt.Sprintf("block's weight metric is "+
					"too high - got %v, max %v",
					blockWeight, MaxBlockWeight)
				return ruleError(ErrBlockWeightTooHigh, str)
			}
		}
	}

	return nil
}

// checkBIP0030 ensures blocks do not contain duplicate transactions which
// 'overwrite' older transactions that are not fully spent.  This prevents an
// attack where a coinbase and all of its dependent transactions could be
// duplicated to effectively revert the overwritten transactions to a single
// confirmation thereby making them vulnerable to a double spend.
//
// For more details, see
// https://github.com/bitcoin/bips/blob/master/bip-0030.mediawiki and
// http://r6.ca/blog/20120206T005236Z.html.
//
// This function MUST be called with the chain state lock held (for reads).
func (b *BlockChain) checkBIP0030(node *blockNode, block *chainutil.Block, view *UtxoViewpoint) error {
	// Fetch utxos for all of the transaction outputs in this block.
	// Typically, there will not be any utxos for any of the outputs.
	fetch := make([]wire.OutPoint, 0, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		prevOut := wire.OutPoint{Hash: *tx.Hash()}
		for txOutIdx := range tx.MsgTx().TxOut {
			prevOut.Index = uint32(txOutIdx)
			fetch = append(fetch, prevOut)
		}
	}
	err := view.fetchUtxos(b.utxoCache, fetch)
	if err != nil {
		return err
	}

	// Duplicate transactions are only allowed if the previous transaction
	// is fully spent.
	for _, outpoint := range fetch {
		utxo := view.LookupEntry(outpoint)
		if utxo != nil && !utxo.IsSpent() {
			str := fmt.Sprintf("tried to overwrite transaction %v "+
				"at block height %d that is not fully spent",
				outpoint.Hash, utxo.BlockHeight())
			return ruleError(ErrOverwriteTx, str)
		}
	}

	return nil
}

// CheckTransactionInputs performs a series of checks on the inputs to a
// transaction to ensure they are valid.  An example of some of the checks
// include verifying all inputs exist, ensuring the coinbase seasoning
// requirements are met, detecting double spends, validating all values and fees
// are in the legal range and the total output amount doesn't exceed the input
// amount, and verifying the signatures to prove the spender was the owner of
// the flokicoins and therefore allowed to spend them.  As it checks the inputs,
// it also calculates the total fees for the transaction and returns that value.
//
// NOTE: The transaction MUST have already been sanity checked with the
// CheckTransactionSanity function prior to calling this function.
func CheckTransactionInputs(tx *chainutil.Tx, txHeight int32, utxoView *UtxoViewpoint, chainParams *chaincfg.Params) (int64, error) {
	// Coinbase transactions have no inputs.
	if IsCoinBase(tx) {
		return 0, nil
	}

	var totalLokiIn int64
	for txInIndex, txIn := range tx.MsgTx().TxIn {
		// Ensure the referenced input transaction is available.
		utxo := utxoView.LookupEntry(txIn.PreviousOutPoint)
		if utxo == nil || utxo.IsSpent() {
			str := fmt.Sprintf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOutPoint,
				tx.Hash(), txInIndex)
			return 0, ruleError(ErrMissingTxOut, str)
		}

		// Ensure the transaction is not spending coins which have not
		// yet reached the required coinbase maturity.
		if utxo.IsCoinBase() {
			originHeight := utxo.BlockHeight()
			blocksSincePrev := txHeight - originHeight
			coinbaseMaturity := int32(chainParams.CoinbaseMaturity)
			if blocksSincePrev < coinbaseMaturity {
				str := fmt.Sprintf("tried to spend coinbase "+
					"transaction output %v from height %v "+
					"at height %v before required maturity "+
					"of %v blocks", txIn.PreviousOutPoint,
					originHeight, txHeight,
					coinbaseMaturity)
				return 0, ruleError(ErrImmatureSpend, str)
			}
		}

		// Ensure the transaction amounts are in range.  Each of the
		// output values of the input transactions must not be negative
		// or more than the max allowed per transaction.  All amounts in
		// a transaction are in a unit value known as a loki.  One
		// flokicoin is a quantity of loki as defined by the
		// LokiPerFlokicoin constant.
		originTxLoki := utxo.Amount()
		if originTxLoki < 0 {
			str := fmt.Sprintf("transaction output has negative "+
				"value of %v", chainutil.Amount(originTxLoki))
			return 0, ruleError(ErrBadTxOutValue, str)
		}
		if originTxLoki > chainutil.MaxLoki {
			str := fmt.Sprintf("transaction output value is "+
				"higher than max allowed value: %v > %v ",
				chainutil.Amount(originTxLoki),
				chainutil.MaxLoki)
			return 0, ruleError(ErrBadTxOutValue, str)
		}

		// The total of all outputs must not be more than the max
		// allowed per transaction.  Also, we could potentially overflow
		// the accumulator so check for overflow.
		lastLokiIn := totalLokiIn
		totalLokiIn += originTxLoki
		if totalLokiIn < lastLokiIn ||
			totalLokiIn > chainutil.MaxLoki {
			str := fmt.Sprintf("total value of all transaction "+
				"inputs is %v which is higher than max "+
				"allowed value of %v", totalLokiIn,
				chainutil.MaxLoki)
			return 0, ruleError(ErrBadTxOutValue, str)
		}
	}

	// Calculate the total output amount for this transaction.  It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	var totalLokiOut int64
	for _, txOut := range tx.MsgTx().TxOut {
		totalLokiOut += txOut.Value
	}

	// Ensure the transaction does not spend more than its inputs.
	if totalLokiIn < totalLokiOut {
		str := fmt.Sprintf("total value of all transaction inputs for "+
			"transaction %v is %v which is less than the amount "+
			"spent of %v", tx.Hash(), totalLokiIn, totalLokiOut)
		return 0, ruleError(ErrSpendTooHigh, str)
	}

	// NOTE: flokicoind checks if the transaction fees are < 0 here, but that
	// is an impossible condition because of the check above that ensures
	// the inputs are >= the outputs.
	txFeeInLoki := totalLokiIn - totalLokiOut
	return txFeeInLoki, nil
}

// checkConnectBlock performs several checks to confirm connecting the passed
// block to the chain represented by the passed view does not violate any rules.
// In addition, the passed view is updated to spend all of the referenced
// outputs and add all of the new utxos created by block.  Thus, the view will
// represent the state of the chain as if the block were actually connected and
// consequently the best hash for the view is also updated to passed block.
//
// An example of some of the checks performed are ensuring connecting the block
// would not cause any duplicate transaction hashes for old transactions that
// aren't already fully spent, double spends, exceeding the maximum allowed
// signature operations per block, invalid values in relation to the expected
// block subsidy, or fail transaction script validation.
//
// The CheckConnectBlockTemplate function makes use of this function to perform
// the bulk of its work.  The only difference is this function accepts a node
// which may or may not require reorganization to connect it to the main chain
// whereas CheckConnectBlockTemplate creates a new node which specifically
// connects to the end of the current main chain and then calls this function
// with that node.
//
// This function MUST be called with the chain state lock held (for writes).
func (b *BlockChain) checkConnectBlock(node *blockNode, block *chainutil.Block, view *UtxoViewpoint, stxos *[]SpentTxOut) error {
	// If the side chain blocks end up in the database, a call to
	// CheckBlockSanity should be done here in case a previous version
	// allowed a block that is no longer valid.  However, since the
	// implementation only currently uses memory for the side chain blocks,
	// it isn't currently necessary.

	// The coinbase for the Genesis block is not spendable, so just return
	// an error now.
	if node.hash.IsEqual(b.chainParams.GenesisHash) {
		str := "the coinbase for the genesis block is not spendable"
		return ruleError(ErrMissingTxOut, str)
	}

	// Ensure the view is for the node being checked.
	parentHash := &block.MsgBlock().Header.PrevBlock
	if !view.BestHash().IsEqual(parentHash) {
		return AssertError(fmt.Sprintf("inconsistent view when "+
			"checking block connection: best hash is %v instead "+
			"of expected %v", view.BestHash(), parentHash))
	}

	// BIP0030 added a rule to prevent blocks which contain duplicate
	// transactions that 'overwrite' older transactions which are not fully
	// spent.  See the documentation for checkBIP0030 for more details.
	//
	// There are two blocks in the chain which violate this rule, so the
	// check must be skipped for those blocks.  The isBIP0030Node function
	// is used to determine if this block is one of the two blocks that must
	// be skipped.
	//
	// In addition, as of BIP0034, duplicate coinbases are no longer
	// possible due to its requirement for including the block height in the
	// coinbase and thus it is no longer possible to create transactions
	// that 'overwrite' older ones.  Therefore, only enforce the rule if
	// BIP0034 is not yet active.  This is a useful optimization because the
	// BIP0030 check is expensive since it involves a ton of cache misses in
	// the utxoset.
	if !isBIP0030Node(node) && (node.height < b.chainParams.BIP0034Height) {
		err := b.checkBIP0030(node, block, view)
		if err != nil {
			return err
		}
	}

	// Load all of the utxos referenced by the inputs for all transactions
	// in the block don't already exist in the utxo view from the cache.
	//
	// These utxo entries are needed for verification of things such as
	// transaction inputs, counting pay-to-script-hashes, and scripts.
	err := view.fetchInputUtxos(b.utxoCache, block)
	if err != nil {
		return err
	}

	// BIP0016 describes a pay-to-script-hash type that is considered a
	// "standard" type.  The rules for this BIP only apply to transactions
	// after the timestamp defined by txscript.Bip16Activation.  See
	// https://en.bitcoin.it/wiki/BIP_0016 for more details.
	enforceBIP0016 := node.timestamp >= txscript.Bip16Activation.Unix()

	// Query for the Version Bits state for the segwit soft-fork
	// deployment. If segwit is active, we'll switch over to enforcing all
	// the new rules.
	segwitState, err := b.deploymentState(node.parent, chaincfg.DeploymentSegwit)
	if err != nil {
		return err
	}
	enforceSegWit := segwitState == ThresholdActive

	// The number of signature operations must be less than the maximum
	// allowed per block.  Note that the preliminary sanity checks on a
	// block also include a check similar to this one, but this check
	// expands the count to include a precise count of pay-to-script-hash
	// signature operations in each of the input transaction public key
	// scripts.
	transactions := block.Transactions()
	totalSigOpCost := 0
	for i, tx := range transactions {
		// Since the first (and only the first) transaction has
		// already been verified to be a coinbase transaction,
		// use i == 0 as an optimization for the flag to
		// countP2SHSigOps for whether or not the transaction is
		// a coinbase transaction rather than having to do a
		// full coinbase check again.
		sigOpCost, err := GetSigOpCost(tx, i == 0, view, enforceBIP0016,
			enforceSegWit)
		if err != nil {
			return err
		}

		// Check for overflow or going over the limits.  We have to do
		// this on every loop iteration to avoid overflow.
		lastSigOpCost := totalSigOpCost
		totalSigOpCost += sigOpCost
		if totalSigOpCost < lastSigOpCost || totalSigOpCost > MaxBlockSigOpsCost {
			str := fmt.Sprintf("block contains too many "+
				"signature operations - got %v, max %v",
				totalSigOpCost, MaxBlockSigOpsCost)
			return ruleError(ErrTooManySigOps, str)
		}
	}

	// Perform several checks on the inputs for each transaction.  Also
	// accumulate the total fees.  This could technically be combined with
	// the loop above instead of running another loop over the transactions,
	// but by separating it we can avoid running the more expensive (though
	// still relatively cheap as compared to running the scripts) checks
	// against all the inputs when the signature operations are out of
	// bounds.
	var totalFees int64
	for _, tx := range transactions {
		txFee, err := CheckTransactionInputs(tx, node.height, view,
			b.chainParams)
		if err != nil {
			return err
		}

		// Sum the total fees and ensure we don't overflow the
		// accumulator.
		lastTotalFees := totalFees
		totalFees += txFee
		if totalFees < lastTotalFees {
			return ruleError(ErrBadFees, "total fees for block "+
				"overflows accumulator")
		}

		// Add all of the outputs for this transaction which are not
		// provably unspendable as available utxos.  Also, the passed
		// spent txos slice is updated to contain an entry for each
		// spent txout in the order each transaction spends them.
		err = view.connectTransaction(tx, node.height, stxos)
		if err != nil {
			return err
		}
	}

	// The total output values of the coinbase transaction must not exceed
	// the expected subsidy value plus total transaction fees gained from
	// mining the block.  It is safe to ignore overflow and out of range
	// errors here because those error conditions would have already been
	// caught by checkTransactionSanity.
	var totalLokiOut int64
	for _, txOut := range transactions[0].MsgTx().TxOut {
		totalLokiOut += txOut.Value
	}
	expectedLokiOut := CalcBlockSubsidy(node.height, b.chainParams) +
		totalFees
	if totalLokiOut > expectedLokiOut {
		str := fmt.Sprintf("coinbase transaction for block pays %v "+
			"which is more than expected value of %v",
			totalLokiOut, expectedLokiOut)
		return ruleError(ErrBadCoinbaseValue, str)
	}

	// Don't run scripts if this node is before the latest known good
	// checkpoint since the validity is verified via the checkpoints (all
	// transactions are included in the merkle root hash and any changes
	// will therefore be detected by the next checkpoint).  This is a huge
	// optimization because running the scripts is the most time consuming
	// portion of block handling.
	checkpoint := b.LatestCheckpoint()
	runScripts := true
	if checkpoint != nil && node.height <= checkpoint.Height {
		runScripts = false
	}

	// Blocks created after the BIP0016 activation time need to have the
	// pay-to-script-hash checks enabled.
	var scriptFlags txscript.ScriptFlags
	if enforceBIP0016 {
		scriptFlags |= txscript.ScriptBip16
	}

	// Enforce DER signatures for block versions 3+ once the historical
	// activation threshold has been reached.  This is part of BIP0066.
	blockHeader := &block.MsgBlock().Header
	if blockHeader.Version >= 3 && node.height >= b.chainParams.BIP0066Height {
		scriptFlags |= txscript.ScriptVerifyDERSignatures
	}

	// Enforce CHECKLOCKTIMEVERIFY for block versions 4+ once the historical
	// activation threshold has been reached.  This is part of BIP0065.
	if blockHeader.Version >= 4 && node.height >= b.chainParams.BIP0065Height {
		scriptFlags |= txscript.ScriptVerifyCheckLockTimeVerify
	}

	// Enforce CHECKSEQUENCEVERIFY during all block validation checks once
	// the soft-fork deployment is fully active.
	csvState, err := b.deploymentState(node.parent, chaincfg.DeploymentCSV)
	if err != nil {
		return err
	}
	if csvState == ThresholdActive {
		// If the CSV soft-fork is now active, then modify the
		// scriptFlags to ensure that the CSV op code is properly
		// validated during the script checks bleow.
		scriptFlags |= txscript.ScriptVerifyCheckSequenceVerify

		// We obtain the MTP of the *previous* block in order to
		// determine if transactions in the current block are final.
		medianTime := CalcPastMedianTime(node.parent)

		// Additionally, if the CSV soft-fork package is now active,
		// then we also enforce the relative sequence number based
		// lock-times within the inputs of all transactions in this
		// candidate block.
		for _, tx := range block.Transactions() {
			// A transaction can only be included within a block
			// once the sequence locks of *all* its inputs are
			// active.
			sequenceLock, err := b.calcSequenceLock(node, tx, view,
				false)
			if err != nil {
				return err
			}
			if !SequenceLockActive(sequenceLock, node.height,
				medianTime) {
				str := fmt.Sprintf("block contains " +
					"transaction whose input sequence " +
					"locks are not met")
				return ruleError(ErrUnfinalizedTx, str)
			}
		}
	}

	// Enforce the segwit soft-fork package once the soft-fork has shifted
	// into the "active" version bits state.
	if enforceSegWit {
		scriptFlags |= txscript.ScriptVerifyWitness
		scriptFlags |= txscript.ScriptStrictMultiSig
	}

	// Before we execute the main scripts, we'll also check to see if
	// taproot is active or not.
	taprootState, err := b.deploymentState(
		node.parent, chaincfg.DeploymentTaproot,
	)
	if err != nil {
		return err
	}
	if taprootState == ThresholdActive {
		scriptFlags |= txscript.ScriptVerifyTaproot
	}

	// Now that the inexpensive checks are done and have passed, verify the
	// transactions are actually allowed to spend the coins by running the
	// expensive ECDSA signature check scripts.  Doing this last helps
	// prevent CPU exhaustion attacks.
	if runScripts {
		err := checkBlockScripts(block, view, scriptFlags, b.sigCache,
			b.hashCache)
		if err != nil {
			return err
		}
	}

	// Update the best hash for view to include this block since all of its
	// transactions have been connected.
	view.SetBestHash(&node.hash)

	return nil
}

// CheckConnectBlockTemplate fully validates that connecting the passed block to
// the main chain does not violate any consensus rules, aside from the proof of
// work requirement. The block must connect to the current tip of the main chain.
//
// This function is safe for concurrent access.
func (b *BlockChain) CheckConnectBlockTemplate(block *chainutil.Block) error {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	// Skip the proof of work check as this is just a block template.
	flags := BFNoPoWCheck

	// This only checks whether the block can be connected to the tip of the
	// current chain.
	tip := b.bestChain.Tip()
	header := block.MsgBlock().Header
	if tip.hash != header.PrevBlock {
		str := fmt.Sprintf("previous block must be the current chain tip %v, "+
			"instead got %v", tip.hash, header.PrevBlock)
		return ruleError(ErrPrevBlockNotBest, str)
	}

	err := checkBlockSanity(block, b.chainParams.PowLimit, b.timeSource, flags)
	if err != nil {
		return err
	}

	err = b.checkBlockContext(block, tip, flags)
	if err != nil {
		return err
	}

	// Leave the spent txouts entry nil in the state since the information
	// is not needed and thus extra work can be avoided.
	view := NewUtxoViewpoint()
	view.SetBestHash(&tip.hash)
	newNode := newBlockNode(&header, tip)
	return b.checkConnectBlock(newNode, block, view, nil)
}

// ChainParams returns the Blockchain's configured chaincfg.Params.
//
// NOTE: Part of the ChainCtx interface.
func (b *BlockChain) ChainParams() *chaincfg.Params {
	return b.chainParams
}

// BlocksPerRetarget returns the number of blocks before retargeting occurs.
//
// NOTE: Part of the ChainCtx interface.
func (b *BlockChain) BlocksPerRetarget() int32 {
	return b.blocksPerRetarget
}

// MinRetargetTimespan returns the minimum amount of time to use in the
// difficulty calculation.
//
// NOTE: Part of the ChainCtx interface.
func (b *BlockChain) MinRetargetTimespan() int64 {
	return b.minRetargetTimespan
}

// MaxRetargetTimespan returns the maximum amount of time to use in the
// difficulty calculation.
//
// NOTE: Part of the ChainCtx interface.
func (b *BlockChain) MaxRetargetTimespan() int64 {
	return b.maxRetargetTimespan
}

// VerifyCheckpoint checks that the height and hash match the stored
// checkpoints.
//
// NOTE: Part of the ChainCtx interface.
func (b *BlockChain) VerifyCheckpoint(height int32,
	hash *chainhash.Hash) bool {

	return b.verifyCheckpoint(height, hash)
}

// FindPreviousCheckpoint finds the checkpoint we've encountered during
// validation.
//
// NOTE: Part of the ChainCtx interface.
func (b *BlockChain) FindPreviousCheckpoint() (HeaderCtx, error) {
	checkpoint, err := b.findPreviousCheckpoint()
	if err != nil {
		return nil, err
	}

	if checkpoint == nil {
		// This check is necessary because if we just return the nil
		// blockNode as a HeaderCtx, a caller performing a nil-check
		// will fail. This is a quirk of go where a nil value stored in
		// an interface is different from the actual nil interface.
		return nil, nil
	}

	return checkpoint, err
}

// A compile-time assertion to ensure BlockChain implements the ChainCtx
// interface.
var _ ChainCtx = (*BlockChain)(nil)
