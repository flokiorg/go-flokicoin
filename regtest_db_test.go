package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flokiorg/go-flokicoin/log"
)

// TestRemoveRegressionDB checks that the regtest block database is deleted on
// start by default and kept when --keepregtestdb is set.
func TestRemoveRegressionDB(t *testing.T) {
	saved, savedLog := cfg, flcdLog
	flcdLog = log.Disabled
	t.Cleanup(func() { cfg, flcdLog = saved, savedLog })

	tests := []struct {
		name     string
		cfg      config
		wantKept bool
	}{
		{"regtest deletes", config{RegressionTest: true}, false},
		{"regtest keeps with flag", config{RegressionTest: true, KeepRegTestDB: true}, true},
		{"mainnet untouched", config{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "blocks_ffldb")
			if err := os.MkdirAll(dbPath, 0700); err != nil {
				t.Fatal(err)
			}
			c := tc.cfg
			cfg = &c
			if err := removeRegressionDB(dbPath); err != nil {
				t.Fatalf("removeRegressionDB: %v", err)
			}
			_, err := os.Stat(dbPath)
			if kept := err == nil; kept != tc.wantKept {
				t.Fatalf("kept = %v, want %v", kept, tc.wantKept)
			}
		})
	}
}
