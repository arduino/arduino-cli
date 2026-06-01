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

import "encoding/json"

func (c Map) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.values)
}

func (c *Map) UnmarshalJSON(data []byte) error {
	in := map[string]any{}
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}

	c.values = map[string]any{}
	for k, v := range flattenMap(in) {
		if err := c.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func flattenMap(in map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		switch v := v.(type) {
		case map[string]any:
			for kk, vv := range flattenMap(v) {
				out[k+"."+kk] = vv
			}
		default:
			out[k] = v
		}
	}
	return out
}
