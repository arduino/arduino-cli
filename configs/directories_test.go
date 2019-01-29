package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultDirs(t *testing.T) {
	// This should fail if cross-compiled for linux/amd64 or compiled natively without CGO enable
	// See: https://github.com/arduino/arduino-cli/issues/133
	os.Setenv("USER", "")
	_, err := getDefaultArduinoDataDir()
	require.NoError(t, err)
}
