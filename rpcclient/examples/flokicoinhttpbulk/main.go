// Copyright (c) 2014-2020 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/flokiorg/go-flokicoin/rpcclient"
)

func main() {
	// Connect to local flokicoin RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:                "localhost:15213",
		User:                "yourrpcuser",
		Pass:                "yourrpcpass",
		DisableConnectOnNew: true,
		HTTPPostMode:        true, // Flokicoin only supports HTTP POST mode
		DisableTLS:          true, // Flokicoin does not provide TLS by default
	}
	batchClient, err := rpcclient.NewBatch(connCfg)
	defer batchClient.Shutdown()

	if err != nil {
		log.Fatal(err)
	}

	// batch mode requires async requests
	blockCount := batchClient.GetBlockCountAsync()
	block1 := batchClient.GetBlockHashAsync(1)
	batchClient.GetBlockHashAsync(2)
	batchClient.GetBlockHashAsync(3)
	block4 := batchClient.GetBlockHashAsync(4)
	difficulty := batchClient.GetDifficultyAsync()

	// sends all queued batch requests
	batchClient.Send()

	fmt.Println(blockCount.Receive())
	fmt.Println(block1.Receive())
	fmt.Println(block4.Receive())
	fmt.Println(difficulty.Receive())
}
