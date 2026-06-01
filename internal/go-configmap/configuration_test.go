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

package configmap_test

import (
	"encoding/json"
	"testing"

	"github.com/arduino/arduino-cli/internal/go-configmap"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfiguration(t *testing.T) {
	c := configmap.New()
	c.Set("foo", "bar")
	c.Set("fooz.bar", "baz")
	c.Set("answer", 42)
	require.Equal(t, "bar", c.Get("foo"))
	require.Equal(t, "baz", c.Get("fooz.bar"))
	require.Equal(t, 42, c.Get("answer"))

	yml, err := yaml.Marshal(c)
	require.NoError(t, err)

	d := configmap.New()
	err = yaml.Unmarshal(yml, &d)
	require.NoError(t, err)

	yml2, err := yaml.Marshal(d)
	require.NoError(t, err)
	require.Equal(t, string(yml), string(yml2))

	d.Set("fooz.abc", "def")
	d.Set("fooz.cde", "fgh")
	require.Equal(t, "def", d.Get("fooz.abc"))
	require.Equal(t, "fgh", d.Get("fooz.cde"))
	d.Delete("fooz.abc")
	require.Nil(t, d.Get("fooz.abc"))
	require.Equal(t, "fgh", d.Get("fooz.cde"))
	d.Delete("fooz")
	require.Nil(t, d.Get("fooz.cde"))
}

func TestYAMLCleanUpOfZeroValues(t *testing.T) {
	inYml := []byte(`
foo: bar
directories:
  builtins: {}
`)
	c := configmap.New()
	outYml1, err := yaml.Marshal(c)
	require.NoError(t, err)
	require.Equal(t, "{}\n", string(outYml1))

	err = yaml.Unmarshal(inYml, &c)
	require.NoError(t, err)

	outYml2, err := yaml.Marshal(c)
	require.NoError(t, err)
	require.Equal(t, "foo: bar\n", string(outYml2))
}

func TestApplyEnvVars(t *testing.T) {
	c := configmap.New()
	c.Set("foo", "bar")
	c.Set("fooz.bar", "baz")
	c.Set("answer", 42)
	c.InjectEnvVars([]string{"APP_FOO=app-bar", "APP_FOOZ_BAR=app-baz"}, "APP")
	require.Equal(t, "app-bar", c.Get("foo"))
	require.Equal(t, "app-baz", c.Get("fooz.bar"))
	require.Equal(t, 42, c.Get("answer"))
}

func TestMerge(t *testing.T) {
	c := configmap.New()
	c.Set("foo", "bar")
	c.Set("fooz.bar", "baz")
	c.Set("answer", 42)

	d := configmap.New()
	d.Set("answer", 24)
	require.NoError(t, c.Merge(d))
	require.Equal(t, "bar", c.Get("foo"))
	require.Equal(t, "baz", c.Get("fooz.bar"))
	require.Equal(t, 24, c.Get("answer"))

	e := configmap.New()
	e.Set("fooz.bar", "barz")
	require.NoError(t, c.Merge(e))
	require.Equal(t, "bar", c.Get("foo"))
	require.Equal(t, "barz", c.Get("fooz.bar"))
	require.Equal(t, 24, c.Get("answer"))

	f := configmap.New()
	f.Set("fooz.bar", 10)
	require.EqualError(t, c.Merge(f), "invalid types for key 'bar': got string but want int")

	g := configmap.New()
	g.Set("fooz.bart", "baz")
	require.EqualError(t, c.Merge(g), "target key do not exist: 'bart'")
}

func TestAllKeys(t *testing.T) {
	{
		c := configmap.New()
		c.Set("foo", "bar")
		c.Set("fooz.bar", "baz")
		c.Set("answer", 42)
		require.ElementsMatch(t, []string{"foo", "fooz.bar", "answer"}, c.AllKeys())
	}
	{
		inYml := []byte(`
foo: bar
dir:
  a: yes
  b: no
  c: {}
  d:
    - 1
    - 2
`)
		c := configmap.New()
		err := yaml.Unmarshal(inYml, &c)
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"foo", "dir.a", "dir.b", "dir.d"}, c.AllKeys())
	}
}

