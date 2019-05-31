/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

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
		HardwareDirs: paths.NewPathList("downloaded_hardware", filepath.Join("..", "hardware"), "hardware"),
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
