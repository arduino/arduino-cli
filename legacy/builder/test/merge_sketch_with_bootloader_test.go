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

package test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestMergeSketchWithBootloader(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")
	defer cleanUpBuilderTestContext(t, ctx)

	buildPath := ctx.Builder.GetBuildPath()
	err := buildPath.Join("sketch").MkdirAll()
	require.NoError(t, err)

	fakeSketchHex := `:100000000C9434000C9446000C9446000C9446006A
:100010000C9446000C9446000C9446000C94460048
:100020000C9446000C9446000C9446000C94460038
:100030000C9446000C9446000C9446000C94460028
:100040000C9448000C9446000C9446000C94460016
:100050000C9446000C9446000C9446000C94460008
:100060000C9446000C94460011241FBECFEFD8E03C
:10007000DEBFCDBF21E0A0E0B1E001C01D92A930FC
:10008000B207E1F70E9492000C94DC000C9400008F
:100090001F920F920FB60F9211242F933F938F93BD
:1000A0009F93AF93BF938091050190910601A0911A
:1000B0000701B09108013091040123E0230F2D378F
:1000C00020F40196A11DB11D05C026E8230F02965C
:1000D000A11DB11D20930401809305019093060199
:1000E000A0930701B0930801809100019091010154
:1000F000A0910201B09103010196A11DB11D809351
:10010000000190930101A0930201B0930301BF91FC
:10011000AF919F918F913F912F910F900FBE0F90B4
:100120001F901895789484B5826084BD84B58160F1
:1001300084BD85B5826085BD85B5816085BD8091B2
:100140006E00816080936E0010928100809181002A
:100150008260809381008091810081608093810022
:10016000809180008160809380008091B1008460E4
:100170008093B1008091B00081608093B000809145
:100180007A00846080937A0080917A008260809304
:100190007A0080917A00816080937A0080917A0061
:1001A000806880937A001092C100C0E0D0E0209770
:0C01B000F1F30E940000FBCFF894FFCF99
:00000001FF
`
	err = buildPath.Join("sketch", "sketch1.ino.hex").WriteFile([]byte(fakeSketchHex))
	require.NoError(t, err)

	err = ctx.Builder.MergeSketchWithBootloader()
	require.NoError(t, err)

	bytes, err := buildPath.Join("sketch", "sketch1.ino.with_bootloader.hex").ReadFile()
	require.NoError(t, err)
	mergedSketchHex := string(bytes)

	require.Contains(t, mergedSketchHex, ":100000000C9434000C9446000C9446000C9446006A\n")
	require.True(t, strings.HasSuffix(mergedSketchHex, ":00000001FF\n"))
}

func TestMergeSketchWithBootloaderSketchInBuildPath(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")
	defer cleanUpBuilderTestContext(t, ctx)

	buildPath := ctx.Builder.GetBuildPath()
	err := buildPath.Join("sketch").MkdirAll()
	require.NoError(t, err)

	fakeSketchHex := `:100000000C9434000C9446000C9446000C9446006A
:100010000C9446000C9446000C9446000C94460048
:100020000C9446000C9446000C9446000C94460038
:100030000C9446000C9446000C9446000C94460028
:100040000C9448000C9446000C9446000C94460016
:100050000C9446000C9446000C9446000C94460008
:100060000C9446000C94460011241FBECFEFD8E03C
:10007000DEBFCDBF21E0A0E0B1E001C01D92A930FC
:10008000B207E1F70E9492000C94DC000C9400008F
:100090001F920F920FB60F9211242F933F938F93BD
:1000A0009F93AF93BF938091050190910601A0911A
:1000B0000701B09108013091040123E0230F2D378F
:1000C00020F40196A11DB11D05C026E8230F02965C
:1000D000A11DB11D20930401809305019093060199
:1000E000A0930701B0930801809100019091010154
:1000F000A0910201B09103010196A11DB11D809351
:10010000000190930101A0930201B0930301BF91FC
:10011000AF919F918F913F912F910F900FBE0F90B4
:100120001F901895789484B5826084BD84B58160F1
:1001300084BD85B5826085BD85B5816085BD8091B2
:100140006E00816080936E0010928100809181002A
:100150008260809381008091810081608093810022
:10016000809180008160809380008091B1008460E4
:100170008093B1008091B00081608093B000809145
:100180007A00846080937A0080917A008260809304
:100190007A0080917A00816080937A0080917A0061
:1001A000806880937A001092C100C0E0D0E0209770
:0C01B000F1F30E940000FBCFF894FFCF99
:00000001FF
`
	err = buildPath.Join("sketch1.ino.hex").WriteFile([]byte(fakeSketchHex))
	require.NoError(t, err)

	err = ctx.Builder.MergeSketchWithBootloader()
	require.NoError(t, err)

	bytes, err := buildPath.Join("sketch1.ino.with_bootloader.hex").ReadFile()
	require.NoError(t, err)
	mergedSketchHex := string(bytes)

	fmt.Println(string(mergedSketchHex))
	require.Contains(t, mergedSketchHex, ":100000000C9434000C9446000C9446000C9446006A\n")
	require.True(t, strings.HasSuffix(mergedSketchHex, ":00000001FF\n"))
}

