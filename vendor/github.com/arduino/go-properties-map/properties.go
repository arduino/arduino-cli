/*
 * This file is part of PropertiesMap library.
 *
 * Copyright 2017 Arduino AG (http://www.arduino.cc/)
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

/*
Package properties is a library for handling maps of hierarchical properties.
This library is mainly used in the Arduino platform software to handle
configurations made of key/value pairs stored in files with an INI like
syntax, for example:

 ...
 uno.name=Arduino/Genuino Uno
 uno.upload.tool=avrdude
 uno.upload.protocol=arduino
 uno.upload.maximum_size=32256
 uno.upload.maximum_data_size=2048
 uno.upload.speed=115200
 uno.build.mcu=atmega328p
 uno.build.f_cpu=16000000L
 uno.build.board=AVR_UNO
 uno.build.core=arduino
 uno.build.variant=standard
 diecimila.name=Arduino Duemilanove or Diecimila
 diecimila.upload.tool=avrdude
 diecimila.upload.protocol=arduino
 diecimila.build.f_cpu=16000000L
 diecimila.build.board=AVR_DUEMILANOVE
 diecimila.build.core=arduino
 diecimila.build.variant=standard
 ...

This library has methods to parse this kind of files into a Map object.

The Map object has many helper methods to accomplish some common operation
on this kind of data like cloning, merging, comparing and also extracting
a submap or generating a map-of-submaps from the first "level" of the hierarchy.

On the Arduino platform the properties are used to populate command line recipes
so there are some methods to help this task like SplitQuotedString or ExpandPropsInString.
*/
package properties

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/arduino/go-paths-helper"
)

// Map is a container of properties
type Map map[string]string

var osSuffix string

func init() {
	switch value := runtime.GOOS; value {
	case "linux", "freebsd", "windows":
		osSuffix = runtime.GOOS
	case "darwin":
		osSuffix = "macosx"
	default:
		panic("Unsupported OS")
	}
}

// Load reads a properties file and makes a Map out of it.
func Load(filepath string) (Map, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("Error reading file: %s", err)
	}

	text := string(bytes)
	text = strings.Replace(text, "\r\n", "\n", -1)
	text = strings.Replace(text, "\r", "\n", -1)

	properties := make(Map)

	for lineNum, line := range strings.Split(text, "\n") {
		if err := properties.parseLine(line); err != nil {
			return nil, fmt.Errorf("Error reading file (%s:%d): %s", filepath, lineNum, err)
		}
	}

	return properties, nil
}

// LoadFromPath reads a properties file and makes a Map out of it.
func LoadFromPath(path *paths.Path) (Map, error) {
	return Load(path.String())
}

// LoadFromSlice reads a properties file from an array of string
// and makes a Map out of it
func LoadFromSlice(lines []string) (Map, error) {
	properties := make(Map)

	for lineNum, line := range lines {
		if err := properties.parseLine(line); err != nil {
			return nil, fmt.Errorf("Error reading from slice (index:%d): %s", lineNum, err)
		}
	}

	return properties, nil
}

func (m Map) parseLine(line string) error {
	line = strings.TrimSpace(line)

	// Skip empty lines or comments
	if len(line) == 0 || line[0] == '#' {
		return nil
	}

	lineParts := strings.SplitN(line, "=", 2)
	if len(lineParts) != 2 {
		return fmt.Errorf("Invalid line format, should be 'key=value'")
	}
	key := strings.TrimSpace(lineParts[0])
	value := strings.TrimSpace(lineParts[1])

	key = strings.Replace(key, "."+osSuffix, "", 1)
	m[key] = value

	return nil
}

// SafeLoadFromPath is like LoadFromPath, except that it returns an empty Map if
// the specified file doesn't exists
func SafeLoadFromPath(path *paths.Path) (Map, error) {
	return SafeLoad(path.String())
}

// SafeLoad is like Load, except that it returns an empty Map if the specified
// file doesn't exists
func SafeLoad(filepath string) (Map, error) {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return make(Map), nil
	}

	properties, err := Load(filepath)
	if err != nil {
		return nil, err
	}
	return properties, nil
}

