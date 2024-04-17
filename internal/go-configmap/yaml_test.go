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
	"fmt"
	"testing"

	"github.com/arduino/arduino-cli/internal/go-configmap"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestYaml(t *testing.T) {
	c := configmap.New()
	c.Set("foo", "bar")
	c.Set("fooz.bar", "baz")
	c.Set("answer", 42)
	require.Equal(t, "bar", c.Get("foo"))
	require.Equal(t, "baz", c.Get("fooz.bar"))
	require.Equal(t, 42, c.Get("answer"))

	y1, err := yaml.Marshal(c)
	require.NoError(t, err)
	fmt.Println(string(y1))

	d := configmap.New()
	err = yaml.Unmarshal(y1, d)
	require.NoError(t, err)
	require.Equal(t, "baz", d.Get("fooz.bar"))

	y2, err := yaml.Marshal(d)
	require.NoError(t, err)
	require.Equal(t, string(y1), string(y2))
}
