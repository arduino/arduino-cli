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
	"errors"
	"fmt"
	"time"
)

func (c Map) GetStringOk(key string) (string, bool, error) {
	v, ok := c.GetOk(key)
	if !ok {
		return "", false, nil
	}
	if s, ok := v.(string); ok {
		return s, true, nil
	}
	return "", false, errors.New(key + " is not a string")
}

func (c Map) GetString(key string) string {
	v, ok, err := c.GetStringOk(key)
	if err != nil {
		panic(err.Error())
	}
	if ok {
		return v
	}
	return ""
}

func (c Map) SetString(key string, value string) {
	c.Set(key, value)
}

func (c Map) GetBoolOk(key string) (bool, bool, error) {
	v, ok := c.GetOk(key)
	if !ok {
		return false, false, nil
	}
	if b, ok := v.(bool); ok {
		return b, true, nil
	}
	return false, false, errors.New(key + " is not a bool")
}

func (c Map) GetBool(key string) bool {
	v, ok, err := c.GetBoolOk(key)
	if err != nil {
		panic(err.Error())
	}
	if ok {
		return v
	}
	return false
}

func (c Map) SetBool(key string, value bool) {
	c.Set(key, value)
}

func (c Map) GetUintOk(key string) (uint, bool, error) {
	v, ok := c.GetOk(key)
	if !ok {
		return 0, false, nil
	}
	if i, ok := v.(uint); ok {
		return i, true, nil
	}
	return 0, false, errors.New(key + " is not a uint")
}

func (c Map) GetUint(key string) uint {
	v, ok, err := c.GetUintOk(key)
	if err != nil {
		panic(err.Error())
	}
	if ok {
		return v
	}
	return 0
}

func (c Map) SetUint(key string, value uint) {
	c.Set(key, value)
}

func (c Map) GetIntOk(key string) (int, bool, error) {
	v, ok := c.GetOk(key)
	if !ok {
		return 0, false, nil
	}
	if i, ok := v.(int); ok {
		return i, true, nil
	}
	return 0, false, errors.New(key + " is not a uint")
}

func (c Map) GetInt(key string) int {
	v, ok, err := c.GetIntOk(key)
	if err != nil {
		panic(err.Error())
	}
	if ok {
		return v
	}
	return 0
}

func (c Map) SetInt(key string, value int) {
	c.Set(key, value)
}

func (c Map) GetUint32Ok(key string) (uint32, bool, error) {
	v, ok := c.GetOk(key)
	if !ok {
		return 0, false, nil
	}
	if i, ok := v.(uint32); ok {
		return i, true, nil
	}
	return 0, false, errors.New(key + " is not a uint32")
}

func (c Map) GetUint32(key string) uint32 {
	v, ok, err := c.GetUint32Ok(key)
	if err != nil {
		panic(err.Error())
	}
	if ok {
		return v
	}
	return 0
}

func (c Map) SetUint32(key string, value uint32) {
	c.Set(key, value)
}

func (c Map) GetStringSliceOk(key string) ([]string, bool, error) {
	v, ok := c.GetOk(key)
	if !ok {
		return nil, false, nil
	}
	if genArray, ok := v.([]string); ok {
		return genArray, true, nil
	}
	if genArray, ok := v.([]interface{}); ok {
		// transform []interface{} to []string
		var strArray []string
		for i, gen := range genArray {
			if str, ok := gen.(string); ok {
				strArray = append(strArray, str)
			} else {
				return nil, false, fmt.Errorf("%s[%d] is not a string", key, i)
			}
		}
		return strArray, true, nil
	}
	return nil, false, fmt.Errorf("%s is not an array of strings", key)
}

func (c Map) GetStringSlice(key string) []string {
	v, ok, err := c.GetStringSliceOk(key)
	if err != nil {
		panic(err.Error())
	}
	if ok {
		return v
	}
	return nil
}

func (c Map) GetDurationOk(key string) (time.Duration, bool, error) {
	v, ok := c.GetOk(key)
	if !ok {
		return 0, false, nil
	}
	if s, ok := v.(string); !ok {
		return 0, false, errors.New(key + " is not a Duration")
	} else if d, err := time.ParseDuration(s); err != nil {
		return 0, false, fmt.Errorf("%s is not a valid Duration: %w", key, err)
	} else {
		return d, true, nil
	}
}

func (c Map) GetDuration(key string) time.Duration {
	v, ok, err := c.GetDurationOk(key)
	if err != nil {
		panic(err.Error())
	}
	if ok {
		return v
	}
	return 0
}

func (c Map) SetDuration(key string, value time.Duration) {
	c.SetString(key, value.String())
}
