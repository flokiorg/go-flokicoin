package main

import (
	"testing"

	"github.com/flokiorg/go-flokicoin/chainjson"
	"github.com/flokiorg/go-flokicoin/mempool"
)

// TestHandleEstimateFee checks that estimatefee answers with its own command
// type instead of panicking, and refuses what it cannot estimate.
func TestHandleEstimateFee(t *testing.T) {
	cmd := func(n int64) *chainjson.EstimateFeeCmd { return &chainjson.EstimateFeeCmd{NumBlocks: n} }

	// No estimator configured.
	s := &rpcServer{}
	if _, err := handleEstimateFee(s, cmd(6), nil); err == nil {
		t.Fatal("want an error with fee estimation disabled")
	}

	// A fresh estimator has no data yet: an error, never a panic.
	s = &rpcServer{cfg: rpcserverConfig{FeeEstimator: mempool.NewFeeEstimator(
		mempool.DefaultEstimateFeeMaxRollback, mempool.DefaultEstimateFeeMinRegisteredBlocks)}}
	if _, err := handleEstimateFee(s, cmd(6), nil); err == nil {
		t.Fatal("want an error from an estimator with no data")
	}

	// A non-positive target is refused.
	if res, err := handleEstimateFee(s, cmd(0), nil); err == nil || res != -1.0 {
		t.Fatalf("got (%v, %v), want (-1, error)", res, err)
	}
}
