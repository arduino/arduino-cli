/*
 * This file is part of PropertiesOrderedMap library.
 *
 * Copyright 2018 Arduino AG (http://www.arduino.cc/)
 *
 * PropertiesOrderedMap library is free software; you can redistribute it and/or modify
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
 */

package properties

import (
	"encoding/json"
)

// XXX: no simple way to preserve ordering in JSON.

// MarshalJSON implements json.Marshaler interface
func (m *Map) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.kv)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (m *Map) UnmarshalJSON(d []byte) error {
	var obj map[string]string
	if err := json.Unmarshal(d, &obj); err != nil {
		return err
	}

	m.kv = map[string]string{}
	m.o = []string{}
	for k, v := range obj {
		m.Set(k, v)
	}
	return nil
}
