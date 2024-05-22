// This file is part of arduino-cli.
//
// Copyright 2020-2022 ARDUINO SA (http://www.arduino.cc/)
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

package sketch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/go-paths-helper"
	"gopkg.in/yaml.v3"
)

// updateOrAddYamlRootEntry updates or adds a new entry to the root of the yaml file.
// If the value is empty the entry is removed.
func updateOrAddYamlRootEntry(path *paths.Path, key, newValue string) error {
	var srcYaml []string
	if path.Exist() {
		src, err := path.ReadFileAsLines()
		if err != nil {
			return err
		}
		lastLine := len(src) - 1
		if lastLine > 0 && src[lastLine] == "" {
			srcYaml = src[:lastLine]
		} else {
			srcYaml = src
		}
	}

	// Generate the new yaml key/value pair
	v, err := yaml.Marshal(newValue)
	if err != nil {
		return err
	}
	updatedLine := key + ": " + strings.TrimSpace(string(v))

	// Update or add the key/value pair into the original yaml
	addMissing := (newValue != "")
	for i, line := range srcYaml {
		if strings.HasPrefix(line, key+": ") {
			if newValue == "" {
				// Remove the key/value pair
				srcYaml = append(srcYaml[:i], srcYaml[i+1:]...)
			} else {
				// Update the key/value pair
				srcYaml[i] = updatedLine
			}
			addMissing = false
			break
		}
	}
	if addMissing {
		lastLine := len(srcYaml) - 1
		if lastLine >= 0 && srcYaml[lastLine] == "" {
			srcYaml[lastLine] = updatedLine
		} else {
			srcYaml = append(srcYaml, updatedLine)
		}
	}

	// Validate the new yaml
	dstYaml := []byte(strings.Join(srcYaml, fmt.Sprintln()) + fmt.Sprintln())
	var dst interface{}
	if err := yaml.Unmarshal(dstYaml, &dst); err != nil {
		return fmt.Errorf("%s: %w", i18n.Tr("could not update sketch project file"), err)
	}
	dstMap, ok := dst.(map[string]interface{})
	if !ok {
		return errors.New(i18n.Tr("could not update sketch project file"))
	}
	writtenValue, notRemoved := dstMap[key]
	if (newValue == "" && notRemoved) || (newValue != "" && newValue != writtenValue) {
		return errors.New(i18n.Tr("could not update sketch project file"))
	}

	// Write back the updated YAML
	return path.WriteFile(dstYaml)
}
