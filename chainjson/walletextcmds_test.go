// Copyright (c) 2014 The btcsuite developers
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
)

// TestWalletdExtCmds tests all of the walletd extended commands marshal and
// unmarshal into valid results include handling of optional fields being
// omitted in the marshalled command, while optional fields with defaults have
// the default assigned on unmarshalled commands.
func TestWalletExtCmds(t *testing.T) {
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
			name: "createnewaccount",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("createnewaccount", "acct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewCreateNewAccountCmd("acct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"createnewaccount","params":["acct"],"id":1}`,
			unmarshalled: &chainjson.CreateNewAccountCmd{
				Account: "acct",
			},
		},
		{
			name: "dumpwallet",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("dumpwallet", "filename")
			},
			staticCmd: func() interface{} {
				return chainjson.NewDumpWalletCmd("filename")
			},
			marshalled: `{"jsonrpc":"1.0","method":"dumpwallet","params":["filename"],"id":1}`,
			unmarshalled: &chainjson.DumpWalletCmd{
				Filename: "filename",
			},
		},
		{
			name: "importaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("importaddress", "1Address", "")
			},
			staticCmd: func() interface{} {
				return chainjson.NewImportAddressCmd("1Address", "", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importaddress","params":["1Address",""],"id":1}`,
			unmarshalled: &chainjson.ImportAddressCmd{
				Address: "1Address",
				Rescan:  chainjson.Bool(true),
			},
		},
		{
			name: "importaddress optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("importaddress", "1Address", "acct", false)
			},
			staticCmd: func() interface{} {
				return chainjson.NewImportAddressCmd("1Address", "acct", chainjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"importaddress","params":["1Address","acct",false],"id":1}`,
			unmarshalled: &chainjson.ImportAddressCmd{
				Address: "1Address",
				Account: "acct",
				Rescan:  chainjson.Bool(false),
			},
		},
		{
			name: "importpubkey",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("importpubkey", "031234")
			},
			staticCmd: func() interface{} {
				return chainjson.NewImportPubKeyCmd("031234", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"importpubkey","params":["031234"],"id":1}`,
			unmarshalled: &chainjson.ImportPubKeyCmd{
				PubKey: "031234",
				Rescan: chainjson.Bool(true),
			},
		},
		{
			name: "importpubkey optional",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("importpubkey", "031234", false)
			},
			staticCmd: func() interface{} {
				return chainjson.NewImportPubKeyCmd("031234", chainjson.Bool(false))
			},
			marshalled: `{"jsonrpc":"1.0","method":"importpubkey","params":["031234",false],"id":1}`,
			unmarshalled: &chainjson.ImportPubKeyCmd{
				PubKey: "031234",
				Rescan: chainjson.Bool(false),
			},
		},
		{
			name: "importwallet",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("importwallet", "filename")
			},
			staticCmd: func() interface{} {
				return chainjson.NewImportWalletCmd("filename")
			},
			marshalled: `{"jsonrpc":"1.0","method":"importwallet","params":["filename"],"id":1}`,
			unmarshalled: &chainjson.ImportWalletCmd{
				Filename: "filename",
			},
		},
		{
			name: "renameaccount",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("renameaccount", "oldacct", "newacct")
			},
			staticCmd: func() interface{} {
				return chainjson.NewRenameAccountCmd("oldacct", "newacct")
			},
			marshalled: `{"jsonrpc":"1.0","method":"renameaccount","params":["oldacct","newacct"],"id":1}`,
			unmarshalled: &chainjson.RenameAccountCmd{
				OldAccount: "oldacct",
				NewAccount: "newacct",
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
