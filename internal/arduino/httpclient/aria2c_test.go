// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/downloader/v3"
)

type progressRecorder struct {
	starts  int
	updates int
	ends    int
	success bool
}

func (r *progressRecorder) callback() rpc.DownloadProgressCB {
	return func(curr *rpc.DownloadProgress) {
		switch {
		case curr.GetStart() != nil:
			r.starts++
		case curr.GetUpdate() != nil:
			r.updates++
		case curr.GetEnd() != nil:
			r.ends++
			r.success = curr.GetEnd().GetSuccess()
		}
	}
}

func TestAria2cOptionsFromContext(t *testing.T) {
	require.Equal(t, Aria2cOptions{}, Aria2cOptionsFromContext(context.Background()))

	opts := Aria2cOptions{Enabled: true, UserAgent: "ua", Proxy: "http://proxy:8080"}
	ctx := ContextWithAria2cOptions(context.Background(), opts)
	require.Equal(t, opts, Aria2cOptionsFromContext(ctx))
}

func TestDownloadFileWithoutAria2cOptIn(t *testing.T) {
	const payload = "built-in downloader payload"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	dest := paths.New(t.TempDir()).Join("file.txt")
	rec := &progressRecorder{}
	err := DownloadFile(context.Background(), dest, srv.URL, "", "label", rec.callback(), downloader.GetDefaultConfig())
	require.NoError(t, err)

	content, err := os.ReadFile(dest.String())
	require.NoError(t, err)
	require.Equal(t, payload, string(content))
	require.Equal(t, 1, rec.starts)
	require.Equal(t, 1, rec.ends)
	require.True(t, rec.success)
}

func TestDownloadFileAria2cEnabledButNotInstalled(t *testing.T) {
	const payload = "fallback payload"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	// Empty PATH guarantees aria2c cannot be found.
	t.Setenv("PATH", t.TempDir())

	ctx := ContextWithAria2cOptions(context.Background(), Aria2cOptions{Enabled: true})
	dest := paths.New(t.TempDir()).Join("file.txt")
	rec := &progressRecorder{}
	err := DownloadFile(ctx, dest, srv.URL, "", "label", rec.callback(), downloader.GetDefaultConfig())
	require.NoError(t, err)

	content, err := os.ReadFile(dest.String())
	require.NoError(t, err)
	require.Equal(t, payload, string(content))
	require.Equal(t, 1, rec.starts)
	require.Equal(t, 1, rec.ends)
	require.True(t, rec.success)
}

// writeFakeAria2c creates an executable fake aria2c that writes marker bytes
// to the --dir/--out destination and records its arguments.
func writeFakeAria2c(t *testing.T, dir, body string) {
	t.Helper()
	script := "#!/bin/sh\n" + body + "\n"
	scriptPath := filepath.Join(dir, "aria2c")
	require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0o755))
}

const fakeAria2cSuccessBody = `
dir=""
out=""
for arg in "$@"; do
  case "$arg" in
    --dir=*) dir="${arg#--dir=}" ;;
    --out=*) out="${arg#--out=}" ;;
  esac
done
echo "$@" > "$FAKE_ARIA2C_ARGS_FILE"
printf 'downloaded-by-fake-aria2c' > "$dir/$out"
`

func TestDownloadFileWithAria2c(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake aria2c shim requires sh")
	}
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte("built-in downloader payload"))
	}))
	defer srv.Close()

	binDir := t.TempDir()
	writeFakeAria2c(t, binDir, fakeAria2cSuccessBody)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	argsFile := filepath.Join(t.TempDir(), "args.txt")
	t.Setenv("FAKE_ARIA2C_ARGS_FILE", argsFile)

	ctx := ContextWithAria2cOptions(context.Background(), Aria2cOptions{
		Enabled:   true,
		UserAgent: "test-user-agent",
		Proxy:     "http://proxy:8080",
	})
	dest := paths.New(t.TempDir()).Join("file.txt")
	rec := &progressRecorder{}
	err := DownloadFile(ctx, dest, srv.URL, "", "label", rec.callback(), downloader.GetDefaultConfig())
	require.NoError(t, err)

	// The marker content proves the download went through (fake) aria2c.
	content, err := os.ReadFile(dest.String())
	require.NoError(t, err)
	require.Equal(t, "downloaded-by-fake-aria2c", string(content))
	require.Zero(t, hits.Load(), "built-in downloader must not be used when aria2c succeeds")
	require.Equal(t, 1, rec.starts)
	require.Equal(t, 1, rec.ends)
	require.True(t, rec.success)

	args, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	require.Contains(t, string(args), "--user-agent=test-user-agent")
	require.Contains(t, string(args), "--all-proxy=http://proxy:8080")
	require.Contains(t, string(args), "--max-connection-per-server=16")
}

func TestDownloadFileAria2cFailureFallsBack(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake aria2c shim requires sh")
	}
	const payload = "fallback payload"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer srv.Close()

	binDir := t.TempDir()
	writeFakeAria2c(t, binDir, "exit 1")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx := ContextWithAria2cOptions(context.Background(), Aria2cOptions{Enabled: true})
	dest := paths.New(t.TempDir()).Join("file.txt")
	rec := &progressRecorder{}
	err := DownloadFile(ctx, dest, srv.URL, "", "label", rec.callback(), downloader.GetDefaultConfig())
	require.NoError(t, err)

	content, err := os.ReadFile(dest.String())
	require.NoError(t, err)
	require.Equal(t, payload, string(content))
	require.Equal(t, 1, rec.starts)
	require.Equal(t, 1, rec.ends)
	require.True(t, rec.success)
	require.NoFileExists(t, dest.String()+".aria2")
}
