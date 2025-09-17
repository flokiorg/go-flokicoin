package main

import (
    "bytes"
    "encoding/hex"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/flokiorg/go-flokicoin/blockchain"
    "github.com/flokiorg/go-flokicoin/database"
    _ "github.com/flokiorg/go-flokicoin/database/ffldb"
    "github.com/flokiorg/go-flokicoin/mining"
    "github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
    "github.com/flokiorg/go-flokicoin/chaincfg"
    "github.com/flokiorg/go-flokicoin/chainjson"
    "github.com/flokiorg/go-flokicoin/chainutil"
    "github.com/flokiorg/go-flokicoin/mempool"
    "github.com/flokiorg/go-flokicoin/wire"
    "github.com/stretchr/testify/require"
)

// TestHandleTestMempoolAcceptFailDecode checks that when invalid hex string is
// used as the raw txns, the corresponding error is returned.
func TestHandleTestMempoolAcceptFailDecode(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	// Create a testing server.
	s := &rpcServer{}

	testCases := []struct {
		name            string
		txns            []string
		expectedErrCode chainjson.RPCErrorCode
	}{
		{
			name:            "hex decode fail",
			txns:            []string{"invalid"},
			expectedErrCode: chainjson.ErrRPCDecodeHexString,
		},
		{
			name:            "tx decode fail",
			txns:            []string{"696e76616c6964"},
			expectedErrCode: chainjson.ErrRPCDeserialization,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a request that uses invalid raw txns.
			cmd := chainjson.NewTestMempoolAcceptCmd(tc.txns, 0)

			// Call the method under test.
			closeChan := make(chan struct{})
			result, err := handleTestMempoolAccept(
				s, cmd, closeChan,
			)

			// Ensure the expected error is returned.
			require.Error(err)
			rpcErr, ok := err.(*chainjson.RPCError)
			require.True(ok)
			require.Equal(tc.expectedErrCode, rpcErr.Code)

			// No result should be returned.
			require.Nil(result)
		})
	}
}

var (
	//
	// txHex1 is taken from `txscript/data/tx_valid.json`.
	txHex1 = "0100000001b14bdcbc3e01bdaad36cc08e81e69c82e1060bc14e518db2b" +
		"49aa43ad90ba26000000000490047304402203f16c6f40162ab686621ef3" +
		"000b04e75418a0c0cb2d8aebeac894ae360ac1e780220ddc15ecdfc3507a" +
		"c48e1681a33eb60996631bf6bf5bc0a0682c4db743ce7ca2b01ffffffff0" +
		"140420f00000000001976a914660d4ef3a743e3e696ad990364e555c271a" +
		"d504b88ac00000000"

	// txHex2 is taken from `txscript/data/tx_valid.json`.
	txHex2 = "0100000001b14bdcbc3e01bdaad36cc08e81e69c82e1060bc14e518db2b" +
		"49aa43ad90ba260000000004a0048304402203f16c6f40162ab686621ef3" +
		"000b04e75418a0c0cb2d8aebeac894ae360ac1e780220ddc15ecdfc3507a" +
		"c48e1681a33eb60996631bf6bf5bc0a0682c4db743ce7ca2bab01fffffff" +
		"f0140420f00000000001976a914660d4ef3a743e3e696ad990364e555c27" +
		"1ad504b88ac00000000"

	// txHex3 is taken from `txscript/data/tx_valid.json`.
	txHex3 = "0100000001b14bdcbc3e01bdaad36cc08e81e69c82e1060bc14e518db2b" +
		"49aa43ad90ba260000000004a01ff47304402203f16c6f40162ab686621e" +
		"f3000b04e75418a0c0cb2d8aebeac894ae360ac1e780220ddc15ecdfc350" +
		"7ac48e1681a33eb60996631bf6bf5bc0a0682c4db743ce7ca2b01fffffff" +
		"f0140420f00000000001976a914660d4ef3a743e3e696ad990364e555c27" +
		"1ad504b88ac00000000"
)

