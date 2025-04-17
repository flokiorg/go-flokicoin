// Copyright (c) 2014-2017 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package hdkeychain

// References:
//   [BIP32]: BIP0032 - Hierarchical Deterministic Wallets
//   https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math"
	"reflect"
	"testing"

	secp_ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/flokiorg/go-flokicoin/chaincfg"
)

// TestBIP0032Vectors tests the vectors provided by [BIP32] to ensure the
// derivation works as intended.
func TestBIP0032Vectors(t *testing.T) {
	// The master seeds for each of the two test vectors in [BIP32].
	testVec1MasterHex := "000102030405060708090a0b0c0d0e0f"
	testVec2MasterHex := "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542"
	testVec3MasterHex := "4b381541583be4423346c643850da4b320e46a87ae3d2a4e6da11eba819cd4acba45d239319ac14f863b8d5ab5a0d0c64d2e8a1e7d1457df2e5a3c51c73235be"
	hkStart := uint32(0x80000000)

	tests := []struct {
		name     string
		master   string
		path     []uint32
		wantPub  string
		wantPriv string
		net      *chaincfg.Params
	}{
		// Test vector 1
		{
			name:     "test vector 1 chain m",
			master:   testVec1MasterHex,
			path:     []uint32{},
			wantPub:  "xpub661MyMwAqRbcG5SaAT7xLF3TzgYfWA4739rpoo48wQXSz19vcVTpr48H7UcGknLrDbNP76Qws33r5mARVAAsTFKYeK5DTZtSdNVz5i4Bg4q",
			wantPriv: "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "xpub692RdDZNz9vqxc84UtNuPNcAd1NYi6K2pLcXYvSLKjUC7JTRaoHpE21yPPUvWuAGRT2LECSxnE7D9XfsM1P3yYmXffN1UGzNCaDVcUbN1Gm",
			wantPriv: "xprv9v35Di2V9nNYk83bNrqu2EfS4yY4JdbBT7gvkY2imPwDEW8H3FyZgDhVY5Jbqicju5EQnkeaQCVrpvyLPoDbAQQbFUS7errGAMejbJRDBYh",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1},
			wantPub:  "xpub6B6DaWu2g47AKZzFKcH6ojBd4qX6vhvQ3evxcfTVGsMfVxDSdbWW7QUNLYioY3Nu1jhQWqoijk55dqrfb6pRRtZpdH7Wz3KiDxM4XV1EaBN",
			wantPriv: "xprv9x6sB1N8qgYs75unDak6SbEtWogcXFCYgS1MpH3siXpgd9tJ64CFZc9tVFjn5V7rKAhiQtnC5kR2uuSxGdFT6iq6Wqjm3bBNrC28sNQiAUy",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2},
			wantPub:  "xpub6D4K2Cbc7DT39TqTngCRN2kupvDMSwfodBGksW8NroBPB23Mz2FRkPJ5tg3YrzJzTFdE4TDm32SyYZhz1QTHgBTQ1UAinJVC1iMBxb7o26s",
			wantPriv: "xprv9z4xch4iGqtjvykzgefQztpBGtNs3UwxFxMA57imJTeQJDiDSUwBCayc3RFABGfVjrFyaLi8xXFM6GsgmY1g9qmkqMDqYpfHzaodAuJwjyG",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2},
			wantPub:  "xpub6EDQjevaMdaHr8DWeVYeRVDdjcPcMkq3tAZYrdvjPr8yfFmX94Z7rhxTu2y2U7bj4GcGssRYnixyTMQU8grfzyT6ykEWaQSa4RufDQkgudN",
			wantPriv: "xprvA1E4L9PgXG1zde93YU1e4MGuBaZ7xJ7CWwdx4FX7qWbznTSNbXEsJudz3ma5VhCM5xoQa1HN27YxabpTCEzjbdra2df13ffKhF9APr5ize4",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2/1000000000",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2, 1000000000},
			wantPub:  "xpub6H2DU4RDR28RF9cNZygR1NRzBERtdbwXTRCtc6LHCWxfjwRUEbATQrqDCRv337uJHu4AzW5eTDffEphyBDcoodrxxb8HQr4KGrpFcvCkgav",
			wantPriv: "xprvA42s4YtKaea82fXuTx9QeEVFdCbQE9Dg6CHHohvfeBRgs96Kh3rCs4WjMAKs774ms4nuznSLchn84T7ziCnbnBcpGKHn6vUuQGXSFVnuTcF",
			net:      &chaincfg.MainNetParams,
		},

		// Test vector 2
		{
			name:     "test vector 2 chain m",
			master:   testVec2MasterHex,
			path:     []uint32{},
			wantPub:  "xpub661MyMwAqRbcFDzkFyxMQfDa214Wz8j51Q4DYxHcEkeqYvJpuxSvsbJShdUsshz4qJ4DKK9ifpTwwdpYbufQmp13SPsNvjuAgxm7HyxYHa6",
			wantPriv: "xprv9s21ZrQH143K2jvH9xRM3XGqTyE2ag1DeB8ckZszgR7rg7ygNR8gKnyxrN26ZEY12McJ4uxJaaFexteUotkHSWr5rt9ekxTQ1kpWwyzi41F",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0",
			master:   testVec2MasterHex,
			path:     []uint32{0},
			wantPub:  "xpub68g7t6wczKHbucmKGyAEo1i3Mftd1eDKUpzmgmqhDdYKjxr89Dtka5g4yj4K8eZMhMK7MiFbazfA9wh6EQtzfrxixthRkJM7X66oZ1FjP5Z",
			wantPriv: "xprv9ugmUbQj9wjJh8grAwdERsmJoe48cBVU7c5AtPS5fJ1LsAWybgaW2HMb8SkrqXNCQCfuR72bA1aQA1T6UWGEU9JCJY6Ra3erqz5eKBscuvF",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647},
			wantPub:  "xpub6Aebfhzi4viR833SkxHVG5Wm6ZMgEgSnQvKa7DYdJwvuXfstJNj1Dp1JAvAGEQZXH4rg4KFd5V1XmdqBgiWEig7MuB5KtbC7ni63UKX1P6f",
			wantPriv: "xprv9wfFGCTpEZA7uYxyevkUtwa2YXXBqDiw3hPyJq91kcPvesYjkqQkg1gpKfViuKbvjCK6saFX1A1t6SAUh1jY8UHphF1g4E2RYN4SqaCqcoS",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1},
			wantPub:  "xpub6CHFwWTBkPBhwywCvmJhzRuSyoXkSQqT9FVPoQaqtX75jiqRxiLneF1NzbRkNRBZMDNoKJxbuiVwaEVYc3XZViU6uC4Fgr57g2b8tRwQWsP",
			wantPriv: "xprv9yHuXzvHv1dQjVrjpjmhdHxiRmhG2x7bn2Zo12BELBa6rvWHRB2Y6Sgu9Mh7DunkJmVx8Td7XrdTg5uAB6fZGotPqvkrFEX2HTyMrG2qcmR",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1/2147483646H",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1, hkStart + 2147483646},
			wantPub:  "xpub6EYEw2b7ZbkeoKHoqKg76JwZHaY7nF4FENc5SpUrQDZ4e6B5pPGQQCC79ffNxGfPaAHUgTPNLSA3PYERaFvdAmiR3CJaArZtbsJFm7dRkH6",
			wantPriv: "xprvA1YtXX4DjECMaqDLjJ96jAzpjYhdNnLPs9gUeS5Eqt25mHqwGqx9rPsdJPGQv4ADU6463HEx25Wh1xhfJGBgsCdq4MTPSUwjTMgTA64S9tK",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 2 chain m/0/2147483647H/1/2147483646H/2",
			master:   testVec2MasterHex,
			path:     []uint32{0, hkStart + 2147483647, 1, hkStart + 2147483646, 2},
			wantPub:  "xpub6FjtKjxLKQuSFZeGDn4brKhTBbNBgbiTLRsuTfjg9n5dZBWk5iNiqMb45x1DtxQ4HDG211wgdrdDsCZ1ESMrQfYqrVjRFyqMzLBHv8KrgfH",
			wantPriv: "xprvA2kXvERSV3M935Zo7kXbVBkidZXhH8zbyCxJfHL4bSYegPBbYB4UHZGaEfgjm5yjJ8a82o7LbUMKG9VpJNVaXFK35Yj8zwjDRhiDNiUQyGH",
			net:      &chaincfg.MainNetParams,
		},

		// Test vector 3
		{
			name:     "test vector 3 chain m",
			master:   testVec3MasterHex,
			path:     []uint32{},
			wantPub:  "xpub661MyMwAqRbcH35jWENdKgKtDsqxvsHvyTzpo8HVTnChRAjrHMpMvqDVX7wBUuVok411N2cofwhjVpmET8kfWVawkHaiHbgsXxZsF4SAezb",
			wantPriv: "xprv9s21ZrQH143K4Z1GQCqcxYP9fr1UXQa5cF5DzjssuSfiYNQhjpW7P2u1fqzbJHcPd6JKriTy4fovQ3hCZXxqPEKwqUvVxKjfhJeFjQCNWXw",
			net:      &chaincfg.MainNetParams,
		},
		{
			name:     "test vector 3 chain m/0H",
			master:   testVec3MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "xpub68TniNGRF35fN6AyP8MgetWPy6tV3GHXCg61PCrM2QDMj83dq3HzkthdsatpVzBc7QcUgXfCqeGb43wQ5SXWThbEykDa9LVtLm12NtwHJWe",
			wantPriv: "xprv9uUSJrjXQfXN9c6WH6pgHkZfR53zdoZfqTAQapSjU4gNrKiVHVykD6PA2KKdJKj9jrTW56xhnGuxTMuSHV5TmnY5FHgAH5fzj76w7gsERbf",
			net:      &chaincfg.MainNetParams,
		},

		// Test vector 1 - Testnet
		{
			name:     "test vector 1 chain m - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{},
			wantPub:  "tpubD6NzVbkrYhZ4XsdRce73YANqKoeKVYZAanT8AM7HHKHLjHpdgiJuzVo7MWh4HHJ61VuHifvpQ8stmwf2t227LgmqhupX96YgDL6Pqffkg9A",
			wantPriv: "tprv8ZgxMBicQKsPeQbdizST8kiikn8PLDNG1UrLsq4ys3UwtoZs4KVKp1BFBP4VaQTemfrW79W8XigVXU62JmD2YhRPnzgFkNSKAzcNLubwEqw",
			net:      &chaincfg.TestNet3Params,
		},
		{
			name:     "test vector 1 chain m/0H - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart},
			wantPub:  "tpubD9Q49TP4hRtJEQJuw5MzbHwXx8UChUp6MyCpuUVUfeE5rb88f28uNTgodRZi3Q7WDMZEqmxqKKwFqiAUjsEHrzDpjG7K9oebnXouNTub21Q",
			wantPriv: "tprv8ci213LpZ4CdLwH83RhQBtHRP6xGY9dBnfc3cxTBFNRh26sN2dKKBy4wTFUFr614GWmBnrGLZZ5fHnX5X1ZXyTgBn7eRKDaK5TQA32R4psZ",
			net:      &chaincfg.TestNet3Params,
		},
		{
			name:     "test vector 1 chain m/0H/1 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1},
			wantPub:  "tpubDBTr6kiiPL4cbNB6moGC1eWzPxckv6RTbHXFyDWdcn7ZFEt9hpMbFr9Caaob4YL8oeEK8RKbGqu8L2MGyxffKL27gsrpfZywouwUHXNsDNb",
			wantPriv: "tprv8emoxLgUExNwhu9Jt9bbcErspw6pkmEZ1yvUghULCWKAQkdP5RY15MXLQRuS5rWAgcEVQzPxF6zqNkzhPqbPun6h3Ux4hwuRmHmZK4yq5BT",
			net:      &chaincfg.TestNet3Params,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2},
			wantPub:  "tpubDDRwYSRHpVQVRG2KEsBWZx6HA3K1SLAsAos4E4BXChwGvJi54F6Wtpxv8i8LPVGEFAA8g2jda8H2EkCbQGJXZcuh54v2Tq9RbfwbicyceWm",
			wantPriv: "tprv8gjuQ2P3g7ipXnzXMDWvAYSAb1o5GzyxbWGGwY9DnS8t5pTJRrGviLM3xbQpBe3p7HnkaSKu7sq9Z8RRtkMcxu3MMzS9DBPLugZ3cc8ByeY",
			net:      &chaincfg.TestNet3Params,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2},
			wantPub:  "tpubDEb3FtkG4uXk7vQN6gXjdQZ14jVGM9L7Ro9rDBysjktsQYSEDHQD19dJ953ozcYxrB9BVSwRKpo29Xu5XYhutQuQ3LypFw6oePW4yT3Kdz6",
			wantPriv: "tprv8hu17Ui1vXr5ETNaD2s9DzttVhyLBp9CrVZ4vfwaKV6Ua4BTatacpf1RxwjjW4afTQLBa6u8BU8m3TNCKTLgQh8AZGsJi2PNcLtaqbFKVnP",
			net:      &chaincfg.TestNet3Params,
		},
		{
			name:     "test vector 1 chain m/0H/1/2H/2/1000000000 - testnet",
			master:   testVec1MasterHex,
			path:     []uint32{hkStart, 1, hkStart + 2, 2, 1000000000},
			wantPub:  "tpubDHPqzJEu8J5sWwoE2AfWDHmMWMXYczSb13oBxePRYRiZVE6BJp1YZJW3STzpZcrY5ob5c5bWzKVhw1Caa5U3h5KG2Bsb6NiYrpQfNyFcKej",
			wantPriv: "tprv8khoqtCeyvQCdUmS8Wzuot7EwL1cTfFgRkCQg8M889vAejqQgRBxNotBGLVX7UT6EWKgzt46n4MvXJfjqR8YbEtQnxW5mHCxKNGrhDrG87b",
			net:      &chaincfg.TestNet3Params,
		},
	}