func TestMergeSketchWithBootloaderWhenNoBootloaderAvailable(t *testing.T) {
	ctx := prepareBuilderTestContext(t, nil, paths.New("sketch1", "sketch1.ino"), "arduino:avr:uno")
	defer cleanUpBuilderTestContext(t, ctx)

	buildPath := ctx.Builder.GetBuildPath()
	buildProperties := ctx.Builder.GetBuildProperties()
	buildProperties.Remove("bootloader.noblink")
	buildProperties.Remove("bootloader.file")

	err := ctx.Builder.MergeSketchWithBootloader()
	require.NoError(t, err)

	exist, err := buildPath.Join("sketch.ino.with_bootloader.hex").ExistCheck()
	require.NoError(t, err)
	require.False(t, exist)
}

// TODO convert in a compile test and we check against the real .hex
func TestMergeSketchWithBootloaderPathIsParameterized(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs:         paths.NewPathList(filepath.Join("..", "hardware"), "downloaded_hardware", "user_hardware"),
		BuiltInToolsDirs:     paths.NewPathList("downloaded_tools"),
		BuiltInLibrariesDirs: paths.New("downloaded_libraries"),
		OtherLibrariesDirs:   paths.NewPathList("libraries"),
	}
	ctx = prepareBuilderTestContext(t, ctx, paths.New("sketch1", "sketch1.ino"), "my_avr_platform:avr:mymega:cpu=atmega2560")
	defer cleanUpBuilderTestContext(t, ctx)

	buildPath := ctx.Builder.GetBuildPath()
	err := buildPath.Join("sketch").MkdirAll()
	require.NoError(t, err)

	fakeSketchHex := `:100000000C9434000C9446000C9446000C9446006A
:100010000C9446000C9446000C9446000C94460048
:100020000C9446000C9446000C9446000C94460038
:100030000C9446000C9446000C9446000C94460028
:100040000C9448000C9446000C9446000C94460016
:100050000C9446000C9446000C9446000C94460008
:100060000C9446000C94460011241FBECFEFD8E03C
:10007000DEBFCDBF21E0A0E0B1E001C01D92A930FC
:10008000B207E1F70E9492000C94DC000C9400008F
:100090001F920F920FB60F9211242F933F938F93BD
:1000A0009F93AF93BF938091050190910601A0911A
:1000B0000701B09108013091040123E0230F2D378F
:1000C00020F40196A11DB11D05C026E8230F02965C
:1000D000A11DB11D20930401809305019093060199
:1000E000A0930701B0930801809100019091010154
:1000F000A0910201B09103010196A11DB11D809351
:10010000000190930101A0930201B0930301BF91FC
:10011000AF919F918F913F912F910F900FBE0F90B4
:100120001F901895789484B5826084BD84B58160F1
:1001300084BD85B5826085BD85B5816085BD8091B2
:100140006E00816080936E0010928100809181002A
:100150008260809381008091810081608093810022
:10016000809180008160809380008091B1008460E4
:100170008093B1008091B00081608093B000809145
:100180007A00846080937A0080917A008260809304
:100190007A0080917A00816080937A0080917A0061
:1001A000806880937A001092C100C0E0D0E0209770
:0C01B000F1F30E940000FBCFF894FFCF99
:00000001FF
`
	err = buildPath.Join("sketch", "sketch1.ino.hex").WriteFile([]byte(fakeSketchHex))
	require.NoError(t, err)

	err = ctx.Builder.MergeSketchWithBootloader()
	require.NoError(t, err)

	bytes, err := buildPath.Join("sketch", "sketch1.ino.with_bootloader.hex").ReadFile()
	require.NoError(t, err)
	mergedSketchHex := string(bytes)

	require.Contains(t, mergedSketchHex, ":100000000C9434000C9446000C9446000C9446006A\n")
	require.True(t, strings.HasSuffix(mergedSketchHex, ":00000001FF\n"))
}