func TestSchema(t *testing.T) {
	c := configmap.New()
	c.SetKeyTypeSchema("string", "")
	c.SetKeyTypeSchema("int", 15)
	c.SetKeyTypeSchema("obj.string", "")
	c.SetKeyTypeSchema("obj.int", 15)
	c.SetKeyTypeSchema("uint", uint(15))
	c.SetKeyTypeSchema("obj.uint", uint(15))
	c.SetKeyTypeSchema("array", []string{})
	c.SetKeyTypeSchema("obj.array", []string{})

	// Set array of string
	require.NoError(t, c.Set("array", []string{"abc", "def"}))
	require.NoError(t, c.Set("obj.array", []string{"abc", "def"}))
	require.Equal(t, []string{"abc", "def"}, c.Get("array"))
	require.Equal(t, []string{"abc", "def"}, c.Get("obj.array"))
	// Set array of string with array of any
	require.NoError(t, c.Set("array", []any{"abc", "def"}))
	require.NoError(t, c.Set("obj.array", []any{"abc", "def"}))
	require.Equal(t, []string{"abc", "def"}, c.Get("array"))
	require.Equal(t, []string{"abc", "def"}, c.Get("obj.array"))
	// Set array of string with array of int
	require.EqualError(t, c.Set("array", []any{"abc", 123}), "invalid type for key 'array': invalid conversion, got int but want string")
	require.EqualError(t, c.Set("obj.array", []any{"abc", 123}), "invalid type for key 'obj.array': invalid conversion, got int but want string")

	// Set string
	require.NoError(t, c.Set("string", "abc"))
	require.NoError(t, c.Set("obj.string", "abc"))
	require.Equal(t, "abc", c.Get("string"))
	require.Equal(t, "abc", c.Get("obj.string"))
	// Set string with int
	require.EqualError(t, c.Set("string", 123), "invalid type for key 'string': invalid conversion, got int but want string")
	require.EqualError(t, c.Set("obj.string", 123), "invalid type for key 'obj.string': invalid conversion, got int but want string")

	// Set int
	require.NoError(t, c.Set("int", 123))
	require.NoError(t, c.Set("obj.int", 123))
	require.Equal(t, 123, c.Get("int"))
	require.Equal(t, 123, c.Get("obj.int"))
	// Set int with string
	require.EqualError(t, c.Set("int", "abc"), "invalid type for key 'int': invalid conversion, got string but want int")
	require.EqualError(t, c.Set("obj.int", "abc"), "invalid type for key 'obj.int': invalid conversion, got string but want int")

	// Set uint
	require.NoError(t, c.Set("uint", uint(234)))
	require.NoError(t, c.Set("obj.uint", uint(234)))
	require.Equal(t, uint(234), c.Get("uint"))
	require.Equal(t, uint(234), c.Get("obj.uint"))
	// Set uint using int
	require.NoError(t, c.Set("uint", 345))
	require.NoError(t, c.Set("obj.uint", 345))
	require.Equal(t, uint(345), c.Get("uint"))
	require.Equal(t, uint(345), c.Get("obj.uint"))
	// Set uint using float
	require.NoError(t, c.Set("uint", 456.0))
	require.NoError(t, c.Set("obj.uint", 456.0))
	require.Equal(t, uint(456), c.Get("uint"))
	require.Equal(t, uint(456), c.Get("obj.uint"))
	// Set uint using string
	require.EqualError(t, c.Set("uint", "567"), "invalid type for key 'uint': invalid conversion, got string but want uint")
	require.EqualError(t, c.Set("obj.uint", "567"), "invalid type for key 'obj.uint': invalid conversion, got string but want uint")
	require.Equal(t, uint(456), c.Get("uint"))
	require.Equal(t, uint(456), c.Get("obj.uint"))

	json1 := []byte(`{"string":"abcd","int":1234,"obj":{"string":"abcd","int":1234}}`)
	require.NoError(t, json.Unmarshal(json1, &c))
	require.Equal(t, "abcd", c.Get("string"))
	require.Equal(t, 1234, c.Get("int"))
	require.Equal(t, "abcd", c.Get("obj.string"))
	require.Equal(t, 1234, c.Get("obj.int"))

	json2 := []byte(`{"string":123,"int":123,"obj":{"string":"abc","int":123}}`)
	require.EqualError(t, json.Unmarshal(json2, &c), "invalid type for key 'string': invalid conversion, got float64 but want string")
	json3 := []byte(`{"string":"avc","int":123,"obj":{"string":123,"int":123}}`)
	require.EqualError(t, json.Unmarshal(json3, &c), "invalid type for key 'obj.string': invalid conversion, got float64 but want string")
}
