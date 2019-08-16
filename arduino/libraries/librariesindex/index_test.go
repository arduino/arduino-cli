/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package librariesindex

import (
	"fmt"
	"testing"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

func TestIndexer(t *testing.T) {
	fail1, err := LoadIndex(paths.New("testdata/inexistent"))
	require.Error(t, err)
	require.Nil(t, fail1)

	fail2, err := LoadIndex(paths.New("testdata/invalid.json"))
	require.Error(t, err)
	require.Nil(t, fail2)

	index, err := LoadIndex(paths.New("testdata/library_index.json"))
	require.NoError(t, err)
	require.Equal(t, 2380, len(index.Libraries), "parsed libraries count")

	alp := index.Libraries["Arduino Low Power"]
	require.NotNil(t, alp)
	require.Equal(t, 4, len(alp.Releases))
	require.Equal(t, "Arduino Low Power@1.2.1", alp.Latest.String())
	require.Len(t, alp.Latest.Dependencies, 1)
	require.Equal(t, "RTCZero", alp.Latest.Dependencies[0].GetName())
	require.Equal(t, "", alp.Latest.Dependencies[0].GetConstraint().String())
	require.Equal(t, "[1.0.0 1.1.0 1.2.0 1.2.1]", fmt.Sprintf("%v", alp.Versions()))

	rtc100ref := &Reference{Name: "RTCZero", Version: semver.MustParse("1.0.0")}
	require.Equal(t, "RTCZero@1.0.0", rtc100ref.String())
	rtc100 := index.FindRelease(rtc100ref)
	require.NotNil(t, rtc100)
	require.Equal(t, "RTCZero@1.0.0", rtc100.String())

	rtcLatestRef := &Reference{Name: "RTCZero"}
	require.Equal(t, "RTCZero", rtcLatestRef.String())
	rtcLatest := index.FindRelease(rtcLatestRef)
	require.NotNil(t, rtcLatest)
	require.Equal(t, "RTCZero@1.6.0", rtcLatest.String())

	rtcInexistent := index.FindRelease(&Reference{
		Name:    "RTCZero",
		Version: semver.MustParse("0.0.0-blah"),
	})
	require.Nil(t, rtcInexistent)

	rtcInexistent = index.FindRelease(&Reference{
		Name: "RTCZero-blah",
	})
	require.Nil(t, rtcInexistent)

	rtc := index.FindIndexedLibrary(&libraries.Library{Name: "RTCZero"})
	require.NotNil(t, rtc)
	require.Equal(t, "RTCZero", rtc.Name)

	rtcUpdate := index.FindLibraryUpdate(&libraries.Library{Name: "RTCZero", Version: semver.MustParse("1.0.0")})
	require.NotNil(t, rtcUpdate)
	require.Equal(t, "RTCZero@1.6.0", rtcUpdate.String())

	rtcNoUpdate := index.FindLibraryUpdate(&libraries.Library{Name: "RTCZero", Version: semver.MustParse("3.0.0")})
	require.Nil(t, rtcNoUpdate)

	rtcInexistent2 := index.FindLibraryUpdate(&libraries.Library{Name: "RTCZero-blah", Version: semver.MustParse("1.0.0")})
	require.Nil(t, rtcInexistent2)
}
