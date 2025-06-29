// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"bytes"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

// TestGenesisBlock tests the genesis block of the main network for validity by
// checking the encoded bytes and hashes.
func TestGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := MainNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), genesisBlockBytes) {
		t.Fatalf("TestGenesisBlock: Genesis block does not appear valid - "+
			"got %v, want %v", spew.Sdump(buf.Bytes()),
			spew.Sdump(genesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := MainNetParams.GenesisBlock.BlockHash()
	if !MainNetParams.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestGenesisBlock: Genesis block hash does not "+
			"appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(MainNetParams.GenesisHash))
	}

}

// TestRegTestGenesisBlock tests the genesis block of the regression test
// network for validity by checking the encoded bytes and hashes.
func TestRegTestGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := RegressionNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestRegTestGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), regTestGenesisBlockBytes) {
		t.Fatalf("TestRegTestGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(regTestGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := RegressionNetParams.GenesisBlock.BlockHash()
	if !RegressionNetParams.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestRegTestGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(RegressionNetParams.GenesisHash))
	}
}

// TestTestNet3GenesisBlock tests the genesis block of the test network (version
// 3) for validity by checking the encoded bytes and hashes.
func TestTestNet3GenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := TestNet3Params.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestTestNet3GenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), testNet3GenesisBlockBytes) {
		t.Fatalf("TestTestNet3GenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(testNet3GenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := TestNet3Params.GenesisBlock.BlockHash()
	if !TestNet3Params.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestTestNet3GenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(TestNet3Params.GenesisHash))
	}
}

// TestSimNetGenesisBlock tests the genesis block of the simulation test network
// for validity by checking the encoded bytes and hashes.
func TestSimNetGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := SimNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestSimNetGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), simNetGenesisBlockBytes) {
		t.Fatalf("TestSimNetGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(simNetGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := SimNetParams.GenesisBlock.BlockHash()
	if !SimNetParams.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestSimNetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(SimNetParams.GenesisHash))
	}
}

// TestSigNetGenesisBlock tests the genesis block of the signet test network for
// validity by checking the encoded bytes and hashes.
func TestSigNetGenesisBlock(t *testing.T) {
	// Encode the genesis block to raw bytes.
	var buf bytes.Buffer
	err := SigNetParams.GenesisBlock.Serialize(&buf)
	if err != nil {
		t.Fatalf("TestSigNetGenesisBlock: %v", err)
	}

	// Ensure the encoded block matches the expected bytes.
	if !bytes.Equal(buf.Bytes(), sigNetGenesisBlockBytes) {
		t.Fatalf("TestSigNetGenesisBlock: Genesis block does not "+
			"appear valid - got %v, want %v",
			spew.Sdump(buf.Bytes()),
			spew.Sdump(sigNetGenesisBlockBytes))
	}

	// Check hash of the block against expected hash.
	hash := SigNetParams.GenesisBlock.BlockHash()
	if !SigNetParams.GenesisHash.IsEqual(&hash) {
		t.Fatalf("TestSigNetGenesisBlock: Genesis block hash does "+
			"not appear valid - got %v, want %v", spew.Sdump(hash),
			spew.Sdump(SigNetParams.GenesisHash))
	}
}

// genesisBlockBytes are the wire encoded bytes for the genesis block of the
// main network as of protocol version 60002.
var genesisBlockBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x04, 0x66, 0xbc, 0xf2, /* |.....f..| */
	0x29, 0x9e, 0x92, 0xab, 0x56, 0x85, 0x26, 0x62, /* |)...V.&b| */
	0x65, 0x80, 0x63, 0xde, 0xfd, 0x6e, 0x20, 0x92, /* |e.c..n .| */
	0x3f, 0xf4, 0xb3, 0x04, 0x53, 0xd1, 0x54, 0xb9, /* |?...S.T.| */
	0x88, 0x31, 0xfb, 0xdc, 0xaf, 0x7d, 0x3e, 0x61, /* |.1...}>a| */
	0xff, 0xff, 0x00, 0x1f, 0x27, 0xf0, 0x2d, 0x7c, /* |....'.-|| */
	0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, /* |........| */
	0xff, 0xff, 0x2d, 0x04, 0xff, 0xff, 0x00, 0x1d, /* |..-.....| */
	0x01, 0x04, 0x25, 0x54, 0x77, 0x69, 0x74, 0x74, /* |..%Twitt| */
	0x65, 0x72, 0x20, 0x31, 0x32, 0x2f, 0x53, 0x65, /* |er 12/Se| */
	0x70, 0x2f, 0x32, 0x30, 0x32, 0x31, 0x20, 0x46, /* |p/2021 F| */
	0x6c, 0x6f, 0x6b, 0x69, 0x20, 0x68, 0x61, 0x73, /* |loki has| */
	0x20, 0x61, 0x72, 0x72, 0x69, 0x76, 0x65, 0x64, /* | arrived| */
	0xff, 0xff, 0xff, 0xff, 0x01, 0x00, 0xe8, 0x76, /* |.......v| */
	0x48, 0x17, 0x00, 0x00, 0x00, 0x43, 0x41, 0x04, /* |H....CA.| */
	0x67, 0x8a, 0xfd, 0xb0, 0xfe, 0x55, 0x48, 0x27, /* |g....UH'| */
	0x19, 0x67, 0xf1, 0xa6, 0x71, 0x30, 0xb7, 0x10, /* |.g..q0..| */
	0x5c, 0xd6, 0xa8, 0x28, 0xe0, 0x39, 0x09, 0xa6, /* |\..(.9..| */
	0x79, 0x62, 0xe0, 0xea, 0x1f, 0x61, 0xde, 0xb6, /* |yb...a..| */
	0x49, 0xf6, 0xbc, 0x3f, 0x4c, 0xef, 0x38, 0xc4, /* |I..?L.8.| */
	0xf3, 0x55, 0x04, 0xe5, 0x1e, 0xc1, 0x12, 0xde, /* |.U......| */
	0x5c, 0x38, 0x4d, 0xf7, 0xba, 0x0b, 0x8d, 0x57, /* |\8M....W| */
	0x8a, 0x4c, 0x70, 0x2b, 0x6b, 0xf1, 0x1d, 0x5f, /* |.Lp+k.._| */
	0xac, 0x00, 0x00, 0x00, 0x00, /* |.....| */
}

