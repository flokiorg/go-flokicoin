// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"math/big"
	"time"

	"github.com/flokiorg/go-flokicoin/blockchain/internal/workmath"
	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
)

const (
	// Move 12.5% toward the observed spacing each block (8 => 1/8).
	DigishieldAmplitudeDivisor int64 = 8

	// Lower clamp as a rational of the target spacing: 3/4 => 0.75 * target.
	DigishieldClampMinNum int64 = 3
	DigishieldClampMinDen int64 = 4

	// Upper clamp as a rational of the target spacing: 3/2 => 1.5 * target.
	DigishieldClampMaxNum int64 = 3
	DigishieldClampMaxDen int64 = 2

	// “Late block” multiple of the target spacing that permits a min-difficulty
	// block when that feature is enabled (2 => later than 2× target spacing).
	DigishieldLateBlockMultiple int64 = 60
)

// HashToBig converts a chainhash.Hash into a big.Int that can be used to
// perform math comparisons.
func HashToBig(hash *chainhash.Hash) *big.Int {
	return workmath.HashToBig(hash)
}

// CompactToBig converts a compact representation of a whole number N to an
// unsigned 32-bit number.  The representation is similar to IEEE754 floating
// point numbers.
//
// Like IEEE754 floating point, there are three basic components: the sign,
// the exponent, and the mantissa.  They are broken out as follows:
//
// - the most significant 8 bits represent the unsigned base 256 exponent
// - bit 23 (the 24th bit) represents the sign bit
// - the least significant 23 bits represent the mantissa
//
//	-------------------------------------------------
//	|   Exponent     |    Sign    |    Mantissa     |
//	-------------------------------------------------
//	| 8 bits [31-24] | 1 bit [23] | 23 bits [22-00] |
//	-------------------------------------------------
//
// The formula to calculate N is:
//
//	N = (-1^sign) * mantissa * 256^(exponent-3)
//
// This compact form is only used in flokicoin to encode unsigned 256-bit numbers
// which represent difficulty targets, thus there really is not a need for a
// sign bit, but it is implemented here to stay consistent with flokicoind.
func CompactToBig(compact uint32) *big.Int {
	return workmath.CompactToBig(compact)
}

// BigToCompact converts a whole number N to a compact representation using
// an unsigned 32-bit number.  The compact representation only provides 23 bits
// of precision, so values larger than (2^23 - 1) only encode the most
// significant digits of the number.  See CompactToBig for details.
func BigToCompact(n *big.Int) uint32 {
	return workmath.BigToCompact(n)
}

// CalcWork calculates a work value from difficulty bits.  Flokicoin increases
// the difficulty for generating a block by decreasing the value which the
// generated hash must be less than.  This difficulty target is stored in each
// block header using a compact representation as described in the documentation
// for CompactToBig.  The main chain is selected by choosing the chain that has
// the most proof of work (highest difficulty).  Since a lower target difficulty
// value equates to higher actual difficulty, the work value which will be
// accumulated must be the inverse of the difficulty.  Also, in order to avoid
// potential division by zero and really small floating point numbers, the
// result adds 1 to the denominator and multiplies the numerator by 2^256.
func CalcWork(bits uint32) *big.Int {
	return workmath.CalcWork(bits)
}

// calcEasiestDifficulty calculates the easiest possible difficulty that a block
// can have given starting difficulty bits and a duration.  It is mainly used to
// verify that claimed proof of work by a block is sane as compared to a
// known good checkpoint.
func (b *BlockChain) calcEasiestDifficulty(bits uint32, duration time.Duration) uint32 {
	// Convert types used in the calculations below.
	durationVal := int64(duration / time.Second)
	adjustmentFactor := big.NewInt(b.chainParams.RetargetAdjustmentFactor)
	targetSeconds := int64(b.chainParams.TargetTimePerBlock / time.Second)

	if b.chainParams.ReduceMinDifficulty {
		if durationVal > DigishieldLateBlockMultiple*targetSeconds {
			return b.chainParams.PowLimitBits
		}
	}

	// Since easier difficulty equates to higher numbers, the easiest
	// difficulty for a given duration is the largest value possible given
	// the number of retargets for the duration and starting difficulty
	// multiplied by the max adjustment factor.
	newTarget := CompactToBig(bits)
	for durationVal > 0 && newTarget.Cmp(b.chainParams.PowLimit) < 0 {
		newTarget.Mul(newTarget, adjustmentFactor)
		durationVal -= b.maxRetargetTimespan
	}

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(b.chainParams.PowLimit) > 0 {
		newTarget.Set(b.chainParams.PowLimit)
	}

	return BigToCompact(newTarget)
}

