package mempool

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFeePersistRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, feeEstimatesFileName)

	est := NewFeeEstimator(DefaultEstimateFeeMaxRollback, DefaultEstimateFeeMinRegisteredBlocks)
	if err := SaveFeeEstimatorToFile(path, est, time.Now()); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadFeeEstimatorFromFile(path, false)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected estimator")
	}
}

func TestFeePersistStale(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, feeEstimatesFileName)

	est := NewFeeEstimator(DefaultEstimateFeeMaxRollback, DefaultEstimateFeeMinRegisteredBlocks)
	old := time.Now().Add(-feeFileMaxAge - time.Hour)
	if err := SaveFeeEstimatorToFile(path, est, old); err != nil {
		t.Fatalf("save: %v", err)
	}

	_, err := LoadFeeEstimatorFromFile(path, false)
	if !errors.Is(err, ErrStaleFeeEstimates) {
		t.Fatalf("expected stale error, got %v", err)
	}

	// accept stale should succeed
	loaded, err := LoadFeeEstimatorFromFile(path, true)
	if err != nil {
		t.Fatalf("load stale with accept: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected estimator when accepting stale data")
	}

	// Ensure the file still exists.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat: %v", err)
	}
}
