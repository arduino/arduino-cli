package librariesmanager

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func Test_RescanLibrariesCallClear(t *testing.T) {
	baseDir := paths.New(t.TempDir())
	lm := NewLibraryManager(baseDir.Join("index_dir"), baseDir.Join("downloads_dir"))
	lm.Libraries["testLibA"] = libraries.List{}
	lm.Libraries["testLibB"] = libraries.List{}
	lm.RescanLibraries()
	require.Len(t, lm.Libraries, 0)
}
