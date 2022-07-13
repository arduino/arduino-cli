// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package integrationtest

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v3"
	"github.com/stretchr/testify/require"
)

// Environment is a test environment for the test suite.
type Environment struct {
	rootDir      *paths.Path
	downloadsDir *paths.Path
	t            *require.Assertions
}

// SharedDownloadDir returns the shared downloads directory.
func SharedDownloadDir(t *testing.T) *paths.Path {
	downloadsDir := paths.TempDir().Join("arduino-cli-test-suite-staging")
	require.NoError(t, downloadsDir.MkdirAll())
	return downloadsDir
}

// NewEnvironment creates a new test environment.
func NewEnvironment(t *testing.T) *Environment {
	downloadsDir := SharedDownloadDir(t)
	rootDir, err := paths.MkTempDir("", "arduino-cli-test-suite")
	require.NoError(t, err)
	return &Environment{
		rootDir:      rootDir,
		downloadsDir: downloadsDir,
		t:            require.New(t),
	}
}

// CleanUp removes the test environment.
func (e *Environment) CleanUp() {
	e.t.NoError(e.rootDir.RemoveAll())
}

// Root returns the root dir of the environment.
func (e *Environment) Root() *paths.Path {
	return e.rootDir
}

// Download downloads a file from a URL and returns the path to the downloaded file.
// The file is saved and cached in a shared downloads directory. If the file already exists, it is not downloaded again.
func (e *Environment) Download(rawURL string) *paths.Path {
	url, err := url.Parse(rawURL)
	e.t.NoError(err)

	filename := filepath.Base(url.Path)
	if filename == "/" {
		filename = ""
	} else {
		filename = "-" + filename
	}

	hash := md5.Sum([]byte(rawURL))
	resource := e.downloadsDir.Join(hex.EncodeToString(hash[:]) + filename)

	// If the resource already exist, return it
	if resource.Exist() {
		return resource
	}

	// Download file
	resp, err := http.Get(rawURL)
	e.t.NoError(err)
	defer resp.Body.Close()

	// Copy data in a temp file
	tmp := resource.Parent().Join(resource.Base() + ".tmp")
	out, err := tmp.Create()
	e.t.NoError(err)
	defer tmp.Remove()
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	e.t.NoError(err)
	e.t.NoError(out.Close())

	// Rename the file to its final destination
	e.t.NoError(tmp.Rename(resource))

	return resource
}

// Extract extracts a tarball to a directory named as the archive
// with the "_content" suffix added. Returns the path to the directory.
func (e *Environment) Extract(archive *paths.Path) *paths.Path {
	destDir := archive.Parent().Join(archive.Base() + "_content")
	if destDir.Exist() {
		return destDir
	}

	file, err := archive.Open()
	e.t.NoError(err)
	defer file.Close()

	err = extract.Archive(context.Background(), file, destDir.String(), nil)
	e.t.NoError(err)

	return destDir
}
