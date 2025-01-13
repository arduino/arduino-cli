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
	"fmt"
	"reflect"
	"strings"

	"fortio.org/safecast"
)

type Map struct {
	values map[string]any
	schema map[string]reflect.Type
}

func New() *Map {
	return &Map{
		values: make(map[string]any),
		schema: make(map[string]reflect.Type),
	}
}

func (c Map) Get(key string) any {
	value, _ := c.GetOk(key)
	return value
}

func (c Map) GetOk(key string) (any, bool) {
	keys := strings.Split(key, ".")
	return c.get(keys)
}

func (c Map) get(keys []string) (any, bool) {
	if len(keys) == 0 {
		return nil, false
	}
	value, ok := c.values[keys[0]]
	if len(keys) == 1 {
		return value, ok
	}

	if subConf, ok := value.(*Map); ok {
		return subConf.get(keys[1:])
	}
	return nil, false
}

func (c Map) Set(key string, value any) error {
	if len(c.schema) > 0 {
		t, ok := c.schema[key]
		if !ok {
			return fmt.Errorf("schema not defined for key '%s'", key)
		}
		newValue, err := tryConversion(value, t)
		if err != nil {
			return fmt.Errorf("invalid type for key '%s': %w", key, err)
		}
		value = newValue
	}
	keys := strings.Split(key, ".")
	c.set(keys, value)
	return nil
}

func tryConversion(current any, desiredType reflect.Type) (any, error) {
	currentType := reflect.TypeOf(current)
	if currentType == desiredType {
		return current, nil
	}

	switch desiredType.Kind() {
	case reflect.Uint:
		// Exception for JSON decoder: json decoder will decode all numbers as float64
		if currentFloat, ok := current.(float64); ok {
			return uint(currentFloat), nil
		}
		if currentInt, ok := current.(int); ok {
			return safecast.Convert[uint](currentInt)
		}
	case reflect.Int:
		// Exception for JSON decoder: json decoder will decode all numbers as float64
		if currentFloat, ok := current.(float64); ok {
			return int(currentFloat), nil
		}
	case reflect.Array, reflect.Slice:
		currentArray, ok := current.([]any)
		if !ok && current != nil {
			break
		}

		resArray := reflect.MakeSlice(desiredType, len(currentArray), len(currentArray))
		for i, elem := range currentArray {
			newElem, err := tryConversion(elem, desiredType.Elem())
			if err != nil {
				return nil, err
			}
			resArray.Index(i).Set(reflect.ValueOf(newElem))
		}
		return resArray.Interface(), nil
	}

	currentTypeString := currentType.String()
	if currentTypeString == "[]interface {}" {
		currentTypeString = "array"
	}
	return nil, fmt.Errorf("invalid conversion, got %s but want %v", currentTypeString, desiredType)
}

func (c Map) set(keys []string, value any) {
	if len(keys) == 0 {
		return
	}
	if len(keys) == 1 {
		c.values[keys[0]] = value
		return
	}

	var subConf *Map
	if subValue, ok := c.values[keys[0]]; !ok {
		subConf = New()
		c.values[keys[0]] = subConf
	} else if conf, ok := subValue.(*Map); !ok {
		subConf = New()
		c.values[keys[0]] = subConf
	} else {
		subConf = conf
	}
	subConf.set(keys[1:], value)
}

func (c Map) Delete(key string) {
	keys := strings.Split(key, ".")
	c.delete(keys)
}

func (c Map) delete(keys []string) {
	if len(keys) == 0 {
		return
	}
	if len(keys) == 1 {
		delete(c.values, keys[0])
		return
	}

	if subValue, ok := c.values[keys[0]]; !ok {
		return
	} else if subConf, ok := subValue.(*Map); !ok {
		return
	} else {
		subConf.delete(keys[1:])
	}
}

func (c *Map) Merge(x *Map) error {
	for xk, xv := range x.values {
		if xSubConf, ok := xv.(*Map); ok {
			if subConf, ok := c.values[xk].(*Map); ok {
				if err := subConf.Merge(xSubConf); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("cannot merge sub-configuration into non sub-configuration: '%s'", xk)
		}

		v, ok := c.values[xk]
		if !ok {
			return fmt.Errorf("target key do not exist: '%s'", xk)
		}
		if reflect.TypeOf(v) != reflect.TypeOf(xv) {
			return fmt.Errorf("invalid types for key '%s': got %T but want %T", xk, v, xv)
		}
		c.values[xk] = xv
	}
	return nil
}

func (c *Map) AllKeys() []string {
	return c.allKeys("")
}

func (c *Map) Schema() map[string]reflect.Type {
	return c.schema
}

func (c *Map) allKeys(prefix string) []string {
	keys := []string{}
	if len(c.schema) > 0 {
		for k := range c.schema {
			keys = append(keys, prefix+k)
		}
	} else {
		for k, v := range c.values {
			if subConf, ok := v.(*Map); ok {
				keys = append(keys, subConf.allKeys(prefix+k+".")...)
			} else {
				keys = append(keys, prefix+k)
			}
		}
	}
	return keys
}

func (c *Map) SetKeyTypeSchema(key string, t any) {
	c.schema[key] = reflect.TypeOf(t)
}
