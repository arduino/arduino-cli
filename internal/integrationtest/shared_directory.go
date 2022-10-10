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
	"sync"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/rogpeppe/go-internal/lockedfile"
	"github.com/stretchr/testify/require"
)

// SharedDir is a directory that is shared between multiple tests.
type SharedDir struct {
	dir      *paths.Path
	t        *testing.T
	mux      sync.Mutex
	fileLock *lockedfile.File
}

// Lock locks the shared directory for exclusive access and return the path to the directory.
func (d *SharedDir) Lock() *paths.Path {
	d.mux.Lock()
	defer d.mux.Unlock()
	if d.fileLock != nil {
		panic("SharedDir already locked")
	}
	fileLock, err := lockedfile.Create(d.dir.Join(".lock").String())
	require.NoError(d.t, err)
	d.fileLock = fileLock
	return d.dir
}

// Unlock unlocks the shared directory.
func (d *SharedDir) Unlock() {
	d.mux.Lock()
	defer d.mux.Unlock()
	if d.fileLock == nil {
		panic("SharedDir already unlocked")
	}
	require.NoError(d.t, d.fileLock.Close())
	d.fileLock = nil
}

// NewSharedDir creates a new shared directory.
func NewSharedDir(t *testing.T, id string) *SharedDir {
	dir := paths.TempDir().Join(ProjectName + "-" + id)
	require.NoError(t, dir.MkdirAll())
	return &SharedDir{
		dir: dir,
		t:   t,
	}
}
