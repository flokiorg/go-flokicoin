// Copyright (c) 2013-2024 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
	"strconv"
	"strings"
)

// XXX pedro: we will probably need to bump this.
const (
	// ProtocolVersion is the latest protocol version this package supports.
	ProtocolVersion uint32 = 70016

	// MultipleAddressVersion is the protocol version which added multiple
	// addresses per message (pver >= MultipleAddressVersion).
	MultipleAddressVersion uint32 = 209

	// NetAddressTimeVersion is the protocol version which added the
	// timestamp field (pver >= NetAddressTimeVersion).
	NetAddressTimeVersion uint32 = 31402

	// BIP0031Version is the protocol version AFTER which a pong message
	// and nonce field in ping were added (pver > BIP0031Version).
	BIP0031Version uint32 = 60000

	// BIP0035Version is the protocol version which added the mempool
	// message (pver >= BIP0035Version).
	BIP0035Version uint32 = 60002

	// BIP0037Version is the protocol version which added new connection
	// bloom filtering related messages and extended the version message
	// with a relay flag (pver >= BIP0037Version).
	BIP0037Version uint32 = 70001

	// RejectVersion is the protocol version which added a new reject
	// message.
	RejectVersion uint32 = 70002

	// BIP0111Version is the protocol version which added the SFNodeBloom
	// service flag.
	BIP0111Version uint32 = 70011

	// SendHeadersVersion is the protocol version which added a new
	// sendheaders message.
	SendHeadersVersion uint32 = 70012

	// FeeFilterVersion is the protocol version which added a new
	// feefilter message.
	FeeFilterVersion uint32 = 70013

	// AddrV2Version is the protocol version which added two new messages.
	// sendaddrv2 is sent during the version-verack handshake and signals
	// support for sending and receiving the addrv2 message. In the future,
	// new messages that occur during the version-verack handshake will not
	// come with a protocol version bump.
	AddrV2Version uint32 = 70016
)

const (
	// NodeNetworkLimitedBlockThreshold is the number of blocks that a node
	// broadcasting SFNodeNetworkLimited MUST be able to serve from the tip.
	NodeNetworkLimitedBlockThreshold = 288
)

// ServiceFlag identifies services supported by a flokicoin peer.
type ServiceFlag uint64

const (
	// SFNodeNetwork is a flag used to indicate a peer is a full node.
	SFNodeNetwork ServiceFlag = 1 << iota

	// SFNodeGetUTXO is a flag used to indicate a peer supports the
	// getutxos and utxos commands (BIP0064).
	SFNodeGetUTXO

	// SFNodeBloom is a flag used to indicate a peer supports bloom
	// filtering.
	SFNodeBloom

	// SFNodeWitness is a flag used to indicate a peer supports blocks
	// and transactions including witness data (BIP0144).
	SFNodeWitness

	// SFNodeXthin is a flag used to indicate a peer supports xthin blocks.
	SFNodeXthin

	// SFNodeBit5 is a flag used to indicate a peer supports a service
	// defined by bit 5.
	SFNodeBit5

	// SFNodeCF is a flag used to indicate a peer supports committed
	// filters (CFs).
	SFNodeCF

	// SFNode2X is a flag used to indicate a peer is running the Segwit2X
	// software.
	SFNode2X

	// SFNodeNetWorkLimited is a flag used to indicate a peer supports serving
	// the last 288 blocks.
	SFNodeNetworkLimited = 1 << 10
)

// Map of service flags back to their constant names for pretty printing.
var sfStrings = map[ServiceFlag]string{
	SFNodeNetwork:        "SFNodeNetwork",
	SFNodeGetUTXO:        "SFNodeGetUTXO",
	SFNodeBloom:          "SFNodeBloom",
	SFNodeWitness:        "SFNodeWitness",
	SFNodeXthin:          "SFNodeXthin",
	SFNodeBit5:           "SFNodeBit5",
	SFNodeCF:             "SFNodeCF",
	SFNode2X:             "SFNode2X",
	SFNodeNetworkLimited: "SFNodeNetworkLimited",
}

// orderedSFStrings is an ordered list of service flags from highest to
// lowest.
var orderedSFStrings = []ServiceFlag{
	SFNodeNetwork,
	SFNodeGetUTXO,
	SFNodeBloom,
	SFNodeWitness,
	SFNodeXthin,
	SFNodeBit5,
	SFNodeCF,
	SFNode2X,
	SFNodeNetworkLimited,
}

// HasFlag returns a bool indicating if the service has the given flag.
func (f ServiceFlag) HasFlag(s ServiceFlag) bool {
	return f&s == s
}

// String returns the ServiceFlag in human-readable form.
func (f ServiceFlag) String() string {
	// No flags are set.
	if f == 0 {
		return "0x0"
	}

	// Add individual bit flags.
	s := ""
	for _, flag := range orderedSFStrings {
		if f&flag == flag {
			s += sfStrings[flag] + "|"
			f -= flag
		}
	}

	// Add any remaining flags which aren't accounted for as hex.
	s = strings.TrimRight(s, "|")
	if f != 0 {
		s += "|0x" + strconv.FormatUint(uint64(f), 16)
	}
	s = strings.TrimLeft(s, "|")
	return s
}

// FlokicoinNet represents which flokicoin network a message belongs to.
type FlokicoinNet uint32

// Constants used to indicate the message flokicoin network.  They can also be
// used to seek to the next message when a stream's state is unknown, but
// this package does not provide that functionality since it's generally a
// better idea to simply disconnect clients that are misbehaving over TCP.
const (
	// MainNet represents the main flokicoin network.
	MainNet FlokicoinNet = 0xd9b4bef9

	// TestNet represents the regression test network.
	TestNet FlokicoinNet = 0xdab5bffa

	// TestNet3 represents the test network (version 3).
	TestNet3 FlokicoinNet = 0x0709110b

	// SimNet represents the simulation test network.
	SimNet FlokicoinNet = 0x12141c16

	// TestNet4 represents the test network (version 4).
	TestNet4 FlokicoinNet = 0x283f161c
)

// bnStrings is a map of flokicoin networks back to their constant names for
// pretty printing.
var bnStrings = map[FlokicoinNet]string{
	MainNet:  "MainNet",
	TestNet:  "TestNet",
	TestNet3: "TestNet3",
	SimNet:   "SimNet",
	TestNet4: "TestNet4",
}

// String returns the FlokicoinNet in human-readable form.
func (n FlokicoinNet) String() string {
	if s, ok := bnStrings[n]; ok {
		return s
	}

	return fmt.Sprintf("Unknown FlokicoinNet (%d)", uint32(n))
}
