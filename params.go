// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/wire"
)

const (
	artReset  = "\033[0m"
	artOrange = "\033[38;2;177;128;10m"
	artBright = "\033[1m"
	artText   = `
 _           _    _      _           _       
| |         | |  (_)    | |         (_)      
| |     ___ | | ___  ___| |__   __ _ _ _ __  
| |    / _ \| |/ / |/ __| '_ \ / _` + "`" + ` | | '_ \ 
| |___| (_) |   <| | (__| | | | (_| | | | | |
|______\___/|_|\_\_|\___|_| |_|\__,_|_|_| |_|
`
)

// activeNetParams is a pointer to the parameters specific to the
// currently active flokicoin network.
var activeNetParams = &mainNetParams

// params is used to group parameters for various networks such as the main
// network and test networks.
type params struct {
	*chaincfg.Params
	rpcPort string
}

// mainNetParams contains parameters specific to the main network
// (wire.MainNet).  NOTE: The RPC port is intentionally different than the
// reference implementation because lokid does not handle wallet requests.  The
// separate wallet process listens on the well-known port and forwards requests
// it does not handle on to lokid.  This approach allows the wallet process
// to emulate the full reference implementation RPC API.
var mainNetParams = params{
	Params:  &chaincfg.MainNetParams,
	rpcPort: "15213",
}

// regressionNetParams contains parameters specific to the regression test
// network (wire.TestNet).  NOTE: The RPC port is intentionally different
// than the reference implementation - see the mainNetParams comment for
// details.
var regressionNetParams = params{
	Params:  &chaincfg.RegressionNetParams,
	rpcPort: "25213",
}

// testNet3Params contains parameters specific to the test network (version 3)
// (wire.TestNet3).  NOTE: The RPC port is intentionally different than the
// reference implementation - see the mainNetParams comment for details.
var testNet3Params = params{
	Params:  &chaincfg.TestNet3Params,
	rpcPort: "35213",
}

// simNetParams contains parameters specific to the simulation test network
// (wire.SimNet).
var simNetParams = params{
	Params:  &chaincfg.SimNetParams,
	rpcPort: "45213",
}

// sigNetParams contains parameters specific to the Signet network
// (wire.SigNet).
var sigNetParams = params{
	Params:  &chaincfg.SigNetParams,
	rpcPort: "55213",
}

// netName returns the name used when referring to a flokicoin network.  At the
// time of writing, lokid currently places blocks for testnet version 3 in the
// data and log directory "testnet", which does not match the Name field of the
// chaincfg parameters.  This function can be used to override this directory
// name as "testnet" when the passed active network matches wire.TestNet3.
//
// A proper upgrade to move the data and log directories for this network to
// "testnet3" is planned for the future, at which point this function can be
// removed and the network parameter's name used instead.
func netName(chainParams *params) string {
	switch chainParams.Net {
	case wire.TestNet3:
		return "testnet"
	default:
		return chainParams.Name
	}
}
