package packageindex

import (
	"fmt"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

func TestIndexParsing(t *testing.T) {
	semver.WarnInvalidVersionWhenParsingRelaxed = true

	list, err := paths.New("testdata").ReadDir()
	require.NoError(t, err)
	for _, indexFile := range list {
		if indexFile.Ext() != ".json" {
			continue
		}
		fmt.Println("Loading:", indexFile)
		_, err := LoadIndex(indexFile)
		require.NoError(t, err)
	}
}
