package rpcclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParseLokidVersion checks that the correct version from lokid's
// `getnetworkinfo` RPC call is parsed.
func TestParseLokidVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		rpcVersion    string
		parsedVersion BitcoindVersion
	}{
		{
			name:          "parse version 0.19 and below",
			rpcVersion:    "/Loki:0.18.0/",
			parsedVersion: LokidPre19,
		},
		{
			name:          "parse version 0.19",
			rpcVersion:    "/Loki:0.19.0/",
			parsedVersion: LokidPre22,
		},
		{
			name:          "parse version 0.19 - 22.0",
			rpcVersion:    "/Loki:0.20.1/",
			parsedVersion: LokidPre22,
		},
		{
			name:          "parse version 22.0",
			rpcVersion:    "/Loki:22.0.0/",
			parsedVersion: LokidPre24,
		},
		{
			name:          "parse version 22.0 - 24.0",
			rpcVersion:    "/Loki:23.1.0/",
			parsedVersion: LokidPre24,
		},
		{
			name:          "parse version 24.0",
			rpcVersion:    "/Loki:24.0.0/",
			parsedVersion: LokidPre25,
		},
		{
			name:          "parse version 25.0",
			rpcVersion:    "/Loki:25.0.0/",
			parsedVersion: LokidPost25,
		},
		{
			name:          "parse version 25.0 and above",
			rpcVersion:    "/Loki:26.0.0/",
			parsedVersion: LokidPost25,
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

// TestParseChainDaemonVersion checks that the correct version from lokid's `getinfo`
// RPC call is parsed.
func TestParseChainDaemonVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		rpcVersion    int32
		parsedVersion LokidVersion
	}{
		{
			name:          "parse version 0.24 and below",
			rpcVersion:    230000,
			parsedVersion: LokidPre2401,
		},
		{
			name:          "parse version 0.24.1",
			rpcVersion:    240100,
			parsedVersion: LokidPost2401,
		},
		{
			name:          "parse version 0.24.1 and above",
			rpcVersion:    250000,
			parsedVersion: LokidPost2401,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			version := parseLokidVersion(tc.rpcVersion)
			require.Equal(t, tc.parsedVersion, version)
		})
	}
}

// TestVersionSupports checks all the versions of lokid and lokid to ensure
// that the RPCs are supported correctly.
func TestVersionSupports(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	// For lokid, unified softforks format is supported in 19.0 and
	// above.
	require.False(LokidPre19.SupportUnifiedSoftForks())
	require.True(LokidPre22.SupportUnifiedSoftForks())
	require.True(LokidPre24.SupportUnifiedSoftForks())
	require.True(LokidPre25.SupportUnifiedSoftForks())
	require.True(LokidPost25.SupportUnifiedSoftForks())

	// For lokid, `testmempoolaccept` is supported in 22.0 and above.
	require.False(LokidPre19.SupportTestMempoolAccept())
	require.False(LokidPre22.SupportTestMempoolAccept())
	require.True(LokidPre24.SupportTestMempoolAccept())
	require.True(LokidPre25.SupportTestMempoolAccept())
	require.True(LokidPost25.SupportTestMempoolAccept())

	// For lokid, `gettxspendingprevout` is supported in 24.0 and above.
	require.False(LokidPre19.SupportGetTxSpendingPrevOut())
	require.False(LokidPre22.SupportGetTxSpendingPrevOut())
	require.False(LokidPre24.SupportGetTxSpendingPrevOut())
	require.True(LokidPre25.SupportGetTxSpendingPrevOut())
	require.True(LokidPost25.SupportGetTxSpendingPrevOut())

	// For lokid, unified softforks format is supported in all versions.
	require.True(LokidPre2401.SupportUnifiedSoftForks())
	require.True(LokidPost2401.SupportUnifiedSoftForks())

	// For lokid, `testmempoolaccept` is supported in 24.1 and above.
	require.False(LokidPre2401.SupportTestMempoolAccept())
	require.True(LokidPost2401.SupportTestMempoolAccept())

	// For lokid, `gettxspendingprevout` is supported in 24.1 and above.
	require.False(LokidPre2401.SupportGetTxSpendingPrevOut())
	require.True(LokidPost2401.SupportGetTxSpendingPrevOut())
}
