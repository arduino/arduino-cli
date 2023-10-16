// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package arguments

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// AddKeyValuePFlag adds a flag to the command that accepts a (possibly repeated) key=value pair.
func AddKeyValuePFlag(cmd *cobra.Command, field *map[string]string, name, shorthand string, value []string, usage string) {
	cmd.Flags().VarP(newKVArrayValue(value, field), name, shorthand, usage)
}

type kvArrayValue struct {
	value   *map[string]string
	changed bool
}

func newKVArrayValue(val []string, p *map[string]string) *kvArrayValue {
	ssv := &kvArrayValue{
		value: p,
	}
	for _, v := range val {
		ssv.Set(v)
	}
	ssv.changed = false
	return ssv
}

func (s *kvArrayValue) Set(arg string) error {
	split := strings.SplitN(arg, "=", 2)
	if len(split) != 2 {
		return errors.New("required format is 'key=value'")
	}
	k, v := split[0], split[1]
	if k == "" {
		return errors.New("key cannot be empty")
	}
	if !s.changed {
		// Remove the default value
		*s.value = make(map[string]string)
		s.changed = true
	}
	if _, ok := (*s.value)[k]; ok {
		return errors.New("duplicate key: " + k)
	}
	(*s.value)[k] = v
	return nil
}

func (s *kvArrayValue) Type() string {
	return "key=value"
}

func (s *kvArrayValue) String() string {
	if len(*s.value) == 0 {
		return ""
	}
	res := "["
	for k, v := range *s.value {
		res += fmt.Sprintf("%s=%s, ", k, v)
	}
	return res[:len(res)-2] + "]"
}
