// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
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

// TestExtCmds tests all of the lokid extended commands marshal and unmarshal
// into valid results include handling of optional fields being omitted in the
// marshalled command, while optional fields with defaults have the default
// assigned on unmarshalled commands.
func TestExtCmds(t *testing.T) {
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
			name: "debuglevel",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("debuglevel", "trace")
			},
			staticCmd: func() interface{} {
				return chainjson.NewDebugLevelCmd("trace")
			},
			marshalled: `{"jsonrpc":"1.0","method":"debuglevel","params":["trace"],"id":1}`,
			unmarshalled: &chainjson.DebugLevelCmd{
				LevelSpec: "trace",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("node", chainjson.NRemove, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return chainjson.NewNodeCmd("remove", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["remove","1.1.1.1"],"id":1}`,
			unmarshalled: &chainjson.NodeCmd{
				SubCmd: chainjson.NRemove,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("node", chainjson.NDisconnect, "1.1.1.1")
			},
			staticCmd: func() interface{} {
				return chainjson.NewNodeCmd("disconnect", "1.1.1.1", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["disconnect","1.1.1.1"],"id":1}`,
			unmarshalled: &chainjson.NodeCmd{
				SubCmd: chainjson.NDisconnect,
				Target: "1.1.1.1",
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("node", chainjson.NConnect, "1.1.1.1", "perm")
			},
			staticCmd: func() interface{} {
				return chainjson.NewNodeCmd("connect", "1.1.1.1", chainjson.String("perm"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","perm"],"id":1}`,
			unmarshalled: &chainjson.NodeCmd{
				SubCmd:        chainjson.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: chainjson.String("perm"),
			},
		},
		{
			name: "node",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("node", chainjson.NConnect, "1.1.1.1", "temp")
			},
			staticCmd: func() interface{} {
				return chainjson.NewNodeCmd("connect", "1.1.1.1", chainjson.String("temp"))
			},
			marshalled: `{"jsonrpc":"1.0","method":"node","params":["connect","1.1.1.1","temp"],"id":1}`,
			unmarshalled: &chainjson.NodeCmd{
				SubCmd:        chainjson.NConnect,
				Target:        "1.1.1.1",
				ConnectSubCmd: chainjson.String("temp"),
			},
		},
		{
			name: "generate",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("generate", 1)
			},
			staticCmd: func() interface{} {
				return chainjson.NewGenerateCmd(1)
			},
			marshalled: `{"jsonrpc":"1.0","method":"generate","params":[1],"id":1}`,
			unmarshalled: &chainjson.GenerateCmd{
				NumBlocks: 1,
			},
		},
		{
			name: "generatetoaddress",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("generatetoaddress", 1, "1Address")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGenerateToAddressCmd(1, "1Address", nil)
			},
			marshalled: `{"jsonrpc":"1.0","method":"generatetoaddress","params":[1,"1Address"],"id":1}`,
			unmarshalled: &chainjson.GenerateToAddressCmd{
				NumBlocks: 1,
				Address:   "1Address",
				MaxTries: func() *int64 {
					var i int64 = 1000000
					return &i
				}(),
			},
		},
		{
			name: "getbestblock",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getbestblock")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetBestBlockCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getbestblock","params":[],"id":1}`,
			unmarshalled: &chainjson.GetBestBlockCmd{},
		},
		{
			name: "getcurrentnet",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getcurrentnet")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetCurrentNetCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"getcurrentnet","params":[],"id":1}`,
			unmarshalled: &chainjson.GetCurrentNetCmd{},
		},
		{
			name: "getheaders",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getheaders", []string{}, "")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetHeadersCmd(
					[]string{},
					"",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getheaders","params":[[],""],"id":1}`,
			unmarshalled: &chainjson.GetHeadersCmd{
				BlockLocators: []string{},
				HashStop:      "",
			},
		},
		{
			name: "getheaders - with arguments",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("getheaders", []string{"000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16", "0000000000000000026f4b7f56eef057b32167eb5ad9ff62006f1807b7336d10"}, "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7")
			},
			staticCmd: func() interface{} {
				return chainjson.NewGetHeadersCmd(
					[]string{
						"000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
						"0000000000000000026f4b7f56eef057b32167eb5ad9ff62006f1807b7336d10",
					},
					"000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
				)
			},
			marshalled: `{"jsonrpc":"1.0","method":"getheaders","params":[["000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16","0000000000000000026f4b7f56eef057b32167eb5ad9ff62006f1807b7336d10"],"000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7"],"id":1}`,
			unmarshalled: &chainjson.GetHeadersCmd{
				BlockLocators: []string{
					"000000000000000001f1739002418e2f9a84c47a4fd2a0eb7a787a6b7dc12f16",
					"0000000000000000026f4b7f56eef057b32167eb5ad9ff62006f1807b7336d10",
				},
				HashStop: "000000000000000000ba33b33e1fad70b69e234fc24414dd47113bff38f523f7",
			},
		},
		{
			name: "version",
			newCmd: func() (interface{}, error) {
				return chainjson.NewCmd("version")
			},
			staticCmd: func() interface{} {
				return chainjson.NewVersionCmd()
			},
			marshalled:   `{"jsonrpc":"1.0","method":"version","params":[],"id":1}`,
			unmarshalled: &chainjson.VersionCmd{},
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
