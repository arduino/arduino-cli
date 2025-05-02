// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package libraries

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDependencyExtract(t *testing.T) {
	check := func(depDefinition string, name []string, ver []string) {
		dep, err := extractDependenciesList(depDefinition)
		require.NoError(t, err)
		require.NotNil(t, dep)
		require.Len(t, dep, len(name))
		for i := range name {
			require.Equal(t, name[i], dep[i].Name, depDefinition)
			require.Equal(t, ver[i], dep[i].VersionConstraint.String(), depDefinition)
		}
	}
	invalid := func(depends string) {
		dep, err := extractDependenciesList(depends)
		require.Nil(t, dep)
		require.Error(t, err)
	}
	check("ciao", []string{"ciao"}, []string{""})
	check("MyLib (>1.2.3)", []string{"MyLib"}, []string{">1.2.3"})
	check("MyLib (>=1.2.3)", []string{"MyLib"}, []string{">=1.2.3"})
	check("MyLib (<1.2.3)", []string{"MyLib"}, []string{"<1.2.3"})
	check("MyLib (<=1.2.3)", []string{"MyLib"}, []string{"<=1.2.3"})
	check("MyLib (!=1.2.3)", []string{"MyLib"}, []string{"!(=1.2.3)"})
	check("MyLib (>1.0.0 && <2.1.0)", []string{"MyLib"}, []string{"(>1.0.0 && <2.1.0)"})
	check("MyLib (<1.0.0 || >2.0.0)", []string{"MyLib"}, []string{"(<1.0.0 || >2.0.0)"})
	check("MyLib ((>0.1.0 && <2.0.0) || >2.1.0)", []string{"MyLib"}, []string{"((>0.1.0 && <2.0.0) || >2.1.0)"})
	check("MyLib ()", []string{"MyLib"}, []string{""})
	check("MyLib (>=1.2.3),AnotherLib, YetAnotherLib (=1.0.0)",
		[]string{"MyLib", "AnotherLib", "YetAnotherLib"},
		[]string{">=1.2.3", "", "=1.0.0"})
	invalid("MyLib,,AnotherLib")
	invalid("(MyLib)")
	invalid("MyLib(=1.2.3)")
	check("Arduino Uno WiFi Dev Ed Library, LoRa Node (^2.1.2)",
		[]string{"Arduino Uno WiFi Dev Ed Library", "LoRa Node"},
		[]string{"", "^2.1.2"})
	check("Arduino Uno WiFi Dev Ed Library   ,   LoRa Node    (^2.1.2)",
		[]string{"Arduino Uno WiFi Dev Ed Library", "LoRa Node"},
		[]string{"", "^2.1.2"})
	check("Arduino_OAuth, ArduinoHttpClient (<0.3.0), NonExistentLib",
		[]string{"Arduino_OAuth", "ArduinoHttpClient", "NonExistentLib"},
		[]string{"", "<0.3.0", ""})
	check("", []string{}, []string{})
}
