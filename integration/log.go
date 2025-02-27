//go:build rpctest
// +build rpctest

package integration

import (
	"os"

	flog "github.com/flokiorg/go-flokicoin/log"
	"github.com/flokiorg/go-flokicoin/rpcclient"
)

type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	return len(p), nil
}

func init() {
	backendLog := flog.NewBackend(logWriter{})
	testLog := backendLog.Logger("ITEST")
	testLog.SetLevel(flog.LevelDebug)

	rpcclient.UseLogger(testLog)
}
