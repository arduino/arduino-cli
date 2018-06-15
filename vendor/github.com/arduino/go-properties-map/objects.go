/*
 * This file is part of PropertiesMap library.
 *
 * Copyright 2018 Arduino AG (http://www.arduino.cc/)
 *
 * PropertiesMap library is free software; you can redistribute it and/or modify
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
	"github.com/arduino/go-paths-helper"
)

// GetBoolean returns true if the map contains the specified key and the value
// equals to the string "true", in any other case returns false.
func (m Map) GetBoolean(key string) bool {
	value, ok := m[key]
	return ok && value == "true"
}

// SetBoolean sets the specified key to the string "true" or "false" if the value
// is respectively true or false.
func (m Map) SetBoolean(key string, value bool) {
	if value {
		m[key] = "true"
	} else {
		m[key] = "false"
	}
}

// GetPath returns a paths.Path object using the map value as path. The function
// returns nil if the key is not present.
func (m Map) GetPath(key string) *paths.Path {
	value, ok := m[key]
	if !ok {
		return nil
	}
	return paths.New(value)
}

// SetPath saves the paths.Path object in the map using the path as value of the map
func (m Map) SetPath(key string, value *paths.Path) {
	if value == nil {
		m[key] = ""
	} else {
		m[key] = value.String()
	}
}
