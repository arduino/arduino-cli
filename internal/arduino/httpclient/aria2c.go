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
	"os/exec"
	"time"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"go.bug.st/downloader/v3"
)

// Aria2cOptions carries the user's preference for the optional external
// aria2c downloader through the download call chain.
type Aria2cOptions struct {
	// Enabled requests aria2c usage when the binary is available.
	Enabled bool
	// UserAgent, when not empty, is forwarded to aria2c.
	UserAgent string
	// Proxy, when not empty, is forwarded to aria2c as --all-proxy.
	Proxy string
}

type aria2cContextKey struct{}

// ContextWithAria2cOptions returns a context carrying the given aria2c options.
func ContextWithAria2cOptions(ctx context.Context, opts Aria2cOptions) context.Context {
	return context.WithValue(ctx, aria2cContextKey{}, opts)
}

// Aria2cOptionsFromContext returns the aria2c options carried by the context,
// or the zero value when the context carries none.
func Aria2cOptionsFromContext(ctx context.Context) Aria2cOptions {
	if opts, ok := ctx.Value(aria2cContextKey{}).(Aria2cOptions); ok {
		return opts
	}
	return Aria2cOptions{}
}

// aria2cLookPath is exec.LookPath, kept as a variable to allow test overrides.
var aria2cLookPath = exec.LookPath

// downloadWithAria2cIfEnabled downloads URL into path using the external
// aria2c downloader when the context requests it and the binary is installed.
// It reports whether the download completed with aria2c; any failure (missing
// binary, spawn error, non-zero exit) is logged and reported as false so the
// caller can fall back to the built-in downloader. config is used to forward
// the User-Agent header when the context carries none.
func downloadWithAria2cIfEnabled(ctx context.Context, path *paths.Path, URL string, downloadCB rpc.DownloadProgressCB, config downloader.Config) bool {
	opts := Aria2cOptionsFromContext(ctx)
	if !opts.Enabled {
		return false
	}
	bin, err := aria2cLookPath("aria2c")
	if err != nil {
		logrus.Info("aria2c downloader requested but aria2c not found in PATH, using built-in downloader")
		return false
	}
	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = config.ExtraHeaders["User-Agent"]
	}
	if err := runAria2c(ctx, bin, path, URL, userAgent, opts.Proxy, downloadCB); err != nil {
		logrus.WithError(err).Info("aria2c download failed, falling back to built-in downloader")
		return false
	}
	return true
}

// runAria2c runs aria2c to download URL into path with multi-connection
// support, periodically reporting progress through downloadCB.
func runAria2c(ctx context.Context, bin string, path *paths.Path, URL, userAgent, proxy string, downloadCB rpc.DownloadProgressCB) error {
	if err := path.Parent().MkdirAll(); err != nil {
		return err
	}
	// Start from a clean slate: a stale partial file (possibly written by the
	// built-in downloader) cannot be resumed by aria2c.
	_ = path.Remove()
	_ = paths.New(path.String() + ".aria2").Remove()

	args := []string{
		"--console-log-level=warn",
		"--allow-overwrite=true",
		"--auto-file-renaming=false",
		"--max-connection-per-server=16",
		"--split=16",
		"--min-split-size=1M",
		"--file-allocation=none",
	}
	if userAgent != "" {
		args = append(args, "--user-agent="+userAgent)
	}
	if proxy != "" {
		args = append(args, "--all-proxy="+proxy)
	}
	args = append(args, "--dir="+path.Parent().String(), "--out="+path.Base(), URL)

	logrus.WithField("url", URL).Info("Starting download with aria2c")
	cmd := exec.CommandContext(ctx, bin, args...)

	// Report indeterminate progress by polling the size of the partial file,
	// mirroring the -1 total size used by the built-in downloader when the
	// remote size is unknown.
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if stat, err := path.Stat(); err == nil {
					downloadCB.Update(stat.Size(), -1)
				}
			}
		}
	}()

	if out, err := cmd.CombinedOutput(); err != nil {
		logrus.WithField("output", string(out)).Debug("aria2c output")
		_ = path.Remove()
		_ = paths.New(path.String() + ".aria2").Remove()
		return err
	}
	_ = paths.New(path.String() + ".aria2").Remove()
	if stat, err := path.Stat(); err == nil {
		downloadCB.Update(stat.Size(), stat.Size())
	}
	return nil
}
