// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainutil_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/chainutil/psbt"
	"github.com/flokiorg/go-flokicoin/crypto"
	"github.com/flokiorg/go-flokicoin/crypto/schnorr"
	"github.com/flokiorg/go-flokicoin/txscript"
	"github.com/flokiorg/go-flokicoin/wire"
	"golang.org/x/crypto/ripemd160"
)

type CustomParamStruct struct {
	Net              wire.FlokicoinNet
	PubKeyHashAddrID byte
	ScriptHashAddrID byte
	Bech32HRPSegwit  string
}

var CustomParams = CustomParamStruct{
	Net:              0xdbb6c0fb, // litecoin mainnet HD version bytes
	PubKeyHashAddrID: 0x30,       // starts with L
	ScriptHashAddrID: 0x32,       // starts with M
	Bech32HRPSegwit:  "ltc",      // starts with ltc
}

// We use this function to be able to test functionality in DecodeAddress for
// defaultNet addresses
func applyCustomParams(params chaincfg.Params, customParams CustomParamStruct) chaincfg.Params {
	params.Net = customParams.Net
	params.PubKeyHashAddrID = customParams.PubKeyHashAddrID
	params.ScriptHashAddrID = customParams.ScriptHashAddrID
	params.Bech32HRPSegwit = customParams.Bech32HRPSegwit
	return params
}

var customParams = applyCustomParams(chaincfg.MainNetParams, CustomParams)

