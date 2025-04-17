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

// BitcoindVersion represents the version of the flokicoind the client is
// currently connected to.
type BitcoindVersion uint8

const (
	// FlokicoindPre19 represents a flokicoind version before 0.19.0.
	FlokicoindPre19 BitcoindVersion = iota

	// BitcoindPre22 represents a flokicoind version equal to or greater than
	// 0.19.0 and smaller than 22.0.0.
	FlokicoindPre22

	// BitcoindPre24 represents a flokicoind version equal to or greater than
	// 22.0.0 and smaller than 24.0.0.
	FlokicoindPre24

	// BitcoindPre25 represents a flokicoind version equal to or greater than
	// 24.0.0 and smaller than 25.0.0.
	FlokicoindPre25

	// BitcoindPre25 represents a flokicoind version equal to or greater than
	// 25.0.0.
	FlokicoindPost25
)

// String returns a human-readable backend version.
func (b BitcoindVersion) String() string {
	switch b {
	case FlokicoindPre19:
		return "flokicoind 0.19 and below"

	case FlokicoindPre22:
		return "flokicoind v0.19.0-v22.0.0"

	case FlokicoindPre24:
		return "flokicoind v22.0.0-v24.0.0"

	case FlokicoindPre25:
		return "flokicoind v24.0.0-v25.0.0"

	case FlokicoindPost25:
		return "flokicoind v25.0.0 and above"

	default:
		return "unknown"
	}
}

// SupportUnifiedSoftForks returns true if the backend supports the unified
// softforks format.
func (b BitcoindVersion) SupportUnifiedSoftForks() bool {
	// Versions of bitcoind on or after v0.19.0 use the unified format.
	return b > FlokicoindPre19
}

// SupportTestMempoolAccept returns true if bitcoind version is 22.0.0 or
// above.
func (b BitcoindVersion) SupportTestMempoolAccept() bool {
	return b > FlokicoindPre22
}

// SupportGetTxSpendingPrevOut returns true if bitcoind version is 24.0.0 or
// above.
func (b BitcoindVersion) SupportGetTxSpendingPrevOut() bool {
	return b > FlokicoindPre24
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
		return FlokicoindPre19

	case version < bitcoind22Str:
		return FlokicoindPre22

	case version < bitcoind24Str:
		return FlokicoindPre24

	case version < bitcoind25Str:
		return FlokicoindPre25

	default:
		return FlokicoindPost25
	}
}

// FlokicoindVersion represents the version of the flokicoind the client is currently
// connected to.
type FlokicoindVersion int32

const (
	// FlokicoindPre2401 describes a flokicoind version before 0.24.1, which doesn't
	// include the `testmempoolaccept` and `gettxspendingprevout` RPCs.
	FlokicoindPre2401 FlokicoindVersion = iota

	// FlokicoindPost2401 describes a flokicoind version equal to or greater than
	// 0.24.1.
	FlokicoindPost2401
)

// String returns a human-readable backend version.
func (b FlokicoindVersion) String() string {
	switch b {
	case FlokicoindPre2401:
		return "flokicoind 24.0.0 and below"

	case FlokicoindPost2401:
		return "flokicoind 24.1.0 and above"

	default:
		return "unknown"
	}
}

// SupportUnifiedSoftForks returns true if the backend supports the unified
// softforks format.
//
// NOTE: always true for flokicoind as we didn't track it before.
func (b FlokicoindVersion) SupportUnifiedSoftForks() bool {
	return true
}

// SupportTestMempoolAccept returns true if flokicoind version is 24.1.0 or above.
func (b FlokicoindVersion) SupportTestMempoolAccept() bool {
	return b > FlokicoindPre2401
}

// SupportGetTxSpendingPrevOut returns true if flokicoind version is 24.1.0 or above.
func (b FlokicoindVersion) SupportGetTxSpendingPrevOut() bool {
	return b > FlokicoindPre2401
}

// Compile-time checks to ensure that FlokicoindVersion satisfy the BackendVersion
// interface.
var _ BackendVersion = FlokicoindVersion(0)

const (
	// flokicoind2401Val is the int representation of flokicoind v0.24.1.
	flokicoind2401Val = 240100
)

// parseFlokicoindVersion parses the flokicoind version from its string representation.
func parseFlokicoindVersion(version int32) FlokicoindVersion {
	switch {
	case version < flokicoind2401Val:
		return FlokicoindPre2401

	default:
		return FlokicoindPost2401
	}
}