// findPrevTestNetDifficulty returns the difficulty of the previous block which
// did not have the special testnet minimum difficulty rule applied.
func findPrevTestNetDifficulty(startNode HeaderCtx, c ChainCtx) uint32 {
	// Search backwards through the chain for the last block without
	// the special rule applied.
	iterNode := startNode
	for iterNode != nil && iterNode.Height()%c.BlocksPerRetarget() != 0 &&
		iterNode.Bits() == c.ChainParams().PowLimitBits {

		iterNode = iterNode.Parent()
	}

	// Return the found difficulty or the minimum difficulty if no
	// appropriate block was found.
	lastBits := c.ChainParams().PowLimitBits
	if iterNode != nil {
		lastBits = iterNode.Bits()
	}
	return lastBits
}

// calcNextRequiredDifficulty calculates the required difficulty for the block
// after the passed previous HeaderCtx based on the difficulty retarget rules.
// This function differs from the exported CalcNextRequiredDifficulty in that
// the exported version uses the current best chain as the previous HeaderCtx
// while this function accepts any block node. This function accepts a ChainCtx
// parameter that gives the necessary difficulty context variables.
func calcNextRequiredDifficulty(lastNode HeaderCtx, newBlockTime time.Time, c ChainCtx) (uint32, error) {

	// Emulate the same behavior as Flokicoin that for regtest there is
	// no difficulty retargeting.
	if c.ChainParams().PoWNoRetargeting {
		return c.ChainParams().PowLimitBits, nil
	}

	// Genesis
	if lastNode == nil {
		return c.ChainParams().PowLimitBits, nil
	}

	heightNext := lastNode.Height() + 1
	if heightNext <= 5 { // the 5 blocks
		return c.ChainParams().PowLimitBits, nil
	}

	// Digishield
	if heightNext >= c.ChainParams().DigishieldActivationHeight {
		return calcNextWorkDigishield(lastNode, newBlockTime, c)
	}

	//  Legacy
	return calcNextWorkLegacy(lastNode, newBlockTime, c)

}

// calcNextWorkDigishield implements Dogecoin's Digishield:
// per-block retarget, raw timestamps, amplitude filter, and clamping.
func calcNextWorkDigishield(lastNode HeaderCtx, newBlockTime time.Time, c ChainCtx) (uint32, error) {
	params := c.ChainParams()
	targetSeconds := int64(params.TargetTimePerBlock / time.Second)
	nextHeight := lastNode.Height() + 1

	// If enabled (typically on testnets), allow a block to be mined at
	// minimum difficulty when its timestamp is more than
	// DigishieldLateBlockMultiple × target spacing after the previous block.
	// This prevents the chain from stalling in low-hash environments.
	if params.ReduceMinDifficulty {
		lateThreshold := lastNode.Timestamp() + DigishieldLateBlockMultiple*targetSeconds
		if newBlockTime.Unix() > lateThreshold {
			return params.PowLimitBits, nil
		}
	}

	// Use the direct parent.
	prevNode := lastNode.RelativeAncestorCtx(1)
	if prevNode == nil {
		return 0, AssertError("unable to obtain previous block for retarget")
	}

	// Raw header timestamps.
	actualTimespan := lastNode.Timestamp() - prevNode.Timestamp()

	// Amplitude filter: modulated = target + (actual - target) / divisor.
	modulatedTimespan := targetSeconds + (actualTimespan-targetSeconds)/DigishieldAmplitudeDivisor

	// Clamp to [0.75T, 1.5T].
	minTimespan := (targetSeconds * DigishieldClampMinNum) / DigishieldClampMinDen
	maxTimespan := (targetSeconds * DigishieldClampMaxNum) / DigishieldClampMaxDen
	if modulatedTimespan < minTimespan {
		modulatedTimespan = minTimespan
	} else if modulatedTimespan > maxTimespan {
		modulatedTimespan = maxTimespan
	}

	// Scale old target and cap to PowLimit.
	oldTarget := CompactToBig(lastNode.Bits())
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(modulatedTimespan))
	newTarget.Div(newTarget, big.NewInt(targetSeconds))
	if newTarget.Cmp(params.PowLimit) > 0 {
		newTarget.Set(params.PowLimit)
	}

	newBits := BigToCompact(newTarget)
	log.Debugf(
		"Digishield @%d old=%08x new=%08x actual=%ds mod=%ds target=%ds ampDiv=%d clamp=[%d/%d..%d/%d] lateMult=%dx",
		nextHeight, lastNode.Bits(), newBits,
		actualTimespan, modulatedTimespan, targetSeconds,
		DigishieldAmplitudeDivisor,
		DigishieldClampMinNum, DigishieldClampMinDen,
		DigishieldClampMaxNum, DigishieldClampMaxDen,
		DigishieldLateBlockMultiple,
	)
	return newBits, nil
}