// decodeTxHex decodes the given hex string into a transaction.
func decodeTxHex(t *testing.T, txHex string) *chainutil.Tx {
	rawBytes, err := hex.DecodeString(txHex)
	require.NoError(t, err)
	tx, err := chainutil.NewTxFromBytes(rawBytes)
	require.NoError(t, err)

	return tx
}

// TestHandleTestMempoolAcceptMixedResults checks that when different txns get
// different responses from calling the mempool method `CheckMempoolAcceptance`
// their results are correctly returned.
func TestHandleTestMempoolAcceptMixedResults(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	// Create a mock mempool.
	mm := &mempool.MockTxMempool{}

	// Create a testing server with the mock mempool.
	s := &rpcServer{cfg: rpcserverConfig{
		TxMemPool: mm,
	}}

	// Decode the hex so we can assert the mock mempool is called with it.
	tx1 := decodeTxHex(t, txHex1)
	tx2 := decodeTxHex(t, txHex2)
	tx3 := decodeTxHex(t, txHex3)

	// Create a slice to hold the expected results. We will use three txns
	// so we expect threeresults.
	expectedResults := make([]*chainjson.TestMempoolAcceptResult, 3)

	// We now mock the first call to `CheckMempoolAcceptance` to return an
	// error.
	dummyErr := errors.New("dummy error")
	mm.On("CheckMempoolAcceptance", tx1).Return(nil, dummyErr).Once()

	// Since the call failed, we expect the first result to give us the
	// error.
	expectedResults[0] = &chainjson.TestMempoolAcceptResult{
		Txid:         tx1.Hash().String(),
		Wtxid:        tx1.WitnessHash().String(),
		Allowed:      false,
		RejectReason: dummyErr.Error(),
	}

	// We mock the second call to `CheckMempoolAcceptance` to return a
	// result saying the tx is missing inputs.
	mm.On("CheckMempoolAcceptance", tx2).Return(
		&mempool.MempoolAcceptResult{
			MissingParents: []*chainhash.Hash{},
		}, nil,
	).Once()

	// We expect the second result to give us the missing-inputs error.
	expectedResults[1] = &chainjson.TestMempoolAcceptResult{
		Txid:         tx2.Hash().String(),
		Wtxid:        tx2.WitnessHash().String(),
		Allowed:      false,
		RejectReason: "missing-inputs",
	}

	// We mock the third call to `CheckMempoolAcceptance` to return a
	// result saying the tx allowed.
	const feeSats = chainutil.Amount(1000)
	mm.On("CheckMempoolAcceptance", tx3).Return(
		&mempool.MempoolAcceptResult{
			TxFee:  feeSats,
			TxSize: 100,
		}, nil,
	).Once()

	// We expect the third result to give us the fee details.
	expectedResults[2] = &chainjson.TestMempoolAcceptResult{
		Txid:    tx3.Hash().String(),
		Wtxid:   tx3.WitnessHash().String(),
		Allowed: true,
		Vsize:   100,
		Fees: &chainjson.TestMempoolAcceptFees{
			Base:             feeSats.ToFLC(),
			EffectiveFeeRate: feeSats.ToFLC() * 1e3 / 100,
		},
	}

	// Create a mock request with default max fee rate of 0.1 FLC/KvB.
	cmd := chainjson.NewTestMempoolAcceptCmd(
		[]string{txHex1, txHex2, txHex3}, 0.1,
	)

	// Call the method handler and assert the expected results are
	// returned.
	closeChan := make(chan struct{})
	results, err := handleTestMempoolAccept(s, cmd, closeChan)
	require.NoError(err)
	require.Equal(expectedResults, results)

	// Assert the mocked method is called as expected.
	mm.AssertExpectations(t)
}