// FirstLevelOf generates a map-of-Maps using the first level of the hierarchy
// of the current Map. For example the following Map:
//
//  properties.Map{
//    "uno.name": "Arduino/Genuino Uno",
//    "uno.upload.tool": "avrdude",
//    "uno.upload.protocol": "arduino",
//    "uno.upload.maximum_size": "32256",
//    "diecimila.name": "Arduino Duemilanove or Diecimila",
//    "diecimila.upload.tool": "avrdude",
//    "diecimila.upload.protocol": "arduino",
//    "diecimila.bootloader.tool": "avrdude",
//    "diecimila.bootloader.low_fuses": "0xFF",
//  }
//
// is transformed into the following map-of-Maps:
//
//  map[string]Map{
//    "uno" : properties.Map{
//      "name": "Arduino/Genuino Uno",
//      "upload.tool": "avrdude",
//      "upload.protocol": "arduino",
//      "upload.maximum_size": "32256",
//    },
//    "diecimila" : properties.Map{
//      "name=Arduino Duemilanove or Diecimila
//      "upload.tool": "avrdude",
//      "upload.protocol": "arduino",
//      "bootloader.tool": "avrdude",
//      "bootloader.low_fuses": "0xFF",
//    }
//  }
func (m Map) FirstLevelOf() map[string]Map {
	newMap := make(map[string]Map)
	for key, value := range m {
		if strings.Index(key, ".") == -1 {
			continue
		}
		keyParts := strings.SplitN(key, ".", 2)
		if newMap[keyParts[0]] == nil {
			newMap[keyParts[0]] = make(Map)
		}
		newMap[keyParts[0]][keyParts[1]] = value
	}
	return newMap
}

// SubTree extracts a sub Map from an existing map using the first level
// of the keys hierarchy as selector.
// For example the following Map:
//
//  properties.Map{
//    "uno.name": "Arduino/Genuino Uno",
//    "uno.upload.tool": "avrdude",
//    "uno.upload.protocol": "arduino",
//    "uno.upload.maximum_size": "32256",
//    "diecimila.name": "Arduino Duemilanove or Diecimila",
//    "diecimila.upload.tool": "avrdude",
//    "diecimila.upload.protocol": "arduino",
//    "diecimila.bootloader.tool": "avrdude",
//    "diecimila.bootloader.low_fuses": "0xFF",
//  }
//
// after calling SubTree("uno") will be transformed in:
//
//  properties.Map{
//    "name": "Arduino/Genuino Uno",
//    "upload.tool": "avrdude",
//    "upload.protocol": "arduino",
//    "upload.maximum_size": "32256",
//  },
func (m Map) SubTree(rootKey string) Map {
	rootKey += "."
	newMap := Map{}
	for key, value := range m {
		if !strings.HasPrefix(key, rootKey) {
			continue
		}
		newMap[key[len(rootKey):]] = value
	}
	return newMap
}

// ExpandPropsInString use the Map to replace values into a format string.
// The format string should contains markers between braces, for example:
//
//  "The selected upload protocol is {upload.protocol}."
//
// Each marker is replaced by the corresponding value of the Map.
// The values in the Map may contains other markers, they are evaluated
// recursively up to 10 times.
func (m Map) ExpandPropsInString(str string) string {
	for i := 0; i < 10; i++ {
		newStr := str
		for key, value := range m {
			newStr = strings.Replace(newStr, "{"+key+"}", value, -1)
		}
		if str == newStr {
			break
		}
		str = newStr
	}
	return str
}

// Merge merges a Map into this one. Each key/value of the merged Maps replaces
// the key/value present in the original Map.
func (m Map) Merge(sources ...Map) Map {
	for _, source := range sources {
		for key, value := range source {
			m[key] = value
		}
	}
	return m
}

// Keys returns an array of the keys contained in the Map
func (m Map) Keys() []string {
	keys := make([]string, len(m))
	i := 0
	for key := range m {
		keys[i] = key
		i++
	}
	return keys
}

// Values returns an array of the values contained in the Map. Duplicated
// values are repeated in the list accordingly.
func (m Map) Values() []string {
	values := make([]string, len(m))
	i := 0
	for _, value := range m {
		values[i] = value
		i++
	}
	return values
}

// Clone makes a copy of the Map
func (m Map) Clone() Map {
	clone := make(Map)
	clone.Merge(m)
	return clone
}

// Equals returns true if the current Map contains the same key/value pairs of
// the Map passed as argument.
func (m Map) Equals(other Map) bool {
	return reflect.DeepEqual(m, other)
}

// MergeMapsOfProperties merge the map-of-Maps (obtained from the method FirstLevelOf()) into the
// target map-of-Maps.
func MergeMapsOfProperties(target map[string]Map, sources ...map[string]Map) map[string]Map {
	for _, source := range sources {
		for key, value := range source {
			target[key] = value
		}
	}
	return target
}

// DeleteUnexpandedPropsFromString removes all the brace markers "{xxx}" that are not expanded
// into a value using the Map.ExpandPropsInString() method.
func DeleteUnexpandedPropsFromString(str string) string {
	rxp := regexp.MustCompile("\\{.+?\\}")
	return rxp.ReplaceAllString(str, "")
}

// Dump returns a representation of the map in golang source format
func (m Map) Dump() string {
	res := "properties.Map{\n"
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		res += fmt.Sprintf("  \"%s\": \"%s\",\n", strings.Replace(k, `"`, `\"`, -1), strings.Replace(m[k], `"`, `\"`, -1))
	}
	res += "}"
	return res
}