// calcNextWorkLegacy: simple per-block scaling with generic clamps.
func calcNextWorkLegacy(lastNode HeaderCtx, _ time.Time, c ChainCtx) (uint32, error) {
	// Per-block retarget: use the direct parent for timing comparison.
	prev := lastNode.RelativeAncestorCtx(1)
	if prev == nil {
		return 0, AssertError("unable to obtain previous block for retarget")
	}

	// Limit the amount of adjustment that can occur to the previous
	// difficulty.
	actualTimespan := lastNode.Timestamp() - prev.Timestamp()

	adjustedTimespan := actualTimespan
	if actualTimespan < c.MinRetargetTimespan() {
		adjustedTimespan = c.MinRetargetTimespan()
	} else if actualTimespan > c.MaxRetargetTimespan() {
		adjustedTimespan = c.MaxRetargetTimespan()
	}

	oldTarget := CompactToBig(lastNode.Bits())

	// Calculate new target difficulty as:
	//  currentDifficulty * (adjustedTimespan / targetTimespan)
	// The result uses integer division which means it will be slightly
	// rounded down.  Flokicoind also uses integer division to calculate this
	// result.
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
	targetTimeSpan := int64(c.ChainParams().TargetTimespan / time.Second)
	newTarget.Div(newTarget, big.NewInt(targetTimeSpan))

	// Limit new value to the proof of work limit.
	if newTarget.Cmp(c.ChainParams().PowLimit) > 0 {
		newTarget.Set(c.ChainParams().PowLimit)
	}

	// Log new target difficulty and return it.  The new target logging is
	// intentionally converting the bits back to a number instead of using
	// newTarget since conversion to the compact representation loses
	// precision.
	newTargetBits := BigToCompact(newTarget)
	log.Debugf("Per-block difficulty retarget at block height %d", lastNode.Height()+1)
	log.Debugf("Old target %08x (%064x)", lastNode.Bits(), oldTarget)
	log.Debugf("New target %08x (%064x)", newTargetBits, CompactToBig(newTargetBits))
	log.Debugf("Actual timespan %v, adjusted timespan %v, target timespan %v",
		time.Duration(actualTimespan)*time.Second,
		time.Duration(adjustedTimespan)*time.Second,
		c.ChainParams().TargetTimespan)

	return newTargetBits, nil
}

// CalcNextRequiredDifficulty calculates the required difficulty for the block
// after the end of the current best chain based on the difficulty retarget
// rules.
//
// This function is safe for concurrent access.
func (b *BlockChain) CalcNextRequiredDifficulty(timestamp time.Time) (uint32, error) {
	b.chainLock.Lock()
	difficulty, err := calcNextRequiredDifficulty(b.bestChain.Tip(), timestamp, b)
	b.chainLock.Unlock()
	return difficulty, err
}