tests:
	for i, test := range tests {
		masterSeed, err := hex.DecodeString(test.master)
		if err != nil {
			t.Errorf("DecodeString #%d (%s): unexpected error: %v",
				i, test.name, err)
			continue
		}

		extKey, err := NewMaster(masterSeed, test.net)
		if err != nil {
			t.Errorf("NewMaster #%d (%s): unexpected error when "+
				"creating new master key: %v", i, test.name,
				err)
			continue
		}

		for _, childNum := range test.path {
			var err error
			extKey, err = extKey.Derive(childNum)
			if err != nil {
				t.Errorf("err: %v", err)
				continue tests
			}
		}

		if extKey.Depth() != uint8(len(test.path)) {
			t.Errorf("Depth of key %d should match fixture path: %v",
				extKey.Depth(), len(test.path))
			continue
		}

		privStr := extKey.String()
		if privStr != test.wantPriv {
			t.Errorf("Serialize #%d (%s): mismatched serialized "+
				"private extended key -- got: %s, want: %s", i,
				test.name, privStr, test.wantPriv)
			continue
		}

		pubKey, err := extKey.Neuter()
		if err != nil {
			t.Errorf("Neuter #%d (%s): unexpected error: %v ", i,
				test.name, err)
			continue
		}

		// Neutering a second time should have no effect.
		pubKey, err = pubKey.Neuter()
		if err != nil {
			t.Errorf("Neuter #%d (%s): unexpected error: %v", i,
				test.name, err)
			return
		}

		pubStr := pubKey.String()
		if pubStr != test.wantPub {
			t.Errorf("Neuter #%d (%s): mismatched serialized "+
				"public extended key -- got: %s, want: %s", i,
				test.name, pubStr, test.wantPub)
			continue
		}
	}
}

