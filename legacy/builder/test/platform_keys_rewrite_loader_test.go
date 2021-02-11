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
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestLoadPlatformKeysRewrite(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("downloaded_hardware", filepath.Join("..", "hardware")),
	}

	commands := []types.Command{
		&builder.PlatformKeysRewriteLoader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	platformKeysRewrite := ctx.PlatformKeyRewrites

	require.Equal(t, 13, len(platformKeysRewrite.Rewrites))
	require.Equal(t, "compiler.path", platformKeysRewrite.Rewrites[0].Key)
	require.Equal(t, "{runtime.ide.path}/hardware/tools/avr/bin/", platformKeysRewrite.Rewrites[0].OldValue)
	require.Equal(t, "{runtime.tools.avr-gcc.path}/bin/", platformKeysRewrite.Rewrites[0].NewValue)

	require.Equal(t, "tools.avrdude.cmd.path", platformKeysRewrite.Rewrites[1].Key)
	require.Equal(t, "{runtime.ide.path}/hardware/tools/avr/bin/avrdude", platformKeysRewrite.Rewrites[1].OldValue)
	require.Equal(t, "{path}/bin/avrdude", platformKeysRewrite.Rewrites[1].NewValue)

	require.Equal(t, "compiler.path", platformKeysRewrite.Rewrites[3].Key)
	require.Equal(t, "{runtime.ide.path}/hardware/tools/gcc-arm-none-eabi-4.8.3-2014q1/bin/", platformKeysRewrite.Rewrites[3].OldValue)
	require.Equal(t, "{runtime.tools.arm-none-eabi-gcc.path}/bin/", platformKeysRewrite.Rewrites[3].NewValue)

	require.Equal(t, "recipe.c.combine.pattern", platformKeysRewrite.Rewrites[5].Key)
	require.Equal(t, "\"{compiler.path}{compiler.c.elf.cmd}\" {compiler.c.elf.flags} -mcpu={build.mcu} \"-T{build.variant.path}/{build.ldscript}\" \"-Wl,-Map,{build.path}/{build.project_name}.map\" {compiler.c.elf.extra_flags} -o \"{build.path}/{build.project_name}.elf\" \"-L{build.path}\" -mthumb -Wl,--cref -Wl,--check-sections -Wl,--gc-sections -Wl,--entry=Reset_Handler -Wl,--unresolved-symbols=report-all -Wl,--warn-common -Wl,--warn-section-align -Wl,--warn-unresolved-symbols -Wl,--start-group \"{build.path}/syscalls_sam3.c.o\" {object_files} \"{build.variant.path}/{build.variant_system_lib}\" \"{build.path}/{archive_file}\" -Wl,--end-group -lm -gcc", platformKeysRewrite.Rewrites[5].OldValue)
	require.Equal(t, "\"{compiler.path}{compiler.c.elf.cmd}\" {compiler.c.elf.flags} -mcpu={build.mcu} \"-T{build.variant.path}/{build.ldscript}\" \"-Wl,-Map,{build.path}/{build.project_name}.map\" {compiler.c.elf.extra_flags} -o \"{build.path}/{build.project_name}.elf\" \"-L{build.path}\" -mthumb -Wl,--cref -Wl,--check-sections -Wl,--gc-sections -Wl,--entry=Reset_Handler -Wl,--unresolved-symbols=report-all -Wl,--warn-common -Wl,--warn-section-align -Wl,--warn-unresolved-symbols -Wl,--start-group \"{build.path}/core/syscalls_sam3.c.o\" {object_files} \"{build.variant.path}/{build.variant_system_lib}\" \"{build.path}/{archive_file}\" -Wl,--end-group -lm -gcc", platformKeysRewrite.Rewrites[5].NewValue)
}
