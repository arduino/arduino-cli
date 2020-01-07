// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package builder_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
)

func tmpDirOrDie() string {
	dir, err := ioutil.TempDir(os.TempDir(), "builder_test")
	if err != nil {
		panic(fmt.Sprintf("error creating tmp dir: %v", err))
	}
	return dir
}

func TestGenBuildPath(t *testing.T) {
	want := filepath.Join(os.TempDir(), "arduino-sketch-ACBD18DB4CC2F85CEDEF654FCCC4A4D8")
	assert.Equal(t, want, builder.GenBuildPath(paths.New("foo")).String())

	want = filepath.Join(os.TempDir(), "arduino-sketch-D41D8CD98F00B204E9800998ECF8427E")
	assert.Equal(t, want, builder.GenBuildPath(nil).String())
}

func TestEnsureBuildPathExists(t *testing.T) {
	tmp := tmpDirOrDie()
	defer os.RemoveAll(tmp)
	bp := filepath.Join(tmp, "build_path")

	assert.Nil(t, builder.EnsureBuildPathExists(bp))
	_, err := os.Stat(bp)
	assert.Nil(t, err)

	// run again over an existing folder
	assert.Nil(t, builder.EnsureBuildPathExists(bp))
	_, err = os.Stat(bp)
	assert.Nil(t, err)
}
