// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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
package librariesmanager

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/stretchr/testify/require"
)

func TestLibrariesBuilderScanCloneRescan(t *testing.T) {
	lmb := NewBuilder()
	lmb.libraries["testLibA"] = libraries.List{}
	lmb.libraries["testLibB"] = libraries.List{}
	lm, warns := lmb.Build()
	require.Empty(t, warns)
	require.Len(t, lm.libraries, 2)

	// Cloning should keep existing libraries
	lm2, warns2 := lm.Clone().Build()
	require.Empty(t, warns2)
	require.Len(t, lm2.libraries, 2)

	// Full rescan should update libs
	{
		lmi2, release := lm2.NewInstaller()
		lmi2.RescanLibraries()
		release()
	}
	require.Len(t, lm.libraries, 2) // Ensure deep-coping worked as expected...
	require.Len(t, lm2.libraries, 0)
}