// TestValidateFeeRate checks that `validateFeeRate` behaves as expected.
func TestValidateFeeRate(t *testing.T) {
	t.Parallel()

	const (
		// testFeeRate is in FLC/kvB.
		testFeeRate = 0.1

		// testTxSize is in vb.
		testTxSize = 100

		// testFeeSats is in sats.
		// We have 0.1FLC/kvB =
		//   0.1 * 1e8 sats/kvB =
		//   0.1 * 1e8 / 1e3 sats/vb = 0.1 * 1e5 sats/vb.
		testFeeSats = chainutil.Amount(testFeeRate * 1e5 * testTxSize)
	)

	testCases := []struct {
		name         string
		feeSats      chainutil.Amount
		txSize       int64
		maxFeeRate   float64
		expectedFees *chainjson.TestMempoolAcceptFees
		allowed      bool
	}{
		{
			// When the fee rate(0.1) is above the max fee
			// rate(0.01), we expect a nil result and false.
			name:         "fee rate above max",
			feeSats:      testFeeSats,
			txSize:       testTxSize,
			maxFeeRate:   testFeeRate / 10,
			expectedFees: nil,
			allowed:      false,
		},
		{
			// When the fee rate(0.1) is no greater than the max
			// fee rate(0.1), we expect a result and true.
			name:       "fee rate below max",
			feeSats:    testFeeSats,
			txSize:     testTxSize,
			maxFeeRate: testFeeRate,
			expectedFees: &chainjson.TestMempoolAcceptFees{
				Base:             testFeeSats.ToFLC(),
				EffectiveFeeRate: testFeeRate,
			},
			allowed: true,
		},
		{
			// When the fee rate(1) is above the default max fee
			// rate(0.1), we expect a nil result and false.
			name:         "fee rate above default max",
			feeSats:      testFeeSats,
			txSize:       testTxSize / 10,
			expectedFees: nil,
			allowed:      false,
		},
		{
			// When the fee rate(0.1) is no greater than the
			// default max fee rate(0.1), we expect a result and
			// true.
			name:    "fee rate below default max",
			feeSats: testFeeSats,
			txSize:  testTxSize,
			expectedFees: &chainjson.TestMempoolAcceptFees{
				Base:             testFeeSats.ToFLC(),
				EffectiveFeeRate: testFeeRate,
			},
			allowed: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			result, allowed := validateFeeRate(
				tc.feeSats, tc.txSize, tc.maxFeeRate,
			)

			require.Equal(tc.expectedFees, result)
			require.Equal(tc.allowed, allowed)
		})
	}
}

// ---------------------------
// AuxPoW RPC handler tests
// ---------------------------

// stubTxSource implements mining.TxSource with minimal behavior.
type stubTxSource struct{ last time.Time }

func (s stubTxSource) LastUpdated() time.Time                         { return s.last }
func (s stubTxSource) MiningDescs() []*mining.TxDesc                 { return nil }
func (s stubTxSource) HaveTransaction(h *chainhash.Hash) bool        { return false }

// mkTempDB creates a temporary ffldb database path and returns path and cleanup.
func mkTempDB(t *testing.T, net wire.FlokicoinNet) (string, func()) {
    t.Helper()
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "ffldb-aux-tests")
    return dbPath, func() { os.RemoveAll(dir) }
}

// mkChain constructs a minimal blockchain with provided params and empty DB.
func mkChain(t *testing.T, params *chaincfg.Params) (*blockchain.BlockChain, func()) {
    t.Helper()
    // Initialize logging rotator to avoid nil-pointer in chain logs during tests.
    initLogRotator(filepath.Join(t.TempDir(), "auxpow-rpc-tests.log"))
    setLogLevels("off")
    dbPath, cleanup := mkTempDB(t, params.Net)
    db, err := database.Create("ffldb", dbPath, params.Net)
    if err != nil {
        t.Fatalf("create db: %v", err)
    }
    // Build chain
    cfg := &blockchain.Config{
        DB:          db,
        ChainParams: params,
        TimeSource:  blockchain.NewMedianTime(),
    }
    chain, err := blockchain.New(cfg)
    if err != nil {
        t.Fatalf("new chain: %v", err)
    }
    return chain, func() {
        db.Close()
        cleanup()
    }
}