// regTestGenesisBlockBytes are the wire encoded bytes for the genesis block of
// the regression test network as of protocol version 60002.
var regTestGenesisBlockBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x04, 0x66, 0xbc, 0xf2, /* |.....f..| */
	0x29, 0x9e, 0x92, 0xab, 0x56, 0x85, 0x26, 0x62, /* |)...V.&b| */
	0x65, 0x80, 0x63, 0xde, 0xfd, 0x6e, 0x20, 0x92, /* |e.c..n .| */
	0x3f, 0xf4, 0xb3, 0x04, 0x53, 0xd1, 0x54, 0xb9, /* |?...S.T.| */
	0x88, 0x31, 0xfb, 0xdc, 0xb6, 0xbc, 0x6f, 0x67, /* |.1....og| */
	0xff, 0xff, 0x7f, 0x20, 0x1e, 0xac, 0x2b, 0x7c, /* |... ..+|| */
	0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, /* |........| */
	0xff, 0xff, 0x2d, 0x04, 0xff, 0xff, 0x00, 0x1d, /* |..-.....| */
	0x01, 0x04, 0x25, 0x54, 0x77, 0x69, 0x74, 0x74, /* |..%Twitt| */
	0x65, 0x72, 0x20, 0x31, 0x32, 0x2f, 0x53, 0x65, /* |er 12/Se| */
	0x70, 0x2f, 0x32, 0x30, 0x32, 0x31, 0x20, 0x46, /* |p/2021 F| */
	0x6c, 0x6f, 0x6b, 0x69, 0x20, 0x68, 0x61, 0x73, /* |loki has| */
	0x20, 0x61, 0x72, 0x72, 0x69, 0x76, 0x65, 0x64, /* | arrived| */
	0xff, 0xff, 0xff, 0xff, 0x01, 0x00, 0xe8, 0x76, /* |.......v| */
	0x48, 0x17, 0x00, 0x00, 0x00, 0x43, 0x41, 0x04, /* |H....CA.| */
	0x67, 0x8a, 0xfd, 0xb0, 0xfe, 0x55, 0x48, 0x27, /* |g....UH'| */
	0x19, 0x67, 0xf1, 0xa6, 0x71, 0x30, 0xb7, 0x10, /* |.g..q0..| */
	0x5c, 0xd6, 0xa8, 0x28, 0xe0, 0x39, 0x09, 0xa6, /* |\..(.9..| */
	0x79, 0x62, 0xe0, 0xea, 0x1f, 0x61, 0xde, 0xb6, /* |yb...a..| */
	0x49, 0xf6, 0xbc, 0x3f, 0x4c, 0xef, 0x38, 0xc4, /* |I..?L.8.| */
	0xf3, 0x55, 0x04, 0xe5, 0x1e, 0xc1, 0x12, 0xde, /* |.U......| */
	0x5c, 0x38, 0x4d, 0xf7, 0xba, 0x0b, 0x8d, 0x57, /* |\8M....W| */
	0x8a, 0x4c, 0x70, 0x2b, 0x6b, 0xf1, 0x1d, 0x5f, /* |.Lp+k.._| */
	0xac, 0x00, 0x00, 0x00, 0x00, /* |.....| */
}

