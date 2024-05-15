// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package configmap

import (
	"gopkg.in/yaml.v3"
)

func (c Map) MarshalYAML() (interface{}, error) {
	return c.values, nil
}

func (c *Map) UnmarshalYAML(node *yaml.Node) error {
	in := map[string]any{}
	if err := node.Decode(&in); err != nil {
		return err
	}

	errs := &UnmarshalErrors{}
	for k, v := range flattenMap(in) {
		if err := c.Set(k, v); err != nil {
			errs.append(err)
		}
	}
	return errs.result()
}
