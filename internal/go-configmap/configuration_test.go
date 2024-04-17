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
	"fmt"
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
	fmt.Println(string(yml))

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
	require.Error(t, c.Merge(f))
	fmt.Println(c.Merge(f))

	g := configmap.New()
	g.Set("fooz.bart", "baz")
	require.Error(t, c.Merge(g))
	fmt.Println(c.Merge(g))
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
	require.NoError(t, c.Set("string", "abc"))
	require.Error(t, c.Set("string", 123))
	require.NoError(t, c.Set("int", 123))
	require.Error(t, c.Set("int", "abc"))
	require.Equal(t, "abc", c.Get("string"))

	json1 := []byte(`{"string":"abc","int":123,"obj":{"string":"abc","int":123}}`)
	require.NoError(t, json.Unmarshal(json1, &c))
	require.Equal(t, "abc", c.Get("string"))
	require.Equal(t, 123, c.Get("int"))

	json2 := []byte(`{"string":123,"int":123,"obj":{"string":"abc","int":123}}`)
	require.Error(t, json.Unmarshal(json2, &c))
	json3 := []byte(`{"string":"avc","int":123,"obj":{"string":123,"int":123}}`)
	require.Error(t, json.Unmarshal(json3, &c))
}
