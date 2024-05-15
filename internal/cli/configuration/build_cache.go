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

package configuration

import "time"

// GetCompilationsBeforeBuildCachePurge returns the number of compilations before the build cache is purged.
func (s *Settings) GetCompilationsBeforeBuildCachePurge() uint {
	if res, ok, _ := s.GetUintOk("build_cache.compilations_before_purge"); ok {
		return res
	}
	return s.Defaults.GetUint("build_cache.compilations_before_purge")
}

// GetBuildCacheTTL returns the time-to-live of the build cache (i.e. the minimum age to wait before purging the cache).
func (s *Settings) GetBuildCacheTTL() time.Duration {
	if res, ok, _ := s.GetDurationOk("build_cache.ttl"); ok {
		return res
	}
	return s.Defaults.GetDuration("build_cache.ttl")
}