// testNet3GenesisBlockBytes are the wire encoded bytes for the genesis block of
// the test network (version 3) as of protocol version 60002.
var testNet3GenesisBlockBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x04, 0x66, 0xbc, 0xf2, /* |.....f..| */
	0x29, 0x9e, 0x92, 0xab, 0x56, 0x85, 0x26, 0x62, /* |)...V.&b| */
	0x65, 0x80, 0x63, 0xde, 0xfd, 0x6e, 0x20, 0x92, /* |e.c..n .| */
	0x3f, 0xf4, 0xb3, 0x04, 0x53, 0xd1, 0x54, 0xb9, /* |?...S.T.| */
	0x88, 0x31, 0xfb, 0xdc, 0xb6, 0xbc, 0x6f, 0x67, /* |.1....og| */
	0xff, 0xff, 0x7f, 0x20, 0x1e, 0xac, 0x2b, 0x7c, /* |... ..+|| */
	0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, /* |........| */
	0xff, 0xff, 0x2d, 0x04, 0xff, 0xff, 0x00, 0x1d, /* |..-.....| */
	0x01, 0x04, 0x25, 0x54, 0x77, 0x69, 0x74, 0x74, /* |..%Twitt| */
	0x65, 0x72, 0x20, 0x31, 0x32, 0x2f, 0x53, 0x65, /* |er 12/Se| */
	0x70, 0x2f, 0x32, 0x30, 0x32, 0x31, 0x20, 0x46, /* |p/2021 F| */
	0x6c, 0x6f, 0x6b, 0x69, 0x20, 0x68, 0x61, 0x73, /* |loki has| */
	0x20, 0x61, 0x72, 0x72, 0x69, 0x76, 0x65, 0x64, /* | arrived| */
	0xff, 0xff, 0xff, 0xff, 0x01, 0x00, 0xe8, 0x76, /* |.......v| */
	0x48, 0x17, 0x00, 0x00, 0x00, 0x43, 0x41, 0x04, /* |H....CA.| */
	0x67, 0x8a, 0xfd, 0xb0, 0xfe, 0x55, 0x48, 0x27, /* |g....UH'| */
	0x19, 0x67, 0xf1, 0xa6, 0x71, 0x30, 0xb7, 0x10, /* |.g..q0..| */
	0x5c, 0xd6, 0xa8, 0x28, 0xe0, 0x39, 0x09, 0xa6, /* |\..(.9..| */
	0x79, 0x62, 0xe0, 0xea, 0x1f, 0x61, 0xde, 0xb6, /* |yb...a..| */
	0x49, 0xf6, 0xbc, 0x3f, 0x4c, 0xef, 0x38, 0xc4, /* |I..?L.8.| */
	0xf3, 0x55, 0x04, 0xe5, 0x1e, 0xc1, 0x12, 0xde, /* |.U......| */
	0x5c, 0x38, 0x4d, 0xf7, 0xba, 0x0b, 0x8d, 0x57, /* |\8M....W| */
	0x8a, 0x4c, 0x70, 0x2b, 0x6b, 0xf1, 0x1d, 0x5f, /* |.Lp+k.._| */
	0xac, 0x00, 0x00, 0x00, 0x00, /* |.....| */
}

// simNetGenesisBlockBytes are the wire encoded bytes for the genesis block of
// the simulation test network as of protocol version 70002.
var simNetGenesisBlockBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x04, 0x66, 0xbc, 0xf2, /* |.....f..| */
	0x29, 0x9e, 0x92, 0xab, 0x56, 0x85, 0x26, 0x62, /* |)...V.&b| */
	0x65, 0x80, 0x63, 0xde, 0xfd, 0x6e, 0x20, 0x92, /* |e.c..n .| */
	0x3f, 0xf4, 0xb3, 0x04, 0x53, 0xd1, 0x54, 0xb9, /* |?...S.T.| */
	0x88, 0x31, 0xfb, 0xdc, 0xb6, 0xbc, 0x6f, 0x67, /* |.1....og| */
	0xff, 0xff, 0x7f, 0x20, 0x1e, 0xac, 0x2b, 0x7c, /* |... ..+|| */
	0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, /* |........| */
	0xff, 0xff, 0x2d, 0x04, 0xff, 0xff, 0x00, 0x1d, /* |..-.....| */
	0x01, 0x04, 0x25, 0x54, 0x77, 0x69, 0x74, 0x74, /* |..%Twitt| */
	0x65, 0x72, 0x20, 0x31, 0x32, 0x2f, 0x53, 0x65, /* |er 12/Se| */
	0x70, 0x2f, 0x32, 0x30, 0x32, 0x31, 0x20, 0x46, /* |p/2021 F| */
	0x6c, 0x6f, 0x6b, 0x69, 0x20, 0x68, 0x61, 0x73, /* |loki has| */
	0x20, 0x61, 0x72, 0x72, 0x69, 0x76, 0x65, 0x64, /* | arrived| */
	0xff, 0xff, 0xff, 0xff, 0x01, 0x00, 0xe8, 0x76, /* |.......v| */
	0x48, 0x17, 0x00, 0x00, 0x00, 0x43, 0x41, 0x04, /* |H....CA.| */
	0x67, 0x8a, 0xfd, 0xb0, 0xfe, 0x55, 0x48, 0x27, /* |g....UH'| */
	0x19, 0x67, 0xf1, 0xa6, 0x71, 0x30, 0xb7, 0x10, /* |.g..q0..| */
	0x5c, 0xd6, 0xa8, 0x28, 0xe0, 0x39, 0x09, 0xa6, /* |\..(.9..| */
	0x79, 0x62, 0xe0, 0xea, 0x1f, 0x61, 0xde, 0xb6, /* |yb...a..| */
	0x49, 0xf6, 0xbc, 0x3f, 0x4c, 0xef, 0x38, 0xc4, /* |I..?L.8.| */
	0xf3, 0x55, 0x04, 0xe5, 0x1e, 0xc1, 0x12, 0xde, /* |.U......| */
	0x5c, 0x38, 0x4d, 0xf7, 0xba, 0x0b, 0x8d, 0x57, /* |\8M....W| */
	0x8a, 0x4c, 0x70, 0x2b, 0x6b, 0xf1, 0x1d, 0x5f, /* |.Lp+k.._| */
	0xac, 0x00, 0x00, 0x00, 0x00, /* |.....| */
}

