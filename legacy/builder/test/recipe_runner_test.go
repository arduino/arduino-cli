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

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

// TODO
// I can't find a command I can run on linux, mac and windows
// and that allows to test if the recipe is actually run
// So this test is pretty useless
func TestRecipeRunner(t *testing.T) {
	ctx := &types.Context{}
	buildProperties := properties.NewMap()
	ctx.BuildProperties = buildProperties

	buildProperties.Set("recipe.hooks.prebuild.1.pattern", "echo")

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.RecipeByPrefixSuffixRunner{Prefix: constants.HOOKS_PREBUILD, Suffix: constants.HOOKS_PATTERN_SUFFIX},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}
}

func TestRecipesComposition(t *testing.T) {
	require.Equal(t, "recipe.hooks.core.postbuild", constants.HOOKS_CORE_POSTBUILD)
	require.Equal(t, "recipe.hooks.postbuild", constants.HOOKS_POSTBUILD)
	require.Equal(t, "recipe.hooks.linking.prelink", constants.HOOKS_LINKING_PRELINK)
	require.Equal(t, "recipe.hooks.objcopy.preobjcopy", constants.HOOKS_OBJCOPY_PREOBJCOPY)
}
