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
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/internal/arduino/builder/logger"
	"github.com/arduino/arduino-cli/internal/i18n"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
)

// RunRecipe fixdoc
func (b *Builder) RunRecipe(prefix, suffix string, skipIfOnlyUpdatingCompilationDatabase bool) error {
	return b.RunRecipeWithProps(prefix, suffix, b.buildProperties, skipIfOnlyUpdatingCompilationDatabase)
}

func (b *Builder) RunRecipeWithProps(prefix, suffix string, buildProperties *properties.Map, skipIfOnlyUpdatingCompilationDatabase bool) error {
	logrus.Debugf("Looking for recipes like %s", prefix+"*"+suffix)

	recipes := findRecipes(buildProperties, prefix, suffix)

	for _, recipe := range recipes {
		logrus.Debugf("Running recipe: %s", recipe)

		command, err := b.prepareCommandForRecipe(buildProperties, recipe, false)
		if err != nil {
			return err
		}

		if b.onlyUpdateCompilationDatabase && skipIfOnlyUpdatingCompilationDatabase {
			if b.logger.VerbosityLevel() == logger.VerbosityVerbose {
				b.logger.Info(i18n.Tr("Skipping: %[1]s", strings.Join(command.GetArgs(), " ")))
			}
			return nil
		}

		if err := b.execCommand(command); err != nil {
			return err
		}
	}

	return nil
}

func findRecipes(buildProperties *properties.Map, patternPrefix string, patternSuffix string) []string {
	var recipes []string

	exactKey := patternPrefix + patternSuffix

	for _, key := range buildProperties.Keys() {
		if key != exactKey && strings.HasPrefix(key, patternPrefix) && strings.HasSuffix(key, patternSuffix) && buildProperties.Get(key) != "" {
			recipes = append(recipes, key)
		}
	}

	// If no recipes were found, check if the exact key exists
	if len(recipes) == 0 && buildProperties.Get(exactKey) != "" {
		recipes = append(recipes, exactKey)
	}

	sort.Strings(recipes)

	return recipes
}
