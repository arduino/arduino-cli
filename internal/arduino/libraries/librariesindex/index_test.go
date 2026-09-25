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

package librariesindex

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	semver "go.bug.st/relaxed-semver"
)

func releaseStrings(releases []*ReleaseReference) []string {
	res := make([]string, len(releases))
	for i, r := range releases {
		res[i] = r.String()
	}
	return res
}

func TestIndexer(t *testing.T) {
	// A missing index file is an error at load time.
	missing, err := LoadIndex(paths.New("testdata/inexistent"))
	require.Error(t, err)
	require.Nil(t, missing)

	// The same holds for an invalid/corrupted index file.
	invalid, err := LoadIndex(paths.New("testdata/invalid.json"))
	require.NoError(t, err)
	require.NotNil(t, invalid)
	_, err = invalid.FindRelease("RTCZero", nil)
	require.Error(t, err)

	index, err := LoadIndex(paths.New("testdata/library_index.json"))
	require.NoError(t, err)

	count := 0
	for range index.Libraries() {
		count++
	}
	require.Equal(t, 4124, count, "parsed libraries count")

	alp := index.FindIndexedLibrary(&libraries.Library{Name: "Arduino Low Power"})
	require.NotNil(t, alp)
	require.Equal(t, 5, len(alp.Releases))
	require.Equal(t, "Arduino Low Power@1.2.2", alp.Latest.String())
	require.Len(t, alp.Latest.Dependencies, 1)
	require.Equal(t, "RTCZero", alp.Latest.Dependencies[0].GetName())
	require.Equal(t, "", alp.Latest.Dependencies[0].GetConstraint().String())
	require.Equal(t, "[1.0.0 1.1.0 1.2.0 1.2.1 1.2.2]", fmt.Sprintf("%v", alp.Versions()))

	rtc100, err := index.FindRelease("RTCZero", semver.MustParse("1.0.0"))
	require.NoError(t, err)
	require.NotNil(t, rtc100)
	require.Equal(t, "RTCZero@1.0.0", rtc100.String())

	rtcLatest, err := index.FindRelease("RTCZero", nil)
	require.NoError(t, err)
	require.NotNil(t, rtcLatest)
	require.Equal(t, "RTCZero@1.6.0", rtcLatest.String())

	rtcInexistent, err := index.FindRelease("RTCZero", semver.MustParse("0.0.0-blah"))
	require.Error(t, err)
	require.Nil(t, rtcInexistent)

	rtcInexistent, err = index.FindRelease("RTCZero-blah", nil)
	require.Error(t, err)
	require.Nil(t, rtcInexistent)

	rtc := index.FindIndexedLibrary(&libraries.Library{Name: "RTCZero"})
	require.NotNil(t, rtc)
	require.Equal(t, "RTCZero", rtc.Name)

	rtcUpdate := index.FindLibraryUpdates(&libraries.Library{Name: "RTCZero", Version: semver.MustParse("1.0.0")})[0]
	require.NotNil(t, rtcUpdate)
	require.Equal(t, "RTCZero@1.6.0", rtcUpdate.String())

	rtcUpdateNoVersion := index.FindLibraryUpdates(&libraries.Library{Name: "RTCZero", Version: nil})[0]
	require.NotNil(t, rtcUpdateNoVersion)
	require.Equal(t, "RTCZero@1.6.0", rtcUpdateNoVersion.String())

	rtcNoUpdate := index.FindLibraryUpdates(&libraries.Library{Name: "RTCZero", Version: semver.MustParse("3.0.0")})[0]
	require.Nil(t, rtcNoUpdate)

	rtcInexistent2 := index.FindLibraryUpdates(&libraries.Library{Name: "RTCZero-blah", Version: semver.MustParse("1.0.0")})[0]
	require.Nil(t, rtcInexistent2)

	resolve1 := index.ResolveDependencies(alp.Releases["1.2.1"], nil)
	require.Len(t, resolve1, 2)
	require.Contains(t, releaseStrings(resolve1), "Arduino Low Power@1.2.1")
	require.Contains(t, releaseStrings(resolve1), "RTCZero@1.6.0")

	oauth010, err := index.FindRelease("Arduino_OAuth", semver.MustParse("0.1.0"))
	require.NoError(t, err)
	require.NotNil(t, oauth010)
	require.Equal(t, "Arduino_OAuth@0.1.0", oauth010.String())
	eccx135, err := index.FindRelease("ArduinoECCX08", semver.MustParse("1.3.5"))
	require.NoError(t, err)
	require.NotNil(t, eccx135)
	require.Equal(t, "ArduinoECCX08@1.3.5", eccx135.String())
	bear172, err := index.FindRelease("ArduinoBearSSL", semver.MustParse("1.7.2"))
	require.NoError(t, err)
	require.NotNil(t, bear172)
	require.Equal(t, "ArduinoBearSSL@1.7.2", bear172.String())
	http040, err := index.FindRelease("ArduinoHttpClient", semver.MustParse("0.4.0"))
	require.NoError(t, err)
	require.NotNil(t, http040)
	require.Equal(t, "ArduinoHttpClient@0.4.0", http040.String())

	resolve2 := index.ResolveDependencies(oauth010, nil)
	require.Len(t, resolve2, 4)
	require.Contains(t, releaseStrings(resolve2), "Arduino_OAuth@0.1.0")
	require.Contains(t, releaseStrings(resolve2), "ArduinoECCX08@1.3.5")
	require.Contains(t, releaseStrings(resolve2), "ArduinoBearSSL@1.7.2")
	require.Contains(t, releaseStrings(resolve2), "ArduinoHttpClient@0.4.0")
}

