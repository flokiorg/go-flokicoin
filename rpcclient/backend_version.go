package rpcclient

import "strings"

// BackendVersion defines an interface to handle the version of the backend
// used by the client.
type BackendVersion interface {
	// String returns a human-readable backend version.
	String() string

	// SupportUnifiedSoftForks returns true if the backend supports the
	// unified softforks format.
	SupportUnifiedSoftForks() bool

	// SupportTestMempoolAccept returns true if the backend supports the
	// testmempoolaccept RPC.
	SupportTestMempoolAccept() bool

	// SupportGetTxSpendingPrevOut returns true if the backend supports the
	// gettxspendingprevout RPC.
	SupportGetTxSpendingPrevOut() bool
}

// BitcoindVersion represents the version of the lokid the client is
// currently connected to.
type BitcoindVersion uint8

const (
	// LokidPre19 represents a lokid version before 0.19.0.
	LokidPre19 BitcoindVersion = iota

	// BitcoindPre22 represents a lokid version equal to or greater than
	// 0.19.0 and smaller than 22.0.0.
	LokidPre22

	// BitcoindPre24 represents a lokid version equal to or greater than
	// 22.0.0 and smaller than 24.0.0.
	LokidPre24

	// BitcoindPre25 represents a lokid version equal to or greater than
	// 24.0.0 and smaller than 25.0.0.
	LokidPre25

	// BitcoindPre25 represents a lokid version equal to or greater than
	// 25.0.0.
	LokidPost25
)

// String returns a human-readable backend version.
func (b BitcoindVersion) String() string {
	switch b {
	case LokidPre19:
		return "lokid 0.19 and below"

	case LokidPre22:
		return "lokid v0.19.0-v22.0.0"

	case LokidPre24:
		return "lokid v22.0.0-v24.0.0"

	case LokidPre25:
		return "lokid v24.0.0-v25.0.0"

	case LokidPost25:
		return "lokid v25.0.0 and above"

	default:
		return "unknown"
	}
}

// SupportUnifiedSoftForks returns true if the backend supports the unified
// softforks format.
func (b BitcoindVersion) SupportUnifiedSoftForks() bool {
	// Versions of bitcoind on or after v0.19.0 use the unified format.
	return b > LokidPre19
}

// SupportTestMempoolAccept returns true if bitcoind version is 22.0.0 or
// above.
func (b BitcoindVersion) SupportTestMempoolAccept() bool {
	return b > LokidPre22
}

// SupportGetTxSpendingPrevOut returns true if bitcoind version is 24.0.0 or
// above.
func (b BitcoindVersion) SupportGetTxSpendingPrevOut() bool {
	return b > LokidPre24
}

// Compile-time checks to ensure that BitcoindVersion satisfy the
// BackendVersion interface.
var _ BackendVersion = BitcoindVersion(0)

const (
	// bitcoind19Str is the string representation of bitcoind v0.19.0.
	bitcoind19Str = "0.19.0"

	// bitcoind22Str is the string representation of bitcoind v22.0.0.
	bitcoind22Str = "22.0.0"

	// bitcoind24Str is the string representation of bitcoind v24.0.0.
	bitcoind24Str = "24.0.0"

	// bitcoind25Str is the string representation of bitcoind v25.0.0.
	bitcoind25Str = "25.0.0"

	// bitcoindVersionPrefix specifies the prefix included in every bitcoind
	// version exposed through GetNetworkInfo.
	bitcoindVersionPrefix = "/Loki:"

	// bitcoindVersionSuffix specifies the suffix included in every bitcoind
	// version exposed through GetNetworkInfo.
	bitcoindVersionSuffix = "/"
)

// parseBitcoindVersion parses the bitcoind version from its string
// representation.
func parseBitcoindVersion(version string) BitcoindVersion {
	// Trim the version of its prefix and suffix to determine the
	// appropriate version number.
	version = strings.TrimPrefix(
		strings.TrimSuffix(version, bitcoindVersionSuffix),
		bitcoindVersionPrefix,
	)
	switch {
	case version < bitcoind19Str:
		return LokidPre19

	case version < bitcoind22Str:
		return LokidPre22

	case version < bitcoind24Str:
		return LokidPre24

	case version < bitcoind25Str:
		return LokidPre25

	default:
		return LokidPost25
	}
}

// LokidVersion represents the version of the lokid the client is currently
// connected to.
type LokidVersion int32

const (
	// LokidPre2401 describes a lokid version before 0.24.1, which doesn't
	// include the `testmempoolaccept` and `gettxspendingprevout` RPCs.
	LokidPre2401 LokidVersion = iota

	// LokidPost2401 describes a lokid version equal to or greater than
	// 0.24.1.
	LokidPost2401
)

// String returns a human-readable backend version.
func (b LokidVersion) String() string {
	switch b {
	case LokidPre2401:
		return "lokid 24.0.0 and below"

	case LokidPost2401:
		return "lokid 24.1.0 and above"

	default:
		return "unknown"
	}
}

// SupportUnifiedSoftForks returns true if the backend supports the unified
// softforks format.
//
// NOTE: always true for lokid as we didn't track it before.
func (b LokidVersion) SupportUnifiedSoftForks() bool {
	return true
}

// SupportTestMempoolAccept returns true if lokid version is 24.1.0 or above.
func (b LokidVersion) SupportTestMempoolAccept() bool {
	return b > LokidPre2401
}

// SupportGetTxSpendingPrevOut returns true if lokid version is 24.1.0 or above.
func (b LokidVersion) SupportGetTxSpendingPrevOut() bool {
	return b > LokidPre2401
}

// Compile-time checks to ensure that LokidVersion satisfy the BackendVersion
// interface.
var _ BackendVersion = LokidVersion(0)

const (
	// lokid2401Val is the int representation of lokid v0.24.1.
	lokid2401Val = 240100
)

// parseLokidVersion parses the lokid version from its string representation.
func parseLokidVersion(version int32) LokidVersion {
	switch {
	case version < lokid2401Val:
		return LokidPre2401

	default:
		return LokidPost2401
	}
}
