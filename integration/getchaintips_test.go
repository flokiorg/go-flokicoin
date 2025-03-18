package integration

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/chainjson"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/integration/rpctest"
	"github.com/stretchr/testify/require"
)

func getBlockFromString(t *testing.T, hexStr string) *chainutil.Block {
	t.Helper()

	serializedBlock, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatalf("couldn't decode hex string of %s", hexStr)
	}

	block, err := chainutil.NewBlockFromBytes(serializedBlock)
	if err != nil {
		t.Fatalf("couldn't make a new block from bytes. "+
			"Decoded hex string: %s", hexStr)
	}

	return block
}

// compareMultipleChainTips checks that all the expected chain tips are included in got chain tips and
// verifies that the got chain tip matches the expected chain tip.
func compareMultipleChainTips(t *testing.T, gotChainTips, expectedChainTips []*chainjson.GetChainTipsResult) error {
	if len(gotChainTips) != len(expectedChainTips) {
		return fmt.Errorf("Expected %d chaintips but got %d", len(expectedChainTips), len(gotChainTips))
	}

	gotChainTipsMap := make(map[string]chainjson.GetChainTipsResult)
	for _, gotChainTip := range gotChainTips {
		gotChainTipsMap[gotChainTip.Hash] = *gotChainTip
	}

	for _, expectedChainTip := range expectedChainTips {
		gotChainTip, found := gotChainTipsMap[expectedChainTip.Hash]
		if !found {
			return fmt.Errorf("Couldn't find expected chaintip with hash %s", expectedChainTip.Hash)
		}

		require.Equal(t, gotChainTip, *expectedChainTip)
	}

	return nil
}

