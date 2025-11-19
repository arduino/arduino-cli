// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package cpp

import (
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestDepFileReader(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		deps, err := ReadDepFile(paths.New("testdata", "depcheck.0.d"))
		require.NoError(t, err)
		require.NotNil(t, deps)
		require.Len(t, deps.Dependencies, 302)
		require.Equal(t, "sketch.ino.cpp.o", deps.ObjectFile)
		require.Equal(t, "/home/megabug/Arduino/sketch/build/sketch/sketch.ino.cpp.merged", deps.Dependencies[0])
		require.Equal(t, "/home/megabug/.arduino15/packages/arduino/hardware/zephyr/0.10.0-rc.10/variants/b_u585i_iot02a_stm32u585xx/llext-edk/include/zephyr/include/generated/zephyr/autoconf.h", deps.Dependencies[1])
		require.Equal(t, "/home/megabug/.arduino15/packages/arduino/hardware/zephyr/0.10.0-rc.10/variants/b_u585i_iot02a_stm32u585xx/llext-edk/include/zephyr/include/zephyr/toolchain/zephyr_stdint.h", deps.Dependencies[2])
		require.Equal(t, "/home/megabug/.arduino15/packages/arduino/hardware/zephyr/0.10.0-rc.10/libraries/Arduino_RPCLite/src/dispatcher.h", deps.Dependencies[301])
	})
	t.Run("1", func(t *testing.T) {
		deps, err := ReadDepFile(paths.New("testdata", "depcheck.1.d"))
		require.NoError(t, err)
		require.NotNil(t, deps)
		require.Equal(t, "sketch.ino.o", deps.ObjectFile)
		require.Len(t, deps.Dependencies, 302)
		require.Equal(t, "/home/megabug/Arduino/sketch/build/sketch/sketch.ino.cpp", deps.Dependencies[0])
		require.Equal(t, "/home/megabug/.arduino15/packages/arduino/hardware/zephyr/0.10.0-rc.10/variants/b_u585i_iot02a_stm32u585xx/llext-edk/include/zephyr/include/generated/zephyr/autoconf.h", deps.Dependencies[1])
		require.Equal(t, "/home/megabug/.arduino15/packages/arduino/hardware/zephyr/0.10.0-rc.10/variants/b_u585i_iot02a_stm32u585xx/llext-edk/include/zephyr/include/zephyr/toolchain/zephyr_stdint.h", deps.Dependencies[2])
		require.Equal(t, "/home/megabug/.arduino15/packages/arduino/hardware/zephyr/0.10.0-rc.10/libraries/Arduino_RPCLite/src/dispatcher.h", deps.Dependencies[301])
	})
	t.Run("2", func(t *testing.T) {
		deps, err := ReadDepFile(paths.New("testdata", "depcheck.2.d"))
		require.NoError(t, err)
		require.NotNil(t, deps)
		require.Equal(t, "ske tch.ino.cpp.o", deps.ObjectFile)
		require.Len(t, deps.Dependencies, 302)
		require.Equal(t, "/home/megabug/Arduino/ske tch/build/sketch/ske tch.ino.cpp.merged", deps.Dependencies[0])
		require.Equal(t, "/home/megabug/.arduino15/packages/arduino/hardware/zephyr/0.10.0-rc.10/variants/b_u585i_iot02a_stm32u585xx/llext-edk/include/zephyr/include/generated/zephyr/autoconf.h", deps.Dependencies[1])
		require.Equal(t, "/home/megabug/.arduino15/packages/arduino/hardware/zephyr/0.10.0-rc.10/variants/b_u585i_iot02a_stm32u585xx/llext-edk/include/zephyr/include/zephyr/toolchain/zephyr_stdint.h", deps.Dependencies[2])
		require.Equal(t, "/home/megabug/.arduino15/packages/arduino/hardware/zephyr/0.10.0-rc.10/libraries/Arduino_RPCLite/src/dispatcher.h", deps.Dependencies[301])
	})
	t.Run("3", func(t *testing.T) {
		deps, err := ReadDepFile(paths.New("testdata", "depcheck.3.d"))
		require.NoError(t, err)
		require.NotNil(t, deps)
		require.Equal(t, "myfile.o", deps.ObjectFile)
		require.Len(t, deps.Dependencies, 3)
		require.Equal(t, "/some/path\\twith/backslash and spaces/file.cpp", deps.Dependencies[0])
		require.Equal(t, "/some/other$/path#/file.h", deps.Dependencies[1])
		require.Equal(t, "/yet/ano\\ther/path/file.h", deps.Dependencies[2])
	})
	t.Run("4", func(t *testing.T) {
		deps, err := ReadDepFile(paths.New("testdata", "depcheck.4.d"))
		require.EqualError(t, err, "invalid dollar sequence: $a")
		require.Nil(t, deps)
	})
	t.Run("6", func(t *testing.T) {
		deps, err := ReadDepFile(paths.New("testdata", "depcheck.6.d"))
		require.EqualError(t, err, "unclosed escape sequence at end of depfile")
		require.Nil(t, deps)
	})
	t.Run("7", func(t *testing.T) {
		deps, err := ReadDepFile(paths.New("testdata", "depcheck.7.d"))
		require.EqualError(t, err, "no colon in first item of depfile")
		require.Nil(t, deps)
	})
	t.Run("8", func(t *testing.T) {
		deps, err := ReadDepFile(paths.New("testdata", "depcheck.8.d"))
		require.NoError(t, err)
		require.Nil(t, deps.Dependencies)
		require.Empty(t, deps.ObjectFile)
	})
	t.Run("9", func(t *testing.T) {
		deps, err := ReadDepFile(paths.New("testdata", "nonexistent.d"))
		require.Error(t, err)
		require.Nil(t, deps)
	})
}