// TestPrivateDerivation tests several vectors which derive private keys from
// other private keys works as intended.
func TestPrivateDerivation(t *testing.T) {
	// The private extended keys for test vectors in [BIP32].
	testVec1MasterPrivKey := "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p"
	testVec2MasterPrivKey := "xprv9s21ZrQH143K2jvH9xRM3XGqTyE2ag1DeB8ckZszgR7rg7ygNR8gKnyxrN26ZEY12McJ4uxJaaFexteUotkHSWr5rt9ekxTQ1kpWwyzi41F"

	tests := []struct {
		name     string
		master   string
		path     []uint32
		wantPriv string
	}{
		// Test vector 1
		{
			name:     "test vector 1 chain m",
			master:   testVec1MasterPrivKey,
			path:     []uint32{},
			wantPriv: "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p",
		},
		{
			name:     "test vector 1 chain m/0",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0},
			wantPriv: "xprv9v35Di2Lp7qaZ9YW2Ssau1a9hvzT4CgCGXCbPXkbhkMFhJS432W3JdPKgv1Qv54JxYDPHpPx4iiBLsYtnd3EaZqX89gjin3Y5GpfUsqX3yv",
		},
		{
			name:     "test vector 1 chain m/0/1",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1},
			wantPriv: "xprv9wPPC8ju2LHiGiKwY1HNC9p5WHHG1VztxcbmXscf9N3YECjunU7NTC76EUv2uNavNMYbJjBGHWzgpAaYMRHPiSbgnDZhEyvjFdbhjeGPgzF",
		},
		{
			name:     "test vector 1 chain m/0/1/2",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2},
			wantPriv: "xprv9xtMCD2ZtT8mVPRtGSFCcx6fGf32oXYuHcsQWrsHJExSqZLm1LTorRe6Fe5khq8baJxc2vkkyu4KzKvzsJPq4Hq2NJd5t9jv2eZcjpu6FvR",
		},
		{
			name:     "test vector 1 chain m/0/1/2/2",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2, 2},
			wantPriv: "xprvA1Yr3zgRZ7jbB5hTyXqMmcPueqXXjLifZdQevNEiv59PpWmAE49Eix9sK3K1D3iTYmFV8gZrCLRdcSo558khs8hGsaEUNpvE8qwN7hLaZ1v",
		},
		{
			name:     "test vector 1 chain m/0/1/2/2/1000000000",
			master:   testVec1MasterPrivKey,
			path:     []uint32{0, 1, 2, 2, 1000000000},
			wantPriv: "xprvA3kTXuXPwXQQtHiS8gREDqw5Zrf3Ae1p2uKu1iPCdhLyw96QadsTN651ve4sBisrvY6SdaagbeN8wnBNFBHmXM8WECucsCcZeSLZ65RapYm",
		},

		// Test vector 2
		{
			name:     "test vector 2 chain m",
			master:   testVec2MasterPrivKey,
			path:     []uint32{},
			wantPriv: "xprv9s21ZrQH143K2jvH9xRM3XGqTyE2ag1DeB8ckZszgR7rg7ygNR8gKnyxrN26ZEY12McJ4uxJaaFexteUotkHSWr5rt9ekxTQ1kpWwyzi41F",
		},
		{
			name:     "test vector 2 chain m/0",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0},
			wantPriv: "xprv9ugmUbQj9wjJh8grAwdERsmJoe48cBVU7c5AtPS5fJ1LsAWybgaW2HMb8SkrqXNCQCfuR72bA1aQA1T6UWGEU9JCJY6Ra3erqz5eKBscuvF",
		},
		{
			name:     "test vector 2 chain m/0/2147483647",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647},
			wantPriv: "xprv9wfFGCTfttd9j3NxgWR9NsWPGCJhccGpR3ev9uzBA2DbC787gxmGALFH5gUJGKiV4anu7xpzQBcSxPSdKqxfmDfbZbBobKVLVikbMmDM3ww",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1},
			wantPriv: "xprv9yKepnxsC8RfANTQmz962TTtvRi1MR1updDQU3DwyK5h1tMFvVLNUDPByPrsAyAu7q828eWa3PbcChbGwsVRWHrgbtxyBK9F8HyBppyW2H8",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1/2147483646",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1, 2147483646},
			wantPriv: "xprv9zc5hJ6yX2TLLRTZ8EGm7SMQdnvYFUE6v3W8QRDYRpWr1AAtLC8am1xhGzdbeQ62QwANaELaEu2ccgseSCpsVFtHAri8DGwM8E4ZhLYMTW7",
		},
		{
			name:     "test vector 2 chain m/0/2147483647/1/2147483646/2",
			master:   testVec2MasterPrivKey,
			path:     []uint32{0, 2147483647, 1, 2147483646, 2},
			wantPriv: "xprvA3n5yDVimCLXaHDBnMDsPbvA4QunKcyEx873HioqvVT5rHxs8FdUhsb1QhUtXYpSztrk4N6naFf3M9Bs6FafEjQTuNHDKsbftv8AeHpG3aZ",
		},

		// Custom tests to trigger specific conditions.
		{
			// Seed 000000000000000000000000000000da.
			name:     "Derived privkey with zero high byte m/0",
			master:   "xprv9s21ZrQH143K4FR6rNeqEK4EBhRgLjWLWhA3pw8iqgAKk82ypz58PXbrzU19opYcxw8JDJQF4id55PwTsN1Zv8Xt6SKvbr2KNU5y8jN8djz",
			path:     []uint32{0},
			wantPriv: "xprv9uC5JqtViMmgcAMUxcsBCBFA7oYCNs4bozPbyvLfddjHou4rMiGEHipz94xNaPb1e4f18TRoPXfiXx4C3cDAcADqxCSRSSWLvMBRWPctSN9",
		},
	}