// mkP2PKH returns a valid P2PKH address string for params.
func mkP2PKH(t *testing.T, params *chaincfg.Params) string {
    t.Helper()
    var h20 [20]byte
    for i := 0; i < 20; i++ {
        h20[i] = byte(i + 1)
    }
    addr, err := chainutil.NewAddressPubKeyHash(h20[:], params)
    require.NoError(t, err)
    return addr.EncodeAddress()
}

// mkTemplate builds a minimal block template for next height with a coinbase.
func mkTemplate(prev chainhash.Hash, height int32, bits uint32, coinbaseValue int64) *mining.BlockTemplate {
    cb := &wire.MsgTx{
        Version: 1,
        TxIn: []*wire.TxIn{{
            PreviousOutPoint: wire.OutPoint{},
            SignatureScript:  []byte{0x51}, // OP_TRUE placeholder
            Sequence:         0xffffffff,
        }},
        TxOut: []*wire.TxOut{{
            Value:    coinbaseValue,
            PkScript: []byte{0x6a}, // OP_RETURN placeholder; set later in handler
        }},
        LockTime: 0,
    }
    msg := &wire.MsgBlock{
        Header: wire.BlockHeader{
            Version:    1,
            PrevBlock:  prev,
            MerkleRoot: chainhash.Hash{},
            Timestamp:  time.Unix(time.Now().Unix(), 0),
            Bits:       bits,
            Nonce:      0,
        },
        Transactions: []*wire.MsgTx{cb},
    }
    return &mining.BlockTemplate{Block: msg, Height: height, ValidPayAddress: false}
}

// mkAuxServer constructs an rpcServer with a prepared template and adjusted params for aux tests.
func mkAuxServer(t *testing.T, params chaincfg.Params, tmpl *mining.BlockTemplate) *rpcServer {
    t.Helper()

    // Lower gating to current height (0) by setting AuxpowHeightEffective to 0.
    params.AuxpowHeightEffective = 0

    chain, cleanup := mkChain(t, &params)
    t.Cleanup(cleanup)

    ts := blockchain.NewMedianTime()
    gen := mining.NewBlkTmplGenerator(&mining.Policy{}, &params, stubTxSource{last: time.Now()}, chain, ts, nil, nil)

    s := &rpcServer{
        cfg: rpcserverConfig{
            Chain:       chain,
            ChainParams: &params,
            Generator:   gen,
            TimeSource:  ts,
            AuxCache:    newAuxCache(2 * time.Minute),
        },
        gbtWorkState: newGbtWorkState(ts),
    }

    // Preload template and state so updateBlockTemplate() goes fast path.
    st := s.gbtWorkState
    st.Lock()
    st.template = tmpl
    bs := s.cfg.Chain.BestSnapshot()
    st.prevHash = &bs.Hash
    st.lastGenerated = time.Now()
    st.lastTxUpdate = gen.TxSource().LastUpdated()
    st.Unlock()

    return s
}