func TestAddresses(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		encoded string
		valid   bool
		result  chainutil.Address
		f       func() (chainutil.Address, error)
		net     *chaincfg.Params
	}{
		// Positive P2PKH tests.
		{
			name:    "mainnet p2pkh",
			addr:    "FNzaa51JKQ7EymKxam9E3PNNxCjgsjGcQi",
			encoded: "FNzaa51JKQ7EymKxam9E3PNNxCjgsjGcQi",
			valid:   true,
			result: chainutil.TstAddressPubKeyHash(
				[ripemd160.Size]byte{
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xc5, 0x4c, 0xe7, 0xd2, 0xa4, 0x91, 0xbb, 0x4a, 0x0e, 0x84},
				chaincfg.MainNetParams.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				pkHash := []byte{
					0xe3, 0x4c, 0xce, 0x70, 0xc8, 0x63, 0x73, 0x27, 0x3e, 0xfc,
					0xc5, 0x4c, 0xe7, 0xd2, 0xa4, 0x91, 0xbb, 0x4a, 0x0e, 0x84}
				return chainutil.NewAddressPubKeyHash(pkHash, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			name:    "mainnet p2pkh 2",
			addr:    "FNjHEib9NH7JvvbcHhG8xXa2vd19aQZhSW",
			encoded: "FNjHEib9NH7JvvbcHhG8xXa2vd19aQZhSW",
			valid:   true,
			result: chainutil.TstAddressPubKeyHash(
				[ripemd160.Size]byte{
					0x0e, 0xf0, 0x30, 0x10, 0x7f, 0xd2, 0x6e, 0x0b, 0x6b, 0xf4,
					0x05, 0x12, 0xbc, 0xa2, 0xce, 0xb1, 0xdd, 0x80, 0xad, 0xaa},
				chaincfg.MainNetParams.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				pkHash := []byte{
					0x0e, 0xf0, 0x30, 0x10, 0x7f, 0xd2, 0x6e, 0x0b, 0x6b, 0xf4,
					0x05, 0x12, 0xbc, 0xa2, 0xce, 0xb1, 0xdd, 0x80, 0xad, 0xaa}
				return chainutil.NewAddressPubKeyHash(pkHash, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			name:    "litecoin mainnet p2pkh",
			addr:    "LM2WMpR1Rp6j3Sa59cMXMs1SPzj9eXpGc1",
			encoded: "LM2WMpR1Rp6j3Sa59cMXMs1SPzj9eXpGc1",
			valid:   true,
			result: chainutil.TstAddressPubKeyHash(
				[ripemd160.Size]byte{
					0x13, 0xc6, 0x0d, 0x8e, 0x68, 0xd7, 0x34, 0x9f, 0x5b, 0x4c,
					0xa3, 0x62, 0xc3, 0x95, 0x4b, 0x15, 0x04, 0x50, 0x61, 0xb1},
				CustomParams.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				pkHash := []byte{
					0x13, 0xc6, 0x0d, 0x8e, 0x68, 0xd7, 0x34, 0x9f, 0x5b, 0x4c,
					0xa3, 0x62, 0xc3, 0x95, 0x4b, 0x15, 0x04, 0x50, 0x61, 0xb1}
				return chainutil.NewAddressPubKeyHash(pkHash, &customParams)
			},
			net: &customParams,
		},
		{
			name:    "testnet p2pkh",
			addr:    "mrX9vMRYLfVy1BnZbc5gZjuyaqH3ZW2ZHz",
			encoded: "mrX9vMRYLfVy1BnZbc5gZjuyaqH3ZW2ZHz",
			valid:   true,
			result: chainutil.TstAddressPubKeyHash(
				[ripemd160.Size]byte{
					0x78, 0xb3, 0x16, 0xa0, 0x86, 0x47, 0xd5, 0xb7, 0x72, 0x83,
					0xe5, 0x12, 0xd3, 0x60, 0x3f, 0x1f, 0x1c, 0x8d, 0xe6, 0x8f},
				chaincfg.TestNet3Params.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				pkHash := []byte{
					0x78, 0xb3, 0x16, 0xa0, 0x86, 0x47, 0xd5, 0xb7, 0x72, 0x83,
					0xe5, 0x12, 0xd3, 0x60, 0x3f, 0x1f, 0x1c, 0x8d, 0xe6, 0x8f}
				return chainutil.NewAddressPubKeyHash(pkHash, &chaincfg.TestNet3Params)
			},
			net: &chaincfg.TestNet3Params,
		},

		// Negative P2PKH tests.
		{
			name:  "p2pkh wrong hash length",
			addr:  "",
			valid: false,
			f: func() (chainutil.Address, error) {
				pkHash := []byte{
					0x00, 0x0e, 0xf0, 0x30, 0x10, 0x7f, 0xd2, 0x6e, 0x0b, 0x6b,
					0xf4, 0x05, 0x12, 0xbc, 0xa2, 0xce, 0xb1, 0xdd, 0x80, 0xad,
					0xaa}
				return chainutil.NewAddressPubKeyHash(pkHash, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			name:  "p2pkh bad checksum",
			addr:  "1MirQ9bwyQcGVJPwKUgapu5ouK2E2Ey4gY",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},

		// Positive P2SH tests.
		{
			// Taken from transactions:
			// output: 3c9018e8d5615c306d72397f8f5eef44308c98fb576a88e030c25456b4f3a7ac
			// input:  837dea37ddc8b1e3ce646f1a656e79bbd8cc7f558ac56a169626d649ebe2a3ba.
			name:    "mainnet p2sh",
			addr:    "3QJmV3qfvL9SuYo34YihAf3sRCW3qSinyC",
			encoded: "3QJmV3qfvL9SuYo34YihAf3sRCW3qSinyC",
			valid:   true,
			result: chainutil.TstAddressScriptHash(
				[ripemd160.Size]byte{
					0xf8, 0x15, 0xb0, 0x36, 0xd9, 0xbb, 0xbc, 0xe5, 0xe9, 0xf2,
					0xa0, 0x0a, 0xbd, 0x1b, 0xf3, 0xdc, 0x91, 0xe9, 0x55, 0x10},
				chaincfg.MainNetParams.ScriptHashAddrID),
			f: func() (chainutil.Address, error) {
				script := []byte{
					0x52, 0x41, 0x04, 0x91, 0xbb, 0xa2, 0x51, 0x09, 0x12, 0xa5,
					0xbd, 0x37, 0xda, 0x1f, 0xb5, 0xb1, 0x67, 0x30, 0x10, 0xe4,
					0x3d, 0x2c, 0x6d, 0x81, 0x2c, 0x51, 0x4e, 0x91, 0xbf, 0xa9,
					0xf2, 0xeb, 0x12, 0x9e, 0x1c, 0x18, 0x33, 0x29, 0xdb, 0x55,
					0xbd, 0x86, 0x8e, 0x20, 0x9a, 0xac, 0x2f, 0xbc, 0x02, 0xcb,
					0x33, 0xd9, 0x8f, 0xe7, 0x4b, 0xf2, 0x3f, 0x0c, 0x23, 0x5d,
					0x61, 0x26, 0xb1, 0xd8, 0x33, 0x4f, 0x86, 0x41, 0x04, 0x86,
					0x5c, 0x40, 0x29, 0x3a, 0x68, 0x0c, 0xb9, 0xc0, 0x20, 0xe7,
					0xb1, 0xe1, 0x06, 0xd8, 0xc1, 0x91, 0x6d, 0x3c, 0xef, 0x99,
					0xaa, 0x43, 0x1a, 0x56, 0xd2, 0x53, 0xe6, 0x92, 0x56, 0xda,
					0xc0, 0x9e, 0xf1, 0x22, 0xb1, 0xa9, 0x86, 0x81, 0x8a, 0x7c,
					0xb6, 0x24, 0x53, 0x2f, 0x06, 0x2c, 0x1d, 0x1f, 0x87, 0x22,
					0x08, 0x48, 0x61, 0xc5, 0xc3, 0x29, 0x1c, 0xcf, 0xfe, 0xf4,
					0xec, 0x68, 0x74, 0x41, 0x04, 0x8d, 0x24, 0x55, 0xd2, 0x40,
					0x3e, 0x08, 0x70, 0x8f, 0xc1, 0xf5, 0x56, 0x00, 0x2f, 0x1b,
					0x6c, 0xd8, 0x3f, 0x99, 0x2d, 0x08, 0x50, 0x97, 0xf9, 0x97,
					0x4a, 0xb0, 0x8a, 0x28, 0x83, 0x8f, 0x07, 0x89, 0x6f, 0xba,
					0xb0, 0x8f, 0x39, 0x49, 0x5e, 0x15, 0xfa, 0x6f, 0xad, 0x6e,
					0xdb, 0xfb, 0x1e, 0x75, 0x4e, 0x35, 0xfa, 0x1c, 0x78, 0x44,
					0xc4, 0x1f, 0x32, 0x2a, 0x18, 0x63, 0xd4, 0x62, 0x13, 0x53,
					0xae}
				return chainutil.NewAddressScriptHash(script, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			name:    "litecoin mainnet P2SH ",
			addr:    "MVcg9uEvtWuP5N6V48EHfEtbz48qR8TKZ9",
			encoded: "MVcg9uEvtWuP5N6V48EHfEtbz48qR8TKZ9",
			valid:   true,
			result: chainutil.TstAddressScriptHash(
				[ripemd160.Size]byte{
					0xee, 0x34, 0xac, 0x67, 0x6b, 0xda, 0xf6, 0xe3, 0x70, 0xc8,
					0xc8, 0x20, 0xb9, 0x48, 0xed, 0xfa, 0xd3, 0xa8, 0x73, 0xd8},
				CustomParams.ScriptHashAddrID),
			f: func() (chainutil.Address, error) {
				pkHash := []byte{
					0xEE, 0x34, 0xAC, 0x67, 0x6B, 0xDA, 0xF6, 0xE3, 0x70, 0xC8,
					0xC8, 0x20, 0xB9, 0x48, 0xED, 0xFA, 0xD3, 0xA8, 0x73, 0xD8}
				return chainutil.NewAddressScriptHashFromHash(pkHash, &customParams)
			},
			net: &customParams,
		},
		{
			// Taken from transactions:
			// output: b0539a45de13b3e0403909b8bd1a555b8cbe45fd4e3f3fda76f3a5f52835c29d
			// input: (not yet redeemed at time test was written)
			name:    "mainnet p2sh 2",
			addr:    "3NukJ6fYZJ5Kk8bPjycAnruZkE5Q7UW7i8",
			encoded: "3NukJ6fYZJ5Kk8bPjycAnruZkE5Q7UW7i8",
			valid:   true,
			result: chainutil.TstAddressScriptHash(
				[ripemd160.Size]byte{
					0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37,
					0xc0, 0x51, 0x99, 0x29, 0x01, 0x9e, 0xf8, 0x6e, 0xb5, 0xb4},
				chaincfg.MainNetParams.ScriptHashAddrID),
			f: func() (chainutil.Address, error) {
				hash := []byte{
					0xe8, 0xc3, 0x00, 0xc8, 0x79, 0x86, 0xef, 0xa8, 0x4c, 0x37,
					0xc0, 0x51, 0x99, 0x29, 0x01, 0x9e, 0xf8, 0x6e, 0xb5, 0xb4}
				return chainutil.NewAddressScriptHashFromHash(hash, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			// Taken from lokid base58_keys_valid.
			name:    "testnet p2sh",
			addr:    "2NBFNJTktNa7GZusGbDbGKRZTxdK9VVez3n",
			encoded: "2NBFNJTktNa7GZusGbDbGKRZTxdK9VVez3n",
			valid:   true,
			result: chainutil.TstAddressScriptHash(
				[ripemd160.Size]byte{
					0xc5, 0x79, 0x34, 0x2c, 0x2c, 0x4c, 0x92, 0x20, 0x20, 0x5e,
					0x2c, 0xdc, 0x28, 0x56, 0x17, 0x04, 0x0c, 0x92, 0x4a, 0x0a},
				chaincfg.TestNet3Params.ScriptHashAddrID),
			f: func() (chainutil.Address, error) {
				hash := []byte{
					0xc5, 0x79, 0x34, 0x2c, 0x2c, 0x4c, 0x92, 0x20, 0x20, 0x5e,
					0x2c, 0xdc, 0x28, 0x56, 0x17, 0x04, 0x0c, 0x92, 0x4a, 0x0a}
				return chainutil.NewAddressScriptHashFromHash(hash, &chaincfg.TestNet3Params)
			},
			net: &chaincfg.TestNet3Params,
		},

		// Negative P2SH tests.
		{
			name:  "p2sh wrong hash length",
			addr:  "",
			valid: false,
			f: func() (chainutil.Address, error) {
				hash := []byte{
					0x00, 0xf8, 0x15, 0xb0, 0x36, 0xd9, 0xbb, 0xbc, 0xe5, 0xe9,
					0xf2, 0xa0, 0x0a, 0xbd, 0x1b, 0xf3, 0xdc, 0x91, 0xe9, 0x55,
					0x10}
				return chainutil.NewAddressScriptHashFromHash(hash, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},

		// Positive P2PK tests.
		{
			name:    "mainnet p2pk compressed (0x02)",
			addr:    "03772d344270fbe13f5d55fd79a6f4500d3b98b14d54bf39d97befd18abeb6ba8c",
			encoded: "FUVymGvCu6Ycg8mQQZzEhcNJ3qQjHMyFMr",
			valid:   true,
			result: chainutil.TstAddressPubKey(
				[]byte{
					0x02, 0x19, 0x2d, 0x74, 0xd0, 0xcb, 0x94, 0x34, 0x4c, 0x95,
					0x69, 0xc2, 0xe7, 0x79, 0x01, 0x57, 0x3d, 0x8d, 0x79, 0x03,
					0xc3, 0xeb, 0xec, 0x3a, 0x95, 0x77, 0x24, 0x89, 0x5d, 0xca,
					0x52, 0xc6, 0xb4},
				chainutil.PKFCompressed, chaincfg.MainNetParams.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				serializedPubKey := []byte{
					0x02, 0x19, 0x2d, 0x74, 0xd0, 0xcb, 0x94, 0x34, 0x4c, 0x95,
					0x69, 0xc2, 0xe7, 0x79, 0x01, 0x57, 0x3d, 0x8d, 0x79, 0x03,
					0xc3, 0xeb, 0xec, 0x3a, 0x95, 0x77, 0x24, 0x89, 0x5d, 0xca,
					0x52, 0xc6, 0xb4}
				return chainutil.NewAddressPubKey(serializedPubKey, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			name:    "mainnet p2pk compressed (0x03)",
			addr:    "021ed307977c099f351d5e5c7e5b7370c21de678503c54ee116c018be162851e7f",
			encoded: "FAaTtV6xSBm3Ls5G4HFnLtUE3kvRL39B72",
			valid:   true,
			result: chainutil.TstAddressPubKey(
				[]byte{
					0x03, 0xb0, 0xbd, 0x63, 0x42, 0x34, 0xab, 0xbb, 0x1b, 0xa1,
					0xe9, 0x86, 0xe8, 0x84, 0x18, 0x5c, 0x61, 0xcf, 0x43, 0xe0,
					0x01, 0xf9, 0x13, 0x7f, 0x23, 0xc2, 0xc4, 0x09, 0x27, 0x3e,
					0xb1, 0x6e, 0x65},
				chainutil.PKFCompressed, chaincfg.MainNetParams.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				serializedPubKey := []byte{
					0x03, 0xb0, 0xbd, 0x63, 0x42, 0x34, 0xab, 0xbb, 0x1b, 0xa1,
					0xe9, 0x86, 0xe8, 0x84, 0x18, 0x5c, 0x61, 0xcf, 0x43, 0xe0,
					0x01, 0xf9, 0x13, 0x7f, 0x23, 0xc2, 0xc4, 0x09, 0x27, 0x3e,
					0xb1, 0x6e, 0x65}
				return chainutil.NewAddressPubKey(serializedPubKey, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			name:    "mainnet p2pk uncompressed (0x04)",
			addr:    "046c970280769cdc7e74fd79ea47363b4a8b2beff66d260db9c5999dfe8868a6f328ccdf751ec4c5db861de26246ad507f10700c469cedcd64b4c676032009291d",
			encoded: "FNdFoUj8voYMgvc4Koy26qyHMtqSLRGjis",
			valid:   true,
			result: chainutil.TstAddressPubKey(
				[]byte{
					0x04, 0xbf, 0x7c, 0xdb, 0xac, 0x0f, 0xa8, 0x71,
					0x2c, 0x6e, 0xd8, 0x54, 0x03, 0xac, 0x22, 0x15,
					0x8d, 0xcf, 0x85, 0xee, 0xe5, 0xfd, 0x49, 0xad,
					0xae, 0xc5, 0x1a, 0xd0, 0x77, 0xb3, 0x0f, 0x4d,
					0x76, 0x68, 0x6a, 0x96, 0x33, 0x59, 0x97, 0x76,
					0xff, 0x1a, 0xa5, 0x35, 0x05, 0xa5, 0x27, 0x82,
					0x26, 0xb5, 0x66, 0xaa, 0x97, 0x88, 0x7a, 0x7f,
					0x7e, 0x64, 0xdd, 0xf7, 0xbf, 0x98, 0x4b, 0xec,
					0xd9,
				},
				chainutil.PKFUncompressed, chaincfg.MainNetParams.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				serializedPubKey := []byte{
					0x04, 0xbf, 0x7c, 0xdb, 0xac, 0x0f, 0xa8, 0x71,
					0x2c, 0x6e, 0xd8, 0x54, 0x03, 0xac, 0x22, 0x15,
					0x8d, 0xcf, 0x85, 0xee, 0xe5, 0xfd, 0x49, 0xad,
					0xae, 0xc5, 0x1a, 0xd0, 0x77, 0xb3, 0x0f, 0x4d,
					0x76, 0x68, 0x6a, 0x96, 0x33, 0x59, 0x97, 0x76,
					0xff, 0x1a, 0xa5, 0x35, 0x05, 0xa5, 0x27, 0x82,
					0x26, 0xb5, 0x66, 0xaa, 0x97, 0x88, 0x7a, 0x7f,
					0x7e, 0x64, 0xdd, 0xf7, 0xbf, 0x98, 0x4b, 0xec,
					0xd9,
				}
				return chainutil.NewAddressPubKey(serializedPubKey, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			name:    "testnet p2pk compressed (0x02)",
			addr:    "02192d74d0cb94344c9569c2e77901573d8d7903c3ebec3a957724895dca52c6b4",
			encoded: "mhiDPVP2nJunaAgTjzWSHCYfAqxxrxzjmo",
			valid:   true,
			result: chainutil.TstAddressPubKey(
				[]byte{
					0x02, 0x19, 0x2d, 0x74, 0xd0, 0xcb, 0x94, 0x34, 0x4c, 0x95,
					0x69, 0xc2, 0xe7, 0x79, 0x01, 0x57, 0x3d, 0x8d, 0x79, 0x03,
					0xc3, 0xeb, 0xec, 0x3a, 0x95, 0x77, 0x24, 0x89, 0x5d, 0xca,
					0x52, 0xc6, 0xb4},
				chainutil.PKFCompressed, chaincfg.TestNet3Params.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				serializedPubKey := []byte{
					0x02, 0x19, 0x2d, 0x74, 0xd0, 0xcb, 0x94, 0x34, 0x4c, 0x95,
					0x69, 0xc2, 0xe7, 0x79, 0x01, 0x57, 0x3d, 0x8d, 0x79, 0x03,
					0xc3, 0xeb, 0xec, 0x3a, 0x95, 0x77, 0x24, 0x89, 0x5d, 0xca,
					0x52, 0xc6, 0xb4}
				return chainutil.NewAddressPubKey(serializedPubKey, &chaincfg.TestNet3Params)
			},
			net: &chaincfg.TestNet3Params,
		},
		{
			name:    "testnet p2pk compressed (0x03)",
			addr:    "03b0bd634234abbb1ba1e986e884185c61cf43e001f9137f23c2c409273eb16e65",
			encoded: "mkPETRTSzU8MZLHkFKBmbKppxmdw9qT42t",
			valid:   true,
			result: chainutil.TstAddressPubKey(
				[]byte{
					0x03, 0xb0, 0xbd, 0x63, 0x42, 0x34, 0xab, 0xbb, 0x1b, 0xa1,
					0xe9, 0x86, 0xe8, 0x84, 0x18, 0x5c, 0x61, 0xcf, 0x43, 0xe0,
					0x01, 0xf9, 0x13, 0x7f, 0x23, 0xc2, 0xc4, 0x09, 0x27, 0x3e,
					0xb1, 0x6e, 0x65},
				chainutil.PKFCompressed, chaincfg.TestNet3Params.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				serializedPubKey := []byte{
					0x03, 0xb0, 0xbd, 0x63, 0x42, 0x34, 0xab, 0xbb, 0x1b, 0xa1,
					0xe9, 0x86, 0xe8, 0x84, 0x18, 0x5c, 0x61, 0xcf, 0x43, 0xe0,
					0x01, 0xf9, 0x13, 0x7f, 0x23, 0xc2, 0xc4, 0x09, 0x27, 0x3e,
					0xb1, 0x6e, 0x65}
				return chainutil.NewAddressPubKey(serializedPubKey, &chaincfg.TestNet3Params)
			},
			net: &chaincfg.TestNet3Params,
		},
		{
			name: "testnet p2pk uncompressed (0x04)",
			addr: "0411db93e1dcdb8a016b49840f8c53bc1eb68a382e97b1482ecad7b148a6909a5" +
				"cb2e0eaddfb84ccf9744464f82e160bfa9b8b64f9d4c03f999b8643f656b412a3",
			encoded: "mh8YhPYEAYs3E7EVyKtB5xrcfMExkkdEMF",
			valid:   true,
			result: chainutil.TstAddressPubKey(
				[]byte{
					0x04, 0x11, 0xdb, 0x93, 0xe1, 0xdc, 0xdb, 0x8a, 0x01, 0x6b,
					0x49, 0x84, 0x0f, 0x8c, 0x53, 0xbc, 0x1e, 0xb6, 0x8a, 0x38,
					0x2e, 0x97, 0xb1, 0x48, 0x2e, 0xca, 0xd7, 0xb1, 0x48, 0xa6,
					0x90, 0x9a, 0x5c, 0xb2, 0xe0, 0xea, 0xdd, 0xfb, 0x84, 0xcc,
					0xf9, 0x74, 0x44, 0x64, 0xf8, 0x2e, 0x16, 0x0b, 0xfa, 0x9b,
					0x8b, 0x64, 0xf9, 0xd4, 0xc0, 0x3f, 0x99, 0x9b, 0x86, 0x43,
					0xf6, 0x56, 0xb4, 0x12, 0xa3},
				chainutil.PKFUncompressed, chaincfg.TestNet3Params.PubKeyHashAddrID),
			f: func() (chainutil.Address, error) {
				serializedPubKey := []byte{
					0x04, 0x11, 0xdb, 0x93, 0xe1, 0xdc, 0xdb, 0x8a, 0x01, 0x6b,
					0x49, 0x84, 0x0f, 0x8c, 0x53, 0xbc, 0x1e, 0xb6, 0x8a, 0x38,
					0x2e, 0x97, 0xb1, 0x48, 0x2e, 0xca, 0xd7, 0xb1, 0x48, 0xa6,
					0x90, 0x9a, 0x5c, 0xb2, 0xe0, 0xea, 0xdd, 0xfb, 0x84, 0xcc,
					0xf9, 0x74, 0x44, 0x64, 0xf8, 0x2e, 0x16, 0x0b, 0xfa, 0x9b,
					0x8b, 0x64, 0xf9, 0xd4, 0xc0, 0x3f, 0x99, 0x9b, 0x86, 0x43,
					0xf6, 0x56, 0xb4, 0x12, 0xa3}
				return chainutil.NewAddressPubKey(serializedPubKey, &chaincfg.TestNet3Params)
			},
			net: &chaincfg.TestNet3Params,
		},
		// Segwit address tests.
		{
			name:    "segwit mainnet p2wpkh v0",
			addr:    "FC1QV4078QVGMF7JTKEYS3DW7QSUGNF5JZ8NH507P3",
			encoded: "fc1qv4078qvgmf7jtkeys3dw7qsugnf5jz8nh507p3",
			valid:   true,
			result: chainutil.TstAddressWitnessPubKeyHash(
				0,
				[20]byte{
					0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4, 0x54, 0x94,
					0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23, 0xf1, 0x43, 0x3b, 0xd6},
				chaincfg.MainNetParams.Bech32HRPSegwit),
			f: func() (chainutil.Address, error) {
				pkHash := []byte{
					0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4, 0x54, 0x94,
					0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23, 0xf1, 0x43, 0x3b, 0xd6}
				return chainutil.NewAddressWitnessPubKeyHash(pkHash, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			name:    "segwit mainnet p2wsh v0",
			addr:    "fc1qcch6lu4j90lt9uc7gwyrzudkjx6nk9zskjvvexsqcxa3tuxwuaeq546tqs",
			encoded: "fc1qcch6lu4j90lt9uc7gwyrzudkjx6nk9zskjvvexsqcxa3tuxwuaeq546tqs",
			valid:   true,
			result: chainutil.TstAddressWitnessScriptHash(
				0,
				[32]byte{
					0x18, 0x63, 0x14, 0x3c, 0x14, 0xc5, 0x16, 0x68,
					0x04, 0xbd, 0x19, 0x20, 0x33, 0x56, 0xda, 0x13,
					0x6c, 0x98, 0x56, 0x78, 0xcd, 0x4d, 0x27, 0xa1,
					0xb8, 0xc6, 0x32, 0x96, 0x04, 0x90, 0x32, 0x62},
				chaincfg.MainNetParams.Bech32HRPSegwit),
			f: func() (chainutil.Address, error) {
				scriptHash := []byte{
					0x18, 0x63, 0x14, 0x3c, 0x14, 0xc5, 0x16, 0x68,
					0x04, 0xbd, 0x19, 0x20, 0x33, 0x56, 0xda, 0x13,
					0x6c, 0x98, 0x56, 0x78, 0xcd, 0x4d, 0x27, 0xa1,
					0xb8, 0xc6, 0x32, 0x96, 0x04, 0x90, 0x32, 0x62}
				return chainutil.NewAddressWitnessScriptHash(scriptHash, &chaincfg.MainNetParams)
			},
			net: &chaincfg.MainNetParams,
		},
		{
			name:    "segwit testnet p2wpkh v0",
			addr:    "tf1qdz70ec5uqady5eyk9tvqugd5sllmajw87e6gft",
			encoded: "tf1qdz70ec5uqady5eyk9tvqugd5sllmajw87e6gft",
			valid:   true,
			result: chainutil.TstAddressWitnessPubKeyHash(
				0,
				[20]byte{
					0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4, 0x54, 0x94,
					0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23, 0xf1, 0x43, 0x3b, 0xd6},
				chaincfg.TestNet3Params.Bech32HRPSegwit),
			f: func() (chainutil.Address, error) {
				pkHash := []byte{
					0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4, 0x54, 0x94,
					0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23, 0xf1, 0x43, 0x3b, 0xd6}
				return chainutil.NewAddressWitnessPubKeyHash(pkHash, &chaincfg.TestNet3Params)
			},
			net: &chaincfg.TestNet3Params,
		},
		{
			name:    "segwit testnet p2wsh v0",
			addr:    "tf1qyhetym3jfqsc6rdgkmw72d66e7et75x95z3e4np8qsuppvgmnymqhsdhhx",
			encoded: "tf1qyhetym3jfqsc6rdgkmw72d66e7et75x95z3e4np8qsuppvgmnymqhsdhhx",
			valid:   true,
			result: chainutil.TstAddressWitnessScriptHash(
				0,
				[32]byte{
					0x18, 0x63, 0x14, 0x3c, 0x14, 0xc5, 0x16, 0x68,
					0x04, 0xbd, 0x19, 0x20, 0x33, 0x56, 0xda, 0x13,
					0x6c, 0x98, 0x56, 0x78, 0xcd, 0x4d, 0x27, 0xa1,
					0xb8, 0xc6, 0x32, 0x96, 0x04, 0x90, 0x32, 0x62},
				chaincfg.TestNet3Params.Bech32HRPSegwit),
			f: func() (chainutil.Address, error) {
				scriptHash := []byte{
					0x18, 0x63, 0x14, 0x3c, 0x14, 0xc5, 0x16, 0x68,
					0x04, 0xbd, 0x19, 0x20, 0x33, 0x56, 0xda, 0x13,
					0x6c, 0x98, 0x56, 0x78, 0xcd, 0x4d, 0x27, 0xa1,
					0xb8, 0xc6, 0x32, 0x96, 0x04, 0x90, 0x32, 0x62}
				return chainutil.NewAddressWitnessScriptHash(scriptHash, &chaincfg.TestNet3Params)
			},
			net: &chaincfg.TestNet3Params,
		},
		{
			name:    "segwit testnet p2wsh witness v0",
			addr:    "tf1qk552chr0u0nz6xa7ggl9aagh6fk0tqfwk6h6z53t0m02skrpajcssjsjs9",
			encoded: "tf1qk552chr0u0nz6xa7ggl9aagh6fk0tqfwk6h6z53t0m02skrpajcssjsjs9",
			valid:   true,
			result: chainutil.TstAddressWitnessScriptHash(
				0,
				[32]byte{
					0x00, 0x00, 0x00, 0xc4, 0xa5, 0xca, 0xd4, 0x62,
					0x21, 0xb2, 0xa1, 0x87, 0x90, 0x5e, 0x52, 0x66,
					0x36, 0x2b, 0x99, 0xd5, 0xe9, 0x1c, 0x6c, 0xe2,
					0x4d, 0x16, 0x5d, 0xab, 0x93, 0xe8, 0x64, 0x33},
				chaincfg.TestNet3Params.Bech32HRPSegwit),
			f: func() (chainutil.Address, error) {
				scriptHash := []byte{
					0x00, 0x00, 0x00, 0xc4, 0xa5, 0xca, 0xd4, 0x62,
					0x21, 0xb2, 0xa1, 0x87, 0x90, 0x5e, 0x52, 0x66,
					0x36, 0x2b, 0x99, 0xd5, 0xe9, 0x1c, 0x6c, 0xe2,
					0x4d, 0x16, 0x5d, 0xab, 0x93, 0xe8, 0x64, 0x33}
				return chainutil.NewAddressWitnessScriptHash(scriptHash, &chaincfg.TestNet3Params)
			},
			net: &chaincfg.TestNet3Params,
		},
		{
			name:    "segwit litecoin mainnet p2wpkh v0",
			addr:    "LTC1QW508D6QEJXTDG4Y5R3ZARVARY0C5XW7KGMN4N9",
			encoded: "ltc1qw508d6qejxtdg4y5r3zarvary0c5xw7kgmn4n9",
			valid:   true,
			result: chainutil.TstAddressWitnessPubKeyHash(
				0,
				[20]byte{
					0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4, 0x54, 0x94,
					0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23, 0xf1, 0x43, 0x3b, 0xd6},
				CustomParams.Bech32HRPSegwit,
			),
			f: func() (chainutil.Address, error) {
				pkHash := []byte{
					0x75, 0x1e, 0x76, 0xe8, 0x19, 0x91, 0x96, 0xd4, 0x54, 0x94,
					0x1c, 0x45, 0xd1, 0xb3, 0xa3, 0x23, 0xf1, 0x43, 0x3b, 0xd6}
				return chainutil.NewAddressWitnessPubKeyHash(pkHash, &customParams)
			},
			net: &customParams,
		},

		// P2TR address tests.
		{
			name:    "segwit v1 mainnet p2tr",
			addr:    "fc1phzajy94pcsuaf40uen5uyhre9g30dxey5cwf5c5vg824cmkaxp6smg7ggs",
			encoded: "fc1phzajy94pcsuaf40uen5uyhre9g30dxey5cwf5c5vg824cmkaxp6smg7ggs",
			valid:   true,
			result: chainutil.TstAddressTaproot(
				1, [32]byte{
					0xef, 0x46, 0xd1, 0xaa, 0x78, 0x10, 0x1e, 0x33,
					0x50, 0x60, 0x0a, 0x5d, 0x36, 0x04, 0x5b, 0xa9,
					0x7c, 0x26, 0x70, 0xda, 0xa9, 0x1e, 0x9f, 0x3a,
					0x48, 0xc4, 0x3c, 0x6e, 0x73, 0x97, 0x54, 0xe6,
				}, chaincfg.MainNetParams.Bech32HRPSegwit,
			),
			f: func() (chainutil.Address, error) {
				scriptHash := []byte{
					0xef, 0x46, 0xd1, 0xaa, 0x78, 0x10, 0x1e, 0x33,
					0x50, 0x60, 0x0a, 0x5d, 0x36, 0x04, 0x5b, 0xa9,
					0x7c, 0x26, 0x70, 0xda, 0xa9, 0x1e, 0x9f, 0x3a,
					0x48, 0xc4, 0x3c, 0x6e, 0x73, 0x97, 0x54, 0xe6,
				}
				return chainutil.NewAddressTaproot(
					scriptHash, &chaincfg.MainNetParams,
				)
			},
			net: &chaincfg.MainNetParams,
		},

		// Invalid bech32m tests. Source:
		// https://github.com/bitcoin/bips/blob/master/bip-0350.mediawiki
		{
			name:  "segwit v1 invalid human-readable part",
			addr:  "tc1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vq5zuyut",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit v1 mainnet bech32 instead of bech32m",
			addr:  "bc1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vqh2y7hd",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit v1 testnet bech32 instead of bech32m",
			addr:  "tb1z0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vqglt7rf",
			valid: false,
			net:   &chaincfg.TestNet3Params,
		},
		{
			name:  "segwit v1 mainnet bech32 instead of bech32m upper case",
			addr:  "BC1S0XLXVLHEMJA6C4DQV22UAPCTQUPFHLXM9H8Z3K2E72Q4K9HCZ7VQ54WELL",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit v0 mainnet bech32m instead of bech32",
			addr:  "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kemeawh",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit v1 testnet bech32 instead of bech32m second test",
			addr:  "tb1q0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vq24jc47",
			valid: false,
			net:   &chaincfg.TestNet3Params,
		},
		{
			name:  "segwit v1 mainnet bech32m invalid character in checksum",
			addr:  "bc1p38j9r5y49hruaue7wxjce0updqjuyyx0kh56v8s25huc6995vvpql3jow4",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit mainnet witness v17",
			addr:  "BC130XLXVLHEMJA6C4DQV22UAPCTQUPFHLXM9H8Z3K2E72Q4K9HCZ7VQ7ZWS8R",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit v1 mainnet bech32m invalid program length (1 byte)",
			addr:  "bc1pw5dgrnzv",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit v1 mainnet bech32m invalid program length (41 bytes)",
			addr:  "bc1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7v8n0nx0muaewav253zgeav",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit v1 testnet bech32m mixed case",
			addr:  "tb1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vq47Zagq",
			valid: false,
			net:   &chaincfg.TestNet3Params,
		},
		{
			name:  "segwit v1 mainnet bech32m zero padding of more than 4 bits",
			addr:  "bc1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7v07qwwzcrf",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit v1 mainnet bech32m non-zero padding in 8-to-5-conversion",
			addr:  "tb1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vpggkg4j",
			valid: false,
			net:   &chaincfg.TestNet3Params,
		},
		{
			name:  "segwit v1 mainnet bech32m empty data section",
			addr:  "bc1gmk9yu",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},

		// Unsupported witness versions (version 0 and 1 only supported at this point)
		{
			name:  "segwit mainnet witness v16",
			addr:  "BC1SW50QA3JX3S",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit mainnet witness v2",
			addr:  "bc1zw508d6qejxtdg4y5r3zarvaryvg6kdaj",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		// Invalid segwit addresses
		{
			name:  "segwit invalid hrp",
			addr:  "tc1qw508d6qejxtdg4y5r3zarvary0c5xw7kg3g4ty",
			valid: false,
			net:   &chaincfg.TestNet3Params,
		},
		{
			name:  "segwit invalid checksum",
			addr:  "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t5",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit invalid witness version",
			addr:  "BC13W508D6QEJXTDG4Y5R3ZARVARY0C5XW7KN40WF2",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit invalid program length",
			addr:  "bc1rw5uspcuh",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit invalid program length",
			addr:  "bc10w508d6qejxtdg4y5r3zarvary0c5xw7kw508d6qejxtdg4y5r3zarvary0c5xw7kw5rljs90",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit invalid program length for witness version 0 (per BIP141)",
			addr:  "BC1QR508D6QEJXTDG4Y5R3ZARVARYV98GJ9P",
			valid: false,
			net:   &chaincfg.MainNetParams,
		},
		{
			name:  "segwit mixed case",
			addr:  "tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sL5k7",
			valid: false,
			net:   &chaincfg.TestNet3Params,
		},
		{
			name:  "segwit zero padding of more than 4 bits",
			addr:  "tb1pw508d6qejxtdg4y5r3zarqfsj6c3",
			valid: false,
			net:   &chaincfg.TestNet3Params,
		},
		{
			name:  "segwit non-zero padding in 8-to-5 conversion",
			addr:  "tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3pjxtptv",
			valid: false,
			net:   &chaincfg.TestNet3Params,
		},
	}

	if err := chaincfg.Register(&customParams); err != nil {
		panic(err)
	}

	for _, test := range tests {
		// Decode addr and compare error against valid.
		decoded, err := chainutil.DecodeAddress(test.addr, test.net)
		if (err == nil) != test.valid {
			t.Errorf("%v: decoding test failed: %v", test.name, err)
			return
		}

		if err == nil {
			// Ensure the stringer returns the same address as the
			// original.
			if decodedStringer, ok := decoded.(fmt.Stringer); ok {
				addr := test.addr

				// For Segwit addresses the string representation
				// will always be lower case, so in that case we
				// convert the original to lower case first.
				if strings.Contains(test.name, "segwit") {
					addr = strings.ToLower(addr)
				}

				if addr != decodedStringer.String() {
					t.Errorf("%v: String on decoded value does not match expected value: %v != %v",
						test.name, test.addr, decodedStringer.String())
					return
				}
			}

			// Encode again and compare against the original.
			encoded := decoded.EncodeAddress()
			if test.encoded != encoded {
				t.Errorf("%v: decoding and encoding produced different addresses: %v != %v",
					test.name, test.encoded, encoded)
				return
			}

			// Perform type-specific calculations.
			var saddr []byte
			switch d := decoded.(type) {
			case *chainutil.AddressPubKeyHash:
				saddr = chainutil.TstAddressSAddr(encoded)

			case *chainutil.AddressScriptHash:
				saddr = chainutil.TstAddressSAddr(encoded)

			case *chainutil.AddressPubKey:
				// Ignore the error here since the script
				// address is checked below.
				saddr, _ = hex.DecodeString(d.String())
			case *chainutil.AddressWitnessPubKeyHash:
				saddr = chainutil.TstAddressSegwitSAddr(encoded)
			case *chainutil.AddressWitnessScriptHash:
				saddr = chainutil.TstAddressSegwitSAddr(encoded)
			case *chainutil.AddressTaproot:
				saddr = chainutil.TstAddressTaprootSAddr(encoded)
			}

			// Check script address, as well as the Hash160 method for P2PKH and
			// P2SH addresses.
			if !bytes.Equal(saddr, decoded.ScriptAddress()) {
				t.Errorf("%v: script addresses do not match:\n%x != \n%x",
					test.name, saddr, decoded.ScriptAddress())
				return
			}
			switch a := decoded.(type) {
			case *chainutil.AddressPubKeyHash:
				if h := a.Hash160()[:]; !bytes.Equal(saddr, h) {
					t.Errorf("%v: hashes do not match:\n%x != \n%x",
						test.name, saddr, h)
					return
				}

			case *chainutil.AddressScriptHash:
				if h := a.Hash160()[:]; !bytes.Equal(saddr, h) {
					t.Errorf("%v: hashes do not match:\n%x != \n%x",
						test.name, saddr, h)
					return
				}

			case *chainutil.AddressWitnessPubKeyHash:
				if hrp := a.Hrp(); test.net.Bech32HRPSegwit != hrp {
					t.Errorf("%v: hrps do not match:\n%x != \n%x",
						test.name, test.net.Bech32HRPSegwit, hrp)
					return
				}

				expVer := test.result.(*chainutil.AddressWitnessPubKeyHash).WitnessVersion()
				if v := a.WitnessVersion(); v != expVer {
					t.Errorf("%v: witness versions do not match:\n%x != \n%x",
						test.name, expVer, v)
					return
				}

				if p := a.WitnessProgram(); !bytes.Equal(saddr, p) {
					t.Errorf("%v: witness programs do not match:\n%x != \n%x",
						test.name, saddr, p)
					return
				}

			case *chainutil.AddressWitnessScriptHash:
				if hrp := a.Hrp(); test.net.Bech32HRPSegwit != hrp {
					t.Errorf("%v: hrps do not match:\n%x != \n%x",
						test.name, test.net.Bech32HRPSegwit, hrp)
					return
				}

				expVer := test.result.(*chainutil.AddressWitnessScriptHash).WitnessVersion()
				if v := a.WitnessVersion(); v != expVer {
					t.Errorf("%v: witness versions do not match:\n%x != \n%x",
						test.name, expVer, v)
					return
				}

				if p := a.WitnessProgram(); !bytes.Equal(saddr, p) {
					t.Errorf("%v: witness programs do not match:\n%x != \n%x",
						test.name, saddr, p)
					return
				}
			}

			// Ensure the address is for the expected network.
			if !decoded.IsForNet(test.net) {
				t.Errorf("%v: calculated network does not match expected",
					test.name)
				return
			}
		} else {
			// If there is an error, make sure we can print it
			// correctly.
			errStr := err.Error()
			if errStr == "" {
				t.Errorf("%v: error was non-nil but message is"+
					"empty: %v", test.name, err)
			}
		}

		if !test.valid {
			// If address is invalid, but a creation function exists,
			// verify that it returns a nil addr and non-nil error.
			if test.f != nil {
				_, err := test.f()
				if err == nil {
					t.Errorf("%v: address is invalid but creating new address succeeded",
						test.name)
					return
				}
			}
			continue
		}

		// Valid test, compare address created with f against expected result.
		addr, err := test.f()
		if err != nil {
			t.Errorf("%v: address is valid but creating new address failed with error %v",
				test.name, err)
			return
		}

		if !reflect.DeepEqual(addr, test.result) {
			t.Errorf("%v: created address does not match expected result",
				test.name)
			return
		}
	}
}

func TestCalcSegwit(t *testing.T) { // #FLOKI_CHANGE
	priv, _ := crypto.NewPrivateKey()

	pubKey := priv.PubKey().SerializeCompressed()

	pubKeyHash := sha256.Sum256(pubKey)

	addr, err := chainutil.NewAddressWitnessPubKeyHash(pubKeyHash[:20], &chaincfg.TestNet3Params)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	log.Printf("encoded: %x addr: %v", priv.PubKey().SerializeCompressed(), addr.EncodeAddress())
}

func TestCalcWitnessScriptHash(t *testing.T) { // #FLOKI_CHANGE
	priv, _ := crypto.NewPrivateKey()

	pubKey := priv.PubKey().SerializeCompressed()

	pubKeyHash := sha256.Sum256(pubKey)

	redeemScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(pubKeyHash[:]).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()
	if err != nil {
		log.Fatalf("Failed to create redeem script: %v", err)
	}

	witnessProg := sha256.Sum256(redeemScript)

	addr, err := chainutil.NewAddressWitnessScriptHash(witnessProg[:], &chaincfg.TestNet3Params)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	log.Printf("encoded: %x addr: %v", priv.PubKey().SerializeCompressed(), addr.EncodeAddress())
}

func TestCalcTaproot(t *testing.T) { // #FLOKI_CHANGE
	priv, _ := crypto.NewPrivateKey()

	pubKey := priv.PubKey().SerializeCompressed()[1:]

	log.Printf("pubKey: %d", len(pubKey))

	tweakedPubKey, err := schnorr.ParsePubKey(pubKey)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	schnorrPubKey := schnorr.SerializePubKey(tweakedPubKey)
	log.Printf("schnorrPubKey: %d", len(schnorrPubKey))

	addr, err := chainutil.NewAddressTaproot(schnorrPubKey, &chaincfg.MainNetParams)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	log.Printf("encoded: %x addr: %v", priv.PubKey().SerializeCompressed(), addr.EncodeAddress())
}

func TestCalcWitness(t *testing.T) { // #FLOKI_CHANGE

	tx := wire.NewMsgTx(wire.TxVersion)

	// Add a dummy input (replace with real UTXO details)
	prevHash, _ := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000001")
	prevOut := wire.NewOutPoint(prevHash, 0)
	tx.AddTxIn(wire.NewTxIn(prevOut, nil, nil))

	// Add an output (replace with real script and value)
	txOut := wire.NewTxOut(10000, []byte{}) // 0.0001 FLC
	tx.AddTxOut(txOut)

	// Step 2: Convert the unsigned transaction into a PSBT
	var buf bytes.Buffer
	err := tx.Serialize(&buf)
	if err != nil {
		log.Fatalf("Failed to serialize transaction: %v", err)
	}

	psbtPacket, err := psbt.NewFromUnsignedTx(tx)
	if err != nil {
		log.Fatalf("Failed to create PSBT: %v", err)
	}

	// Step 3: Serialize the PSBT for storage or signing
	var psbtBuf bytes.Buffer
	err = psbtPacket.Serialize(&psbtBuf)
	if err != nil {
		log.Fatalf("Failed to serialize PSBT: %v", err)
	}
	txHex := hex.EncodeToString(buf.Bytes())

	log.Printf("Transaction Hex: %s", txHex)

	log.Printf("PSBT Base64: %s", psbtBuf.String())

}
