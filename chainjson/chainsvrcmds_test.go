// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainjson_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/chainjson"
	"github.com/flokiorg/go-flokicoin/wire"
)

// TestChainSvrCmds tests all of the chain server commands marshal and unmarshal
// into valid results include handling of optional fields being omitted in the
// marshalled command, while optional fields with defaults have the default
// assigned on unmarshalled commands.
func TestChainSvrCmds(t *testing.T) {
	t.Parallel()

	testID := int(1)
	tests := []struct {
		name         string
		newCmd       func() (interface{}, error)
		staticCmd    func() interface{}
		marshalled   string
		unmarshalled interface{}
	}{
		{
			name: "addnode",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("addnode", "127.0.0.1", chainjson.ANRemove)
			},
			staticCmd: func() interface{} {
				return chainjson.NewAddNodeCmd("127.0.0.1", chainjson.ANRemove)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"addnode","params":["127.0.0.1","remove"],"id":1}`,
			unmarshalled: &chainjson.AddNodeCmd{Addr: "127.0.0.1", SubCmd: chainjson.ANRemove},
		},
		{
			name: "createrawtransaction",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("createrawtransaction", `[{"txid":"123","vout":1}]`,
					`{"456":0.0123}`)
			},
			staticCmd: func() interface{} {
				txInputs := []chainjson.TransactionInput{
					{Txid: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return chainjson.NewCreateRawTransactionCmd(txInputs, amounts, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createrawtransaction","params":[[{"txid":"123","vout":1}],{"456":0.0123}],"id":1}`,
			unmarshalled: &chainjson.CreateRawTransactionCmd{
				Inputs:  []chainjson.TransactionInput{{Txid: "123", Vout: 1}},
				Amounts: map[string]float64{"456": .0123},
			},
		},
		{
			name: "createrawtransaction - no inputs",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("createrawtransaction", `[]`, `{"456":0.0123}`)
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"456": .0123}
				return chainjson.NewCreateRawTransactionCmd(nil, amounts, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createrawtransaction","params":[[],{"456":0.0123}],"id":1}`,
			unmarshalled: &chainjson.CreateRawTransactionCmd{
				Inputs:  []chainjson.TransactionInput{},
				Amounts: map[string]float64{"456": .0123},
			},
		},
		{
			name: "createrawtransaction optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("createrawtransaction", `[{"txid":"123","vout":1}]`,
					`{"456":0.0123}`, int64(12312333333))
			},
			staticCmd: func() interface{} {
				txInputs := []chainjson.TransactionInput{
					{Txid: "123", Vout: 1},
				}
				amounts := map[string]float64{"456": .0123}
				return chainjson.NewCreateRawTransactionCmd(txInputs, amounts, chainjson.Int64(12312333333))
			},
			marshalled: `{"jsonrpc":"1.0","method":"createrawtransaction","params":[[{"txid":"123","vout":1}],{"456":0.0123},12312333333],"id":1}`,
			unmarshalled: &chainjson.CreateRawTransactionCmd{
				Inputs:   []chainjson.TransactionInput{{Txid: "123", Vout: 1}},
				Amounts:  map[string]float64{"456": .0123},
				LockTime: chainjson.Int64(12312333333),
			},
		},
		{
			name: "fundrawtransaction - empty opts",
			newCmd: func() (i interface{}, e error) {
				return chainjson.NewCmd("fundrawtransaction", "deadbeef", "{}")
			},
			staticCmd: func() interface{} {
				deadbeef, err := hex.DecodeString("deadbeef")
				if err != nil {
					panic(err)
				}
				return chainjson.NewFundRawTransactionCmd(deadbeef, chainjson.FundRawTransactionOpts{}, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"fundrawtransaction","params":["deadbeef",{}],"id":1}`,
			unmarshalled: &chainjson.FundRawTransactionCmd{
				HexTx:     "deadbeef",
				Options:   chainjson.FundRawTransactionOpts{},
				IsWitness: nil,
			},
		},
		{
			name: "fundrawtransaction - full opts",
			newCmd: func() (i interface{}, e error) {
				return chainjson.NewCmd("fundrawtransaction", "deadbeef", `{"changeAddress":"bcrt1qeeuctq9wutlcl5zatge7rjgx0k45228cxez655","changePosition":1,"change_type":"legacy","includeWatching":true,"lockUnspents":true,"feeRate":0.7,"subtractFeeFromOutputs":[0],"replaceable":true,"conf_target":8,"estimate_mode":"ECONOMICAL"}`)
			},
			staticCmd: func() interface{} {
				deadbeef, err := hex.DecodeString("deadbeef")
				if err != nil {
					panic(err)
				}
				changeAddress := "bcrt1qeeuctq9wutlcl5zatge7rjgx0k45228cxez655"
				change := 1
				changeType := chainjson.ChangeTypeLegacy
				watching := true
				lockUnspents := true
				feeRate := 0.7
				replaceable := true
				confTarget := 8

				return chainjson.NewFundRawTransactionCmd(deadbeef, chainjson.FundRawTransactionOpts{
					ChangeAddress:          &changeAddress,
					ChangePosition:         &change,
					ChangeType:             &changeType,
					IncludeWatching:        &watching,
					LockUnspents:           &lockUnspents,
					FeeRate:                &feeRate,
					SubtractFeeFromOutputs: []int{0},
					Replaceable:            &replaceable,
					ConfTarget:             &confTarget,
					EstimateMode:           &chainjson.EstimateModeEconomical,
				}, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"fundrawtransaction","params":["deadbeef",{"changeAddress":"bcrt1qeeuctq9wutlcl5zatge7rjgx0k45228cxez655","changePosition":1,"change_type":"legacy","includeWatching":true,"lockUnspents":true,"feeRate":0.7,"subtractFeeFromOutputs":[0],"replaceable":true,"conf_target":8,"estimate_mode":"ECONOMICAL"}],"id":1}`,
			unmarshalled: func() interface{} {
				changeAddress := "bcrt1qeeuctq9wutlcl5zatge7rjgx0k45228cxez655"
				change := 1
				changeType := chainjson.ChangeTypeLegacy
				watching := true
				lockUnspents := true
				feeRate := 0.7
				replaceable := true
				confTarget := 8
				return &chainjson.FundRawTransactionCmd{
					HexTx: "deadbeef",
					Options: chainjson.FundRawTransactionOpts{
						ChangeAddress:          &changeAddress,
						ChangePosition:         &change,
						ChangeType:             &changeType,
						IncludeWatching:        &watching,
						LockUnspents:           &lockUnspents,
						FeeRate:                &feeRate,
						SubtractFeeFromOutputs: []int{0},
						Replaceable:            &replaceable,
						ConfTarget:             &confTarget,
						EstimateMode:           &chainjson.EstimateModeEconomical,
					},
					IsWitness: nil,
				}
			}(),
		},
		{
			name: "fundrawtransaction - iswitness",
			newCmd: func() (i interface{}, e error) {
				return chainjson.NewCmd("fundrawtransaction", "deadbeef", "{}", true)
			},
			staticCmd: func() interface{} {
				deadbeef, err := hex.DecodeString("deadbeef")
				if err != nil {
					panic(err)
				}
				t := true
				return chainjson.NewFundRawTransactionCmd(deadbeef, chainjson.FundRawTransactionOpts{}, &t)
			},
			marshalled: `{"jsonrpc":"1.0","method":"fundrawtransaction","params":["deadbeef",{},true],"id":1}`,
			unmarshalled: &chainjson.FundRawTransactionCmd{
				HexTx:   "deadbeef",
				Options: chainjson.FundRawTransactionOpts{},
				IsWitness: func() *bool {
					t := true
					return &t
				}(),
			},
		},
		{
			name: "decoderawtransaction",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("decoderawtransaction", "123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewDecodeRawTransactionCmd("123")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decoderawtransaction","params":["123"],"id":1}`,
			unmarshalled: &chainjson.DecodeRawTransactionCmd{HexTx: "123"},
		},
		{
			name: "decodescript",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("decodescript", "00")
			},
			staticCmd: func() interface{} {
				return chainjson.NewDecodeScriptCmd("00")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"decodescript","params":["00"],"id":1}`,
			unmarshalled: &chainjson.DecodeScriptCmd{HexScript: "00"},
		},
		{
			name: "deriveaddresses no range",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("deriveaddresses", "00")
			},
			staticCmd: func() interface{} {
				return chainjson.NewDeriveAddressesCmd("00", nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"deriveaddresses","params":["00"],"id":1}`,
			unmarshalled: &chainjson.DeriveAddressesCmd{Descriptor: "00"},
		},
		{
			name: "deriveaddresses int range",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"deriveaddresses", "00", chainjson.DescriptorRange{Value: 2})
			},
			staticCmd: func() interface{} {
				return chainjson.NewDeriveAddressesCmd(
					"00", &chainjson.DescriptorRange{Value: 2})
			},
			marshalled: `{"jsonrpc":"1.0","method":"deriveaddresses","params":["00",2],"id":1}`,
			unmarshalled: &chainjson.DeriveAddressesCmd{
				Descriptor: "00",
				Range:      &chainjson.DescriptorRange{Value: 2},
			},
		},
		{
			name: "deriveaddresses slice range",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"deriveaddresses", "00",
					chainjson.DescriptorRange{Value: []int{0, 2}},
				)
			},
			staticCmd: func() interface{} {
				return chainjson.NewDeriveAddressesCmd(
					"00", &chainjson.DescriptorRange{Value: []int{0, 2}})
			},
			marshalled: `{"jsonrpc":"1.0","method":"deriveaddresses","params":["00",[0,2]],"id":1}`,
			unmarshalled: &chainjson.DeriveAddressesCmd{
				Descriptor: "00",
				Range:      &chainjson.DescriptorRange{Value: []int{0, 2}},
			},
		},
		{
			name: "getaddednodeinfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getaddednodeinfo", true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetAddedNodeInfoCmd(true, nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getaddednodeinfo","params":[true],"id":1}`,
			unmarshalled: &chainjson.GetAddedNodeInfoCmd{DNS: true, Node: nil},
		},
		{
			name: "getaddednodeinfo optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getaddednodeinfo", true, "127.0.0.1")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetAddedNodeInfoCmd(true, chainjson.String("127.0.0.1"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getaddednodeinfo","params":[true,"127.0.0.1"],"id":1}`,
			unmarshalled: &chainjson.GetAddedNodeInfoCmd{
				DNS:  true,
				Node: chainjson.String("127.0.0.1"),
			},
		},
		{
			name: "getbestblockhash",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getbestblockhash")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBestBlockHashCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getbestblockhash","params":[],"id":1}`,
			unmarshalled: &chainjson.GetBestBlockHashCmd{},
		},
		{
			name: "getblock",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblock", "123", chainjson.Int(0))
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockCmd("123", chainjson.Int(0))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblock","params":["123",0],"id":1}`,
			unmarshalled: &chainjson.GetBlockCmd{
				Hash:      "123",
				Verbosity: chainjson.Int(0),
			},
		},
		{
			name: "getblock default verbosity",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblock", "123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblock","params":["123"],"id":1}`,
			unmarshalled: &chainjson.GetBlockCmd{
				Hash:      "123",
				Verbosity: chainjson.Int(1),
			},
		},
		{
			name: "getblock required optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblock", "123", chainjson.Int(1))
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockCmd("123", chainjson.Int(1))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblock","params":["123",1],"id":1}`,
			unmarshalled: &chainjson.GetBlockCmd{
				Hash:      "123",
				Verbosity: chainjson.Int(1),
			},
		},
		{
			name: "getblock required optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblock", "123", chainjson.Int(2))
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockCmd("123", chainjson.Int(2))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblock","params":["123",2],"id":1}`,
			unmarshalled: &chainjson.GetBlockCmd{
				Hash:      "123",
				Verbosity: chainjson.Int(2),
			},
		},
		{
			name: "getblockchaininfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockchaininfo")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockChainInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblockchaininfo","params":[],"id":1}`,
			unmarshalled: &chainjson.GetBlockChainInfoCmd{},
		},
		{
			name: "getblockcount",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockcount")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblockcount","params":[],"id":1}`,
			unmarshalled: &chainjson.GetBlockCountCmd{},
		},
		{
			name: "getblockfilter",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockfilter", "0000afaf")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockFilterCmd("0000afaf", nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblockfilter","params":["0000afaf"],"id":1}`,
			unmarshalled: &chainjson.GetBlockFilterCmd{"0000afaf", nil},
		},
		{
			name: "getblockfilter optional filtertype",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockfilter", "0000afaf", "basic")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockFilterCmd("0000afaf", chainjson.NewFilterTypeName(chainjson.FilterTypeBasic))
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblockfilter","params":["0000afaf","basic"],"id":1}`,
			unmarshalled: &chainjson.GetBlockFilterCmd{"0000afaf", chainjson.NewFilterTypeName(chainjson.FilterTypeBasic)},
		},
		{
			name: "getblockhash",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockhash", 123)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockHashCmd(123)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblockhash","params":[123],"id":1}`,
			unmarshalled: &chainjson.GetBlockHashCmd{Index: 123},
		},
		{
			name: "getblockheader",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockheader", "123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockHeaderCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblockheader","params":["123"],"id":1}`,
			unmarshalled: &chainjson.GetBlockHeaderCmd{
				Hash:    "123",
				Verbose: chainjson.Bool(true),
			},
		},
		{
			name: "getblockstats height",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockstats", chainjson.HashOrHeight{Value: 123})
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockStatsCmd(chainjson.HashOrHeight{Value: 123}, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblockstats","params":[123],"id":1}`,
			unmarshalled: &chainjson.GetBlockStatsCmd{
				HashOrHeight: chainjson.HashOrHeight{Value: 123},
			},
		},
		{
			name: "getblockstats hash",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockstats", chainjson.HashOrHeight{Value: "deadbeef"})
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockStatsCmd(chainjson.HashOrHeight{Value: "deadbeef"}, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblockstats","params":["deadbeef"],"id":1}`,
			unmarshalled: &chainjson.GetBlockStatsCmd{
				HashOrHeight: chainjson.HashOrHeight{Value: "deadbeef"},
			},
		},
		{
			name: "getblockstats height optional stats",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockstats", chainjson.HashOrHeight{Value: 123}, []string{"avgfee", "maxfee"})
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockStatsCmd(chainjson.HashOrHeight{Value: 123}, &[]string{"avgfee", "maxfee"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblockstats","params":[123,["avgfee","maxfee"]],"id":1}`,
			unmarshalled: &chainjson.GetBlockStatsCmd{
				HashOrHeight: chainjson.HashOrHeight{Value: 123},
				Stats:        &[]string{"avgfee", "maxfee"},
			},
		},
		{
			name: "getblockstats hash optional stats",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblockstats", chainjson.HashOrHeight{Value: "deadbeef"}, []string{"avgfee", "maxfee"})
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockStatsCmd(chainjson.HashOrHeight{Value: "deadbeef"}, &[]string{"avgfee", "maxfee"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblockstats","params":["deadbeef",["avgfee","maxfee"]],"id":1}`,
			unmarshalled: &chainjson.GetBlockStatsCmd{
				HashOrHeight: chainjson.HashOrHeight{Value: "deadbeef"},
				Stats:        &[]string{"avgfee", "maxfee"},
			},
		},
		{
			name: "getblocktemplate",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblocktemplate")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBlockTemplateCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getblocktemplate","params":[],"id":1}`,
			unmarshalled: &chainjson.GetBlockTemplateCmd{Request: nil},
		},
		{
			name: "getblocktemplate optional - template request",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblocktemplate", `{"mode":"template","capabilities":["longpoll","coinbasetxn"]}`)
			},
			staticCmd: func() interface{} {
				template := chainjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				}
				return chainjson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblocktemplate","params":[{"mode":"template","capabilities":["longpoll","coinbasetxn"]}],"id":1}`,
			unmarshalled: &chainjson.GetBlockTemplateCmd{
				Request: &chainjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
				},
			},
		},
		{
			name: "getblocktemplate optional - template request with tweaks",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblocktemplate", `{"mode":"template","capabilities":["longpoll","coinbasetxn"],"sigoplimit":500,"sizelimit":100000000,"maxversion":2}`)
			},
			staticCmd: func() interface{} {
				template := chainjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
					SigOpLimit:   500,
					SizeLimit:    100000000,
					MaxVersion:   2,
				}
				return chainjson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblocktemplate","params":[{"mode":"template","capabilities":["longpoll","coinbasetxn"],"sigoplimit":500,"sizelimit":100000000,"maxversion":2}],"id":1}`,
			unmarshalled: &chainjson.GetBlockTemplateCmd{
				Request: &chainjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
					SigOpLimit:   int64(500),
					SizeLimit:    int64(100000000),
					MaxVersion:   2,
				},
			},
		},
		{
			name: "getblocktemplate optional - template request with tweaks 2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getblocktemplate", `{"mode":"template","capabilities":["longpoll","coinbasetxn"],"sigoplimit":true,"sizelimit":100000000,"maxversion":2}`)
			},
			staticCmd: func() interface{} {
				template := chainjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
					SigOpLimit:   true,
					SizeLimit:    100000000,
					MaxVersion:   2,
				}
				return chainjson.NewGetBlockTemplateCmd(&template)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getblocktemplate","params":[{"mode":"template","capabilities":["longpoll","coinbasetxn"],"sigoplimit":true,"sizelimit":100000000,"maxversion":2}],"id":1}`,
			unmarshalled: &chainjson.GetBlockTemplateCmd{
				Request: &chainjson.TemplateRequest{
					Mode:         "template",
					Capabilities: []string{"longpoll", "coinbasetxn"},
					SigOpLimit:   true,
					SizeLimit:    int64(100000000),
					MaxVersion:   2,
				},
			},
		},
		{
			name: "getcfilter",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getcfilter", "123",
					wire.GCSFilterRegular)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetCFilterCmd("123",
					wire.GCSFilterRegular)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getcfilter","params":["123",0],"id":1}`,
			unmarshalled: &chainjson.GetCFilterCmd{
				Hash:       "123",
				FilterType: wire.GCSFilterRegular,
			},
		},
		{
			name: "getcfilterheader",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getcfilterheader", "123",
					wire.GCSFilterRegular)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetCFilterHeaderCmd("123",
					wire.GCSFilterRegular)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getcfilterheader","params":["123",0],"id":1}`,
			unmarshalled: &chainjson.GetCFilterHeaderCmd{
				Hash:       "123",
				FilterType: wire.GCSFilterRegular,
			},
		},
		{
			name: "getchaintips",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getchaintips")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetChainTipsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getchaintips","params":[],"id":1}`,
			unmarshalled: &chainjson.GetChainTipsCmd{},
		},
		{
			name: "getchaintxstats",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getchaintxstats")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetChainTxStatsCmd(nil, nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getchaintxstats","params":[],"id":1}`,
			unmarshalled: &chainjson.GetChainTxStatsCmd{},
		},
		{
			name: "getchaintxstats optional nblocks",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getchaintxstats", chainjson.Int32(1000))
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetChainTxStatsCmd(chainjson.Int32(1000), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getchaintxstats","params":[1000],"id":1}`,
			unmarshalled: &chainjson.GetChainTxStatsCmd{
				NBlocks: chainjson.Int32(1000),
			},
		},
		{
			name: "getchaintxstats optional nblocks and blockhash",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getchaintxstats", chainjson.Int32(1000), chainjson.String("0000afaf"))
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetChainTxStatsCmd(chainjson.Int32(1000), chainjson.String("0000afaf"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getchaintxstats","params":[1000,"0000afaf"],"id":1}`,
			unmarshalled: &chainjson.GetChainTxStatsCmd{
				NBlocks:   chainjson.Int32(1000),
				BlockHash: chainjson.String("0000afaf"),
			},
		},
		{
			name: "getconnectioncount",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getconnectioncount")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetConnectionCountCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getconnectioncount","params":[],"id":1}`,
			unmarshalled: &chainjson.GetConnectionCountCmd{},
		},
		{
			name: "getdifficulty",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getdifficulty")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetDifficultyCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getdifficulty","params":[],"id":1}`,
			unmarshalled: &chainjson.GetDifficultyCmd{},
		},
		{
			name: "getgenerate",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getgenerate")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetGenerateCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getgenerate","params":[],"id":1}`,
			unmarshalled: &chainjson.GetGenerateCmd{},
		},
		{
			name: "gethashespersec",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("gethashespersec")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetHashesPerSecCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"gethashespersec","params":[],"id":1}`,
			unmarshalled: &chainjson.GetHashesPerSecCmd{},
		},
		{
			name: "getinfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getinfo")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getinfo","params":[],"id":1}`,
			unmarshalled: &chainjson.GetInfoCmd{},
		},
		{
			name: "getmempoolentry",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getmempoolentry", "txhash")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetMempoolEntryCmd("txhash")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getmempoolentry","params":["txhash"],"id":1}`,
			unmarshalled: &chainjson.GetMempoolEntryCmd{
				TxID: "txhash",
			},
		},
		{
			name: "getmempoolinfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getmempoolinfo")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetMempoolInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getmempoolinfo","params":[],"id":1}`,
			unmarshalled: &chainjson.GetMempoolInfoCmd{},
		},
		{
			name: "getmininginfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getmininginfo")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetMiningInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getmininginfo","params":[],"id":1}`,
			unmarshalled: &chainjson.GetMiningInfoCmd{},
		},
		{
			name: "getnetworkinfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnetworkinfo")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNetworkInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getnetworkinfo","params":[],"id":1}`,
			unmarshalled: &chainjson.GetNetworkInfoCmd{},
		},
		{
			name: "getnettotals",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnettotals")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNetTotalsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getnettotals","params":[],"id":1}`,
			unmarshalled: &chainjson.GetNetTotalsCmd{},
		},
		{
			name: "getnetworkhashps",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnetworkhashps")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNetworkHashPSCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnetworkhashps","params":[],"id":1}`,
			unmarshalled: &chainjson.GetNetworkHashPSCmd{
				Blocks: chainjson.Int(120),
				Height: chainjson.Int(-1),
			},
		},
		{
			name: "getnetworkhashps optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnetworkhashps", 200)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNetworkHashPSCmd(chainjson.Int(200), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnetworkhashps","params":[200],"id":1}`,
			unmarshalled: &chainjson.GetNetworkHashPSCmd{
				Blocks: chainjson.Int(200),
				Height: chainjson.Int(-1),
			},
		},
		{
			name: "getnetworkhashps optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnetworkhashps", 200, 123)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNetworkHashPSCmd(chainjson.Int(200), chainjson.Int(123))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnetworkhashps","params":[200,123],"id":1}`,
			unmarshalled: &chainjson.GetNetworkHashPSCmd{
				Blocks: chainjson.Int(200),
				Height: chainjson.Int(123),
			},
		},
		{
			name: "getnodeaddresses",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnodeaddresses")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNodeAddressesCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnodeaddresses","params":[],"id":1}`,
			unmarshalled: &chainjson.GetNodeAddressesCmd{
				Count: chainjson.Int32(1),
			},
		},
		{
			name: "getnodeaddresses optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnodeaddresses", 10)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNodeAddressesCmd(chainjson.Int32(10))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnodeaddresses","params":[10],"id":1}`,
			unmarshalled: &chainjson.GetNodeAddressesCmd{
				Count: chainjson.Int32(10),
			},
		},
		{
			name: "getpeerinfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getpeerinfo")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetPeerInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getpeerinfo","params":[],"id":1}`,
			unmarshalled: &chainjson.GetPeerInfoCmd{},
		},
		{
			name: "getrawmempool",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getrawmempool")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetRawMempoolCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawmempool","params":[],"id":1}`,
			unmarshalled: &chainjson.GetRawMempoolCmd{
				Verbose: chainjson.Bool(false),
			},
		},
		{
			name: "getrawmempool optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getrawmempool", false)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetRawMempoolCmd(chainjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawmempool","params":[false],"id":1}`,
			unmarshalled: &chainjson.GetRawMempoolCmd{
				Verbose: chainjson.Bool(false),
			},
		},
		{
			name: "getrawtransaction",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getrawtransaction", "123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetRawTransactionCmd("123", nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawtransaction","params":["123"],"id":1}`,
			unmarshalled: &chainjson.GetRawTransactionCmd{
				Txid:    "123",
				Verbose: chainjson.Int(0),
			},
		},
		{
			name: "getrawtransaction optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getrawtransaction", "123", 1)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetRawTransactionCmd("123", chainjson.Int(1), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawtransaction","params":["123",1],"id":1}`,
			unmarshalled: &chainjson.GetRawTransactionCmd{
				Txid:    "123",
				Verbose: chainjson.Int(1),
			},
		},
		{
			name: "gettxout",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("gettxout", "123", 1)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetTxOutCmd("123", 1, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettxout","params":["123",1],"id":1}`,
			unmarshalled: &chainjson.GetTxOutCmd{
				Txid:           "123",
				Vout:           1,
				IncludeMempool: chainjson.Bool(true),
			},
		},
		{
			name: "gettxout optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("gettxout", "123", 1, true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetTxOutCmd("123", 1, chainjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettxout","params":["123",1,true],"id":1}`,
			unmarshalled: &chainjson.GetTxOutCmd{
				Txid:           "123",
				Vout:           1,
				IncludeMempool: chainjson.Bool(true),
			},
		},
		{
			name: "gettxoutproof",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("gettxoutproof", []string{"123", "456"})
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetTxOutProofCmd([]string{"123", "456"}, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettxoutproof","params":[["123","456"]],"id":1}`,
			unmarshalled: &chainjson.GetTxOutProofCmd{
				TxIDs: []string{"123", "456"},
			},
		},
		{
			name: "gettxoutproof optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("gettxoutproof", []string{"123", "456"},
					chainjson.String("000000000000034a7dedef4a161fa058a2d67a173a90155f3a2fe6fc132e0ebf"))
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetTxOutProofCmd([]string{"123", "456"},
					chainjson.String("000000000000034a7dedef4a161fa058a2d67a173a90155f3a2fe6fc132e0ebf"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettxoutproof","params":[["123","456"],` +
				`"000000000000034a7dedef4a161fa058a2d67a173a90155f3a2fe6fc132e0ebf"],"id":1}`,
			unmarshalled: &chainjson.GetTxOutProofCmd{
				TxIDs:     []string{"123", "456"},
				BlockHash: chainjson.String("000000000000034a7dedef4a161fa058a2d67a173a90155f3a2fe6fc132e0ebf"),
			},
		},
		{
			name: "gettxoutsetinfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("gettxoutsetinfo")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetTxOutSetInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"gettxoutsetinfo","params":[],"id":1}`,
			unmarshalled: &chainjson.GetTxOutSetInfoCmd{},
		},
		{
			name: "getwork",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getwork")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetWorkCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getwork","params":[],"id":1}`,
			unmarshalled: &chainjson.GetWorkCmd{
				Data: nil,
			},
		},
		{
			name: "getwork optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getwork", "00112233")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetWorkCmd(chainjson.String("00112233"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getwork","params":["00112233"],"id":1}`,
			unmarshalled: &chainjson.GetWorkCmd{
				Data: chainjson.String("00112233"),
			},
		},
		{
			name: "help",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("help")
			},
			staticCmd: func() interface{} {
				return chainjson.NewHelpCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":[],"id":1}`,
			unmarshalled: &chainjson.HelpCmd{
				Command: nil,
			},
		},
		{
			name: "help optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("help", "getblock")
			},
			staticCmd: func() interface{} {
				return chainjson.NewHelpCmd(chainjson.String("getblock"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"help","params":["getblock"],"id":1}`,
			unmarshalled: &chainjson.HelpCmd{
				Command: chainjson.String("getblock"),
			},
		},
		{
			name: "invalidateblock",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("invalidateblock", "123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewInvalidateBlockCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"invalidateblock","params":["123"],"id":1}`,
			unmarshalled: &chainjson.InvalidateBlockCmd{
				BlockHash: "123",
			},
		},
		{
			name: "ping",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("ping")
			},
			staticCmd: func() interface{} {
				return chainjson.NewPingCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"ping","params":[],"id":1}`,
			unmarshalled: &chainjson.PingCmd{},
		},
		{
			name: "preciousblock",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("preciousblock", "0123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewPreciousBlockCmd("0123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"preciousblock","params":["0123"],"id":1}`,
			unmarshalled: &chainjson.PreciousBlockCmd{
				BlockHash: "0123",
			},
		},
		{
			name: "reconsiderblock",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("reconsiderblock", "123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewReconsiderBlockCmd("123")
			},
			marshalled: `{"jsonrpc":"1.0","method":"reconsiderblock","params":["123"],"id":1}`,
			unmarshalled: &chainjson.ReconsiderBlockCmd{
				BlockHash: "123",
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("searchrawtransactions", "1Address")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSearchRawTransactionsCmd("1Address", nil, nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address"],"id":1}`,
			unmarshalled: &chainjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     chainjson.Int(1),
				Skip:        chainjson.Int(0),
				Count:       chainjson.Int(100),
				VinExtra:    chainjson.Int(0),
				Reverse:     chainjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("searchrawtransactions", "1Address", 0)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSearchRawTransactionsCmd("1Address",
					chainjson.Int(0), nil, nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0],"id":1}`,
			unmarshalled: &chainjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     chainjson.Int(0),
				Skip:        chainjson.Int(0),
				Count:       chainjson.Int(100),
				VinExtra:    chainjson.Int(0),
				Reverse:     chainjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("searchrawtransactions", "1Address", 0, 5)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSearchRawTransactionsCmd("1Address",
					chainjson.Int(0), chainjson.Int(5), nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5],"id":1}`,
			unmarshalled: &chainjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     chainjson.Int(0),
				Skip:        chainjson.Int(5),
				Count:       chainjson.Int(100),
				VinExtra:    chainjson.Int(0),
				Reverse:     chainjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("searchrawtransactions", "1Address", 0, 5, 10)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSearchRawTransactionsCmd("1Address",
					chainjson.Int(0), chainjson.Int(5), chainjson.Int(10), nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5,10],"id":1}`,
			unmarshalled: &chainjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     chainjson.Int(0),
				Skip:        chainjson.Int(5),
				Count:       chainjson.Int(10),
				VinExtra:    chainjson.Int(0),
				Reverse:     chainjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("searchrawtransactions", "1Address", 0, 5, 10, 1)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSearchRawTransactionsCmd("1Address",
					chainjson.Int(0), chainjson.Int(5), chainjson.Int(10), chainjson.Int(1), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5,10,1],"id":1}`,
			unmarshalled: &chainjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     chainjson.Int(0),
				Skip:        chainjson.Int(5),
				Count:       chainjson.Int(10),
				VinExtra:    chainjson.Int(1),
				Reverse:     chainjson.Bool(false),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("searchrawtransactions", "1Address", 0, 5, 10, 1, true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSearchRawTransactionsCmd("1Address",
					chainjson.Int(0), chainjson.Int(5), chainjson.Int(10), chainjson.Int(1), chainjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5,10,1,true],"id":1}`,
			unmarshalled: &chainjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     chainjson.Int(0),
				Skip:        chainjson.Int(5),
				Count:       chainjson.Int(10),
				VinExtra:    chainjson.Int(1),
				Reverse:     chainjson.Bool(true),
				FilterAddrs: nil,
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("searchrawtransactions", "1Address", 0, 5, 10, 1, true, []string{"1Address"})
			},
			staticCmd: func() interface{} {
				return chainjson.NewSearchRawTransactionsCmd("1Address",
					chainjson.Int(0), chainjson.Int(5), chainjson.Int(10), chainjson.Int(1), chainjson.Bool(true), &[]string{"1Address"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5,10,1,true,["1Address"]],"id":1}`,
			unmarshalled: &chainjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     chainjson.Int(0),
				Skip:        chainjson.Int(5),
				Count:       chainjson.Int(10),
				VinExtra:    chainjson.Int(1),
				Reverse:     chainjson.Bool(true),
				FilterAddrs: &[]string{"1Address"},
			},
		},
		{
			name: "searchrawtransactions",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("searchrawtransactions", "1Address", 0, 5, 10, "null", true, []string{"1Address"})
			},
			staticCmd: func() interface{} {
				return chainjson.NewSearchRawTransactionsCmd("1Address",
					chainjson.Int(0), chainjson.Int(5), chainjson.Int(10), nil, chainjson.Bool(true), &[]string{"1Address"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"searchrawtransactions","params":["1Address",0,5,10,null,true,["1Address"]],"id":1}`,
			unmarshalled: &chainjson.SearchRawTransactionsCmd{
				Address:     "1Address",
				Verbose:     chainjson.Int(0),
				Skip:        chainjson.Int(5),
				Count:       chainjson.Int(10),
				VinExtra:    nil,
				Reverse:     chainjson.Bool(true),
				FilterAddrs: &[]string{"1Address"},
			},
		},
		{
			name: "sendrawtransaction",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendrawtransaction", "1122", &chainjson.AllowHighFeesOrMaxFeeRate{})
			},
			staticCmd: func() interface{} {
				return chainjson.NewSendRawTransactionCmd("1122", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendrawtransaction","params":["1122",false],"id":1}`,
			unmarshalled: &chainjson.SendRawTransactionCmd{
				HexTx: "1122",
				FeeSetting: &chainjson.AllowHighFeesOrMaxFeeRate{
					Value: chainjson.Bool(false),
				},
			},
		},
		{
			name: "sendrawtransaction optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendrawtransaction", "1122", &chainjson.AllowHighFeesOrMaxFeeRate{Value: chainjson.Bool(false)})
			},
			staticCmd: func() interface{} {
				return chainjson.NewSendRawTransactionCmd("1122", chainjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendrawtransaction","params":["1122",false],"id":1}`,
			unmarshalled: &chainjson.SendRawTransactionCmd{
				HexTx: "1122",
				FeeSetting: &chainjson.AllowHighFeesOrMaxFeeRate{
					Value: chainjson.Bool(false),
				},
			},
		},
		{
			name: "sendrawtransaction optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendrawtransaction", "1122", &chainjson.AllowHighFeesOrMaxFeeRate{Value: chainjson.Float64(0.1234)})
			},
			staticCmd: func() interface{} {
				return chainjson.NewFlokicoindSendRawTransactionCmd("1122", 0.1234)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendrawtransaction","params":["1122",0.1234],"id":1}`,
			unmarshalled: &chainjson.SendRawTransactionCmd{
				HexTx: "1122",
				FeeSetting: &chainjson.AllowHighFeesOrMaxFeeRate{
					Value: chainjson.Float64(0.1234),
				},
			},
		},
		{
			name: "setgenerate",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("setgenerate", true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSetGenerateCmd(true, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"setgenerate","params":[true],"id":1}`,
			unmarshalled: &chainjson.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: chainjson.Int(-1),
			},
		},
		{
			name: "setgenerate optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("setgenerate", true, 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSetGenerateCmd(true, chainjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"setgenerate","params":[true,6],"id":1}`,
			unmarshalled: &chainjson.SetGenerateCmd{
				Generate:     true,
				GenProcLimit: chainjson.Int(6),
			},
		},
		{
			name: "signmessagewithprivkey",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signmessagewithprivkey", "5Hue", "Hey")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSignMessageWithPrivKey("5Hue", "Hey")
			},
			marshalled: `{"jsonrpc":"1.0","method":"signmessagewithprivkey","params":["5Hue","Hey"],"id":1}`,
			unmarshalled: &chainjson.SignMessageWithPrivKeyCmd{
				PrivKey: "5Hue",
				Message: "Hey",
			},
		},
		{
			name: "stop",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("stop")
			},
			staticCmd: func() interface{} {
				return chainjson.NewStopCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"stop","params":[],"id":1}`,
			unmarshalled: &chainjson.StopCmd{},
		},
		{
			name: "submitblock",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("submitblock", "112233")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSubmitBlockCmd("112233", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitblock","params":["112233"],"id":1}`,
			unmarshalled: &chainjson.SubmitBlockCmd{
				HexBlock: "112233",
				Options:  nil,
			},
		},
		{
			name: "submitblock optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("submitblock", "112233", `{"workid":"12345"}`)
			},
			staticCmd: func() interface{} {
				options := chainjson.SubmitBlockOptions{
					WorkID: "12345",
				}
				return chainjson.NewSubmitBlockCmd("112233", &options)
			},
			marshalled: `{"jsonrpc":"1.0","method":"submitblock","params":["112233",{"workid":"12345"}],"id":1}`,
			unmarshalled: &chainjson.SubmitBlockCmd{
				HexBlock: "112233",
				Options: &chainjson.SubmitBlockOptions{
					WorkID: "12345",
				},
			},
		},
		{
			name: "uptime",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("uptime")
			},
			staticCmd: func() interface{} {
				return chainjson.NewUptimeCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"uptime","params":[],"id":1}`,
			unmarshalled: &chainjson.UptimeCmd{},
		},
		{
			name: "validateaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("validateaddress", "1Address")
			},
			staticCmd: func() interface{} {
				return chainjson.NewValidateAddressCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"validateaddress","params":["1Address"],"id":1}`,
			unmarshalled: &chainjson.ValidateAddressCmd{
				Address: "1Address",
			},
		},
		{
			name: "verifychain",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("verifychain")
			},
			staticCmd: func() interface{} {
				return chainjson.NewVerifyChainCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifychain","params":[],"id":1}`,
			unmarshalled: &chainjson.VerifyChainCmd{
				CheckLevel: chainjson.Int32(3),
				CheckDepth: chainjson.Int32(288),
			},
		},
		{
			name: "verifychain optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("verifychain", 2)
			},
			staticCmd: func() interface{} {
				return chainjson.NewVerifyChainCmd(chainjson.Int32(2), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifychain","params":[2],"id":1}`,
			unmarshalled: &chainjson.VerifyChainCmd{
				CheckLevel: chainjson.Int32(2),
				CheckDepth: chainjson.Int32(288),
			},
		},
		{
			name: "verifychain optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("verifychain", 2, 500)
			},
			staticCmd: func() interface{} {
				return chainjson.NewVerifyChainCmd(chainjson.Int32(2), chainjson.Int32(500))
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifychain","params":[2,500],"id":1}`,
			unmarshalled: &chainjson.VerifyChainCmd{
				CheckLevel: chainjson.Int32(2),
				CheckDepth: chainjson.Int32(500),
			},
		},
		{
			name: "verifymessage",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("verifymessage", "1Address", "301234", "test")
			},
			staticCmd: func() interface{} {
				return chainjson.NewVerifyMessageCmd("1Address", "301234", "test")
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifymessage","params":["1Address","301234","test"],"id":1}`,
			unmarshalled: &chainjson.VerifyMessageCmd{
				Address:   "1Address",
				Signature: "301234",
				Message:   "test",
			},
		},
		{
			name: "verifytxoutproof",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("verifytxoutproof", "test")
			},
			staticCmd: func() interface{} {
				return chainjson.NewVerifyTxOutProofCmd("test")
			},
			marshalled: `{"jsonrpc":"1.0","method":"verifytxoutproof","params":["test"],"id":1}`,
			unmarshalled: &chainjson.VerifyTxOutProofCmd{
				Proof: "test",
			},
		},
		{
			name: "getdescriptorinfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getdescriptorinfo", "123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetDescriptorInfoCmd("123")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getdescriptorinfo","params":["123"],"id":1}`,
			unmarshalled: &chainjson.GetDescriptorInfoCmd{Descriptor: "123"},
		},
		{
			name: "getzmqnotifications",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getzmqnotifications")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetZmqNotificationsCmd()
			},

			marshalled:   `{"jsonrpc":"1.0","method":"getzmqnotifications","params":[],"id":1}`,
			unmarshalled: &chainjson.GetZmqNotificationsCmd{},
		},
		{
			name: "testmempoolaccept",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("testmempoolaccept", []string{"rawhex"}, 0.1)
			},
			staticCmd: func() interface{} {
				return chainjson.NewTestMempoolAcceptCmd([]string{"rawhex"}, 0.1)
			},
			marshalled: `{"jsonrpc":"1.0","method":"testmempoolaccept","params":[["rawhex"],0.1],"id":1}`,
			unmarshalled: &chainjson.TestMempoolAcceptCmd{
				RawTxns:    []string{"rawhex"},
				MaxFeeRate: 0.1,
			},
		},
		{
			name: "testmempoolaccept with maxfeerate",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("testmempoolaccept", []string{"rawhex"}, 0.01)
			},
			staticCmd: func() interface{} {
				return chainjson.NewTestMempoolAcceptCmd([]string{"rawhex"}, 0.01)
			},
			marshalled: `{"jsonrpc":"1.0","method":"testmempoolaccept","params":[["rawhex"],0.01],"id":1}`,
			unmarshalled: &chainjson.TestMempoolAcceptCmd{
				RawTxns:    []string{"rawhex"},
				MaxFeeRate: 0.01,
			},
		},
		{
			name: "gettxspendingprevout",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"gettxspendingprevout",
					[]*chainjson.GetTxSpendingPrevOutCmdOutput{
						{Txid: "0000000000000000000000000000000000000000000000000000000000000001", Vout: 0},
					})
			},
			staticCmd: func() interface{} {
				outputs := []wire.OutPoint{
					{Hash: chainhash.Hash{1}, Index: 0},
				}
				return chainjson.NewGetTxSpendingPrevOutCmd(outputs)
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettxspendingprevout","params":[[{"txid":"0000000000000000000000000000000000000000000000000000000000000001","vout":0}]],"id":1}`,
			unmarshalled: &chainjson.GetTxSpendingPrevOutCmd{
				Outputs: []*chainjson.GetTxSpendingPrevOutCmdOutput{{
					Txid: "0000000000000000000000000000000000000000000000000000000000000001",
					Vout: 0,
				}},
			},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Marshal the command as created by the new static command
		// creation function.
		marshalled, err := chainjson.MarshalCmd(chainjson.RpcVersion1, testID, test.staticCmd())
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			t.Errorf("\n%s\n%s", marshalled, test.marshalled)
			continue
		}

		// Ensure the command is created without error via the generic
		// new command creation function.
		cmd, err := test.newCmd()
		if err != nil {
			t.Errorf("Test #%d (%s) unexpected NewCmd error: %v ",
				i, test.name, err)
		}

		// Marshal the command as created by the generic new command
		// creation function.
		marshalled, err = chainjson.MarshalCmd(chainjson.RpcVersion1, testID, cmd)
		if err != nil {
			t.Errorf("MarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !bytes.Equal(marshalled, []byte(test.marshalled)) {
			t.Errorf("Test #%d (%s) unexpected marshalled data - "+
				"got %s, want %s", i, test.name, marshalled,
				test.marshalled)
			continue
		}

		var request chainjson.Request
		if err := json.Unmarshal(marshalled, &request); err != nil {
			t.Errorf("Test #%d (%s) unexpected error while "+
				"unmarshalling JSON-RPC request: %v", i,
				test.name, err)
			continue
		}

		cmd, err = chainjson.UnmarshalCmd(&request)
		if err != nil {
			t.Errorf("UnmarshalCmd #%d (%s) unexpected error: %v", i,
				test.name, err)
			continue
		}

		if !reflect.DeepEqual(cmd, test.unmarshalled) {
			t.Errorf("Test #%d (%s) unexpected unmarshalled command "+
				"- got %s, want %s", i, test.name,
				fmt.Sprintf("(%T) %+[1]v", cmd),
				fmt.Sprintf("(%T) %+[1]v\n", test.unmarshalled))
			continue
		}
	}
}

