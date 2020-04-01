package lib

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/assert"
)

var customIndexPath = paths.New("testdata")

func TestSearchLibrary(t *testing.T) {
	lm := librariesmanager.NewLibraryManager(customIndexPath, nil)
	lm.LoadIndex()

	req := &rpc.LibrarySearchReq{
		Instance: &rpc.Instance{Id: 1},
		Query:    "test",
	}

	resp, err := searchLibrary(req, lm)
	if err != nil {
		t.Fatal(err)
	}

	assert := assert.New(t)
	assert.Equal(resp.GetStatus(), rpc.LibrarySearchStatus_success)
	assert.Equal(len(resp.GetLibraries()), 2)
	assert.Equal(resp.GetLibraries()[0].Name, "ArduinoTestPackage")
	assert.Equal(resp.GetLibraries()[1].Name, "Test")
}

func TestSearchLibrarySimilar(t *testing.T) {
	lm := librariesmanager.NewLibraryManager(customIndexPath, nil)
	lm.LoadIndex()

	req := &rpc.LibrarySearchReq{
		Instance: &rpc.Instance{Id: 1},
		Query:    "ardino",
	}

	resp, err := searchLibrary(req, lm)
	if err != nil {
		t.Fatal(err)
	}

	assert := assert.New(t)
	assert.Equal(resp.GetStatus(), rpc.LibrarySearchStatus_failed)
	assert.Equal(len(resp.GetLibraries()), 1)
	assert.Equal(resp.GetLibraries()[0].Name, "Arduino")
}
