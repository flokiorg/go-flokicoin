// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// TestPing tests the MsgPing API against the latest protocol version.
func TestPing(t *testing.T) {
	pver := ProtocolVersion

	// Ensure we get the same nonce back out.
	nonce, err := RandomUint64()
	if err != nil {
		t.Errorf("RandomUint64: Error generating nonce: %v", err)
	}
	msg := NewMsgPing(nonce)
	if msg.Nonce != nonce {
		t.Errorf("NewMsgPing: wrong nonce - got %v, want %v",
			msg.Nonce, nonce)
	}

	// Ensure the command is expected value.
	wantCmd := "ping"
	if cmd := msg.Command(); cmd != wantCmd {
		t.Errorf("NewMsgPing: wrong command - got %v want %v",
			cmd, wantCmd)
	}

	// Ensure max payload is expected value for latest protocol version.
	wantPayload := uint32(8)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}
}

// TestPingBIP0031 tests the MsgPing API against the protocol version
// BIP0031Version.
func TestPingBIP0031(t *testing.T) {
	// Use the protocol version just prior to BIP0031Version changes.
	pver := BIP0031Version
	enc := BaseEncoding

	nonce, err := RandomUint64()
	if err != nil {
		t.Errorf("RandomUint64: Error generating nonce: %v", err)
	}
	msg := NewMsgPing(nonce)
	if msg.Nonce != nonce {
		t.Errorf("NewMsgPing: wrong nonce - got %v, want %v",
			msg.Nonce, nonce)
	}

	// Ensure max payload is expected value for old protocol version.
	wantPayload := uint32(0)
	maxPayload := msg.MaxPayloadLength(pver)
	if maxPayload != wantPayload {
		t.Errorf("MaxPayloadLength: wrong max payload length for "+
			"protocol version %d - got %v, want %v", pver,
			maxPayload, wantPayload)
	}

	// Test encode with old protocol version.
	var buf bytes.Buffer
	err = msg.FlcEncode(&buf, pver, enc)
	if err != nil {
		t.Errorf("encode of MsgPing failed %v err <%v>", msg, err)
	}

	// Test decode with old protocol version.
	readmsg := NewMsgPing(0)
	err = readmsg.FlcDecode(&buf, pver, enc)
	if err != nil {
		t.Errorf("decode of MsgPing failed [%v] err <%v>", buf, err)
	}

	// Since this protocol version doesn't support the nonce, make sure
	// it didn't get encoded and decoded back out.
	if msg.Nonce == readmsg.Nonce {
		t.Errorf("Should not get same nonce for protocol version %d", pver)
	}
}

// TestPingCrossProtocol tests the MsgPing API when encoding with the latest
// protocol version and decoding with BIP0031Version.
func TestPingCrossProtocol(t *testing.T) {
	nonce, err := RandomUint64()
	if err != nil {
		t.Errorf("RandomUint64: Error generating nonce: %v", err)
	}
	msg := NewMsgPing(nonce)
	if msg.Nonce != nonce {
		t.Errorf("NewMsgPing: wrong nonce - got %v, want %v",
			msg.Nonce, nonce)
	}

	// Encode with latest protocol version.
	var buf bytes.Buffer
	err = msg.FlcEncode(&buf, ProtocolVersion, BaseEncoding)
	if err != nil {
		t.Errorf("encode of MsgPing failed %v err <%v>", msg, err)
	}

	// Decode with old protocol version.
	readmsg := NewMsgPing(0)
	err = readmsg.FlcDecode(&buf, BIP0031Version, BaseEncoding)
	if err != nil {
		t.Errorf("decode of MsgPing failed [%v] err <%v>", buf, err)
	}

	// Since one of the protocol versions doesn't support the nonce, make
	// sure it didn't get encoded and decoded back out.
	if msg.Nonce == readmsg.Nonce {
		t.Error("Should not get same nonce for cross protocol")
	}
}

// TestPingWire tests the MsgPing wire encode and decode for various protocol
// versions.
func TestPingWire(t *testing.T) {
	tests := []struct {
		in   MsgPing         // Message to encode
		out  MsgPing         // Expected decoded message
		buf  []byte          // Wire encoding
		pver uint32          // Protocol version for wire encoding
		enc  MessageEncoding // Message encoding format
	}{
		// Latest protocol version.
		{
			MsgPing{Nonce: 123123}, // 0x1e0f3
			MsgPing{Nonce: 123123}, // 0x1e0f3
			[]byte{0xf3, 0xe0, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
			ProtocolVersion,
			BaseEncoding,
		},

		// Protocol version BIP0031Version+1
		{
			MsgPing{Nonce: 456456}, // 0x6f708
			MsgPing{Nonce: 456456}, // 0x6f708
			[]byte{0x08, 0xf7, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00},
			BIP0031Version + 1,
			BaseEncoding,
		},

		// Protocol version BIP0031Version
		{
			MsgPing{Nonce: 789789}, // 0xc0d1d
			MsgPing{Nonce: 0},      // No nonce for pver
			[]byte{},               // No nonce for pver
			BIP0031Version,
			BaseEncoding,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode the message to wire format.
		var buf bytes.Buffer
		err := test.in.FlcEncode(&buf, test.pver, test.enc)
		if err != nil {
			t.Errorf("FlcEncode #%d error %v", i, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.buf) {
			t.Errorf("FlcEncode #%d\n got: %s want: %s", i,
				spew.Sdump(buf.Bytes()), spew.Sdump(test.buf))
			continue
		}

		// Decode the message from wire format.
		var msg MsgPing
		rbuf := bytes.NewReader(test.buf)
		err = msg.FlcDecode(rbuf, test.pver, test.enc)
		if err != nil {
			t.Errorf("FlcDecode #%d error %v", i, err)
			continue
		}
		if !reflect.DeepEqual(msg, test.out) {
			t.Errorf("FlcDecode #%d\n got: %s want: %s", i,
				spew.Sdump(msg), spew.Sdump(test.out))
			continue
		}
	}
}

// TestPingWireErrors performs negative tests against wire encode and decode
// of MsgPing to confirm error paths work correctly.
func TestPingWireErrors(t *testing.T) {
	pver := ProtocolVersion

	tests := []struct {
		in       *MsgPing        // Value to encode
		buf      []byte          // Wire encoding
		pver     uint32          // Protocol version for wire encoding
		enc      MessageEncoding // Message encoding format
		max      int             // Max size of fixed buffer to induce errors
		writeErr error           // Expected write error
		readErr  error           // Expected read error
	}{
		// Latest protocol version with intentional read/write errors.
		{
			&MsgPing{Nonce: 123123}, // 0x1e0f3
			[]byte{0xf3, 0xe0, 0x01, 0x00},
			pver,
			BaseEncoding,
			2,
			io.ErrShortWrite,
			io.ErrUnexpectedEOF,
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		// Encode to wire format.
		w := newFixedWriter(test.max)
		err := test.in.FlcEncode(w, test.pver, test.enc)
		if err != test.writeErr {
			t.Errorf("FlcEncode #%d wrong error got: %v, want: %v",
				i, err, test.writeErr)
			continue
		}

		// Decode from wire format.
		var msg MsgPing
		r := newFixedReader(test.max, test.buf)
		err = msg.FlcDecode(r, test.pver, test.enc)
		if err != test.readErr {
			t.Errorf("FlcDecode #%d wrong error got: %v, want: %v",
				i, err, test.readErr)
			continue
		}
	}
}