func TestCreateAuxBlock_Table(t *testing.T) {
    type wantErr struct{ code chainjson.RPCErrorCode }

    cases := []struct {
        name      string
        setup     func(t *testing.T) *rpcServer
        address   string
        wantErr   *wantErr
        verifyOK  func(t *testing.T, s *rpcServer, res *chainjson.CreateAuxBlockResult)
    }{
        {
            name: "success",
            setup: func(t *testing.T) *rpcServer {
                params := chaincfg.RegressionNetParams
                prev := *params.GenesisHash
                tmpl := mkTemplate(prev, 1, params.PowLimitBits, 50*1e8)
                return mkAuxServer(t, params, tmpl)
            },
            address: "<valid>",
            verifyOK: func(t *testing.T, s *rpcServer, res *chainjson.CreateAuxBlockResult) {
                require := require.New(t)
                tmpl := s.gbtWorkState.template
                require.NotNil(tmpl)
                require.Equal(int(s.cfg.ChainParams.AuxpowChainId), res.ChainID)
                require.Equal(tmpl.Block.Header.PrevBlock.String(), res.PreviousBlockHash)
                require.NotEmpty(res.Hash)
                require.NotEmpty(res.Target)
                require.EqualValues(tmpl.Height, res.Height)
                require.Equal(fmt.Sprintf("%08x", tmpl.Block.Header.Bits), res.Bits)
                require.EqualValues(tmpl.Block.Transactions[0].TxOut[0].Value, res.CoinbaseValue)
                h, err := chainhash.NewHashFromStr(res.Hash)
                require.NoError(err)
                cand, ok := s.cfg.AuxCache.get(*h)
                require.True(ok)
                require.Equal(*h, cand.Hash)
            },
        },
        {
            name: "not-supported",
            setup: func(t *testing.T) *rpcServer {
                params := chaincfg.RegressionNetParams
                params.AuxpowHeightEffective = 100
                chain, cleanup := mkChain(t, &params)
                t.Cleanup(cleanup)
                ts := blockchain.NewMedianTime()
                return &rpcServer{cfg: rpcserverConfig{
                    Chain:       chain,
                    ChainParams: &params,
                    Generator:   mining.NewBlkTmplGenerator(&mining.Policy{}, &params, stubTxSource{last: time.Now()}, chain, ts, nil, nil),
                    TimeSource:  ts,
                    AuxCache:    newAuxCache(2 * time.Minute),
                }, gbtWorkState: newGbtWorkState(ts)}
            },
            address: mkP2PKH(t, &chaincfg.RegressionNetParams),
            wantErr: &wantErr{chainjson.ErrRPCAuxNotSupported},
        },
        {
            name: "invalid-address",
            setup: func(t *testing.T) *rpcServer {
                params := chaincfg.RegressionNetParams
                prev := *params.GenesisHash
                tmpl := mkTemplate(prev, 1, params.PowLimitBits, 25*1e8)
                return mkAuxServer(t, params, tmpl)
            },
            address: "invalid",
            wantErr: &wantErr{chainjson.ErrRPCInvalidAddressOrKey},
        },
        {
            name: "address-wrong-net",
            setup: func(t *testing.T) *rpcServer {
                params := chaincfg.RegressionNetParams
                prev := *params.GenesisHash
                tmpl := mkTemplate(prev, 1, params.PowLimitBits, 25*1e8)
                return mkAuxServer(t, params, tmpl)
            },
            address: mkP2PKH(t, &chaincfg.MainNetParams),
            wantErr: &wantErr{chainjson.ErrRPCInvalidAddressOrKey},
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            s := tc.setup(t)
            addr := tc.address
            if addr == "<valid>" {
                addr = mkP2PKH(t, s.cfg.ChainParams)
            }
            cmd := &chainjson.CreateAuxBlockCmd{Address: addr}
            resAny, err := handleCreateAuxBlock(s, cmd, make(chan struct{}))
            if tc.wantErr != nil {
                require.Error(t, err)
                rpcErr := err.(*chainjson.RPCError)
                require.Equal(t, tc.wantErr.code, rpcErr.Code)
                return
            }
            require.NoError(t, err)
            res := resAny.(*chainjson.CreateAuxBlockResult)
            if tc.verifyOK != nil {
                tc.verifyOK(t, s, res)
            }
        })
    }
}

