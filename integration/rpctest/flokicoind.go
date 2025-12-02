// Copyright (c) 2017 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpctest

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	// compileMtx guards access to the executable path so that the project is
	// only compiled once.
	compileMtx sync.Mutex

	// executablePath is the path to the compiled executable. This is the empty
	// string until lokid is compiled. This should not be accessed directly;
	// instead use the function lokidExecutablePath().
	executablePath string
)

// lokidExecutablePath returns a path to the lokid executable to be used by
// rpctests. To ensure the code tests against the most up-to-date version of
// lokid, this method compiles lokid the first time it is called. After that, the
// generated binary is used for subsequent test harnesses. The executable file
// is not cleaned up, but since it lives at a static path in a temp directory,
// it is not a big deal.
func lokidExecutablePath() (string, error) {
	compileMtx.Lock()
	defer compileMtx.Unlock()

	// If lokid has already been compiled, just use that.
	if len(executablePath) != 0 {
		return executablePath, nil
	}

	testDir, err := baseDir()
	if err != nil {
		return "", err
	}

	// Build lokid and output an executable in a static temp path.
	outputPath := filepath.Join(testDir, "lokid")
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	cmd := exec.Command(
		"go", "build", "-o", outputPath, "github.com/flokiorg/go-flokicoin",
	)
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Failed to build lokid: %v", err)
	}

	// Save executable path so future calls do not recompile.
	executablePath = outputPath
	return executablePath, nil
}
