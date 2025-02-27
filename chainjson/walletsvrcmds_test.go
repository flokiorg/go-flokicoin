// Copyright (c) 2014-2020 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainjson_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/flokiorg/go-flokicoin/chainjson"
	"github.com/flokiorg/go-flokicoin/chainutil"
)

// TestWalletSvrCmds tests all of the wallet server commands marshal and
// unmarshal into valid results include handling of optional fields being
// omitted in the marshalled command, while optional fields with defaults have
// the default assigned on unmarshalled commands.
func TestWalletSvrCmds(t *testing.T) {
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
			name: "addmultisigaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("addmultisigaddress", 2, []string{"031234", "035678"})
			},
			staticCmd: func() interface{} {
				keys := []string{"031234", "035678"}
				return chainjson.NewAddMultisigAddressCmd(2, keys, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"addmultisigaddress","params":[2,["031234","035678"]],"id":1}`,
			unmarshalled: &chainjson.AddMultisigAddressCmd{
				NRequired: 2,
				Keys:      []string{"031234", "035678"},
				Account:   nil,
			},
		},
		{
			name: "addmultisigaddress optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("addmultisigaddress", 2, []string{"031234", "035678"}, "test")
			},
			staticCmd: func() interface{} {
				keys := []string{"031234", "035678"}
				return chainjson.NewAddMultisigAddressCmd(2, keys, chainjson.String("test"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"addmultisigaddress","params":[2,["031234","035678"],"test"],"id":1}`,
			unmarshalled: &chainjson.AddMultisigAddressCmd{
				NRequired: 2,
				Keys:      []string{"031234", "035678"},
				Account:   chainjson.String("test"),
			},
		},
		{
			name: "createwallet",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("createwallet", "mywallet", true, true, "secret", true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewCreateWalletCmd("mywallet",
					chainjson.Bool(true), chainjson.Bool(true),
					chainjson.String("secret"), chainjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"createwallet","params":["mywallet",true,true,"secret",true],"id":1}`,
			unmarshalled: &chainjson.CreateWalletCmd{
				WalletName:         "mywallet",
				DisablePrivateKeys: chainjson.Bool(true),
				Blank:              chainjson.Bool(true),
				Passphrase:         chainjson.String("secret"),
				AvoidReuse:         chainjson.Bool(true),
			},
		},
		{
			name: "createwallet - optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("createwallet", "mywallet")
			},
			staticCmd: func() interface{} {
				return chainjson.NewCreateWalletCmd("mywallet",
					nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createwallet","params":["mywallet"],"id":1}`,
			unmarshalled: &chainjson.CreateWalletCmd{
				WalletName:         "mywallet",
				DisablePrivateKeys: chainjson.Bool(false),
				Blank:              chainjson.Bool(false),
				Passphrase:         chainjson.String(""),
				AvoidReuse:         chainjson.Bool(false),
			},
		},
		{
			name: "createwallet - optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("createwallet", "mywallet", "null", "null", "secret")
			},
			staticCmd: func() interface{} {
				return chainjson.NewCreateWalletCmd("mywallet",
					nil, nil, chainjson.String("secret"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createwallet","params":["mywallet",null,null,"secret"],"id":1}`,
			unmarshalled: &chainjson.CreateWalletCmd{
				WalletName:         "mywallet",
				DisablePrivateKeys: nil,
				Blank:              nil,
				Passphrase:         chainjson.String("secret"),
				AvoidReuse:         chainjson.Bool(false),
			},
		},
		{
			name: "addwitnessaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("addwitnessaddress", "1address")
			},
			staticCmd: func() interface{} {
				return chainjson.NewAddWitnessAddressCmd("1address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"addwitnessaddress","params":["1address"],"id":1}`,
			unmarshalled: &chainjson.AddWitnessAddressCmd{
				Address: "1address",
			},
		},
		{
			name: "backupwallet",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("backupwallet", "backup.dat")
			},
			staticCmd: func() interface{} {
				return chainjson.NewBackupWalletCmd("backup.dat")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"backupwallet","params":["backup.dat"],"id":1}`,
			unmarshalled: &chainjson.BackupWalletCmd{Destination: "backup.dat"},
		},
		{
			name: "loadwallet",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("loadwallet", "wallet.dat")
			},
			staticCmd: func() interface{} {
				return chainjson.NewLoadWalletCmd("wallet.dat")
			},
			marshalled:   `{"jsonrpc":"1.0","method":"loadwallet","params":["wallet.dat"],"id":1}`,
			unmarshalled: &chainjson.LoadWalletCmd{WalletName: "wallet.dat"},
		},
		{
			name: "unloadwallet",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("unloadwallet", "default")
			},
			staticCmd: func() interface{} {
				return chainjson.NewUnloadWalletCmd(chainjson.String("default"))
			},
			marshalled:   `{"jsonrpc":"1.0","method":"unloadwallet","params":["default"],"id":1}`,
			unmarshalled: &chainjson.UnloadWalletCmd{WalletName: chainjson.String("default")},
		},
		{name: "unloadwallet - nil arg",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("unloadwallet")
			},
			staticCmd: func() interface{} {
				return chainjson.NewUnloadWalletCmd(nil)
			},
			marshalled:   `{"jsonrpc":"1.0","method":"unloadwallet","params":[],"id":1}`,
			unmarshalled: &chainjson.UnloadWalletCmd{WalletName: nil},
		},
		{
			name: "createmultisig",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("createmultisig", 2, []string{"031234", "035678"})
			},
			staticCmd: func() interface{} {
				keys := []string{"031234", "035678"}
				return chainjson.NewCreateMultisigCmd(2, keys)
			},
			marshalled: `{"jsonrpc":"1.0","method":"createmultisig","params":[2,["031234","035678"]],"id":1}`,
			unmarshalled: &chainjson.CreateMultisigCmd{
				NRequired: 2,
				Keys:      []string{"031234", "035678"},
			},
		},
		{
			name: "dumpprivkey",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("dumpprivkey", "1Address")
			},
			staticCmd: func() interface{} {
				return chainjson.NewDumpPrivKeyCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"dumpprivkey","params":["1Address"],"id":1}`,
			unmarshalled: &chainjson.DumpPrivKeyCmd{
				Address: "1Address",
			},
		},
		{
			name: "encryptwallet",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("encryptwallet", "pass")
			},
			staticCmd: func() interface{} {
				return chainjson.NewEncryptWalletCmd("pass")
			},
			marshalled: `{"jsonrpc":"1.0","method":"encryptwallet","params":["pass"],"id":1}`,
			unmarshalled: &chainjson.EncryptWalletCmd{
				Passphrase: "pass",
			},
		},
		{
			name: "estimatefee",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("estimatefee", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewEstimateFeeCmd(6)
			},
			marshalled: `{"jsonrpc":"1.0","method":"estimatefee","params":[6],"id":1}`,
			unmarshalled: &chainjson.EstimateFeeCmd{
				NumBlocks: 6,
			},
		},
		{
			name: "estimatesmartfee - no mode",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("estimatesmartfee", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewEstimateSmartFeeCmd(6, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"estimatesmartfee","params":[6],"id":1}`,
			unmarshalled: &chainjson.EstimateSmartFeeCmd{
				ConfTarget:   6,
				EstimateMode: &chainjson.EstimateModeConservative,
			},
		},
		{
			name: "estimatesmartfee - economical mode",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("estimatesmartfee", 6, chainjson.EstimateModeEconomical)
			},
			staticCmd: func() interface{} {
				return chainjson.NewEstimateSmartFeeCmd(6, &chainjson.EstimateModeEconomical)
			},
			marshalled: `{"jsonrpc":"1.0","method":"estimatesmartfee","params":[6,"ECONOMICAL"],"id":1}`,
			unmarshalled: &chainjson.EstimateSmartFeeCmd{
				ConfTarget:   6,
				EstimateMode: &chainjson.EstimateModeEconomical,
			},
		},
		{
			name: "estimatepriority",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("estimatepriority", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewEstimatePriorityCmd(6)
			},
			marshalled: `{"jsonrpc":"1.0","method":"estimatepriority","params":[6],"id":1}`,
			unmarshalled: &chainjson.EstimatePriorityCmd{
				NumBlocks: 6,
			},
		},
		{
			name: "getaccount",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getaccount", "1Address")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetAccountCmd("1Address")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getaccount","params":["1Address"],"id":1}`,
			unmarshalled: &chainjson.GetAccountCmd{
				Address: "1Address",
			},
		},
		{
			name: "getaccountaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getaccountaddress", "acct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetAccountAddressCmd("acct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getaccountaddress","params":["acct"],"id":1}`,
			unmarshalled: &chainjson.GetAccountAddressCmd{
				Account: "acct",
			},
		},
		{
			name: "getaddressesbyaccount",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getaddressesbyaccount", "acct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetAddressesByAccountCmd("acct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getaddressesbyaccount","params":["acct"],"id":1}`,
			unmarshalled: &chainjson.GetAddressesByAccountCmd{
				Account: "acct",
			},
		},
		{
			name: "getaddressinfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getaddressinfo", "1234")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetAddressInfoCmd("1234")
			},
			marshalled: `{"jsonrpc":"1.0","method":"getaddressinfo","params":["1234"],"id":1}`,
			unmarshalled: &chainjson.GetAddressInfoCmd{
				Address: "1234",
			},
		},
		{
			name: "getbalance",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getbalance")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBalanceCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getbalance","params":[],"id":1}`,
			unmarshalled: &chainjson.GetBalanceCmd{
				Account: nil,
				MinConf: chainjson.Int(1),
			},
		},
		{
			name: "getbalance optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getbalance", "acct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBalanceCmd(chainjson.String("acct"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getbalance","params":["acct"],"id":1}`,
			unmarshalled: &chainjson.GetBalanceCmd{
				Account: chainjson.String("acct"),
				MinConf: chainjson.Int(1),
			},
		},
		{
			name: "getbalance optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getbalance", "acct", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBalanceCmd(chainjson.String("acct"), chainjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getbalance","params":["acct",6],"id":1}`,
			unmarshalled: &chainjson.GetBalanceCmd{
				Account: chainjson.String("acct"),
				MinConf: chainjson.Int(6),
			},
		},
		{
			name: "getbalances",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getbalances")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBalancesCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getbalances","params":[],"id":1}`,
			unmarshalled: &chainjson.GetBalancesCmd{},
		},
		{
			name: "getnewaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnewaddress")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNewAddressCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnewaddress","params":[],"id":1}`,
			unmarshalled: &chainjson.GetNewAddressCmd{
				Account:     nil,
				AddressType: nil,
			},
		},
		{
			name: "getnewaddress optional acct",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnewaddress", "acct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNewAddressCmd(chainjson.String("acct"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnewaddress","params":["acct"],"id":1}`,
			unmarshalled: &chainjson.GetNewAddressCmd{
				Account:     chainjson.String("acct"),
				AddressType: nil,
			},
		},
		{
			name: "getnewaddress optional acct and type",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getnewaddress", "acct", "legacy")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetNewAddressCmd(chainjson.String("acct"), chainjson.String("legacy"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getnewaddress","params":["acct","legacy"],"id":1}`,
			unmarshalled: &chainjson.GetNewAddressCmd{
				Account:     chainjson.String("acct"),
				AddressType: chainjson.String("legacy"),
			},
		},
		{
			name: "getrawchangeaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getrawchangeaddress")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetRawChangeAddressCmd(nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawchangeaddress","params":[],"id":1}`,
			unmarshalled: &chainjson.GetRawChangeAddressCmd{
				Account:     nil,
				AddressType: nil,
			},
		},
		{
			name: "getrawchangeaddress optional acct",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getrawchangeaddress", "acct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetRawChangeAddressCmd(chainjson.String("acct"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawchangeaddress","params":["acct"],"id":1}`,
			unmarshalled: &chainjson.GetRawChangeAddressCmd{
				Account:     chainjson.String("acct"),
				AddressType: nil,
			},
		},
		{
			name: "getrawchangeaddress optional acct and type",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getrawchangeaddress", "acct", "legacy")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetRawChangeAddressCmd(chainjson.String("acct"), chainjson.String("legacy"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getrawchangeaddress","params":["acct","legacy"],"id":1}`,
			unmarshalled: &chainjson.GetRawChangeAddressCmd{
				Account:     chainjson.String("acct"),
				AddressType: chainjson.String("legacy"),
			},
		},
		{
			name: "getreceivedbyaccount",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getreceivedbyaccount", "acct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetReceivedByAccountCmd("acct", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getreceivedbyaccount","params":["acct"],"id":1}`,
			unmarshalled: &chainjson.GetReceivedByAccountCmd{
				Account: "acct",
				MinConf: chainjson.Int(1),
			},
		},
		{
			name: "getreceivedbyaccount optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getreceivedbyaccount", "acct", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetReceivedByAccountCmd("acct", chainjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getreceivedbyaccount","params":["acct",6],"id":1}`,
			unmarshalled: &chainjson.GetReceivedByAccountCmd{
				Account: "acct",
				MinConf: chainjson.Int(6),
			},
		},
		{
			name: "getreceivedbyaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getreceivedbyaddress", "1Address")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetReceivedByAddressCmd("1Address", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getreceivedbyaddress","params":["1Address"],"id":1}`,
			unmarshalled: &chainjson.GetReceivedByAddressCmd{
				Address: "1Address",
				MinConf: chainjson.Int(1),
			},
		},
		{
			name: "getreceivedbyaddress optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getreceivedbyaddress", "1Address", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetReceivedByAddressCmd("1Address", chainjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"getreceivedbyaddress","params":["1Address",6],"id":1}`,
			unmarshalled: &chainjson.GetReceivedByAddressCmd{
				Address: "1Address",
				MinConf: chainjson.Int(6),
			},
		},
		{
			name: "gettransaction",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("gettransaction", "123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetTransactionCmd("123", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettransaction","params":["123"],"id":1}`,
			unmarshalled: &chainjson.GetTransactionCmd{
				Txid:             "123",
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "gettransaction optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("gettransaction", "123", true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetTransactionCmd("123", chainjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"gettransaction","params":["123",true],"id":1}`,
			unmarshalled: &chainjson.GetTransactionCmd{
				Txid:             "123",
				IncludeWatchOnly: chainjson.Bool(true),
			},
		},
		{
			name: "getwalletinfo",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getwalletinfo")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetWalletInfoCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getwalletinfo","params":[],"id":1}`,
			unmarshalled: &chainjson.GetWalletInfoCmd{},
		},
		{
			name: "importprivkey",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("importprivkey", "abc")
			},
			staticCmd: func() interface{} {
				return chainjson.NewImportPrivKeyCmd("abc", nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importprivkey","params":["abc"],"id":1}`,
			unmarshalled: &chainjson.ImportPrivKeyCmd{
				PrivKey: "abc",
				Label:   nil,
				Rescan:  chainjson.Bool(true),
			},
		},
		{
			name: "importprivkey optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("importprivkey", "abc", "label")
			},
			staticCmd: func() interface{} {
				return chainjson.NewImportPrivKeyCmd("abc", chainjson.String("label"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importprivkey","params":["abc","label"],"id":1}`,
			unmarshalled: &chainjson.ImportPrivKeyCmd{
				PrivKey: "abc",
				Label:   chainjson.String("label"),
				Rescan:  chainjson.Bool(true),
			},
		},
		{
			name: "importprivkey optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("importprivkey", "abc", "label", false)
			},
			staticCmd: func() interface{} {
				return chainjson.NewImportPrivKeyCmd("abc", chainjson.String("label"), chainjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"importprivkey","params":["abc","label",false],"id":1}`,
			unmarshalled: &chainjson.ImportPrivKeyCmd{
				PrivKey: "abc",
				Label:   chainjson.String("label"),
				Rescan:  chainjson.Bool(false),
			},
		},
		{
			name: "keypoolrefill",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("keypoolrefill")
			},
			staticCmd: func() interface{} {
				return chainjson.NewKeyPoolRefillCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"keypoolrefill","params":[],"id":1}`,
			unmarshalled: &chainjson.KeyPoolRefillCmd{
				NewSize: chainjson.Uint(100),
			},
		},
		{
			name: "keypoolrefill optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("keypoolrefill", 200)
			},
			staticCmd: func() interface{} {
				return chainjson.NewKeyPoolRefillCmd(chainjson.Uint(200))
			},
			marshalled: `{"jsonrpc":"1.0","method":"keypoolrefill","params":[200],"id":1}`,
			unmarshalled: &chainjson.KeyPoolRefillCmd{
				NewSize: chainjson.Uint(200),
			},
		},
		{
			name: "listaccounts",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listaccounts")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListAccountsCmd(nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listaccounts","params":[],"id":1}`,
			unmarshalled: &chainjson.ListAccountsCmd{
				MinConf: chainjson.Int(1),
			},
		},
		{
			name: "listaccounts optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listaccounts", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListAccountsCmd(chainjson.Int(6))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listaccounts","params":[6],"id":1}`,
			unmarshalled: &chainjson.ListAccountsCmd{
				MinConf: chainjson.Int(6),
			},
		},
		{
			name: "listaddressgroupings",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listaddressgroupings")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListAddressGroupingsCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"listaddressgroupings","params":[],"id":1}`,
			unmarshalled: &chainjson.ListAddressGroupingsCmd{},
		},
		{
			name: "listlockunspent",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listlockunspent")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListLockUnspentCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"listlockunspent","params":[],"id":1}`,
			unmarshalled: &chainjson.ListLockUnspentCmd{},
		},
		{
			name: "listreceivedbyaccount",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listreceivedbyaccount")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListReceivedByAccountCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaccount","params":[],"id":1}`,
			unmarshalled: &chainjson.ListReceivedByAccountCmd{
				MinConf:          chainjson.Int(1),
				IncludeEmpty:     chainjson.Bool(false),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaccount optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listreceivedbyaccount", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListReceivedByAccountCmd(chainjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaccount","params":[6],"id":1}`,
			unmarshalled: &chainjson.ListReceivedByAccountCmd{
				MinConf:          chainjson.Int(6),
				IncludeEmpty:     chainjson.Bool(false),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaccount optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listreceivedbyaccount", 6, true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListReceivedByAccountCmd(chainjson.Int(6), chainjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaccount","params":[6,true],"id":1}`,
			unmarshalled: &chainjson.ListReceivedByAccountCmd{
				MinConf:          chainjson.Int(6),
				IncludeEmpty:     chainjson.Bool(true),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaccount optional3",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listreceivedbyaccount", 6, true, false)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListReceivedByAccountCmd(chainjson.Int(6), chainjson.Bool(true), chainjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaccount","params":[6,true,false],"id":1}`,
			unmarshalled: &chainjson.ListReceivedByAccountCmd{
				MinConf:          chainjson.Int(6),
				IncludeEmpty:     chainjson.Bool(true),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listreceivedbyaddress")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListReceivedByAddressCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaddress","params":[],"id":1}`,
			unmarshalled: &chainjson.ListReceivedByAddressCmd{
				MinConf:          chainjson.Int(1),
				IncludeEmpty:     chainjson.Bool(false),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaddress optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listreceivedbyaddress", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListReceivedByAddressCmd(chainjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaddress","params":[6],"id":1}`,
			unmarshalled: &chainjson.ListReceivedByAddressCmd{
				MinConf:          chainjson.Int(6),
				IncludeEmpty:     chainjson.Bool(false),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaddress optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listreceivedbyaddress", 6, true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListReceivedByAddressCmd(chainjson.Int(6), chainjson.Bool(true), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaddress","params":[6,true],"id":1}`,
			unmarshalled: &chainjson.ListReceivedByAddressCmd{
				MinConf:          chainjson.Int(6),
				IncludeEmpty:     chainjson.Bool(true),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listreceivedbyaddress optional3",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listreceivedbyaddress", 6, true, false)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListReceivedByAddressCmd(chainjson.Int(6), chainjson.Bool(true), chainjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listreceivedbyaddress","params":[6,true,false],"id":1}`,
			unmarshalled: &chainjson.ListReceivedByAddressCmd{
				MinConf:          chainjson.Int(6),
				IncludeEmpty:     chainjson.Bool(true),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listsinceblock",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listsinceblock")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListSinceBlockCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listsinceblock","params":[],"id":1}`,
			unmarshalled: &chainjson.ListSinceBlockCmd{
				BlockHash:           nil,
				TargetConfirmations: chainjson.Int(1),
				IncludeWatchOnly:    chainjson.Bool(false),
			},
		},
		{
			name: "listsinceblock optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listsinceblock", "123")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListSinceBlockCmd(chainjson.String("123"), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listsinceblock","params":["123"],"id":1}`,
			unmarshalled: &chainjson.ListSinceBlockCmd{
				BlockHash:           chainjson.String("123"),
				TargetConfirmations: chainjson.Int(1),
				IncludeWatchOnly:    chainjson.Bool(false),
			},
		},
		{
			name: "listsinceblock optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listsinceblock", "123", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListSinceBlockCmd(chainjson.String("123"), chainjson.Int(6), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listsinceblock","params":["123",6],"id":1}`,
			unmarshalled: &chainjson.ListSinceBlockCmd{
				BlockHash:           chainjson.String("123"),
				TargetConfirmations: chainjson.Int(6),
				IncludeWatchOnly:    chainjson.Bool(false),
			},
		},
		{
			name: "listsinceblock optional3",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listsinceblock", "123", 6, true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListSinceBlockCmd(chainjson.String("123"), chainjson.Int(6), chainjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listsinceblock","params":["123",6,true],"id":1}`,
			unmarshalled: &chainjson.ListSinceBlockCmd{
				BlockHash:           chainjson.String("123"),
				TargetConfirmations: chainjson.Int(6),
				IncludeWatchOnly:    chainjson.Bool(true),
			},
		},
		{
			name: "listsinceblock pad null",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listsinceblock", "null", 1, false)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListSinceBlockCmd(nil, chainjson.Int(1), chainjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listsinceblock","params":[null,1,false],"id":1}`,
			unmarshalled: &chainjson.ListSinceBlockCmd{
				BlockHash:           nil,
				TargetConfirmations: chainjson.Int(1),
				IncludeWatchOnly:    chainjson.Bool(false),
			},
		},
		{
			name: "listtransactions",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listtransactions")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListTransactionsCmd(nil, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":[],"id":1}`,
			unmarshalled: &chainjson.ListTransactionsCmd{
				Account:          nil,
				Count:            chainjson.Int(10),
				From:             chainjson.Int(0),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listtransactions optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listtransactions", "acct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListTransactionsCmd(chainjson.String("acct"), nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":["acct"],"id":1}`,
			unmarshalled: &chainjson.ListTransactionsCmd{
				Account:          chainjson.String("acct"),
				Count:            chainjson.Int(10),
				From:             chainjson.Int(0),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listtransactions optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listtransactions", "acct", 20)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListTransactionsCmd(chainjson.String("acct"), chainjson.Int(20), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":["acct",20],"id":1}`,
			unmarshalled: &chainjson.ListTransactionsCmd{
				Account:          chainjson.String("acct"),
				Count:            chainjson.Int(20),
				From:             chainjson.Int(0),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listtransactions optional3",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listtransactions", "acct", 20, 1)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListTransactionsCmd(chainjson.String("acct"), chainjson.Int(20),
					chainjson.Int(1), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":["acct",20,1],"id":1}`,
			unmarshalled: &chainjson.ListTransactionsCmd{
				Account:          chainjson.String("acct"),
				Count:            chainjson.Int(20),
				From:             chainjson.Int(1),
				IncludeWatchOnly: chainjson.Bool(false),
			},
		},
		{
			name: "listtransactions optional4",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listtransactions", "acct", 20, 1, true)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListTransactionsCmd(chainjson.String("acct"), chainjson.Int(20),
					chainjson.Int(1), chainjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"listtransactions","params":["acct",20,1,true],"id":1}`,
			unmarshalled: &chainjson.ListTransactionsCmd{
				Account:          chainjson.String("acct"),
				Count:            chainjson.Int(20),
				From:             chainjson.Int(1),
				IncludeWatchOnly: chainjson.Bool(true),
			},
		},
		{
			name: "listunspent",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listunspent")
			},
			staticCmd: func() interface{} {
				return chainjson.NewListUnspentCmd(nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listunspent","params":[],"id":1}`,
			unmarshalled: &chainjson.ListUnspentCmd{
				MinConf:   chainjson.Int(1),
				MaxConf:   chainjson.Int(9999999),
				Addresses: nil,
			},
		},
		{
			name: "listunspent optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listunspent", 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListUnspentCmd(chainjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listunspent","params":[6],"id":1}`,
			unmarshalled: &chainjson.ListUnspentCmd{
				MinConf:   chainjson.Int(6),
				MaxConf:   chainjson.Int(9999999),
				Addresses: nil,
			},
		},
		{
			name: "listunspent optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listunspent", 6, 100)
			},
			staticCmd: func() interface{} {
				return chainjson.NewListUnspentCmd(chainjson.Int(6), chainjson.Int(100), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"listunspent","params":[6,100],"id":1}`,
			unmarshalled: &chainjson.ListUnspentCmd{
				MinConf:   chainjson.Int(6),
				MaxConf:   chainjson.Int(100),
				Addresses: nil,
			},
		},
		{
			name: "listunspent optional3",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("listunspent", 6, 100, []string{"1Address", "1Address2"})
			},
			staticCmd: func() interface{} {
				return chainjson.NewListUnspentCmd(chainjson.Int(6), chainjson.Int(100),
					&[]string{"1Address", "1Address2"})
			},
			marshalled: `{"jsonrpc":"1.0","method":"listunspent","params":[6,100,["1Address","1Address2"]],"id":1}`,
			unmarshalled: &chainjson.ListUnspentCmd{
				MinConf:   chainjson.Int(6),
				MaxConf:   chainjson.Int(100),
				Addresses: &[]string{"1Address", "1Address2"},
			},
		},
		{
			name: "lockunspent",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("lockunspent", true, `[{"txid":"123","vout":1}]`)
			},
			staticCmd: func() interface{} {
				txInputs := []chainjson.TransactionInput{
					{Txid: "123", Vout: 1},
				}
				return chainjson.NewLockUnspentCmd(true, txInputs)
			},
			marshalled: `{"jsonrpc":"1.0","method":"lockunspent","params":[true,[{"txid":"123","vout":1}]],"id":1}`,
			unmarshalled: &chainjson.LockUnspentCmd{
				Unlock: true,
				Transactions: []chainjson.TransactionInput{
					{Txid: "123", Vout: 1},
				},
			},
		},
		{
			name: "move",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("move", "from", "to", 0.5)
			},
			staticCmd: func() interface{} {
				return chainjson.NewMoveCmd("from", "to", 0.5, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"move","params":["from","to",0.5],"id":1}`,
			unmarshalled: &chainjson.MoveCmd{
				FromAccount: "from",
				ToAccount:   "to",
				Amount:      0.5,
				MinConf:     chainjson.Int(1),
				Comment:     nil,
			},
		},
		{
			name: "move optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("move", "from", "to", 0.5, 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewMoveCmd("from", "to", 0.5, chainjson.Int(6), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"move","params":["from","to",0.5,6],"id":1}`,
			unmarshalled: &chainjson.MoveCmd{
				FromAccount: "from",
				ToAccount:   "to",
				Amount:      0.5,
				MinConf:     chainjson.Int(6),
				Comment:     nil,
			},
		},
		{
			name: "move optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("move", "from", "to", 0.5, 6, "comment")
			},
			staticCmd: func() interface{} {
				return chainjson.NewMoveCmd("from", "to", 0.5, chainjson.Int(6), chainjson.String("comment"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"move","params":["from","to",0.5,6,"comment"],"id":1}`,
			unmarshalled: &chainjson.MoveCmd{
				FromAccount: "from",
				ToAccount:   "to",
				Amount:      0.5,
				MinConf:     chainjson.Int(6),
				Comment:     chainjson.String("comment"),
			},
		},
		{
			name: "sendfrom",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendfrom", "from", "1Address", 0.5)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSendFromCmd("from", "1Address", 0.5, nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendfrom","params":["from","1Address",0.5],"id":1}`,
			unmarshalled: &chainjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     chainjson.Int(1),
				Comment:     nil,
				CommentTo:   nil,
			},
		},
		{
			name: "sendfrom optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendfrom", "from", "1Address", 0.5, 6)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSendFromCmd("from", "1Address", 0.5, chainjson.Int(6), nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendfrom","params":["from","1Address",0.5,6],"id":1}`,
			unmarshalled: &chainjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     chainjson.Int(6),
				Comment:     nil,
				CommentTo:   nil,
			},
		},
		{
			name: "sendfrom optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendfrom", "from", "1Address", 0.5, 6, "comment")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSendFromCmd("from", "1Address", 0.5, chainjson.Int(6),
					chainjson.String("comment"), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendfrom","params":["from","1Address",0.5,6,"comment"],"id":1}`,
			unmarshalled: &chainjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     chainjson.Int(6),
				Comment:     chainjson.String("comment"),
				CommentTo:   nil,
			},
		},
		{
			name: "sendfrom optional3",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendfrom", "from", "1Address", 0.5, 6, "comment", "commentto")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSendFromCmd("from", "1Address", 0.5, chainjson.Int(6),
					chainjson.String("comment"), chainjson.String("commentto"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendfrom","params":["from","1Address",0.5,6,"comment","commentto"],"id":1}`,
			unmarshalled: &chainjson.SendFromCmd{
				FromAccount: "from",
				ToAddress:   "1Address",
				Amount:      0.5,
				MinConf:     chainjson.Int(6),
				Comment:     chainjson.String("comment"),
				CommentTo:   chainjson.String("commentto"),
			},
		},
		{
			name: "sendmany",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendmany", "from", `{"1Address":0.5}`)
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"1Address": 0.5}
				return chainjson.NewSendManyCmd("from", amounts, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendmany","params":["from",{"1Address":0.5}],"id":1}`,
			unmarshalled: &chainjson.SendManyCmd{
				FromAccount: "from",
				Amounts:     map[string]float64{"1Address": 0.5},
				MinConf:     chainjson.Int(1),
				Comment:     nil,
			},
		},
		{
			name: "sendmany optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendmany", "from", `{"1Address":0.5}`, 6)
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"1Address": 0.5}
				return chainjson.NewSendManyCmd("from", amounts, chainjson.Int(6), nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendmany","params":["from",{"1Address":0.5},6],"id":1}`,
			unmarshalled: &chainjson.SendManyCmd{
				FromAccount: "from",
				Amounts:     map[string]float64{"1Address": 0.5},
				MinConf:     chainjson.Int(6),
				Comment:     nil,
			},
		},
		{
			name: "sendmany optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendmany", "from", `{"1Address":0.5}`, 6, "comment")
			},
			staticCmd: func() interface{} {
				amounts := map[string]float64{"1Address": 0.5}
				return chainjson.NewSendManyCmd("from", amounts, chainjson.Int(6), chainjson.String("comment"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendmany","params":["from",{"1Address":0.5},6,"comment"],"id":1}`,
			unmarshalled: &chainjson.SendManyCmd{
				FromAccount: "from",
				Amounts:     map[string]float64{"1Address": 0.5},
				MinConf:     chainjson.Int(6),
				Comment:     chainjson.String("comment"),
			},
		},
		{
			name: "sendtoaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendtoaddress", "1Address", 0.5)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSendToAddressCmd("1Address", 0.5, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendtoaddress","params":["1Address",0.5],"id":1}`,
			unmarshalled: &chainjson.SendToAddressCmd{
				Address:   "1Address",
				Amount:    0.5,
				Comment:   nil,
				CommentTo: nil,
			},
		},
		{
			name: "sendtoaddress optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("sendtoaddress", "1Address", 0.5, "comment", "commentto")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSendToAddressCmd("1Address", 0.5, chainjson.String("comment"),
					chainjson.String("commentto"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"sendtoaddress","params":["1Address",0.5,"comment","commentto"],"id":1}`,
			unmarshalled: &chainjson.SendToAddressCmd{
				Address:   "1Address",
				Amount:    0.5,
				Comment:   chainjson.String("comment"),
				CommentTo: chainjson.String("commentto"),
			},
		},
		{
			name: "setaccount",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("setaccount", "1Address", "acct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSetAccountCmd("1Address", "acct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"setaccount","params":["1Address","acct"],"id":1}`,
			unmarshalled: &chainjson.SetAccountCmd{
				Address: "1Address",
				Account: "acct",
			},
		},
		{
			name: "settxfee",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("settxfee", 0.0001)
			},
			staticCmd: func() interface{} {
				return chainjson.NewSetTxFeeCmd(0.0001)
			},
			marshalled: `{"jsonrpc":"1.0","method":"settxfee","params":[0.0001],"id":1}`,
			unmarshalled: &chainjson.SetTxFeeCmd{
				Amount: 0.0001,
			},
		},
		{
			name: "signmessage",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signmessage", "1Address", "message")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSignMessageCmd("1Address", "message")
			},
			marshalled: `{"jsonrpc":"1.0","method":"signmessage","params":["1Address","message"],"id":1}`,
			unmarshalled: &chainjson.SignMessageCmd{
				Address: "1Address",
				Message: "message",
			},
		},
		{
			name: "signrawtransaction",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signrawtransaction", "001122")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSignRawTransactionCmd("001122", nil, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransaction","params":["001122"],"id":1}`,
			unmarshalled: &chainjson.SignRawTransactionCmd{
				RawTx:    "001122",
				Inputs:   nil,
				PrivKeys: nil,
				Flags:    chainjson.String("ALL"),
			},
		},
		{
			name: "signrawtransaction optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signrawtransaction", "001122", `[{"txid":"123","vout":1,"scriptPubKey":"00","redeemScript":"01"}]`)
			},
			staticCmd: func() interface{} {
				txInputs := []chainjson.RawTxInput{
					{
						Txid:         "123",
						Vout:         1,
						ScriptPubKey: "00",
						RedeemScript: "01",
					},
				}

				return chainjson.NewSignRawTransactionCmd("001122", &txInputs, nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransaction","params":["001122",[{"txid":"123","vout":1,"scriptPubKey":"00","redeemScript":"01"}]],"id":1}`,
			unmarshalled: &chainjson.SignRawTransactionCmd{
				RawTx: "001122",
				Inputs: &[]chainjson.RawTxInput{
					{
						Txid:         "123",
						Vout:         1,
						ScriptPubKey: "00",
						RedeemScript: "01",
					},
				},
				PrivKeys: nil,
				Flags:    chainjson.String("ALL"),
			},
		},
		{
			name: "signrawtransaction optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signrawtransaction", "001122", `[]`, `["abc"]`)
			},
			staticCmd: func() interface{} {
				txInputs := []chainjson.RawTxInput{}
				privKeys := []string{"abc"}
				return chainjson.NewSignRawTransactionCmd("001122", &txInputs, &privKeys, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransaction","params":["001122",[],["abc"]],"id":1}`,
			unmarshalled: &chainjson.SignRawTransactionCmd{
				RawTx:    "001122",
				Inputs:   &[]chainjson.RawTxInput{},
				PrivKeys: &[]string{"abc"},
				Flags:    chainjson.String("ALL"),
			},
		},
		{
			name: "signrawtransaction optional3",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signrawtransaction", "001122", `[]`, `[]`, "ALL")
			},
			staticCmd: func() interface{} {
				txInputs := []chainjson.RawTxInput{}
				privKeys := []string{}
				return chainjson.NewSignRawTransactionCmd("001122", &txInputs, &privKeys,
					chainjson.String("ALL"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransaction","params":["001122",[],[],"ALL"],"id":1}`,
			unmarshalled: &chainjson.SignRawTransactionCmd{
				RawTx:    "001122",
				Inputs:   &[]chainjson.RawTxInput{},
				PrivKeys: &[]string{},
				Flags:    chainjson.String("ALL"),
			},
		},
		{
			name: "signrawtransactionwithwallet",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signrawtransactionwithwallet", "001122")
			},
			staticCmd: func() interface{} {
				return chainjson.NewSignRawTransactionWithWalletCmd("001122", nil, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransactionwithwallet","params":["001122"],"id":1}`,
			unmarshalled: &chainjson.SignRawTransactionWithWalletCmd{
				RawTx:       "001122",
				Inputs:      nil,
				SigHashType: chainjson.String("ALL"),
			},
		},
		{
			name: "signrawtransactionwithwallet optional1",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signrawtransactionwithwallet", "001122", `[{"txid":"123","vout":1,"scriptPubKey":"00","redeemScript":"01","witnessScript":"02","amount":1.5}]`)
			},
			staticCmd: func() interface{} {
				txInputs := []chainjson.RawTxWitnessInput{
					{
						Txid:          "123",
						Vout:          1,
						ScriptPubKey:  "00",
						RedeemScript:  chainjson.String("01"),
						WitnessScript: chainjson.String("02"),
						Amount:        chainjson.Float64(1.5),
					},
				}

				return chainjson.NewSignRawTransactionWithWalletCmd("001122", &txInputs, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransactionwithwallet","params":["001122",[{"txid":"123","vout":1,"scriptPubKey":"00","redeemScript":"01","witnessScript":"02","amount":1.5}]],"id":1}`,
			unmarshalled: &chainjson.SignRawTransactionWithWalletCmd{
				RawTx: "001122",
				Inputs: &[]chainjson.RawTxWitnessInput{
					{
						Txid:          "123",
						Vout:          1,
						ScriptPubKey:  "00",
						RedeemScript:  chainjson.String("01"),
						WitnessScript: chainjson.String("02"),
						Amount:        chainjson.Float64(1.5),
					},
				},
				SigHashType: chainjson.String("ALL"),
			},
		},
		{
			name: "signrawtransactionwithwallet optional1 with blank fields in input",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signrawtransactionwithwallet", "001122", `[{"txid":"123","vout":1,"scriptPubKey":"00","redeemScript":"01"}]`)
			},
			staticCmd: func() interface{} {
				txInputs := []chainjson.RawTxWitnessInput{
					{
						Txid:         "123",
						Vout:         1,
						ScriptPubKey: "00",
						RedeemScript: chainjson.String("01"),
					},
				}

				return chainjson.NewSignRawTransactionWithWalletCmd("001122", &txInputs, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransactionwithwallet","params":["001122",[{"txid":"123","vout":1,"scriptPubKey":"00","redeemScript":"01"}]],"id":1}`,
			unmarshalled: &chainjson.SignRawTransactionWithWalletCmd{
				RawTx: "001122",
				Inputs: &[]chainjson.RawTxWitnessInput{
					{
						Txid:         "123",
						Vout:         1,
						ScriptPubKey: "00",
						RedeemScript: chainjson.String("01"),
					},
				},
				SigHashType: chainjson.String("ALL"),
			},
		},
		{
			name: "signrawtransactionwithwallet optional2",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("signrawtransactionwithwallet", "001122", `[]`, "ALL")
			},
			staticCmd: func() interface{} {
				txInputs := []chainjson.RawTxWitnessInput{}
				return chainjson.NewSignRawTransactionWithWalletCmd("001122", &txInputs, chainjson.String("ALL"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"signrawtransactionwithwallet","params":["001122",[],"ALL"],"id":1}`,
			unmarshalled: &chainjson.SignRawTransactionWithWalletCmd{
				RawTx:       "001122",
				Inputs:      &[]chainjson.RawTxWitnessInput{},
				SigHashType: chainjson.String("ALL"),
			},
		},
		{
			name: "walletlock",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("walletlock")
			},
			staticCmd: func() interface{} {
				return chainjson.NewWalletLockCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"walletlock","params":[],"id":1}`,
			unmarshalled: &chainjson.WalletLockCmd{},
		},
		{
			name: "walletpassphrase",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("walletpassphrase", "pass", 60)
			},
			staticCmd: func() interface{} {
				return chainjson.NewWalletPassphraseCmd("pass", 60)
			},
			marshalled: `{"jsonrpc":"1.0","method":"walletpassphrase","params":["pass",60],"id":1}`,
			unmarshalled: &chainjson.WalletPassphraseCmd{
				Passphrase: "pass",
				Timeout:    60,
			},
		},
		{
			name: "walletpassphrasechange",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("walletpassphrasechange", "old", "new")
			},
			staticCmd: func() interface{} {
				return chainjson.NewWalletPassphraseChangeCmd("old", "new")
			},
			marshalled: `{"jsonrpc":"1.0","method":"walletpassphrasechange","params":["old","new"],"id":1}`,
			unmarshalled: &chainjson.WalletPassphraseChangeCmd{
				OldPassphrase: "old",
				NewPassphrase: "new",
			},
		},
		{
			name: "importmulti with descriptor + options",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"importmulti",
					// Cannot use a native string, due to special types like timestamp.
					[]chainjson.ImportMultiRequest{
						{Descriptor: chainjson.String("123"), Timestamp: chainjson.TimestampOrNow{Value: 0}},
					},
					`{"rescan": true}`,
				)
			},
			staticCmd: func() interface{} {
				requests := []chainjson.ImportMultiRequest{
					{Descriptor: chainjson.String("123"), Timestamp: chainjson.TimestampOrNow{Value: 0}},
				}
				options := chainjson.ImportMultiOptions{Rescan: true}
				return chainjson.NewImportMultiCmd(requests, &options)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importmulti","params":[[{"desc":"123","timestamp":0}],{"rescan":true}],"id":1}`,
			unmarshalled: &chainjson.ImportMultiCmd{
				Requests: []chainjson.ImportMultiRequest{
					{
						Descriptor: chainjson.String("123"),
						Timestamp:  chainjson.TimestampOrNow{Value: 0},
					},
				},
				Options: &chainjson.ImportMultiOptions{Rescan: true},
			},
		},
		{
			name: "importmulti with descriptor + no options",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"importmulti",
					// Cannot use a native string, due to special types like timestamp.
					[]chainjson.ImportMultiRequest{
						{
							Descriptor: chainjson.String("123"),
							Timestamp:  chainjson.TimestampOrNow{Value: 0},
							WatchOnly:  chainjson.Bool(false),
							Internal:   chainjson.Bool(true),
							Label:      chainjson.String("aaa"),
							KeyPool:    chainjson.Bool(false),
						},
					},
				)
			},
			staticCmd: func() interface{} {
				requests := []chainjson.ImportMultiRequest{
					{
						Descriptor: chainjson.String("123"),
						Timestamp:  chainjson.TimestampOrNow{Value: 0},
						WatchOnly:  chainjson.Bool(false),
						Internal:   chainjson.Bool(true),
						Label:      chainjson.String("aaa"),
						KeyPool:    chainjson.Bool(false),
					},
				}
				return chainjson.NewImportMultiCmd(requests, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importmulti","params":[[{"desc":"123","timestamp":0,"internal":true,"watchonly":false,"label":"aaa","keypool":false}]],"id":1}`,
			unmarshalled: &chainjson.ImportMultiCmd{
				Requests: []chainjson.ImportMultiRequest{
					{
						Descriptor: chainjson.String("123"),
						Timestamp:  chainjson.TimestampOrNow{Value: 0},
						WatchOnly:  chainjson.Bool(false),
						Internal:   chainjson.Bool(true),
						Label:      chainjson.String("aaa"),
						KeyPool:    chainjson.Bool(false),
					},
				},
			},
		},
		{
			name: "importmulti with descriptor + string timestamp",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"importmulti",
					// Cannot use a native string, due to special types like timestamp.
					[]chainjson.ImportMultiRequest{
						{
							Descriptor: chainjson.String("123"),
							Timestamp:  chainjson.TimestampOrNow{Value: "now"},
						},
					},
				)
			},
			staticCmd: func() interface{} {
				requests := []chainjson.ImportMultiRequest{
					{Descriptor: chainjson.String("123"), Timestamp: chainjson.TimestampOrNow{Value: "now"}},
				}
				return chainjson.NewImportMultiCmd(requests, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importmulti","params":[[{"desc":"123","timestamp":"now"}]],"id":1}`,
			unmarshalled: &chainjson.ImportMultiCmd{
				Requests: []chainjson.ImportMultiRequest{
					{Descriptor: chainjson.String("123"), Timestamp: chainjson.TimestampOrNow{Value: "now"}},
				},
			},
		},
		{
			name: "importmulti with scriptPubKey script",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"importmulti",
					// Cannot use a native string, due to special types like timestamp and scriptPubKey
					[]chainjson.ImportMultiRequest{
						{
							ScriptPubKey: &chainjson.ScriptPubKey{Value: "script"},
							RedeemScript: chainjson.String("123"),
							Timestamp:    chainjson.TimestampOrNow{Value: 0},
							PubKeys:      &[]string{"aaa"},
						},
					},
				)
			},
			staticCmd: func() interface{} {
				requests := []chainjson.ImportMultiRequest{
					{
						ScriptPubKey: &chainjson.ScriptPubKey{Value: "script"},
						RedeemScript: chainjson.String("123"),
						Timestamp:    chainjson.TimestampOrNow{Value: 0},
						PubKeys:      &[]string{"aaa"},
					},
				}
				return chainjson.NewImportMultiCmd(requests, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importmulti","params":[[{"scriptPubKey":"script","timestamp":0,"redeemscript":"123","pubkeys":["aaa"]}]],"id":1}`,
			unmarshalled: &chainjson.ImportMultiCmd{
				Requests: []chainjson.ImportMultiRequest{
					{
						ScriptPubKey: &chainjson.ScriptPubKey{Value: "script"},
						RedeemScript: chainjson.String("123"),
						Timestamp:    chainjson.TimestampOrNow{Value: 0},
						PubKeys:      &[]string{"aaa"},
					},
				},
			},
		},
		{
			name: "importmulti with scriptPubKey address",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"importmulti",
					// Cannot use a native string, due to special types like timestamp and scriptPubKey
					[]chainjson.ImportMultiRequest{
						{
							ScriptPubKey:  &chainjson.ScriptPubKey{Value: chainjson.ScriptPubKeyAddress{Address: "addr"}},
							WitnessScript: chainjson.String("123"),
							Timestamp:     chainjson.TimestampOrNow{Value: 0},
							Keys:          &[]string{"aaa"},
						},
					},
				)
			},
			staticCmd: func() interface{} {
				requests := []chainjson.ImportMultiRequest{
					{
						ScriptPubKey:  &chainjson.ScriptPubKey{Value: chainjson.ScriptPubKeyAddress{Address: "addr"}},
						WitnessScript: chainjson.String("123"),
						Timestamp:     chainjson.TimestampOrNow{Value: 0},
						Keys:          &[]string{"aaa"},
					},
				}
				return chainjson.NewImportMultiCmd(requests, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importmulti","params":[[{"scriptPubKey":{"address":"addr"},"timestamp":0,"witnessscript":"123","keys":["aaa"]}]],"id":1}`,
			unmarshalled: &chainjson.ImportMultiCmd{
				Requests: []chainjson.ImportMultiRequest{
					{
						ScriptPubKey:  &chainjson.ScriptPubKey{Value: chainjson.ScriptPubKeyAddress{Address: "addr"}},
						WitnessScript: chainjson.String("123"),
						Timestamp:     chainjson.TimestampOrNow{Value: 0},
						Keys:          &[]string{"aaa"},
					},
				},
			},
		},
		{
			name: "importmulti with ranged (int) descriptor",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"importmulti",
					// Cannot use a native string, due to special types like timestamp.
					[]chainjson.ImportMultiRequest{
						{
							Descriptor: chainjson.String("123"),
							Timestamp:  chainjson.TimestampOrNow{Value: 0},
							Range:      &chainjson.DescriptorRange{Value: 7},
						},
					},
				)
			},
			staticCmd: func() interface{} {
				requests := []chainjson.ImportMultiRequest{
					{
						Descriptor: chainjson.String("123"),
						Timestamp:  chainjson.TimestampOrNow{Value: 0},
						Range:      &chainjson.DescriptorRange{Value: 7},
					},
				}
				return chainjson.NewImportMultiCmd(requests, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importmulti","params":[[{"desc":"123","timestamp":0,"range":7}]],"id":1}`,
			unmarshalled: &chainjson.ImportMultiCmd{
				Requests: []chainjson.ImportMultiRequest{
					{
						Descriptor: chainjson.String("123"),
						Timestamp:  chainjson.TimestampOrNow{Value: 0},
						Range:      &chainjson.DescriptorRange{Value: 7},
					},
				},
			},
		},
		{
			name: "importmulti with ranged (slice) descriptor",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"importmulti",
					// Cannot use a native string, due to special types like timestamp.
					[]chainjson.ImportMultiRequest{
						{
							Descriptor: chainjson.String("123"),
							Timestamp:  chainjson.TimestampOrNow{Value: 0},
							Range:      &chainjson.DescriptorRange{Value: []int{1, 7}},
						},
					},
				)
			},
			staticCmd: func() interface{} {
				requests := []chainjson.ImportMultiRequest{
					{
						Descriptor: chainjson.String("123"),
						Timestamp:  chainjson.TimestampOrNow{Value: 0},
						Range:      &chainjson.DescriptorRange{Value: []int{1, 7}},
					},
				}
				return chainjson.NewImportMultiCmd(requests, nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importmulti","params":[[{"desc":"123","timestamp":0,"range":[1,7]}]],"id":1}`,
			unmarshalled: &chainjson.ImportMultiCmd{
				Requests: []chainjson.ImportMultiRequest{
					{
						Descriptor: chainjson.String("123"),
						Timestamp:  chainjson.TimestampOrNow{Value: 0},
						Range:      &chainjson.DescriptorRange{Value: []int{1, 7}},
					},
				},
			},
		},
		{
			name: "walletcreatefundedpsbt",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"walletcreatefundedpsbt",
					[]chainjson.PsbtInput{
						{
							Txid:     "1234",
							Vout:     0,
							Sequence: 0,
						},
					},
					[]chainjson.PsbtOutput{
						chainjson.NewPsbtOutput("1234", chainutil.Amount(1234)),
						chainjson.NewPsbtDataOutput([]byte{1, 2, 3, 4}),
					},
					chainjson.Uint32(1),
					chainjson.WalletCreateFundedPsbtOpts{},
					chainjson.Bool(true),
				)
			},
			staticCmd: func() interface{} {
				return chainjson.NewWalletCreateFundedPsbtCmd(
					[]chainjson.PsbtInput{
						{
							Txid:     "1234",
							Vout:     0,
							Sequence: 0,
						},
					},
					[]chainjson.PsbtOutput{
						chainjson.NewPsbtOutput("1234", chainutil.Amount(1234)),
						chainjson.NewPsbtDataOutput([]byte{1, 2, 3, 4}),
					},
					chainjson.Uint32(1),
					&chainjson.WalletCreateFundedPsbtOpts{},
					chainjson.Bool(true),
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"walletcreatefundedpsbt","params":[[{"txid":"1234","vout":0,"sequence":0}],[{"1234":0.00001234},{"data":"01020304"}],1,{},true],"id":1}`,
			unmarshalled: &chainjson.WalletCreateFundedPsbtCmd{
				Inputs: []chainjson.PsbtInput{
					{
						Txid:     "1234",
						Vout:     0,
						Sequence: 0,
					},
				},
				Outputs: []chainjson.PsbtOutput{
					chainjson.NewPsbtOutput("1234", chainutil.Amount(1234)),
					chainjson.NewPsbtDataOutput([]byte{1, 2, 3, 4}),
				},
				Locktime:    chainjson.Uint32(1),
				Options:     &chainjson.WalletCreateFundedPsbtOpts{},
				Bip32Derivs: chainjson.Bool(true),
			},
		},
		{
			name: "walletprocesspsbt",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd(
					"walletprocesspsbt", "1234", chainjson.Bool(true), chainjson.String("ALL"), chainjson.Bool(true))
			},
			staticCmd: func() interface{} {
				return chainjson.NewWalletProcessPsbtCmd(
					"1234", chainjson.Bool(true), chainjson.String("ALL"), chainjson.Bool(true))
			},
			marshalled: `{"jsonrpc":"1.0","method":"walletprocesspsbt","params":["1234",true,"ALL",true],"id":1}`,
			unmarshalled: &chainjson.WalletProcessPsbtCmd{
				Psbt:        "1234",
				Sign:        chainjson.Bool(true),
				SighashType: chainjson.String("ALL"),
				Bip32Derivs: chainjson.Bool(true),
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