// TestChainSvrCmdErrors ensures any errors that occur in the command during
// custom mashal and unmarshal are as expected.
func TestChainSvrCmdErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     interface{}
		marshalled string
		err        error
	}{
		{
			name:       "template request with invalid type",
			result:     &chainjson.TemplateRequest{},
			marshalled: `{"mode":1}`,
			err:        &json.UnmarshalTypeError{},
		},
		{
			name:       "invalid template request sigoplimit field",
			result:     &chainjson.TemplateRequest{},
			marshalled: `{"sigoplimit":"invalid"}`,
			err:        chainjson.Error{ErrorCode: chainjson.ErrInvalidType},
		},
		{
			name:       "invalid template request sizelimit field",
			result:     &chainjson.TemplateRequest{},
			marshalled: `{"sizelimit":"invalid"}`,
			err:        chainjson.Error{ErrorCode: chainjson.ErrInvalidType},
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		err := json.Unmarshal([]byte(test.marshalled), &test.result)
		if reflect.TypeOf(err) != reflect.TypeOf(test.err) {
			t.Errorf("Test #%d (%s) wrong error - got %T (%v), "+
				"want %T", i, test.name, err, err, test.err)
			continue
		}

		if terr, ok := test.err.(chainjson.Error); ok {
			gotErrorCode := err.(chainjson.Error).ErrorCode
			if gotErrorCode != terr.ErrorCode {
				t.Errorf("Test #%d (%s) mismatched error code "+
					"- got %v (%v), want %v", i, test.name,
					gotErrorCode, terr, terr.ErrorCode)
				continue
			}
		}
	}
}
