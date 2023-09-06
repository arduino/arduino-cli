package utils_test

import (
	"testing"

	"github.com/arduino/arduino-cli/arduino/builder/utils"
	"github.com/stretchr/testify/require"
)

func TestPrintableCommand(t *testing.T) {
	parts := []string{
		"/path/to/dir with spaces/cmd",
		"arg1",
		"arg-\"with\"-quotes",
		"specialchar-`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?-argument",
		"arg   with spaces",
		"arg\twith\t\ttabs",
		"lastarg",
	}
	correct := "\"/path/to/dir with spaces/cmd\"" +
		" arg1 \"arg-\\\"with\\\"-quotes\"" +
		" \"specialchar-`~!@#$%^&*()-_=+[{]}\\\\|;:'\\\",<.>/?-argument\"" +
		" \"arg   with spaces\" \"arg\twith\t\ttabs\"" +
		" lastarg"
	result := utils.PrintableCommand(parts)
	require.Equal(t, correct, result)
}
