// Copyright (c) 2024 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
	"io"
)

// MsgWTxIdRelay defines a flokicoin wtxidrelay message which is used for a peer
// to signal support for relaying witness transaction id (BIP141). It
// implements the Message interface.
//
// This message has no payload.
type MsgWTxIdRelay struct{}

// FlcDecode decodes r using the flokicoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgWTxIdRelay) FlcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	if pver < AddrV2Version {
		str := fmt.Sprintf("wtxidrelay message invalid for protocol "+
			"version %d", pver)
		return messageError("MsgWTxIdRelay.FlcDecode", str)
	}

	return nil
}

// FlcEncode encodes the receiver to w using the flokicoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgWTxIdRelay) FlcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	if pver < AddrV2Version {
		str := fmt.Sprintf("wtxidrelay message invalid for protocol "+
			"version %d", pver)
		return messageError("MsgWTxIdRelay.FlcEncode", str)
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgWTxIdRelay) Command() string {
	return CmdWTxIdRelay
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgWTxIdRelay) MaxPayloadLength(pver uint32) uint32 {
	return 0
}

// NewMsgWTxIdRelay returns a new flokicoin wtxidrelay message that conforms
// to the Message interface.
func NewMsgWTxIdRelay() *MsgWTxIdRelay {
	return &MsgWTxIdRelay{}
}
