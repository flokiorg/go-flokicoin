package main

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/connmgr"
	"github.com/flokiorg/go-flokicoin/netaddr"
	"github.com/flokiorg/go-flokicoin/peer"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) (*server, func()) {
	t.Helper()

	connManager, err := connmgr.New(&connmgr.Config{
		Dial:           func(net.Addr) (net.Conn, error) { return nil, errors.New("disabled") },
		TargetOutbound: 0,
	})
	require.NoError(t, err)

	connManager.Start()

	s := &server{
		connManager: connManager,
	}

	cleanup := func() {
		connManager.Stop()
		connManager.Wait()
	}

	return s, cleanup
}

func newOutboundServerPeer(t *testing.T, s *server, persistent bool) *serverPeer {
	t.Helper()

	cfg := &peer.Config{
		UserAgentName:    "flokicoind",
		UserAgentVersion: "0.0.1",
		ChainParams:      &chaincfg.MainNetParams,
		AllowSelfConns:   true,
	}

	p, err := peer.NewOutboundPeer(cfg, "127.0.0.1:35212")
	require.NoError(t, err)

	sp := newServerPeer(s, persistent)
	sp.Peer = p
	sp.connReq = &connmgr.ConnReq{}

	return sp
}

func newInboundServerPeer(t *testing.T, s *server) *serverPeer {
	t.Helper()

	cfg := &peer.Config{
		UserAgentName:    "flokicoind",
		UserAgentVersion: "0.0.1",
		ChainParams:      &chaincfg.MainNetParams,
		AllowSelfConns:   true,
	}

	p := peer.NewInboundPeer(cfg)

	sp := newServerPeer(s, false)
	sp.Peer = p

	return sp
}

func newPeerState() *peerState {
	return &peerState{
		inboundPeers:    make(map[int32]*serverPeer),
		outboundPeers:   make(map[int32]*serverPeer),
		persistentPeers: make(map[int32]*serverPeer),
		banned:          make(map[string]time.Time),
		outboundGroups:  make(map[string]int),
	}
}

// TestHandleDonePeerMsgGroupAccounting exercises the peer tear-down paths for
// a variety of scenarios to ensure outbound group bookkeeping remains
// consistent.
func TestHandleDonePeerMsgGroupAccounting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		makePeer func(*testing.T, *server) *serverPeer
		setup    func(*testing.T, *serverPeer, *peerState) string
		check    func(*testing.T, *serverPeer, *peerState, string)
	}{
		{
			name: "outbound single slot released on disconnect",
			makePeer: func(t *testing.T, s *server) *serverPeer {
				return newOutboundServerPeer(t, s, false)
			},
			setup: func(t *testing.T, sp *serverPeer, state *peerState) string {
				key := netaddr.GroupKey(sp.NA())
				sp.groupKey = key
				sp.groupCounted = true
				state.outboundPeers[sp.ID()] = sp
				state.outboundGroups[key] = 1
				return key
			},
			check: func(t *testing.T, sp *serverPeer, state *peerState, key string) {
				require.NotContains(t, state.outboundGroups, key)
				require.NotContains(t, state.outboundPeers, sp.ID())
				require.False(t, sp.groupCounted)
				require.Empty(t, sp.groupKey)
			},
		},
		{
			name: "outbound shared group decremented",
			makePeer: func(t *testing.T, s *server) *serverPeer {
				return newOutboundServerPeer(t, s, false)
			},
			setup: func(t *testing.T, sp *serverPeer, state *peerState) string {
				key := netaddr.GroupKey(sp.NA())
				sp.groupKey = key
				sp.groupCounted = true
				state.outboundPeers[sp.ID()] = sp
				state.outboundGroups[key] = 2
				return key
			},
			check: func(t *testing.T, sp *serverPeer, state *peerState, key string) {
				require.Contains(t, state.outboundGroups, key)
				require.Equal(t, 1, state.outboundGroups[key])
				require.NotContains(t, state.outboundPeers, sp.ID())
				require.False(t, sp.groupCounted)
				require.Empty(t, sp.groupKey)
			},
		},
		{
			name: "outbound not counted leaves group untouched",
			makePeer: func(t *testing.T, s *server) *serverPeer {
				return newOutboundServerPeer(t, s, false)
			},
			setup: func(t *testing.T, sp *serverPeer, state *peerState) string {
				key := netaddr.GroupKey(sp.NA())
				sp.groupKey = key
				sp.groupCounted = false
				state.outboundPeers[sp.ID()] = sp
				state.outboundGroups[key] = 3
				return key
			},
			check: func(t *testing.T, sp *serverPeer, state *peerState, key string) {
				require.Contains(t, state.outboundGroups, key)
				require.Equal(t, 3, state.outboundGroups[key])
				require.NotContains(t, state.outboundPeers, sp.ID())
				require.False(t, sp.groupCounted)
				require.Equal(t, key, sp.groupKey)
			},
		},
		{
			name: "inbound peer disconnect does not touch outbound groups",
			makePeer: func(t *testing.T, s *server) *serverPeer {
				return newInboundServerPeer(t, s)
			},
			setup: func(t *testing.T, sp *serverPeer, state *peerState) string {
				state.inboundPeers[sp.ID()] = sp
				state.outboundGroups["other"] = 5
				return "other"
			},
			check: func(t *testing.T, sp *serverPeer, state *peerState, key string) {
				require.Contains(t, state.outboundGroups, key)
				require.Equal(t, 5, state.outboundGroups[key])
				require.NotContains(t, state.inboundPeers, sp.ID())
			},
		},
		{
			name: "persistent outbound releases group slot",
			makePeer: func(t *testing.T, s *server) *serverPeer {
				return newOutboundServerPeer(t, s, true)
			},
			setup: func(t *testing.T, sp *serverPeer, state *peerState) string {
				key := netaddr.GroupKey(sp.NA())
				sp.groupKey = key
				sp.groupCounted = true
				state.persistentPeers[sp.ID()] = sp
				state.outboundGroups[key] = 1
				return key
			},
			check: func(t *testing.T, sp *serverPeer, state *peerState, key string) {
				require.NotContains(t, state.outboundGroups, key)
				require.NotContains(t, state.persistentPeers, sp.ID())
				require.False(t, sp.groupCounted)
				require.Empty(t, sp.groupKey)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s, cleanup := newTestServer(t)
			defer cleanup()

			state := newPeerState()

			sp := tc.makePeer(t, s)
			key := tc.setup(t, sp, state)

			s.handleDonePeerMsg(state, sp)

			tc.check(t, sp, state, key)
		})
	}
}
