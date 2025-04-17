// Copyright (c) 2023 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// This file is ignored during the regular tests due to the following build tag.
//go:build rpctest
// +build rpctest

package integration

import (
	"testing"

	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/integration/rpctest"
	"github.com/stretchr/testify/require"
)

func TestPrune(t *testing.T) {
	t.Parallel()

	// Boilerplate code to make a pruned node.
	flokicoindCfg := []string{"--prune=1536"}
	r, err := rpctest.New(&chaincfg.SimNetParams, nil, flokicoindCfg, "")
	require.NoError(t, err)

	if err := r.SetUp(false, 0); err != nil {
		require.NoError(t, err)
	}
	t.Cleanup(func() { r.TearDown() })

	// Check that the rpc call for block chain info comes back correctly.
	chainInfo, err := r.Client.GetBlockChainInfo()
	require.NoError(t, err)

	if !chainInfo.Pruned {
		t.Fatalf("expected the node to be pruned but the pruned "+
			"boolean was %v", chainInfo.Pruned)
	}
}
