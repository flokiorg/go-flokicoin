package rpcclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParseFlokicoindVersion checks that the correct version from flokicoind's
// `getnetworkinfo` RPC call is parsed.
func TestParseFlokicoindVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		rpcVersion    string
		parsedVersion BitcoindVersion
	}{
		{
			name:          "parse version 0.19 and below",
			rpcVersion:    "/Loki:0.18.0/",
			parsedVersion: FlokicoindPre19,
		},
		{
			name:          "parse version 0.19",
			rpcVersion:    "/Loki:0.19.0/",
			parsedVersion: FlokicoindPre22,
		},
		{
			name:          "parse version 0.19 - 22.0",
			rpcVersion:    "/Loki:0.20.1/",
			parsedVersion: FlokicoindPre22,
		},
		{
			name:          "parse version 22.0",
			rpcVersion:    "/Loki:22.0.0/",
			parsedVersion: FlokicoindPre24,
		},
		{
			name:          "parse version 22.0 - 24.0",
			rpcVersion:    "/Loki:23.1.0/",
			parsedVersion: FlokicoindPre24,
		},
		{
			name:          "parse version 24.0",
			rpcVersion:    "/Loki:24.0.0/",
			parsedVersion: FlokicoindPre25,
		},
		{
			name:          "parse version 25.0",
			rpcVersion:    "/Loki:25.0.0/",
			parsedVersion: FlokicoindPost25,
		},
		{
			name:          "parse version 25.0 and above",
			rpcVersion:    "/Loki:26.0.0/",
			parsedVersion: FlokicoindPost25,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			version := parseBitcoindVersion(tc.rpcVersion)
			require.Equal(t, tc.parsedVersion, version)
		})
	}
}

// TestParseChainDaemonVersion checks that the correct version from flokicoind's `getinfo`
// RPC call is parsed.
func TestParseChainDaemonVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		rpcVersion    int32
		parsedVersion FlokicoindVersion
	}{
		{
			name:          "parse version 0.24 and below",
			rpcVersion:    230000,
			parsedVersion: FlokicoindPre2401,
		},
		{
			name:          "parse version 0.24.1",
			rpcVersion:    240100,
			parsedVersion: FlokicoindPost2401,
		},
		{
			name:          "parse version 0.24.1 and above",
			rpcVersion:    250000,
			parsedVersion: FlokicoindPost2401,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			version := parseFlokicoindVersion(tc.rpcVersion)
			require.Equal(t, tc.parsedVersion, version)
		})
	}
}

// TestVersionSupports checks all the versions of flokicoind and flokicoind to ensure
// that the RPCs are supported correctly.
func TestVersionSupports(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	// For flokicoind, unified softforks format is supported in 19.0 and
	// above.
	require.False(FlokicoindPre19.SupportUnifiedSoftForks())
	require.True(FlokicoindPre22.SupportUnifiedSoftForks())
	require.True(FlokicoindPre24.SupportUnifiedSoftForks())
	require.True(FlokicoindPre25.SupportUnifiedSoftForks())
	require.True(FlokicoindPost25.SupportUnifiedSoftForks())

	// For flokicoind, `testmempoolaccept` is supported in 22.0 and above.
	require.False(FlokicoindPre19.SupportTestMempoolAccept())
	require.False(FlokicoindPre22.SupportTestMempoolAccept())
	require.True(FlokicoindPre24.SupportTestMempoolAccept())
	require.True(FlokicoindPre25.SupportTestMempoolAccept())
	require.True(FlokicoindPost25.SupportTestMempoolAccept())

	// For flokicoind, `gettxspendingprevout` is supported in 24.0 and above.
	require.False(FlokicoindPre19.SupportGetTxSpendingPrevOut())
	require.False(FlokicoindPre22.SupportGetTxSpendingPrevOut())
	require.False(FlokicoindPre24.SupportGetTxSpendingPrevOut())
	require.True(FlokicoindPre25.SupportGetTxSpendingPrevOut())
	require.True(FlokicoindPost25.SupportGetTxSpendingPrevOut())

	// For flokicoind, unified softforks format is supported in all versions.
	require.True(FlokicoindPre2401.SupportUnifiedSoftForks())
	require.True(FlokicoindPost2401.SupportUnifiedSoftForks())

	// For flokicoind, `testmempoolaccept` is supported in 24.1 and above.
	require.False(FlokicoindPre2401.SupportTestMempoolAccept())
	require.True(FlokicoindPost2401.SupportTestMempoolAccept())

	// For flokicoind, `gettxspendingprevout` is supported in 24.1 and above.
	require.False(FlokicoindPre2401.SupportGetTxSpendingPrevOut())
	require.True(FlokicoindPost2401.SupportGetTxSpendingPrevOut())
}