// These benchmarks use only the public API that also exists on the master
// (full-load) implementation, and include LoadIndex in the timed loop, so they
// can be run unchanged on either implementation to compare one-shot
// "load + query" behaviour (like a single CLI invocation).

func benchLoadAndFindRelease(tb testing.TB) {
	idx, _ := LoadIndex(paths.New("testdata/library_index.json"))
	if r, _ := idx.FindRelease("Arduino_OAuth", semver.MustParse("0.1.0")); r == nil {
		tb.Fatal("release not found")
	}
}

func benchLoadAndResolve(tb testing.TB) {
	idx, _ := LoadIndex(paths.New("testdata/library_index.json"))
	r, _ := idx.FindRelease("Arduino_OAuth", semver.MustParse("0.1.0"))
	if deps := idx.ResolveDependencies(r, nil); len(deps) != 4 {
		tb.Fatalf("expected 4 deps, got %d", len(deps))
	}
}

// reportPeakMemory runs fn b.N times while sampling the live heap, and reports
// the peak HeapInuse as a "peakMem-MiB" metric. It uses runtime.ReadMemStats
// (portable, unlike ru_maxrss) so it can live alongside the cross-platform
// tests. ReadMemStats stops the world, so treat the memory metric as the truth
// here and read timings from the plain benchmarks.
func reportPeakMemory(b *testing.B, fn func()) {
	var peak uint64
	stop, done := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		var m runtime.MemStats
		t := time.NewTicker(time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				if runtime.ReadMemStats(&m); m.HeapInuse > peak {
					peak = m.HeapInuse
				}
			}
		}
	}()
	b.ResetTimer()
	for range b.N {
		fn()
	}
	b.StopTimer()
	close(stop)
	<-done
	b.ReportMetric(float64(peak)/(1<<20), "peakMem-MiB")
}

func BenchmarkLoadAndFindRelease(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		benchLoadAndFindRelease(b)
	}
}

func BenchmarkLoadAndResolve(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		benchLoadAndResolve(b)
	}
}

func BenchmarkLoadAndFindReleaseMem(b *testing.B) {
	reportPeakMemory(b, func() { benchLoadAndFindRelease(b) })
}

func BenchmarkLoadAndResolveMem(b *testing.B) {
	reportPeakMemory(b, func() { benchLoadAndResolve(b) })
}
