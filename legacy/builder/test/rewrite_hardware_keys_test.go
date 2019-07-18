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
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

func TestRewriteHardwareKeys(t *testing.T) {
	ctx := &types.Context{}

	packages := cores.Packages{}
	aPackage := &cores.Package{Name: "dummy"}
	packages["dummy"] = aPackage
	aPackage.Platforms = map[string]*cores.Platform{}

	platform := &cores.PlatformRelease{
		Properties: properties.NewFromHashmap(map[string]string{
			"name":          "A test platform",
			"compiler.path": "{runtime.ide.path}/hardware/tools/avr/bin/",
		}),
	}
	aPackage.Platforms["dummy"] = &cores.Platform{
		Architecture: "dummy",
		Releases: map[string]*cores.PlatformRelease{
			"": platform,
		},
	}

	ctx.Hardware = packages

	rewrite := types.PlatforKeyRewrite{Key: "compiler.path", OldValue: "{runtime.ide.path}/hardware/tools/avr/bin/", NewValue: "{runtime.tools.avr-gcc.path}/bin/"}
	platformKeysRewrite := types.PlatforKeysRewrite{Rewrites: []types.PlatforKeyRewrite{rewrite}}
	ctx.PlatformKeyRewrites = platformKeysRewrite

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.RewriteHardwareKeys{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Equal(t, "{runtime.tools.avr-gcc.path}/bin/", platform.Properties.Get("compiler.path"))
}

func TestRewriteHardwareKeysWithRewritingDisabled(t *testing.T) {
	ctx := &types.Context{}

	packages := cores.Packages{}
	aPackage := &cores.Package{Name: "dummy"}
	packages["dummy"] = aPackage
	aPackage.Platforms = make(map[string]*cores.Platform)

	platform := &cores.PlatformRelease{
		Properties: properties.NewFromHashmap(map[string]string{
			"name":          "A test platform",
			"compiler.path": "{runtime.ide.path}/hardware/tools/avr/bin/",
			"rewriting":     "disabled",
		}),
	}
	aPackage.Platforms["dummy"] = &cores.Platform{
		Architecture: "dummy",
		Releases: map[string]*cores.PlatformRelease{
			"": platform,
		},
	}

	ctx.Hardware = packages

	rewrite := types.PlatforKeyRewrite{Key: "compiler.path", OldValue: "{runtime.ide.path}/hardware/tools/avr/bin/", NewValue: "{runtime.tools.avr-gcc.path}/bin/"}
	platformKeysRewrite := types.PlatforKeysRewrite{Rewrites: []types.PlatforKeyRewrite{rewrite}}

	ctx.PlatformKeyRewrites = platformKeysRewrite

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.RewriteHardwareKeys{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	require.Equal(t, "{runtime.ide.path}/hardware/tools/avr/bin/", platform.Properties.Get("compiler.path"))
}
