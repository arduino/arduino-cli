// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package builder

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/builder/logger"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

func TestSizerWithAVRData(t *testing.T) {
	output := []byte(`/tmp/test597119152/sketch.ino.elf  :
section           size      addr
.data               36   8388864
.text             3966         0
.bss               112   8388900
.comment            17         0
.debug_aranges     704         0
.debug_info      21084         0
.debug_abbrev     4704         0
.debug_line       5456         0
.debug_frame      1964         0
.debug_str        8251         0
.debug_loc        7747         0
.debug_ranges      856         0
Total            54897
`)

	size, err := computeSize(`^(?:\.text|\.data|\.bootloader)\s+([0-9]+).*`, output)
	require.NoError(t, err)
	require.Equal(t, 4002, size)

	size, err = computeSize(`^(?:\.data|\.bss|\.noinit)\s+([0-9]+).*`, output)
	require.NoError(t, err)
	require.Equal(t, 148, size)

	size, err = computeSize(`^(?:\.eeprom)\s+([0-9]+).*`, output)
	require.NoError(t, err)
	require.Equal(t, 0, size)
}

func TestSizerWithSAMDData(t *testing.T) {
	output := []byte(`/tmp/test737785204/sketch_usbhost.ino.elf  :
section             size        addr
.text               8076        8192
.data                152   536870912
.bss                1984   536871064
.ARM.attributes       40           0
.comment             128           0
.debug_info       178841           0
.debug_abbrev      14968           0
.debug_aranges      2080           0
.debug_ranges       3840           0
.debug_line        26068           0
.debug_str         55489           0
.debug_frame        5036           0
.debug_loc         20978           0
Total             317680
`)

	size, err := computeSize(`\.text\s+([0-9]+).*`, output)
	require.NoError(t, err)
	require.Equal(t, 8076, size)
}

func TestSizerEmptyRegexpReturnsMinusOne(t *testing.T) {
	size, err := computeSize(``, []byte(`xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`))
	require.NoError(t, err)
	require.Equal(t, -1, size)
}

func TestSizerWithInvalidRegexp(t *testing.T) {
	_, err := computeSize(`[xx`, []byte(`xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`))
	require.Error(t, err)
}

func TestCheckSizeMemoryLimitFormats(t *testing.T) {
	executable, err := os.Executable()
	require.NoError(t, err)

	for _, tc := range []struct {
		name, flash, ram string
	}{
		{"decimal", "32", "8"},
		{"decimal with leading zeros", "032", "008"},
		{"hexadecimal flash", "0x20", "8"},
		{"hexadecimal RAM", "32", "0X8"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			props := properties.NewMap()
			props.Set("upload.maximum_size", tc.flash)
			props.Set("upload.maximum_data_size", tc.ram)
			props.Set("recipe.size.pattern", fmt.Sprintf(`"%s" -test.run=^TestSizerRecipeOutput$`, executable))
			props.Set("recipe.size.regex", `^\.text\s+([0-9]+)`)
			props.Set("recipe.size.regex.data", `^\.data\s+([0-9]+)`)
			b := Builder{
				buildProperties: props,
				logger:          logger.New(io.Discard, io.Discard, logger.VerbosityQuiet, ""),
				toolEnv:         []string{"ARDUINO_CLI_TEST_SIZER_RECIPE=1"},
			}

			sizes, err := b.checkSize()
			require.NoError(t, err)
			require.Equal(t, ExecutablesFileSections{
				{Name: "text", Size: 16, MaxSize: 32},
				{Name: "data", Size: 4, MaxSize: 8},
			}, sizes)
		})
	}
}

// TestSizerRecipeOutput acts as an offline size tool when invoked by the recipe.
func TestSizerRecipeOutput(t *testing.T) {
	if os.Getenv("ARDUINO_CLI_TEST_SIZER_RECIPE") == "1" {
		fmt.Print(".text 16\n.data 4\n")
	}
}