func TestSubmitAuxBlock_Table(t *testing.T) {
    type wantErr struct{ code chainjson.RPCErrorCode }

    cases := []struct {
        name    string
        setup   func(t *testing.T) (*rpcServer, *chainjson.SubmitAuxBlockCmd)
        wantRes *chainjson.SubmitAuxBlockResult
        wantErr *wantErr
    }{
        {
            name: "not-supported",
            setup: func(t *testing.T) (*rpcServer, *chainjson.SubmitAuxBlockCmd) {
                params := chaincfg.RegressionNetParams
                params.AuxpowHeightEffective = 10
                chain, cleanup := mkChain(t, &params)
                t.Cleanup(cleanup)
                s := &rpcServer{cfg: rpcserverConfig{Chain: chain, ChainParams: &params, AuxCache: newAuxCache(2 * time.Minute)}}
                cmd := &chainjson.SubmitAuxBlockCmd{Hash: chainhash.Hash{}.String(), AuxPow: ""}
                return s, cmd
            },
            wantErr: &wantErr{chainjson.ErrRPCAuxNotSupported},
        },
        {
            name: "no-cache",
            setup: func(t *testing.T) (*rpcServer, *chainjson.SubmitAuxBlockCmd) {
                params := chaincfg.RegressionNetParams
                params.AuxpowHeightEffective = 0
                chain, cleanup := mkChain(t, &params)
                t.Cleanup(cleanup)
                s := &rpcServer{cfg: rpcserverConfig{Chain: chain, ChainParams: &params, AuxCache: nil}}
                cmd := &chainjson.SubmitAuxBlockCmd{Hash: chainhash.Hash{}.String(), AuxPow: ""}
                return s, cmd
            },
            wantRes: ptrSubmit(false),
        },
        {
            name: "invalid-hash",
            setup: func(t *testing.T) (*rpcServer, *chainjson.SubmitAuxBlockCmd) {
                params := chaincfg.RegressionNetParams
                params.AuxpowHeightEffective = 0
                chain, cleanup := mkChain(t, &params)
                t.Cleanup(cleanup)
                s := &rpcServer{cfg: rpcserverConfig{Chain: chain, ChainParams: &params, AuxCache: newAuxCache(2 * time.Minute)}}
                cmd := &chainjson.SubmitAuxBlockCmd{Hash: "not-a-hash", AuxPow: ""}
                return s, cmd
            },
            wantErr: &wantErr{chainjson.ErrRPCAuxUnknownHash},
        },
        {
            name: "expired-candidate",
            setup: func(t *testing.T) (*rpcServer, *chainjson.SubmitAuxBlockCmd) {
                params := chaincfg.RegressionNetParams
                params.AuxpowHeightEffective = 0
                chain, cleanup := mkChain(t, &params)
                t.Cleanup(cleanup)
                s := &rpcServer{cfg: rpcserverConfig{Chain: chain, ChainParams: &params, AuxCache: newAuxCache(2 * time.Minute)}}
                cmd := &chainjson.SubmitAuxBlockCmd{Hash: chainhash.Hash{}.String(), AuxPow: ""}
                return s, cmd
            },
            wantErr: &wantErr{chainjson.ErrRPCAuxCandidateExpired},
        },
        {
            name: "invalid-auxpow-hex",
            setup: func(t *testing.T) (*rpcServer, *chainjson.SubmitAuxBlockCmd) {
                params := chaincfg.RegressionNetParams
                params.AuxpowHeightEffective = 0
                chain, cleanup := mkChain(t, &params)
                t.Cleanup(cleanup)
                s := &rpcServer{cfg: rpcserverConfig{Chain: chain, ChainParams: &params, AuxCache: newAuxCache(2 * time.Minute)}}
                blk := &wire.MsgBlock{Header: wire.BlockHeader{PrevBlock: *params.GenesisHash, Bits: params.PowLimitBits}}
                cand := &AuxCandidate{Hash: chainhash.DoubleHashH([]byte("X")), Height: 1, Block: blk}
                s.cfg.AuxCache.put(cand)
                cmd := &chainjson.SubmitAuxBlockCmd{Hash: cand.Hash.String(), AuxPow: "zz-not-hex"}
                return s, cmd
            },
            wantErr: &wantErr{chainjson.ErrRPCAuxInvalidAuxPow},
        },
        {
            name: "parse-auxpow-fail",
            setup: func(t *testing.T) (*rpcServer, *chainjson.SubmitAuxBlockCmd) {
                params := chaincfg.RegressionNetParams
                params.AuxpowHeightEffective = 0
                chain, cleanup := mkChain(t, &params)
                t.Cleanup(cleanup)
                s := &rpcServer{cfg: rpcserverConfig{Chain: chain, ChainParams: &params, AuxCache: newAuxCache(2 * time.Minute)}}
                blk := &wire.MsgBlock{Header: wire.BlockHeader{PrevBlock: *params.GenesisHash, Bits: params.PowLimitBits}}
                cand := &AuxCandidate{Hash: chainhash.DoubleHashH([]byte("Y")), Height: 1, Block: blk}
                s.cfg.AuxCache.put(cand)
                cmd := &chainjson.SubmitAuxBlockCmd{Hash: cand.Hash.String(), AuxPow: hex.EncodeToString([]byte{})}
                return s, cmd
            },
            wantErr: &wantErr{chainjson.ErrRPCAuxInvalidAuxPow},
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            s, cmd := tc.setup(t)
            res, err := handleSubmitAuxBlock(s, cmd, make(chan struct{}))
            if tc.wantErr != nil {
                require.Error(t, err)
                rpcErr := err.(*chainjson.RPCError)
                require.Equal(t, tc.wantErr.code, rpcErr.Code)
                return
            }
            require.NoError(t, err)
            if tc.wantRes != nil {
                require.Equal(t, *tc.wantRes, res)
            }
        })
    }
}

