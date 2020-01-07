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
