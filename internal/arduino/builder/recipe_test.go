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
	"testing"

	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

func TestRecipeFinder(t *testing.T) {
	t.Run("NumberedRecipes", func(t *testing.T) {
		buildProperties := properties.NewMap()
		buildProperties.Set("recipe.test", "test")
		buildProperties.Set("recipe.1.test", "test2")
		buildProperties.Set("recipe.2.test", "test3")
		recipes := findRecipes(buildProperties, "recipe", ".test")
		require.Equal(t, []string{"recipe.1.test", "recipe.2.test"}, recipes)
	})
	t.Run("NumberedRecipesWithGaps", func(t *testing.T) {
		buildProperties := properties.NewMap()
		buildProperties.Set("recipe.test", "test")
		buildProperties.Set("recipe.2.test", "test3")
		buildProperties.Set("recipe.0.test", "test2")
		recipes := findRecipes(buildProperties, "recipe", ".test")
		require.Equal(t, []string{"recipe.0.test", "recipe.2.test"}, recipes)
	})
	t.Run("NumberedRecipesWithGapsAndDifferentLenghtNumbers", func(t *testing.T) {
		buildProperties := properties.NewMap()
		buildProperties.Set("recipe.test", "test")
		buildProperties.Set("recipe.12.test", "test3")
		buildProperties.Set("recipe.2.test", "test2")
		recipes := findRecipes(buildProperties, "recipe", ".test")
		// The order is sorted alphabetically, not numerically
		require.Equal(t, []string{"recipe.12.test", "recipe.2.test"}, recipes)
	})
	t.Run("NumberedRecipesWithGapsAndNumbers", func(t *testing.T) {
		buildProperties := properties.NewMap()
		buildProperties.Set("recipe.test", "test")
		buildProperties.Set("recipe.12.test", "test3")
		buildProperties.Set("recipe.02.test", "test2")
		buildProperties.Set("recipe.09.test", "test2")
		recipes := findRecipes(buildProperties, "recipe", ".test")
		require.Equal(t, []string{"recipe.02.test", "recipe.09.test", "recipe.12.test"}, recipes)
	})
	t.Run("UnnumberedRecipies", func(t *testing.T) {
		buildProperties := properties.NewMap()
		buildProperties.Set("recipe.test", "test")
		buildProperties.Set("recipe.a.test", "test3")
		buildProperties.Set("recipe.b.test", "test2")
		recipes := findRecipes(buildProperties, "recipe", ".test")
		require.Equal(t, []string{"recipe.a.test", "recipe.b.test"}, recipes)
	})
	t.Run("ObjcopyRecipies/1", func(t *testing.T) {
		buildProperties := properties.NewMap()
		buildProperties.Set("recipe.objcopy.eep.pattern", "test")
		buildProperties.Set("recipe.objcopy.hex.pattern", "test")
		recipes := findRecipes(buildProperties, "recipe.objcopy", ".pattern")
		require.Equal(t, []string{"recipe.objcopy.eep.pattern", "recipe.objcopy.hex.pattern"}, recipes)
	})
	t.Run("ObjcopyRecipies/2", func(t *testing.T) {
		buildProperties := properties.NewMap()
		buildProperties.Set("recipe.objcopy.partitions.bin.pattern", "test")
		recipes := findRecipes(buildProperties, "recipe.objcopy", ".pattern")
		require.Equal(t, []string{"recipe.objcopy.partitions.bin.pattern"}, recipes)
	})
}