func ptrSubmit(b bool) *chainjson.SubmitAuxBlockResult { r := chainjson.SubmitAuxBlockResult(b); return &r }

// Optional: build a minimal auxpow header just to go past parse for future tests.
func buildMinimalAuxPowHex(t *testing.T) string {
    var aph wire.AuxPowHeader
    var buf bytes.Buffer
    // empty fields will fail Check later, but parse succeeds only if encoding ok.
    require.NoError(t, aph.Serialize(&buf))
    return hex.EncodeToString(buf.Bytes())
}


// TestHandleTestMempoolAcceptFees checks that the `Fees` field is correctly
// populated based on the max fee rate and the tx being checked.
func TestHandleTestMempoolAcceptFees(t *testing.T) {
	t.Parallel()

	// Create a mock mempool.
	mm := &mempool.MockTxMempool{}

	// Create a testing server with the mock mempool.
	s := &rpcServer{cfg: rpcserverConfig{
		TxMemPool: mm,
	}}

	const (
		// Set transaction's fee rate to be 0.2FLC/kvB.
		feeRate = defaultMaxFeeRate * 2

		// txSize is 100vb.
		txSize = 100

		// feeSats is 2e6 sats.
		feeSats = feeRate * 1e8 * txSize / 1e3
	)

	testCases := []struct {
		name         string
		maxFeeRate   float64
		txHex        string
		rejectReason string
		allowed      bool
	}{
		{
			// When the fee rate(0.2) used by the tx is below the
			// max fee rate(2) specified, the result should allow
			// it.
			name:       "below max fee rate",
			maxFeeRate: feeRate * 10,
			txHex:      txHex1,
			allowed:    true,
		},
		{
			// When the fee rate(0.2) used by the tx is above the
			// max fee rate(0.02) specified, the result should
			// disallow it.
			name:         "above max fee rate",
			maxFeeRate:   feeRate / 10,
			txHex:        txHex1,
			allowed:      false,
			rejectReason: "max-fee-exceeded",
		},
		{
			// When the max fee rate is not set, the default
			// 0.1FLC/kvB is used and the fee rate(0.2) used by the
			// tx is above it, the result should disallow it.
			name:         "above default max fee rate",
			txHex:        txHex1,
			allowed:      false,
			rejectReason: "max-fee-exceeded",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			// Decode the hex so we can assert the mock mempool is
			// called with it.
			tx := decodeTxHex(t, txHex1)

			// We mock the call to `CheckMempoolAcceptance` to
			// return the result.
			mm.On("CheckMempoolAcceptance", tx).Return(
				&mempool.MempoolAcceptResult{
					TxFee:  feeSats,
					TxSize: txSize,
				}, nil,
			).Once()

			// We expect the third result to give us the fee
			// details.
			expected := &chainjson.TestMempoolAcceptResult{
				Txid:    tx.Hash().String(),
				Wtxid:   tx.WitnessHash().String(),
				Allowed: tc.allowed,
			}

			if tc.allowed {
				expected.Vsize = txSize
				expected.Fees = &chainjson.TestMempoolAcceptFees{
					Base:             feeSats / 1e8,
					EffectiveFeeRate: feeRate,
				}
			} else {
				expected.RejectReason = tc.rejectReason
			}

			// Create a mock request with specified max fee rate.
			cmd := chainjson.NewTestMempoolAcceptCmd(
				[]string{txHex1}, tc.maxFeeRate,
			)

			// Call the method handler and assert the expected
			// result is returned.
			closeChan := make(chan struct{})
			r, err := handleTestMempoolAccept(s, cmd, closeChan)
			require.NoError(err)

			// Check the interface type.
			results, ok := r.([]*chainjson.TestMempoolAcceptResult)
			require.True(ok)

			// Expect exactly one result.
			require.Len(results, 1)

			// Check the result is returned as expected.
			require.Equal(expected, results[0])

			// Assert the mocked method is called as expected.
			mm.AssertExpectations(t)
		})
	}
}

