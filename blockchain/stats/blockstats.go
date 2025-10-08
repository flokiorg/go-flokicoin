package stats

import (
	"fmt"
	"math"
	"sort"

	"github.com/flokiorg/go-flokicoin/blockchain"
	"github.com/flokiorg/go-flokicoin/chainutil"
)

var feeRatePercentilesTargets = []float64{10, 25, 50, 75, 90}

// BlockStats aggregates commonly used statistics for a block.
type BlockStats struct {
	TotalSize            int64
	TotalWeight          int64
	TotalFees            int64
	TotalOutputValue     int64
	TotalInputs          int64
	TotalOutputs         int64
	SegWitTotalSize      int64
	SegWitTotalWeight    int64
	SegWitTxs            int64
	UTXOIncrease         int64
	UTXOSizeIncrease     int64
	NonCoinbaseCount     int64
	TotalNonCoinbaseSize int64
	MinFee               int64
	MaxFee               int64
	MinFeeRate           int64
	MaxFeeRate           int64
	MinTxSize            int64
	MaxTxSize            int64
	TxCount              int64
	Fees                 []int64
	FeeRates             []int64
	TxSizes              []int64
}

// ComputeBlockStats returns aggregated statistics for the provided block using
// the associated spend journal entries.
func ComputeBlockStats(block *chainutil.Block, stxos []blockchain.SpentTxOut) (*BlockStats, error) {
	stats := &BlockStats{
		MinFee:     math.MaxInt64,
		MinFeeRate: math.MaxInt64,
		MinTxSize:  math.MaxInt64,
		TxCount:    int64(len(block.Transactions())),
	}

	var stxoIndex int

	for _, tx := range block.Transactions() {
		msgTx := tx.MsgTx()
		txSize := int64(msgTx.SerializeSize())

		txWeight := blockchain.GetTransactionWeight(tx)
		stats.TotalSize += txSize
		stats.TotalWeight += txWeight
		stats.TxSizes = append(stats.TxSizes, txSize)
		if txSize < stats.MinTxSize {
			stats.MinTxSize = txSize
		}
		if txSize > stats.MaxTxSize {
			stats.MaxTxSize = txSize
		}

		outputCount := int64(len(msgTx.TxOut))
		stats.TotalOutputs += outputCount
		stats.UTXOIncrease += outputCount

		var txOutputValue int64
		var outputSizeSum int64
		for _, txOut := range msgTx.TxOut {
			txOutputValue += txOut.Value
			outputSizeSum += int64(len(txOut.PkScript)) + 8 // 8 bytes for value
		}
		stats.TotalOutputValue += txOutputValue
		stats.UTXOSizeIncrease += outputSizeSum

		if tx.HasWitness() {
			stats.SegWitTotalSize += txSize
			stats.SegWitTotalWeight += txWeight
			stats.SegWitTxs++
		}

		inputCount := int64(len(msgTx.TxIn))
		stats.TotalInputs += inputCount

		if blockchain.IsCoinBaseTx(msgTx) {
			continue
		}

		if stxoIndex+int(inputCount) > len(stxos) {
			return nil, fmt.Errorf("spend journal incomplete for tx %s", tx.Hash())
		}

		stats.NonCoinbaseCount++
		stats.TotalNonCoinbaseSize += txSize
		stats.UTXOIncrease -= inputCount

		var inputValue int64
		var spentSizeSum int64
		for i := int64(0); i < inputCount; i++ {
			stxo := stxos[stxoIndex]
			stxoIndex++
			inputValue += stxo.Amount
			spentSizeSum += int64(len(stxo.PkScript)) + 8
		}
		stats.UTXOSizeIncrease -= spentSizeSum

		fee := inputValue - txOutputValue
		if fee < 0 {
			fee = 0
		}

		var feeRate int64
		if txSize > 0 {
			feeRate = fee * 1000 / txSize
		}

		stats.TotalFees += fee
		stats.Fees = append(stats.Fees, fee)
		stats.FeeRates = append(stats.FeeRates, feeRate)

		if fee < stats.MinFee {
			stats.MinFee = fee
		}
		if fee > stats.MaxFee {
			stats.MaxFee = fee
		}
		if feeRate < stats.MinFeeRate {
			stats.MinFeeRate = feeRate
		}
		if feeRate > stats.MaxFeeRate {
			stats.MaxFeeRate = feeRate
		}
	}

	if stxoIndex != len(stxos) {
		return nil, fmt.Errorf("spend journal contains %d entries, used %d", len(stxos), stxoIndex)
	}

	if stats.MinTxSize == math.MaxInt64 {
		stats.MinTxSize = 0
	}
	if stats.MinFee == math.MaxInt64 {
		stats.MinFee = 0
	}
	if stats.MinFeeRate == math.MaxInt64 {
		stats.MinFeeRate = 0
	}

	return stats, nil
}

// AverageFee returns the average fee paid by non-coinbase transactions.
func (bs *BlockStats) AverageFee() int64 {
	if bs.NonCoinbaseCount == 0 {
		return 0
	}
	return bs.TotalFees / bs.NonCoinbaseCount
}

// AverageFeeRate returns the average fee rate in lokis/kB for non-coinbase transactions.
func (bs *BlockStats) AverageFeeRate() int64 {
	if bs.TotalNonCoinbaseSize == 0 {
		return 0
	}
	return bs.TotalFees * 1000 / bs.TotalNonCoinbaseSize
}

// AverageTxSize returns the average serialized transaction size in bytes.
func (bs *BlockStats) AverageTxSize() int64 {
	if bs.TxCount == 0 {
		return 0
	}
	return bs.TotalSize / bs.TxCount
}

// MedianFee returns the median transaction fee.
func (bs *BlockStats) MedianFee() int64 {
	return medianInt64(bs.Fees)
}

// MedianTxSize returns the median transaction size.
func (bs *BlockStats) MedianTxSize() int64 {
	return medianInt64(bs.TxSizes)
}

// FeeRatePercentiles returns the default fee rate percentiles.
func (bs *BlockStats) FeeRatePercentiles() []int64 {
	return percentilesInt64(bs.FeeRates, feeRatePercentilesTargets)
}

func medianInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func percentilesInt64(values []int64, targets []float64) []int64 {
	results := make([]int64, len(targets))
	if len(values) == 0 {
		return results
	}

	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	for i, p := range targets {
		idx := int(math.Ceil(p/100*float64(len(sorted)))) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		results[i] = sorted[idx]
	}
	return results
}