tests:
	for i, test := range tests {
		extKey, err := NewKeyFromString(test.master)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected error "+
				"creating extended key: %v", i, test.name,
				err)
			continue
		}

		for _, childNum := range test.path {
			var err error
			extKey, err = extKey.Derive(childNum)
			if err != nil {
				t.Errorf("err: %v", err)
				continue tests
			}
		}

		privStr := extKey.String()
		if privStr != test.wantPriv {
			t.Errorf("Derive #%d (%s): mismatched serialized "+
				"private extended key -- got: %s, want: %s", i,
				test.name, privStr, test.wantPriv)
			continue
		}
	}
}

// TestPublicDerivation tests several vectors which derive public keys from
// other public keys works as intended.
func TestPublicDerivation(t *testing.T) {
	// The public extended keys for test vectors in [BIP32].
	testVec1MasterPubKey := "xpub661MyMwAqRbcG5SaAT7xLF3TzgYfWA4739rpoo48wQXSz19vcVTpr48H7UcGknLrDbNP76Qws33r5mARVAAsTFKYeK5DTZtSdNVz5i4Bg4q"
	testVec2MasterPubKey := "xpub661MyMwAqRbcFDzkFyxMQfDa214Wz8j51Q4DYxHcEkeqYvJpuxSvsbJShdUsshz4qJ4DKK9ifpTwwdpYbufQmp13SPsNvjuAgxm7HyxYHa6"

	tests := []struct {
		name    string
		master  string
		path    []uint32
		wantPub string
	}{
		// Test vector 1
		{
			name:    "test vector 1 chain m",
			master:  testVec1MasterPubKey,
			path:    []uint32{},
			wantPub: "xpub661MyMwAqRbcG5SaAT7xLF3TzgYfWA4739rpoo48wQXSz19vcVTpr48H7UcGknLrDbNP76Qws33r5mARVAAsTFKYeK5DTZtSdNVz5i4Bg4q",
		},
		{
			name:    "test vector 1 chain m/0",
			master:  testVec1MasterPubKey,
			path:    []uint32{0},
			wantPub: "xpub692RdDZEeVPsmdcy8UQbG9WtFxpwTfQ3dk8CBvADG5tEa6mCaZpHrRhoYDi56vKh2APefHpaFkw426wsPZ8u8mpPkARyGr7xVUcqw4F2idG",
		},
		{
			name:    "test vector 1 chain m/0/1",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1},
			wantPub: "xpub6ANjbeGnrhr1VCQQe2pNZHkp4K7kQxikKqXNLG2GhhaX7154L1RczzRa5k9AoPXvzM9TJpq7R8pdyB549dNCAkp2rbfw5L6jhG1vdnEYjFD",
		},
		{
			name:    "test vector 1 chain m/0/1/2",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2},
			wantPub: "xpub6BshbiZTiph4hsWMNTnCz63PpgsXCzGkeqo1KFGtraVRiMfuYsn4QDxa6vy8jXunduHKvSN4EPHuByiPGpoLqp54CeAsZt4aMDpxiayofFa",
		},
		{
			name:    "test vector 1 chain m/0/1/2/2",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2, 2},
			wantPub: "xpub6EYCTWDKPVHtPZmw5ZNN8kLeCsN28oSWvrLFikeLUQgNhK6JmbTVGkUMAHujifmTn5x6MLLALToYHfkhLcY6zuWykukfERjQUVKZDbULmTG",
		},
		{
			name:    "test vector 1 chain m/0/1/2/2/1000000000",
			master:  testVec1MasterPubKey,
			path:    []uint32{0, 1, 2, 2, 1000000000},
			wantPub: "xpub6GjowR4Hmtxi6mnuEhxEaysp7tVXa6jfQ8FVp6npC2sxowRZ8BBhutPVmtF9uRTkS3prvHuVP3qQrvKwezSfkTPBnEuKz6pwWKLzEcjxBy8",
		},

		// Test vector 2
		{
			name:    "test vector 2 chain m",
			master:  testVec2MasterPubKey,
			path:    []uint32{},
			wantPub: "xpub661MyMwAqRbcFDzkFyxMQfDa214Wz8j51Q4DYxHcEkeqYvJpuxSvsbJShdUsshz4qJ4DKK9ifpTwwdpYbufQmp13SPsNvjuAgxm7HyxYHa6",
		},
		{
			name:    "test vector 2 chain m/0",
			master:  testVec2MasterPubKey,
			path:    []uint32{0},
			wantPub: "xpub68g7t6wczKHbucmKGyAEo1i3Mftd1eDKUpzmgmqhDdYKjxr89Dtka5g4yj4K8eZMhMK7MiFbazfA9wh6EQtzfrxixthRkJM7X66oZ1FjP5Z",
		},
		{
			name:    "test vector 2 chain m/0/2147483647",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647},
			wantPub: "xpub6AebfhzZjGBSwXTRnXx9k1T7pE9C24zfnGaWxJPniMka4uTGEW5Wi8ZkvwMUYBv6V7HmDLJy53Bcs9zFkGGzA5isEDkU3pDD2rmrq54Vi9h",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1},
			wantPub: "xpub6CK1EJVm2VyxNrXst1g6PbQdUTYVksjmBr91GRdZXecftggQU2ed21hfpfdGY53QZe8SvTphGQctirKhmeuGzqMK6XUwZPGR1YHGCT2FgiE",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1/2147483646",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1, 2147483646},
			wantPub: "xpub6DbS6odsMQ1dYuY2EFomUaJ9Bpm2evwxHGRjCod9zA3psxW2sjSqJpHB8EGTpaW3wsncPsUq6uSdVkKLxTFsKwfHj6LXBdE8HF9Lta5kVD6",
		},
		{
			name:    "test vector 2 chain m/0/2147483647/1/2147483646/2",
			master:  testVec2MasterPubKey,
			path:    []uint32{0, 2147483647, 1, 2147483646, 2},
			wantPub: "xpub6GmSNj2cbZtpnmHetNkskjrtcSkGj5h6KM2e67DTUpz4j6J1fnwjFfuVG16a1DDggULvpbamDxRCW2w2AyqT7DYj98nA97iCZZf6RJACn6L",
		},
	}

tests:
	for i, test := range tests {
		extKey, err := NewKeyFromString(test.master)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected error "+
				"creating extended key: %v", i, test.name,
				err)
			continue
		}

		for _, childNum := range test.path {
			var err error
			extKey, err = extKey.Derive(childNum)
			if err != nil {
				t.Errorf("err: %v", err)
				continue tests
			}
		}

		pubStr := extKey.String()
		if pubStr != test.wantPub {
			t.Errorf("Derive #%d (%s): mismatched serialized "+
				"public extended key -- got: %s, want: %s", i,
				test.name, pubStr, test.wantPub)
			continue
		}
	}
}