// TestGetTxSpendingPrevOut checks that handleGetTxSpendingPrevOut handles the
// cmd as expected.
func TestGetTxSpendingPrevOut(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	// Create a mock mempool.
	mm := &mempool.MockTxMempool{}
	defer mm.AssertExpectations(t)

	// Create a testing server with the mock mempool.
	s := &rpcServer{cfg: rpcserverConfig{
		TxMemPool: mm,
	}}

	// First, check the error case.
	//
	// Create a request that will cause an error.
	cmd := &chainjson.GetTxSpendingPrevOutCmd{
		Outputs: []*chainjson.GetTxSpendingPrevOutCmdOutput{
			{Txid: "invalid"},
		},
	}

	// Call the method handler and assert the error is returned.
	closeChan := make(chan struct{})
	results, err := handleGetTxSpendingPrevOut(s, cmd, closeChan)
	require.Error(err)
	require.Nil(results)

	// We now check the normal case. Two outputs will be tested - one found
	// in mempool and other not.
	//
	// Decode the hex so we can assert the mock mempool is called with it.
	tx := decodeTxHex(t, txHex1)

	// Create testing outpoints.
	opInMempool := wire.OutPoint{Hash: chainhash.Hash{1}, Index: 1}
	opNotInMempool := wire.OutPoint{Hash: chainhash.Hash{2}, Index: 1}

	// We only expect to see one output being found as spent in mempool.
	expectedResults := []*chainjson.GetTxSpendingPrevOutResult{
		{
			Txid:         opInMempool.Hash.String(),
			Vout:         opInMempool.Index,
			SpendingTxid: tx.Hash().String(),
		},
		{
			Txid: opNotInMempool.Hash.String(),
			Vout: opNotInMempool.Index,
		},
	}

	// We mock the first call to `CheckSpend` to return a result saying the
	// output is found.
	mm.On("CheckSpend", opInMempool).Return(tx).Once()

	// We mock the second call to `CheckSpend` to return a result saying the
	// output is NOT found.
	mm.On("CheckSpend", opNotInMempool).Return(nil).Once()

	// Create a request with the above outputs.
	cmd = &chainjson.GetTxSpendingPrevOutCmd{
		Outputs: []*chainjson.GetTxSpendingPrevOutCmdOutput{
			{
				Txid: opInMempool.Hash.String(),
				Vout: opInMempool.Index,
			},
			{
				Txid: opNotInMempool.Hash.String(),
				Vout: opNotInMempool.Index,
			},
		},
	}

	// Call the method handler and assert the expected result is returned.
	closeChan = make(chan struct{})
	results, err = handleGetTxSpendingPrevOut(s, cmd, closeChan)
	require.NoError(err)
	require.Equal(expectedResults, results)
}
