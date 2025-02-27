// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"encoding/hex"
	"time"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/wire"
)

func generateGenesisCoinbaseTx() *wire.MsgTx {
	pszTimestamp := "Twitter 12/Sep/2021 Floki has arrived"
	pszTimestampBytes := []byte(pszTimestamp)
	timestampLength := byte(len(pszTimestampBytes))

	// Construct the coinbase script signature
	coinbaseScriptSig := append(
		[]byte{0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04},         // Initial script prefix
		append([]byte{timestampLength}, pszTimestampBytes...)..., // Length prefix + pszTimestamp
	)

	genesisReward := int64(1000 * 1e8)

	genesisOutputScript := "4104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac"
	outputScript, _ := hex.DecodeString(genesisOutputScript)

	return &wire.MsgTx{
		Version: 1,
		TxIn: []*wire.TxIn{
			{
				PreviousOutPoint: wire.OutPoint{
					Hash:  chainhash.Hash{},
					Index: 0xffffffff,
				},
				SignatureScript: coinbaseScriptSig,
				Sequence:        0xffffffff,
			},
		},
		TxOut: []*wire.TxOut{
			{
				Value:    genesisReward,
				PkScript: outputScript,
			},
		},
		LockTime: 0,
	}
}

// c99abcc69738ab0bb2e65d6e08c48a4fae0f529bbaf1aa6bfb55b12d6d8ec47e
var mainGenesisHash = chainhash.Hash([chainhash.HashSize]byte{
	0x7e, 0xc4, 0x8e, 0x6d, 0x2d, 0xb1, 0x55, 0xfb,
	0x6b, 0xaa, 0xf1, 0xba, 0x9b, 0x52, 0x0f, 0xae,
	0x4f, 0x8a, 0xc4, 0x08, 0x6e, 0x5d, 0xe6, 0xb2,
	0x0b, 0xab, 0x38, 0x97, 0xc6, 0xbc, 0x9a, 0xc9,
})

// dcfb3188b954d15304b3f43f92206efdde63806562268556ab929e29f2bc6604
var mainGenesisMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{
	0x04, 0x66, 0xbc, 0xf2, 0x29, 0x9e, 0x92, 0xab,
	0x56, 0x85, 0x26, 0x62, 0x65, 0x80, 0x63, 0xde,
	0xfd, 0x6e, 0x20, 0x92, 0x3f, 0xf4, 0xb3, 0x04,
	0x53, 0xd1, 0x54, 0xb9, 0x88, 0x31, 0xfb, 0xdc,
})

// mainGenesisBlock defines the genesis block of the block chain which serves as the
// public transaction ledger for the main network.
var mainGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},         // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: mainGenesisMerkleRoot,    // dcfb3188b954d15304b3f43f92206efdde63806562268556ab929e29f2bc6604
		Timestamp:  time.Unix(1631485359, 0), // 2021-09-12 22:22:39 +0000 UTC
		Bits:       0x1e56b940,               // 486604799 [00000000ffff0000000000000000000000000000000000000000000000000000]
		Nonce:      2083390567,               // 2083390567
	},
	Transactions: []*wire.MsgTx{generateGenesisCoinbaseTx()},
}

// regTestGenesisHash is the hash of the first block in the block chain for the
// regression test network (genesis block).
var regTestGenesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
	0x1c, 0xec, 0xed, 0x25, 0xae, 0xe2, 0x1f, 0x66,
	0xba, 0x75, 0xd9, 0x55, 0x24, 0x66, 0xf1, 0xd7,
	0x18, 0x7f, 0x21, 0x64, 0x72, 0x15, 0xa9, 0x3c,
	0x56, 0x98, 0x9d, 0x92, 0xff, 0xec, 0x35, 0xfe,
}) // fe35ecff929d98563ca9157264217f18d7f1662455d975ba661fe2ae25edec1c

// regTestGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the regression test network.  It is the same as the merkle root for
// the main network.
var regTestGenesisMerkleRoot = mainGenesisMerkleRoot

// regTestGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the regression test network.
var regTestGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},         // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: regTestGenesisMerkleRoot, // merkleroot
		Timestamp:  time.Unix(1735376054, 0), // 28 Dec 2024 08:54:14 +0000 UTC
		Bits:       0x207fffff,               // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:      2083236894,
	},
	Transactions: []*wire.MsgTx{generateGenesisCoinbaseTx()},
}

// testNet3GenesisHash is the hash of the first block in the block chain for the
// test network (version 3).
var testNet3GenesisHash = regTestGenesisHash

// testNet3GenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the test network (version 3).  It is the same as the merkle root
// for the main network.
var testNet3GenesisMerkleRoot = regTestGenesisMerkleRoot

// testNet3GenesisBlock defines the genesis block of the block chain which
// serves as the public transaction ledger for the test network (version 3).
var testNet3GenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},          // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: testNet3GenesisMerkleRoot, // merkleroot
		Timestamp:  time.Unix(1735376054, 0),  // 28 Dec 2024 08:54:14 +0000 UTC
		Bits:       0x207fffff,                // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:      2083236894,
	},
	Transactions: []*wire.MsgTx{generateGenesisCoinbaseTx()},
}

// simNetGenesisHash is the hash of the first block in the block chain for the
// simulation test network.
var simNetGenesisHash = regTestGenesisHash

// simNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the simulation test network.  It is the same as the merkle root for
// the main network.
var simNetGenesisMerkleRoot = regTestGenesisMerkleRoot

// simNetGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the simulation test network.
var simNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},         // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: simNetGenesisMerkleRoot,  // merkleroot
		Timestamp:  time.Unix(1735376054, 0), // 28 Dec 2024 08:54:14 +0000 UTC
		Bits:       0x207fffff,               // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:      2083236894,
	},
	Transactions: []*wire.MsgTx{generateGenesisCoinbaseTx()},
}

// sigNetGenesisHash is the hash of the first block in the block chain for the
// signet test network.
var sigNetGenesisHash = regTestGenesisHash

// sigNetGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the signet test network. It is the same as the merkle root for
// the main network.
var sigNetGenesisMerkleRoot = regTestGenesisMerkleRoot

// sigNetGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the signet test network.
var sigNetGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{},         // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: sigNetGenesisMerkleRoot,  // xxx
		Timestamp:  time.Unix(1735376054, 0), // 28 Dec 2024 08:54:14 +0000 UTC
		Bits:       0x207fffff,               // 545259519 [7fffff0000000000000000000000000000000000000000000000000000000000]
		Nonce:      2083236894,
	},
	Transactions: []*wire.MsgTx{generateGenesisCoinbaseTx()},
}
