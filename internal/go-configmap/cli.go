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
	"strconv"
	"strings"
)

func (c *Map) SetFromCLIArgs(key string, args ...string) error {
	if len(args) == 0 {
		c.Delete(key)
		return nil
	}

	// Some args might be coming from env vars that are specifying multiple values
	// in a single string. We expand those cases by splitting every args with whitespace
	// Example: args=["e1 e2", "e3"] -> args=["e1","e2","e3"]
	argsExpantion := func(values []string) []string {
		result := []string{}
		for _, v := range values {
			result = append(result, strings.Split(v, " ")...)
		}
		return result
	}
	args = argsExpantion(args)
	// in case of schemaless configuration, we don't know the type of the setting
	// we will save it as a string or array of strings
	if len(c.schema) == 0 {
		switch len(args) {
		case 1:
			c.Set(key, args[0])
		default:
			c.Set(key, args)
		}
		return nil
	}

	// Find the correct type for the given setting
	valueType, ok := c.schema[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	var value any
	isArray := false
	{
		var conversionError error
		switch valueType.String() {
		case "uint":
			value, conversionError = strconv.Atoi(args[0])
		case "bool":
			value, conversionError = strconv.ParseBool(args[0])
		case "string":
			value = args[0]
		case "[]string":
			value = args
			isArray = true
		default:
			return fmt.Errorf("unhandled type: %s", valueType)
		}
		if conversionError != nil {
			return fmt.Errorf("error setting value: %v", conversionError)
		}
	}
	if !isArray && len(args) != 1 {
		return fmt.Errorf("error setting value: key is not an array, but multiple values were provided")
	}

	return c.Set(key, value)
}

func (c *Map) InjectEnvVars(env []string, prefix string) []error {
	if prefix != "" {
		prefix = strings.ToUpper(prefix) + "_"
	}

	errs := []error{}

	envKeyToConfigKey := map[string]string{}
	for _, k := range c.AllKeys() {
		normalizedKey := prefix + strings.ToUpper(k)
		normalizedKey = strings.ReplaceAll(normalizedKey, ".", "_")
		envKeyToConfigKey[normalizedKey] = k
	}

	for _, e := range env {
		// Extract key and value from env
		envKey, envValue, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}

		// Check if the configuration has a matching key
		key, ok := envKeyToConfigKey[strings.ToUpper(envKey)]
		if !ok {
			continue
		}

		// Update the configuration value
		if err := c.SetFromCLIArgs(key, envValue); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
