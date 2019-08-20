/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package configs_test

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/configs"
	paths "github.com/arduino/go-paths-helper"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestNavigate(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Test only runs on Linux")
	}

	tests := []string{
		"noconfig",
		"local",
		"inheritance",
	}
	for _, tt := range tests {
		_tt := tt
		t.Run(_tt, func(t *testing.T) {
			pwd := paths.New("testdata", "navigate", _tt, "first", "second")
			golden := filepath.Join("testdata", "navigate", _tt, "golden.yaml")

			config, _ := configs.NewConfiguration()

			config.Navigate(pwd)
			data, _ := config.SerializeToYAML()

			diff(t, data, golden)
		})
	}
}

func diff(t *testing.T, data []byte, goldenFile string) {
	golden, err := ioutil.ReadFile(goldenFile)
	if err != nil {
		t.Error(err)
		return
	}

	dataStr := strings.TrimSpace(string(data))
	goldenStr := strings.TrimSpace(string(golden))

	// Substitute home folder
	homedir, _ := homedir.Dir()
	dataStr = strings.Replace(dataStr, homedir, "$HOME", -1)

	if dataStr != goldenStr {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(goldenStr, dataStr, false)
		t.Errorf(dmp.DiffPrettyText(diffs))
	}
}
