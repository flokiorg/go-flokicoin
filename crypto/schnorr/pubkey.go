// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package schnorr

import (
	"fmt"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/flokiorg/go-flokicoin/crypto"
)

// These constants define the lengths of serialized public keys.
const (
	PubKeyBytesLen = 32
)

// ParsePubKey parses a public key for a koblitz curve from a bytestring into a
// crypto.Publickey, verifying that it is valid. It only supports public keys in
// the BIP-340 32-byte format.
func ParsePubKey(pubKeyStr []byte) (*crypto.PublicKey, error) {
	if pubKeyStr == nil {
		err := fmt.Errorf("nil pubkey byte string")
		return nil, err
	}
	if len(pubKeyStr) != PubKeyBytesLen {
		err := fmt.Errorf("bad pubkey byte string size (want %v, have %v)",
			PubKeyBytesLen, len(pubKeyStr))
		return nil, err
	}

	// We'll manually prepend the compressed byte so we can re-use the
	// existing pubkey parsing routine of the main crypto package.
	var keyCompressed [crypto.PubKeyBytesLenCompressed]byte
	keyCompressed[0] = secp.PubKeyFormatCompressedEven
	copy(keyCompressed[1:], pubKeyStr)

	return crypto.ParsePubKey(keyCompressed[:])
}

// SerializePubKey serializes a public key as specified by BIP 340. Public keys
// in this format are 32 bytes in length, and are assumed to have an even y
// coordinate.
func SerializePubKey(pub *crypto.PublicKey) []byte {
	pBytes := pub.SerializeCompressed()
	return pBytes[1:]
}
