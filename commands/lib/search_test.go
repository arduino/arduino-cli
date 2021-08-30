package lib

import (
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
)

var customIndexPath = paths.New("testdata")

func TestSearchLibrary(t *testing.T) {
	lm := librariesmanager.NewLibraryManager(customIndexPath, nil)
	lm.LoadIndex()

	req := &rpc.LibrarySearchRequest{
		Instance: &rpc.Instance{Id: 1},
		Query:    "test",
	}

	resp := searchLibrary(req, lm)
	assert := assert.New(t)
	assert.Equal(resp.GetStatus(), rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_SUCCESS)
	assert.Equal(len(resp.GetLibraries()), 2)
	assert.True(strings.Contains(resp.GetLibraries()[0].Name, "Test"))
	assert.True(strings.Contains(resp.GetLibraries()[1].Name, "Test"))
}

func TestSearchLibrarySimilar(t *testing.T) {
	lm := librariesmanager.NewLibraryManager(customIndexPath, nil)
	lm.LoadIndex()

	req := &rpc.LibrarySearchRequest{
		Instance: &rpc.Instance{Id: 1},
		Query:    "arduino",
	}

	resp := searchLibrary(req, lm)
	assert := assert.New(t)
	assert.Equal(resp.GetStatus(), rpc.LibrarySearchStatus_LIBRARY_SEARCH_STATUS_SUCCESS)
	assert.Equal(len(resp.GetLibraries()), 2)
	libs := map[string]*rpc.SearchedLibrary{}
	for _, l := range resp.GetLibraries() {
		libs[l.Name] = l
	}
	assert.Contains(libs, "ArduinoTestPackage")
	assert.Contains(libs, "Arduino")
}
