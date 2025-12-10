package mempool

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	feeEstimatesFileName = "fee_estimates.dat"
	feeFileMagic         = "FEE1"
	feeFileVersion       = uint32(1)
	feeFileMaxAge        = 60 * time.Hour
)

// ErrStaleFeeEstimates indicates persisted estimates are too old to trust.
var ErrStaleFeeEstimates = errors.New("stale fee estimates")

// FeeEstimatesPath returns the default path for the fee estimates file inside
// a data directory.
func FeeEstimatesPath(dataDir string) string {
	return filepath.Join(dataDir, feeEstimatesFileName)
}

// SaveFeeEstimatorToFile writes the estimator state to disk using a small
// versioned binary format. The write is atomic (temp + rename).
func SaveFeeEstimatorToFile(path string, ef *FeeEstimator, now time.Time) error {
	if ef == nil {
		return errors.New("nil fee estimator")
	}

	state := ef.Save()
	buf := bytes.NewBuffer(make([]byte, 0, len(state)+32))
	// magic
	if _, err := buf.Write([]byte(feeFileMagic)); err != nil {
		return err
	}
	// version
	if err := binary.Write(buf, binary.BigEndian, feeFileVersion); err != nil {
		return err
	}
	// timestamp
	if err := binary.Write(buf, binary.BigEndian, now.Unix()); err != nil {
		return err
	}
	// length-prefixed payload
	if err := binary.Write(buf, binary.BigEndian, uint32(len(state))); err != nil {
		return err
	}
	if _, err := buf.Write(state); err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadFeeEstimatorFromFile attempts to restore an estimator from a file. If
// the file is missing, stale, or invalid, an error is returned.
func LoadFeeEstimatorFromFile(path string, acceptStale bool) (*FeeEstimator, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(raw)
	magic := make([]byte, len(feeFileMagic))
	if _, err := io.ReadFull(reader, magic); err != nil {
		return nil, err
	}
	if string(magic) != feeFileMagic {
		return nil, fmt.Errorf("invalid fee estimates magic: %q", string(magic))
	}

	var version uint32
	if err := binary.Read(reader, binary.BigEndian, &version); err != nil {
		return nil, err
	}
	if version != feeFileVersion {
		return nil, fmt.Errorf("unexpected fee estimates version %d", version)
	}

	var ts int64
	if err := binary.Read(reader, binary.BigEndian, &ts); err != nil {
		return nil, err
	}

	var payloadLen uint32
	if err := binary.Read(reader, binary.BigEndian, &payloadLen); err != nil {
		return nil, err
	}
	if payloadLen > uint32(len(raw)) {
		return nil, fmt.Errorf("fee estimates payload length invalid (%d)", payloadLen)
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}

	if !acceptStale {
		age := time.Since(time.Unix(ts, 0))
		if age > feeFileMaxAge {
			return nil, ErrStaleFeeEstimates
		}
	}

	return RestoreFeeEstimator(payload)
}