// TestGenerateSeed ensures the GenerateSeed function works as intended.
func TestGenerateSeed(t *testing.T) {
	wantErr := errors.New("seed length must be between 128 and 512 bits")

	tests := []struct {
		name   string
		length uint8
		err    error
	}{
		// Test various valid lengths.
		{name: "16 bytes", length: 16},
		{name: "17 bytes", length: 17},
		{name: "20 bytes", length: 20},
		{name: "32 bytes", length: 32},
		{name: "64 bytes", length: 64},

		// Test invalid lengths.
		{name: "15 bytes", length: 15, err: wantErr},
		{name: "65 bytes", length: 65, err: wantErr},
	}

	for i, test := range tests {
		seed, err := GenerateSeed(test.length)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("GenerateSeed #%d (%s): unexpected error -- "+
				"want %v, got %v", i, test.name, test.err, err)
			continue
		}

		if test.err == nil && len(seed) != int(test.length) {
			t.Errorf("GenerateSeed #%d (%s): length mismatch -- "+
				"got %d, want %d", i, test.name, len(seed),
				test.length)
			continue
		}
	}
}

// TestExtendedKeyAPI ensures the API on the ExtendedKey type works as intended.
func TestExtendedKeyAPI(t *testing.T) {
	tests := []struct {
		name       string
		extKey     string
		isPrivate  bool
		parentFP   uint32
		chainCode  []byte
		childNum   uint32
		privKey    string
		privKeyErr error
		pubKey     string
		address    string
	}{
		{
			name:      "test vector 1 master node private",
			extKey:    "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p",
			isPrivate: true,
			parentFP:  0,
			chainCode: []byte{154, 35, 145, 116, 145, 184, 40, 153, 91, 122, 61, 31, 110, 34, 164, 216, 37, 21, 168, 132, 135, 52, 168, 188, 223, 167, 91, 102, 246, 108, 216, 38},
			childNum:  0,
			privKey:   "16ff3b756fc8026de38f879a095de8d904b18589e06eda4948c7d86b4f1e17da",
			pubKey:    "0271116998daf962850126e6d8d33cac3e9fc3be0dba948ad7eef62ce9d4f46b53",
			address:   "FKvoqJvo4qijci7p3oTJeYh8nwaU5Wdb4t",
		},
		{
			name:       "test vector 1 chain m/0H/1/2H public",
			extKey:     "xpub6D4K2Cbc7DT39TqTngCRN2kupvDMSwfodBGksW8NroBPB23Mz2FRkPJ5tg3YrzJzTFdE4TDm32SyYZhz1QTHgBTQ1UAinJVC1iMBxb7o26s",
			isPrivate:  false,
			parentFP:   3209066374,
			chainCode:  []byte{125, 214, 30, 206, 177, 78, 100, 16, 248, 99, 80, 253, 244, 19, 183, 20, 197, 138, 36, 238, 15, 68, 120, 82, 65, 41, 113, 197, 208, 11, 164, 101},
			childNum:   2147483650,
			privKeyErr: ErrNotPrivExtKey,
			pubKey:     "02218e665cdbda1cb9607771a97754ac840e19a57099ab36c97323498bbae4252f",
			address:    "FEH6jypszySPg2DMJ17rCF2vKB99iQQPdv",
		},
	}

	for i, test := range tests {
		key, err := NewKeyFromString(test.extKey)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected "+
				"error: %v", i, test.name, err)
			continue
		}

		if key.IsPrivate() != test.isPrivate {
			t.Errorf("IsPrivate #%d (%s): mismatched key type -- "+
				"want private %v, got private %v", i, test.name,
				test.isPrivate, key.IsPrivate())
			continue
		}

		parentFP := key.ParentFingerprint()
		if parentFP != test.parentFP {
			t.Errorf("ParentFingerprint #%d (%s): mismatched "+
				"parent fingerprint -- want %d, got %d", i,
				test.name, test.parentFP, parentFP)
			continue
		}

		chainCode := key.ChainCode()
		if !bytes.Equal(chainCode, test.chainCode) {
			t.Errorf("ChainCode #%d (%s): want %v, got %v", i,
				test.name, test.chainCode, chainCode)
			continue
		}

		childIndex := key.ChildIndex()
		if childIndex != test.childNum {
			t.Errorf("ChildIndex #%d (%s): want %d, got %d", i,
				test.name, test.childNum, childIndex)
			continue
		}

		serializedKey := key.String()
		if serializedKey != test.extKey {
			t.Errorf("String #%d (%s): mismatched serialized key "+
				"-- want %s, got %s", i, test.name, test.extKey,
				serializedKey)
			continue
		}

		privKey, err := key.ECPrivKey()
		if !reflect.DeepEqual(err, test.privKeyErr) {
			t.Errorf("ECPrivKey #%d (%s): mismatched error: want "+
				"%v, got %v", i, test.name, test.privKeyErr, err)
			continue
		}
		if test.privKeyErr == nil {
			privKeyStr := hex.EncodeToString(privKey.Serialize())
			if privKeyStr != test.privKey {
				t.Errorf("ECPrivKey #%d (%s): mismatched "+
					"private key -- want %s, got %s", i,
					test.name, test.privKey, privKeyStr)
				continue
			}
		}

		pubKey, err := key.ECPubKey()
		if err != nil {
			t.Errorf("ECPubKey #%d (%s): unexpected error: %v", i,
				test.name, err)
			continue
		}
		pubKeyStr := hex.EncodeToString(pubKey.SerializeCompressed())
		if pubKeyStr != test.pubKey {
			t.Errorf("ECPubKey #%d (%s): mismatched public key -- "+
				"want %s, got %s", i, test.name, test.pubKey,
				pubKeyStr)
			continue
		}

		addr, err := key.Address(&chaincfg.MainNetParams)
		if err != nil {
			t.Errorf("Address #%d (%s): unexpected error: %v", i,
				test.name, err)
			continue
		}
		if addr.EncodeAddress() != test.address {
			t.Errorf("Address #%d (%s): mismatched address -- want "+
				"%s, got %s", i, test.name, test.address,
				addr.EncodeAddress())
			continue
		}
	}
}

