// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"
)

// MsgPing implements the Message interface and represents a flokicoin ping
// message.
//
// For versions BIP0031Version and earlier, it is used primarily to confirm
// that a connection is still valid.  A transmission error is typically
// interpreted as a closed connection and that the peer should be removed.
// For versions AFTER BIP0031Version it contains an identifier which can be
// returned in the pong message to determine network timing.
//
// The payload for this message just consists of a nonce used for identifying
// it later.
type MsgPing struct {
	// Unique value associated with message that is used to identify
	// specific ping message.
	Nonce uint64
}

// FlcDecode decodes r using the flokicoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgPing) FlcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	// There was no nonce for BIP0031Version and earlier.
	// NOTE: > is not a mistake here.  The BIP0031 was defined as AFTER
	// the version unlike most others.
	if pver > BIP0031Version {
		nonce, err := binarySerializer.Uint64(r, littleEndian)
		if err != nil {
			return err
		}
		msg.Nonce = nonce
	}

	return nil
}

// FlcEncode encodes the receiver to w using the flokicoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgPing) FlcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	// There was no nonce for BIP0031Version and earlier.
	// NOTE: > is not a mistake here.  The BIP0031 was defined as AFTER
	// the version unlike most others.
	if pver > BIP0031Version {
		err := binarySerializer.PutUint64(w, littleEndian, msg.Nonce)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgPing) Command() string {
	return CmdPing
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgPing) MaxPayloadLength(pver uint32) uint32 {
	plen := uint32(0)
	// There was no nonce for BIP0031Version and earlier.
	// NOTE: > is not a mistake here.  The BIP0031 was defined as AFTER
	// the version unlike most others.
	if pver > BIP0031Version {
		// Nonce 8 bytes.
		plen += 8
	}

	return plen
}

// NewMsgPing returns a new flokicoin ping message that conforms to the Message
// interface.  See MsgPing for details.
func NewMsgPing(nonce uint64) *MsgPing {
	return &MsgPing{
		Nonce: nonce,
	}
}
