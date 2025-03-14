// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
	"io"
)

// MsgNotFound defines a flokicoin notfound message which is sent in response to
// a getdata message if any of the requested data in not available on the peer.
// Each message is limited to a maximum number of inventory vectors, which is
// currently 50,000.
//
// Use the AddInvVect function to build up the list of inventory vectors when
// sending a notfound message to another peer.
type MsgNotFound struct {
	InvList []*InvVect
}

// AddInvVect adds an inventory vector to the message.
func (msg *MsgNotFound) AddInvVect(iv *InvVect) error {
	if len(msg.InvList)+1 > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [max %v]",
			MaxInvPerMsg)
		return messageError("MsgNotFound.AddInvVect", str)
	}

	msg.InvList = append(msg.InvList, iv)
	return nil
}

// FlcDecode decodes r using the flokicoin protocol encoding into the receiver.
// This is part of the Message interface implementation.
func (msg *MsgNotFound) FlcDecode(r io.Reader, pver uint32, enc MessageEncoding) error {
	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)

	count, err := ReadVarIntBuf(r, pver, buf)
	if err != nil {
		return err
	}

	// Limit to max inventory vectors per message.
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgNotFound.FlcDecode", str)
	}

	// Create a contiguous slice of inventory vectors to deserialize into in
	// order to reduce the number of allocations.
	invList := make([]InvVect, count)
	msg.InvList = make([]*InvVect, 0, count)
	for i := uint64(0); i < count; i++ {
		iv := &invList[i]
		err := readInvVectBuf(r, pver, iv, buf)
		if err != nil {
			return err
		}
		msg.AddInvVect(iv)
	}

	return nil
}

// FlcEncode encodes the receiver to w using the flokicoin protocol encoding.
// This is part of the Message interface implementation.
func (msg *MsgNotFound) FlcEncode(w io.Writer, pver uint32, enc MessageEncoding) error {
	// Limit to max inventory vectors per message.
	count := len(msg.InvList)
	if count > MaxInvPerMsg {
		str := fmt.Sprintf("too many invvect in message [%v]", count)
		return messageError("MsgNotFound.FlcEncode", str)
	}

	buf := binarySerializer.Borrow()
	defer binarySerializer.Return(buf)

	err := WriteVarIntBuf(w, pver, uint64(count), buf)
	if err != nil {
		return err
	}

	for _, iv := range msg.InvList {
		err := writeInvVectBuf(w, pver, iv, buf)
		if err != nil {
			return err
		}
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the Message interface implementation.
func (msg *MsgNotFound) Command() string {
	return CmdNotFound
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the Message interface implementation.
func (msg *MsgNotFound) MaxPayloadLength(pver uint32) uint32 {
	// Max var int 9 bytes + max InvVects at 36 bytes each.
	// Num inventory vectors (varInt) + max allowed inventory vectors.
	return MaxVarIntPayload + (MaxInvPerMsg * maxInvVectPayload)
}

// NewMsgNotFound returns a new flokicoin notfound message that conforms to the
// Message interface.  See MsgNotFound for details.
func NewMsgNotFound() *MsgNotFound {
	return &MsgNotFound{
		InvList: make([]*InvVect, 0, defaultInvListAlloc),
	}
}
