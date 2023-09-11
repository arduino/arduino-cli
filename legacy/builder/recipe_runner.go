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
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder/utils"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func RecipeByPrefixSuffixRunner(
	prefix, suffix string,
	skipIfOnlyUpdatingCompilationDatabase, onlyUpdateCompilationDatabase, verbose bool,
	buildProps *properties.Map,
	stdoutWriter, stderrWriter io.Writer,
	printInfoFn func(string),
) error {
	logrus.Debugf(fmt.Sprintf("Looking for recipes like %s", prefix+"*"+suffix))

	// TODO is it necessary to use Clone?
	buildProperties := buildProps.Clone()
	recipes := findRecipes(buildProperties, prefix, suffix)

	// TODO is it necessary to use Clone?
	properties := buildProperties.Clone()
	for _, recipe := range recipes {
		logrus.Debugf(fmt.Sprintf("Running recipe: %s", recipe))

		command, err := utils.PrepareCommandForRecipe(properties, recipe, false)
		if err != nil {
			return errors.WithStack(err)
		}

		if onlyUpdateCompilationDatabase && skipIfOnlyUpdatingCompilationDatabase {
			if verbose {
				printInfoFn(tr("Skipping: %[1]s", strings.Join(command.GetArgs(), " ")))
			}
			return nil
		}

		verboseInfo, _, _, err := utils.ExecCommand(verbose, stdoutWriter, stderrWriter, command, utils.ShowIfVerbose /* stdout */, utils.Show /* stderr */)
		if verbose {
			printInfoFn(string(verboseInfo))
		}
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func findRecipes(buildProperties *properties.Map, patternPrefix string, patternSuffix string) []string {
	var recipes []string
	for _, key := range buildProperties.Keys() {
		if strings.HasPrefix(key, patternPrefix) && strings.HasSuffix(key, patternSuffix) && buildProperties.Get(key) != "" {
			recipes = append(recipes, key)
		}
	}

	sort.Strings(recipes)

	return recipes
}