// TestNet ensures the network related APIs work as intended.
func TestNet(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		origNet   *chaincfg.Params
		newNet    *chaincfg.Params
		newPriv   string
		newPub    string
		isPrivate bool
	}{
		// Private extended keys.
		{
			name:      "mainnet -> simnet",
			key:       "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p",
			origNet:   &chaincfg.MainNetParams,
			newNet:    &chaincfg.SimNetParams,
			newPriv:   "sprv8Erh3X3hFeKuo7QWtdepvfdDaoywPbNm6NL86SB6V374qhkdofKmBwX7AB6uaVojN2jhMfua54KweBUTb4o1kTjnveyqATcpuGYixMGKMuJ",
			newPub:    "spub4Tr3T2ab61tD1bUyzfBqHoZx8qpRo46cTbFitpai3Ne3iW5nMCe1jjqb1SpLmF5FBPnMMiS9ZjH6fL6btg6oU4uYJcb6Y1o1GkBm8k21WEn",
			isPrivate: true,
		},
		{
			name:      "simnet -> mainnet",
			key:       "sprv8Erh3X3hFeKuo7QWtdepvfdDaoywPbNm6NL86SB6V374qhkdofKmBwX7AB6uaVojN2jhMfua54KweBUTb4o1kTjnveyqATcpuGYixMGKMuJ",
			origNet:   &chaincfg.SimNetParams,
			newNet:    &chaincfg.MainNetParams,
			newPriv:   "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p",
			newPub:    "xpub661MyMwAqRbcG5SaAT7xLF3TzgYfWA4739rpoo48wQXSz19vcVTpr48H7UcGknLrDbNP76Qws33r5mARVAAsTFKYeK5DTZtSdNVz5i4Bg4q",
			isPrivate: true,
		},
		{
			name:      "mainnet -> regtest",
			key:       "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p",
			origNet:   &chaincfg.MainNetParams,
			newNet:    &chaincfg.RegressionNetParams,
			newPriv:   "tprv8ZgxMBicQKsPeQbdizST8kiikn8PLDNG1UrLsq4ys3UwtoZs4KVKp1BFBP4VaQTemfrW79W8XigVXU62JmD2YhRPnzgFkNSKAzcNLubwEqw",
			newPub:    "tpubD6NzVbkrYhZ4XsdRce73YANqKoeKVYZAanT8AM7HHKHLjHpdgiJuzVo7MWh4HHJ61VuHifvpQ8stmwf2t227LgmqhupX96YgDL6Pqffkg9A",
			isPrivate: true,
		},
		{
			name:      "regtest -> mainnet",
			key:       "tprv8ZgxMBicQKsPeQbdizST8kiikn8PLDNG1UrLsq4ys3UwtoZs4KVKp1BFBP4VaQTemfrW79W8XigVXU62JmD2YhRPnzgFkNSKAzcNLubwEqw",
			origNet:   &chaincfg.RegressionNetParams,
			newNet:    &chaincfg.MainNetParams,
			newPriv:   "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p",
			newPub:    "xpub661MyMwAqRbcG5SaAT7xLF3TzgYfWA4739rpoo48wQXSz19vcVTpr48H7UcGknLrDbNP76Qws33r5mARVAAsTFKYeK5DTZtSdNVz5i4Bg4q",
			isPrivate: true,
		},

		// Public extended keys.
		{
			name:      "mainnet -> simnet",
			key:       "xpub661MyMwAqRbcG5SaAT7xLF3TzgYfWA4739rpoo48wQXSz19vcVTpr48H7UcGknLrDbNP76Qws33r5mARVAAsTFKYeK5DTZtSdNVz5i4Bg4q",
			origNet:   &chaincfg.MainNetParams,
			newNet:    &chaincfg.SimNetParams,
			newPub:    "spub4Tr3T2ab61tD1bUyzfBqHoZx8qpRo46cTbFitpai3Ne3iW5nMCe1jjqb1SpLmF5FBPnMMiS9ZjH6fL6btg6oU4uYJcb6Y1o1GkBm8k21WEn",
			isPrivate: false,
		},
		{
			name:      "simnet -> mainnet",
			key:       "spub4Tr3T2ab61tD1Qa6GHwRFiiKyRRJdfEZpSpXfqgFYCEyaPsqKysqHDjzSzMJSiUEGbcsG3w2SLMoTqn44B8x6u3MLRRkYfACTUBnHK79THk",
			origNet:   &chaincfg.SimNetParams,
			newNet:    &chaincfg.MainNetParams,
			newPub:    "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
			isPrivate: false,
		},
		{
			name:      "mainnet -> regtest",
			key:       "xpub661MyMwAqRbcG5SaAT7xLF3TzgYfWA4739rpoo48wQXSz19vcVTpr48H7UcGknLrDbNP76Qws33r5mARVAAsTFKYeK5DTZtSdNVz5i4Bg4q",
			origNet:   &chaincfg.MainNetParams,
			newNet:    &chaincfg.RegressionNetParams,
			newPub:    "tpubD6NzVbkrYhZ4XsdRce73YANqKoeKVYZAanT8AM7HHKHLjHpdgiJuzVo7MWh4HHJ61VuHifvpQ8stmwf2t227LgmqhupX96YgDL6Pqffkg9A",
			isPrivate: false,
		},
		{
			name:      "regtest -> mainnet",
			key:       "tpubD6NzVbkrYhZ4XsdRce73YANqKoeKVYZAanT8AM7HHKHLjHpdgiJuzVo7MWh4HHJ61VuHifvpQ8stmwf2t227LgmqhupX96YgDL6Pqffkg9A",
			origNet:   &chaincfg.RegressionNetParams,
			newNet:    &chaincfg.MainNetParams,
			newPub:    "xpub661MyMwAqRbcG5SaAT7xLF3TzgYfWA4739rpoo48wQXSz19vcVTpr48H7UcGknLrDbNP76Qws33r5mARVAAsTFKYeK5DTZtSdNVz5i4Bg4q",
			isPrivate: false,
		},
	}

	for i, test := range tests {
		extKey, err := NewKeyFromString(test.key)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected error "+
				"creating extended key: %v", i, test.name,
				err)
			continue
		}

		if !extKey.IsForNet(test.origNet) {
			t.Errorf("IsForNet #%d (%s): key is not for expected "+
				"network %v", i, test.name, test.origNet.Name)
			continue
		}

		extKey.SetNet(test.newNet)
		if !extKey.IsForNet(test.newNet) {
			t.Errorf("SetNet/IsForNet #%d (%s): key is not for "+
				"expected network %v", i, test.name,
				test.newNet.Name)
			continue
		}

		if test.isPrivate {
			privStr := extKey.String()
			if privStr != test.newPriv {
				t.Errorf("Serialize #%d (%s): mismatched serialized "+
					"private extended key -- got: %s, want: %s", i,
					test.name, privStr, test.newPriv)
				continue
			}

			extKey, err = extKey.Neuter()
			if err != nil {
				t.Errorf("Neuter #%d (%s): unexpected error: %v ", i,
					test.name, err)
				continue
			}
		}

		pubStr := extKey.String()
		if pubStr != test.newPub {
			t.Errorf("Neuter #%d (%s): mismatched serialized "+
				"public extended key -- got: %s, want: %s", i,
				test.name, pubStr, test.newPub)
			continue
		}
	}
}

