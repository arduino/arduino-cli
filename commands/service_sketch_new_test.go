// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/stretchr/testify/require"
)

func Test_SketchNameWrongPattern(t *testing.T) {
	invalidNames := []string{
		"&",
		".hello",
		"-hello",
		"hello*",
		"hello.",
		"||||||||||||||",
		",`hack[}attempt{];",
	}
	for _, name := range invalidNames {
		_, err := NewSketch(context.Background(), &commands.NewSketchRequest{
			SketchName: name,
			SketchDir:  t.TempDir(),
		})

		require.EqualError(t, err, fmt.Sprintf(`Can't create sketch: invalid sketch name "%s": the first character must be alphanumeric or "_", the following ones can also contain "-" and ".". The last one cannot be ".".`,
			name))
	}
}

func Test_SketchNameEmpty(t *testing.T) {
	emptyName := ""
	_, err := NewSketch(context.Background(), &commands.NewSketchRequest{
		SketchName: emptyName,
		SketchDir:  t.TempDir(),
	})

	require.EqualError(t, err, `Can't create sketch: sketch name cannot be empty`)
}

func Test_SketchNameTooLong(t *testing.T) {
	tooLongName := make([]byte, sketchNameMaxLength+1)
	for i := range tooLongName {
		tooLongName[i] = 'a'
	}
	_, err := NewSketch(context.Background(), &commands.NewSketchRequest{
		SketchName: string(tooLongName),
		SketchDir:  t.TempDir(),
	})

	require.EqualError(t, err, fmt.Sprintf(`Can't create sketch: sketch name too long (%d characters). Maximum allowed length is %d`,
		len(tooLongName),
		sketchNameMaxLength))
}

func Test_SketchNameOk(t *testing.T) {
	lengthLimitName := make([]byte, sketchNameMaxLength)
	for i := range lengthLimitName {
		lengthLimitName[i] = 'a'
	}
	validNames := []string{
		"h",
		"h.ello",
		"h..ello-world",
		"hello_world__",
		"_hello_world",
		string(lengthLimitName),
	}
	for _, name := range validNames {
		_, err := NewSketch(context.Background(), &commands.NewSketchRequest{
			SketchName: name,
			SketchDir:  t.TempDir(),
		})
		require.Nil(t, err)
	}
}

func Test_SketchNameReserved(t *testing.T) {
	invalidNames := []string{"CON", "PRN", "AUX", "NUL", "COM0", "COM1", "COM2", "COM3", "COM4", "COM5",
		"COM6", "COM7", "COM8", "COM9", "LPT0", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	for _, name := range invalidNames {
		_, err := NewSketch(context.Background(), &commands.NewSketchRequest{
			SketchName: name,
			SketchDir:  t.TempDir(),
		})
		require.EqualError(t, err, fmt.Sprintf(`Can't create sketch: sketch name cannot be the reserved name "%s"`, name))
	}
}