// sigNetGenesisBlockBytes are the wire encoded bytes for the genesis block of
// the signet test network as of protocol version 70002.
var sigNetGenesisBlockBytes = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x04, 0x66, 0xbc, 0xf2, /* |.....f..| */
	0x29, 0x9e, 0x92, 0xab, 0x56, 0x85, 0x26, 0x62, /* |)...V.&b| */
	0x65, 0x80, 0x63, 0xde, 0xfd, 0x6e, 0x20, 0x92, /* |e.c..n .| */
	0x3f, 0xf4, 0xb3, 0x04, 0x53, 0xd1, 0x54, 0xb9, /* |?...S.T.| */
	0x88, 0x31, 0xfb, 0xdc, 0xb6, 0xbc, 0x6f, 0x67, /* |.1....og| */
	0xff, 0xff, 0x7f, 0x20, 0x1e, 0xac, 0x2b, 0x7c, /* |... ..+|| */
	0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, /* |........| */
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, /* |........| */
	0xff, 0xff, 0x2d, 0x04, 0xff, 0xff, 0x00, 0x1d, /* |..-.....| */
	0x01, 0x04, 0x25, 0x54, 0x77, 0x69, 0x74, 0x74, /* |..%Twitt| */
	0x65, 0x72, 0x20, 0x31, 0x32, 0x2f, 0x53, 0x65, /* |er 12/Se| */
	0x70, 0x2f, 0x32, 0x30, 0x32, 0x31, 0x20, 0x46, /* |p/2021 F| */
	0x6c, 0x6f, 0x6b, 0x69, 0x20, 0x68, 0x61, 0x73, /* |loki has| */
	0x20, 0x61, 0x72, 0x72, 0x69, 0x76, 0x65, 0x64, /* | arrived| */
	0xff, 0xff, 0xff, 0xff, 0x01, 0x00, 0xe8, 0x76, /* |.......v| */
	0x48, 0x17, 0x00, 0x00, 0x00, 0x43, 0x41, 0x04, /* |H....CA.| */
	0x67, 0x8a, 0xfd, 0xb0, 0xfe, 0x55, 0x48, 0x27, /* |g....UH'| */
	0x19, 0x67, 0xf1, 0xa6, 0x71, 0x30, 0xb7, 0x10, /* |.g..q0..| */
	0x5c, 0xd6, 0xa8, 0x28, 0xe0, 0x39, 0x09, 0xa6, /* |\..(.9..| */
	0x79, 0x62, 0xe0, 0xea, 0x1f, 0x61, 0xde, 0xb6, /* |yb...a..| */
	0x49, 0xf6, 0xbc, 0x3f, 0x4c, 0xef, 0x38, 0xc4, /* |I..?L.8.| */
	0xf3, 0x55, 0x04, 0xe5, 0x1e, 0xc1, 0x12, 0xde, /* |.U......| */
	0x5c, 0x38, 0x4d, 0xf7, 0xba, 0x0b, 0x8d, 0x57, /* |\8M....W| */
	0x8a, 0x4c, 0x70, 0x2b, 0x6b, 0xf1, 0x1d, 0x5f, /* |.Lp+k.._| */
	0xac, 0x00, 0x00, 0x00, 0x00, /* |.....| */
}
