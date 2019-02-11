package configs_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/configs"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestNavigate(t *testing.T) {
	tests := []string{
		"noconfig",
		"local",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			root := filepath.Join("testdata", "navigate", tt)
			pwd := filepath.Join("testdata", "navigate", tt, "first", "second")
			golden := filepath.Join("testdata", "navigate", tt, "golden.yaml")

			got := configs.Navigate(root, pwd)
			data, _ := got.SerializeToYAML()

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