// TestErrors performs some negative tests for various invalid cases to ensure
// the errors are handled properly.
func TestErrors(t *testing.T) {
	// Should get an error when seed has too few bytes.
	net := &chaincfg.MainNetParams
	_, err := NewMaster(bytes.Repeat([]byte{0x00}, 15), net)
	if err != ErrInvalidSeedLen {
		t.Fatalf("NewMaster: mismatched error -- got: %v, want: %v",
			err, ErrInvalidSeedLen)
	}

	// Should get an error when seed has too many bytes.
	_, err = NewMaster(bytes.Repeat([]byte{0x00}, 65), net)
	if err != ErrInvalidSeedLen {
		t.Fatalf("NewMaster: mismatched error -- got: %v, want: %v",
			err, ErrInvalidSeedLen)
	}

	// Generate a new key and neuter it to a public extended key.
	seed, err := GenerateSeed(RecommendedSeedLen)
	if err != nil {
		t.Fatalf("GenerateSeed: unexpected error: %v", err)
	}
	extKey, err := NewMaster(seed, net)
	if err != nil {
		t.Fatalf("NewMaster: unexpected error: %v", err)
	}
	pubKey, err := extKey.Neuter()
	if err != nil {
		t.Fatalf("Neuter: unexpected error: %v", err)
	}

	// Deriving a hardened child extended key should fail from a public key.
	_, err = pubKey.Derive(HardenedKeyStart)
	if err != ErrDeriveHardFromPublic {
		t.Fatalf("Derive: mismatched error -- got: %v, want: %v",
			err, ErrDeriveHardFromPublic)
	}

	// NewKeyFromString failure tests.
	tests := []struct {
		name      string
		key       string
		err       error
		neuter    bool
		neuterErr error
	}{
		{
			name: "invalid key length",
			key:  "xpub1234",
			err:  ErrInvalidKeyLen,
		},
		{
			name: "bad checksum",
			key:  "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EBygr15",
			err:  ErrBadChecksum,
		},
		{
			name: "pubkey not on curve",
			key:  "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ1hr9Rwbk95YadvBkQXxzHBSngB8ndpW6QH7zhhsXZ2jHyZqPjk",
			err:  secp_ecdsa.ErrPubKeyNotOnCurve,
		},
		{
			name:      "unsupported version",
			key:       "xbad4LfUL9eKmA66w2GJdVMqhvDmYGJpTGjWRAtjHqoUY17sGaymoMV9Cm3ocn9Ud6Hh2vLFVC7KSKCRVVrqc6dsEdsTjRV1WUmkK85YEUujAPX",
			err:       nil,
			neuter:    true,
			neuterErr: chaincfg.ErrUnknownHDKeyID,
		},
	}

	for i, test := range tests {
		extKey, err := NewKeyFromString(test.key)
		if !errors.Is(err, test.err) {
			t.Errorf("NewKeyFromString #%d (%s): mismatched error "+
				"-- got: %v, want: %v", i, test.name, err,
				test.err)
			continue
		}

		if test.neuter {
			_, err := extKey.Neuter()
			if !errors.Is(err, test.neuterErr) {
				t.Errorf("Neuter #%d (%s): mismatched error "+
					"-- got: %v, want: %v", i, test.name,
					err, test.neuterErr)
				continue
			}
		}
	}
}

// TestZero ensures that zeroing an extended key works as intended.
func TestZero(t *testing.T) {
	tests := []struct {
		name   string
		master string
		extKey string
		net    *chaincfg.Params
	}{
		// Test vector 1
		{
			name:   "test vector 1 chain m",
			master: "000102030405060708090a0b0c0d0e0f",
			extKey: "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p",
			net:    &chaincfg.MainNetParams,
		},

		// Test vector 2
		{
			name:   "test vector 2 chain m",
			master: "fffcf9f6f3f0edeae7e4e1dedbd8d5d2cfccc9c6c3c0bdbab7b4b1aeaba8a5a29f9c999693908d8a8784817e7b7875726f6c696663605d5a5754514e4b484542",
			extKey: "xprv9s21ZrQH143K2jvH9xRM3XGqTyE2ag1DeB8ckZszgR7rg7ygNR8gKnyxrN26ZEY12McJ4uxJaaFexteUotkHSWr5rt9ekxTQ1kpWwyzi41F",
			net:    &chaincfg.MainNetParams,
		},
	}

	// Use a closure to test that a key is zeroed since the tests create
	// keys in different ways and need to test the same things multiple
	// times.
	testZeroed := func(i int, testName string, key *ExtendedKey) bool {
		// Zeroing a key should result in it no longer being private
		if key.IsPrivate() {
			t.Errorf("IsPrivate #%d (%s): mismatched key type -- "+
				"want private %v, got private %v", i, testName,
				false, key.IsPrivate())
			return false
		}

		parentFP := key.ParentFingerprint()
		if parentFP != 0 {
			t.Errorf("ParentFingerprint #%d (%s): mismatched "+
				"parent fingerprint -- want %d, got %d", i,
				testName, 0, parentFP)
			return false
		}

		wantKey := "zeroed extended key"
		serializedKey := key.String()
		if serializedKey != wantKey {
			t.Errorf("String #%d (%s): mismatched serialized key "+
				"-- want %s, got %s", i, testName, wantKey,
				serializedKey)
			return false
		}

		wantErr := ErrNotPrivExtKey
		_, err := key.ECPrivKey()
		if !reflect.DeepEqual(err, wantErr) {
			t.Errorf("ECPrivKey #%d (%s): mismatched error: want "+
				"%v, got %v", i, testName, wantErr, err)
			return false
		}

		wantErr = secp_ecdsa.ErrPubKeyInvalidLen
		_, err = key.ECPubKey()
		if !errors.Is(err, wantErr) {
			t.Errorf("ECPubKey #%d (%s): mismatched error: want "+
				"%v, got %v", i, testName, wantErr, err)
			return false
		}

		wantAddr := "FNHERGTTXy1KjNx1fJeBQZ9KPzX8i5ZfbG"
		addr, err := key.Address(&chaincfg.MainNetParams)
		if err != nil {
			t.Errorf("Address #%d (%s): unexpected error: %v", i,
				testName, err)
			return false
		}
		if addr.EncodeAddress() != wantAddr {
			t.Errorf("Address #%d (%s): mismatched address -- want "+
				"%s, got %s", i, testName, wantAddr,
				addr.EncodeAddress())
			return false
		}

		return true
	}

	for i, test := range tests {
		// Create new key from seed and get the neutered version.
		masterSeed, err := hex.DecodeString(test.master)
		if err != nil {
			t.Errorf("DecodeString #%d (%s): unexpected error: %v",
				i, test.name, err)
			continue
		}
		key, err := NewMaster(masterSeed, test.net)
		if err != nil {
			t.Errorf("NewMaster #%d (%s): unexpected error when "+
				"creating new master key: %v", i, test.name,
				err)
			continue
		}
		neuteredKey, err := key.Neuter()
		if err != nil {
			t.Errorf("Neuter #%d (%s): unexpected error: %v", i,
				test.name, err)
			continue
		}

		// Ensure both non-neutered and neutered keys are zeroed
		// properly.
		key.Zero()
		if !testZeroed(i, test.name+" from seed not neutered", key) {
			continue
		}
		neuteredKey.Zero()
		if !testZeroed(i, test.name+" from seed neutered", key) {
			continue
		}

		// Deserialize key and get the neutered version.
		key, err = NewKeyFromString(test.extKey)
		if err != nil {
			t.Errorf("NewKeyFromString #%d (%s): unexpected "+
				"error: %v", i, test.name, err)
			continue
		}
		neuteredKey, err = key.Neuter()
		if err != nil {
			t.Errorf("Neuter #%d (%s): unexpected error: %v", i,
				test.name, err)
			continue
		}

		// Ensure both non-neutered and neutered keys are zeroed
		// properly.
		key.Zero()
		if !testZeroed(i, test.name+" deserialized not neutered", key) {
			continue
		}
		neuteredKey.Zero()
		if !testZeroed(i, test.name+" deserialized neutered", key) {
			continue
		}
	}
}