func TestGetChainTips(t *testing.T) {
	// Has blockhash of "8e81b69bfebe8a4b959d7d8e983c18329f2e32eeed7b95094f6b68218566404f".
	block1Hex := "010000001ceced25aee21f66ba75d9552466f1d7187f21647215a93c56989d92ffec35fe222197599f0281fba135f0352c5acf9b78023e09f57a89dd3aaf9c34ce984aeac0bc6f67ffff7f201eac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1104ffff001d0104096865696768743a2031ffffffff0100e8764817000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

	// Has blockhash of "1d589b7f67400efc6f584ced1fb0a9be0a19613125c7ab27c0d30c5f72c9e762".
	block2Hex := "010000004f40668521686b4f09957bedee322e9f32183c988e7d9d954b8abefe9bb6818e5e06e490b5f4d1d9ff0348036ca225f9828b607c0ced56a61a9e5fe1e46df2e8cabc6f67ffff7f201eac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1104ffff001d0104096865696768743a2032ffffffff0100e8764817000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

	// Has blockhash of "b7d10b0df1d20181615a071e6f3e424a7219058b2d7a8ab090f3e640cb3600de".
	block3Hex := "0100000062e7c9725f0cd3c027abc7253161190abea9b01fed4c586ffc0e40677f9b581dfa5d87ec07526892f30040072ad9469fe5b8ec4609d29f7bda7ee57d43e91b20d4bc6f67ffff7f201fac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1104ffff001d0104096865696768743a2033ffffffff0100e8764817000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

	// Has blockhash of "ed9c7b0f0ad8fe85c3079f8e50bf158d8ee1c0d97855232373388aa6f09cbd75".
	block4Hex := "01000000de0036cb40e6f390b08a7a2d8b0519724a423e6f1e075a618101d2f10d0bd1b71f035ca07326773d27a5f2c07c8dee11cb414074ae0b3d1aab653cc631fe79addebc6f67ffff7f201dac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1104ffff001d0104096865696768743a2034ffffffff0100e8764817000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

	// Has blockhash of "55b160b6538642086d6d58eaf539ef2d70192aa556fb7463cb6ee8f3486cd3d4".
	block2aHex := "010000004f40668521686b4f09957bedee322e9f32183c988e7d9d954b8abefe9bb6818e5e06e490b5f4d1d9ff0348036ca225f9828b607c0ced56a61a9e5fe1e46df2e8debc6f67ffff7f201eac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1104ffff001d0104096865696768743a2032ffffffff0100e8764817000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

	// Has blockhash of "60461e03ae62699c9f5183358ed8a11f7c91f087a6ebb1f0de114e3d8c031fa2".
	block3aHex := "01000000d4d36c48f3e86ecb6374fb56a52a19702def39f5ea586d6d08428653b660b155fa5d87ec07526892f30040072ad9469fe5b8ec4609d29f7bda7ee57d43e91b20f2bc6f67ffff7f201dac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1104ffff001d0104096865696768743a2033ffffffff0100e8764817000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

	// Has blockhash of "3847e92584756de9632eb36b8a45a38b3e8bc6829dfec028b718300894588e91".
	block4aHex := "01000000a21f038c3d4e11def0b1eba687f0917c1fa1d88e3583519f9c6962ae031e46601f035ca07326773d27a5f2c07c8dee11cb414074ae0b3d1aab653cc631fe79ad06bd6f67ffff7f201dac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1104ffff001d0104096865696768743a2034ffffffff0100e8764817000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

	// Has blockhash of "3fb8e2e06e2d91a06cad4c554b1e4ea9e8a824c43615b6861f29ffe425e179c0".
	block5aHex := "01000000918e5894083018b728c0fe9d82c68b3e8ba3458a6bb32e63e96d758425e947383436f9068c79612ddd31abcdee959ea2574d88802f0cda31482ed95375dfe0511abd6f67ffff7f201dac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1104ffff001d0104096865696768743a2035ffffffff0100e8764817000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

	// Has blockhash of "ca488d6c0a3d4449524fb40e65af219eaa1c3ccc8d1090d0d4b3044595f02316".
	block4bHex := "01000000a21f038c3d4e11def0b1eba687f0917c1fa1d88e3583519f9c6962ae031e46601f035ca07326773d27a5f2c07c8dee11cb414074ae0b3d1aab653cc631fe79ad2ebd6f67ffff7f201dac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1104ffff001d0104096865696768743a2034ffffffff0100e8764817000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

	// Set up regtest chain.
	r, err := rpctest.New(&chaincfg.RegressionNetParams, nil, nil, "")
	if err != nil {
		t.Fatal("TestGetChainTips fail. Unable to create primary harness: ", err)
	}
	if err := r.SetUp(true, 0); err != nil {
		t.Fatalf("TestGetChainTips fail. Unable to setup test chain: %v", err)
	}
	defer r.TearDown()

	// Immediately call getchaintips after setting up regtest.
	gotChainTips, err := r.Client.GetChainTips()
	if err != nil {
		t.Fatal(err)
	}
	// We expect a single genesis block.
	expectedChainTips := []*chainjson.GetChainTipsResult{
		{
			Height:    0,
			Hash:      chaincfg.RegressionNetParams.GenesisHash.String(),
			BranchLen: 0,
			Status:    "active",
		},
	}

	err = compareMultipleChainTips(t, gotChainTips, expectedChainTips)
	if err != nil {
		t.Fatalf("TestGetChainTips fail. Error: %v", err)
	}

	// Submit 4 blocks.
	//
	// Our chain view looks like so:
	// (genesis block) -> 1 -> 2 -> 3 -> 4
	blockStrings := []string{block1Hex, block2Hex, block3Hex, block4Hex}
	for _, blockString := range blockStrings {
		block := getBlockFromString(t, blockString)
		err = r.Client.SubmitBlock(block, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	gotChainTips, err = r.Client.GetChainTips()
	if err != nil {
		t.Fatal(err)
	}
	expectedChainTips = []*chainjson.GetChainTipsResult{
		{
			Height:    4,
			Hash:      getBlockFromString(t, blockStrings[len(blockStrings)-1]).Hash().String(),
			BranchLen: 0,
			Status:    "active",
		},
	}
	err = compareMultipleChainTips(t, gotChainTips, expectedChainTips)
	if err != nil {
		t.Fatalf("TestGetChainTips fail. Error: %v", err)
	}

	// Submit 2 blocks that don't build on top of the current active tip.
	//
	// Our chain view looks like so:
	// (genesis block) -> 1 -> 2  -> 3  -> 4  (active)
	//                    \ -> 2a -> 3a       (valid-fork)
	blockStrings = []string{block2aHex, block3aHex}
	for _, blockString := range blockStrings {
		block := getBlockFromString(t, blockString)
		err = r.Client.SubmitBlock(block, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	gotChainTips, err = r.Client.GetChainTips()
	if err != nil {
		t.Fatal(err)
	}
	expectedChainTips = []*chainjson.GetChainTipsResult{
		{
			Height:    4,
			Hash:      getBlockFromString(t, block4Hex).Hash().String(),
			BranchLen: 0,
			Status:    "active",
		},
		{
			Height:    3,
			Hash:      getBlockFromString(t, block3aHex).Hash().String(),
			BranchLen: 2,
			Status:    "valid-fork",
		},
	}
	err = compareMultipleChainTips(t, gotChainTips, expectedChainTips)
	if err != nil {
		t.Fatalf("TestGetChainTips fail. Error: %v", err)
	}

	// Submit a single block that don't build on top of the current active tip.
	//
	// Our chain view looks like so:
	// (genesis block) -> 1 -> 2  -> 3  -> 4   (active)
	//                    \ -> 2a -> 3a -> 4a  (valid-fork)
	block := getBlockFromString(t, block4aHex)
	err = r.Client.SubmitBlock(block, nil)
	if err != nil {
		t.Fatal(err)
	}

	gotChainTips, err = r.Client.GetChainTips()
	if err != nil {
		t.Fatal(err)
	}
	expectedChainTips = []*chainjson.GetChainTipsResult{
		{
			Height:    4,
			Hash:      getBlockFromString(t, block4Hex).Hash().String(),
			BranchLen: 0,
			Status:    "active",
		},
		{
			Height:    4,
			Hash:      getBlockFromString(t, block4aHex).Hash().String(),
			BranchLen: 3,
			Status:    "valid-fork",
		},
	}
	err = compareMultipleChainTips(t, gotChainTips, expectedChainTips)
	if err != nil {
		t.Fatalf("TestGetChainTips fail. Error: %v", err)
	}

	// Submit a single block that changes the active branch to 5a.
	//
	// Our chain view looks like so:
	// (genesis block) -> 1 -> 2  -> 3  -> 4         (valid-fork)
	//                    \ -> 2a -> 3a -> 4a -> 5a  (active)
	block = getBlockFromString(t, block5aHex)
	err = r.Client.SubmitBlock(block, nil)
	if err != nil {
		t.Fatal(err)
	}
	gotChainTips, err = r.Client.GetChainTips()
	if err != nil {
		t.Fatal(err)
	}
	expectedChainTips = []*chainjson.GetChainTipsResult{
		{
			Height:    4,
			Hash:      getBlockFromString(t, block4Hex).Hash().String(),
			BranchLen: 3,
			Status:    "valid-fork",
		},
		{
			Height:    5,
			Hash:      getBlockFromString(t, block5aHex).Hash().String(),
			BranchLen: 0,
			Status:    "active",
		},
	}
	err = compareMultipleChainTips(t, gotChainTips, expectedChainTips)
	if err != nil {
		t.Fatalf("TestGetChainTips fail. Error: %v", err)
	}

	// Submit a single block that builds on top of 3a.
	//
	// Our chain view looks like so:
	// (genesis block) -> 1 -> 2  -> 3  -> 4         (valid-fork)
	//                    \ -> 2a -> 3a -> 4a -> 5a  (active)
	//                                \ -> 4b        (valid-fork)
	block = getBlockFromString(t, block4bHex)
	err = r.Client.SubmitBlock(block, nil)
	if err != nil {
		t.Fatal(err)
	}
	gotChainTips, err = r.Client.GetChainTips()
	if err != nil {
		t.Fatal(err)
	}
	expectedChainTips = []*chainjson.GetChainTipsResult{
		{
			Height:    4,
			Hash:      getBlockFromString(t, block4Hex).Hash().String(),
			BranchLen: 3,
			Status:    "valid-fork",
		},
		{
			Height:    5,
			Hash:      getBlockFromString(t, block5aHex).Hash().String(),
			BranchLen: 0,
			Status:    "active",
		},
		{
			Height:    4,
			Hash:      getBlockFromString(t, block4bHex).Hash().String(),
			BranchLen: 1,
			Status:    "valid-fork",
		},
	}

	err = compareMultipleChainTips(t, gotChainTips, expectedChainTips)
	if err != nil {
		t.Fatalf("TestGetChainTips fail. Error: %v", err)
	}
}
