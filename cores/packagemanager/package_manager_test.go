package packagemanager_test

import (
	"testing"

	"github.com/bcmi-labs/arduino-cli/cores/packagemanager"
	"github.com/stretchr/testify/require"
)

func TestFindBoardWithFQBN(t *testing.T) {
	pm := packagemanager.NewPackageManager()
	pm.LoadHardwareFromDirectory("testdata")

	board, err := pm.FindBoardWithFQBN("arduino:avr:uno")
	require.Nil(t, err)
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino/Genuino Uno")

	board, err = pm.FindBoardWithFQBN("arduino:avr:mega")
	require.Nil(t, err)
	require.NotNil(t, board)
	require.Equal(t, board.Name(), "Arduino/Genuino Mega or Mega 2560")
}