// TestMaximumDepth ensures that attempting to retrieve a child key when already
// at the maximum depth is not allowed.  The serialization of a BIP32 key uses
// uint8 to encode the depth.  This implicitly bounds the depth of the tree to
// 255 derivations.  Here we test that an error is returned after 'max uint8'.
func TestMaximumDepth(t *testing.T) {
	net := &chaincfg.MainNetParams
	extKey, err := NewMaster([]byte(`abcd1234abcd1234abcd1234abcd1234`), net)
	if err != nil {
		t.Fatalf("NewMaster: unexpected error: %v", err)
	}

	for i := uint8(0); i < math.MaxUint8; i++ {
		if extKey.Depth() != i {
			t.Fatalf("extendedkey depth %d should match expected value %d",
				extKey.Depth(), i)
		}
		newKey, err := extKey.Derive(1)
		if err != nil {
			t.Fatalf("Derive: unexpected error: %v", err)
		}
		extKey = newKey
	}

	noKey, err := extKey.Derive(1)
	if err != ErrDeriveBeyondMaxDepth {
		t.Fatalf("Derive: mismatched error: want %v, got %v",
			ErrDeriveBeyondMaxDepth, err)
	}
	if noKey != nil {
		t.Fatal("Derive: deriving 256th key should not succeed")
	}
}

// TestCloneWithVersion ensures proper conversion between standard and SLIP132
// extended keys.
//
// The following tool was used for generating the tests:
//
//	https://jlopp.github.io/xpub-converter
func TestCloneWithVersion(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		version []byte
		want    string
		wantErr error
	}{
		{
			name:    "test xpub to zpub",
			key:     "xpub661MyMwAqRbcG5SaAT7xLF3TzgYfWA4739rpoo48wQXSz19vcVTpr48H7UcGknLrDbNP76Qws33r5mARVAAsTFKYeK5DTZtSdNVz5i4Bg4q",
			version: []byte{0x04, 0xb2, 0x47, 0x46},
			want:    "zpub6jftahH18ngZxfpoqAhCkREULcqZPQ36sNuGNaquhRHD6CnP7onx6BSZ9tXSkbeh2sbzc3c4nMkwrLPYvYzu3igkNzU4dPXRApdGrqoLi3L",
		},
		{
			name:    "test zpub to xpub",
			key:     "zpub6jftahH18ngZxUuv6oSniLNrBCSSE1B4EEU59bwTCEt8x6aS6b2mdfLxbS4QS53g85SWWP6wexqeer516433gYpZQoJie2tcMYdJ1SYYYAL",
			version: []byte{0x04, 0x88, 0xb2, 0x1e},
			want:    "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8",
		},
		{
			name:    "test xprv to zprv",
			key:     "xprv9s21ZrQH143K3bN74Rawy76jSeiB6hLFfvwE1QeXP4zU7Cpn4x9aJFooGCtqa35LQEKj73tNNN6h4cYHBYs5je9oGMTx61iGFtrwuGFTZ8p",
			version: []byte{0x04, 0xb2, 0x43, 0x0c},
			want:    "zprvAWgYBBk7JR8GkBkLj9ACPHHjnb14ywKFW9yfaCSJ95kEDQTEaGUhYP85Jcp1ZrPBDWZLc15VHgonqBmQcwh7L7X112roFqMEoLzEgPJ4fwr",
		},
		{
			name:    "test zprv to xprv",
			key:     "zprvAWgYBBk7JR8GjzqSzmunMCS7dAbwpYTCs1YUMDXqduMA5JFHZ3iX5s2UkAR6vBdcCYYa1S5o1fVLrKsrnpCQ4WpUd6aVUWP1bS2Yy5DoaKv",
			version: []byte{0x04, 0x88, 0xad, 0xe4},
			want:    "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
		},
		{
			name:    "test invalid key id",
			key:     "zprvAWgYBBk7JR8GjzqSzmunMCS7dAbwpYTCs1YUMDXqduMA5JFHZ3iX5s2UkAR6vBdcCYYa1S5o1fVLrKsrnpCQ4WpUd6aVUWP1bS2Yy5DoaKv",
			version: []byte{0x4B, 0x1D},
			wantErr: chaincfg.ErrUnknownHDKeyID,
		},
	}

	for i, test := range tests {
		extKey, err := NewKeyFromString(test.key)
		if err != nil {
			panic(err) // This is never expected to fail.
		}

		got, err := extKey.CloneWithVersion(test.version)
		if !reflect.DeepEqual(err, test.wantErr) {
			t.Errorf("CloneWithVersion #%d (%s): unexpected error -- "+
				"want %v, got %v", i, test.name, test.wantErr, err)
			continue
		}

		if test.wantErr == nil {
			if k := got.String(); k != test.want {
				t.Errorf("CloneWithVersion #%d (%s): "+
					"got %s, want %s", i, test.name, k, test.want)
				continue
			}
		}
	}
}

// TestLeadingZero ensures that deriving children from keys with a leading zero byte is done according
// to the BIP-32 standard and that the legacy method generates a backwards-compatible result.
func TestLeadingZero(t *testing.T) {
	// The 400th seed results in a m/0' public key with a leading zero, allowing us to test
	// the desired behavior.
	ii := 399
	seed := make([]byte, 32)
	binary.BigEndian.PutUint32(seed[28:], uint32(ii))
	masterKey, err := NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("hdkeychain.NewMaster failed: %v", err)
	}
	child0, err := masterKey.Derive(0 + HardenedKeyStart)
	if err != nil {
		t.Fatalf("masterKey.Derive failed: %v", err)
	}
	if child0.IsAffectedByIssue172() {
		t.Fatalf("expected child0 to NOT be affected by issue 172, but it was")
	}
	child1, err := child0.Derive(0 + HardenedKeyStart)
	if err != nil {
		t.Fatalf("child0.Derive failed: %v", err)
	}
	if child1.IsAffectedByIssue172() {
		t.Fatalf("did not expect child1 to be affected by issue 172")
	}

	// This is the correct result based on BIP32
	expected := "515210f0271acefcd01d8b66a3c4e51ef39ab79fb91aa8e28b44ea628834f9d5"
	if ret := hex.EncodeToString(child1.key); ret != expected {
		t.Errorf("incorrect standard BIP32 derivation, got: %s want: %s", ret, expected)
	}

	if child1.IsAffectedByIssue172() {
		t.Error("child 1 should not be affected by issue 172")
	}
}
