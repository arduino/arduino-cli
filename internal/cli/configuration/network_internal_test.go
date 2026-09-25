// This file is part of arduino-cli.
//
// Copyright (c) Arduino s.r.l. and/or its affiliated companies
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

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsGitHubArtifactURL(t *testing.T) {
	allowed := []string{
		"https://api.github.com/repos/owner/repo/actions/artifacts/12345/zip",
		"https://api.github.com/repos/owner/repo/actions/artifacts/99/json",
		"https://api.github.com/repos/my-org/my-repo/actions/artifacts/1/zip",
	}
	for _, raw := range allowed {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		require.True(t, isGitHubArtifactURL(u), "should be allowed: %s", raw)
	}

	rejected := []string{
		// Wrong scheme
		"http://api.github.com/repos/owner/repo/actions/artifacts/12345/zip",
		// Wrong host
		"https://github.com/repos/owner/repo/actions/artifacts/12345/zip",
		"https://raw.githubusercontent.com/repos/owner/repo/actions/artifacts/12345/zip",
		// Wrong path
		"https://api.github.com/repos/owner/repo",
		"https://api.github.com/repos/owner/repo/actions/artifacts",
		"https://api.github.com/user",
	}
	for _, raw := range rejected {
		u, err := url.Parse(raw)
		require.NoError(t, err)
		require.False(t, isGitHubArtifactURL(u), "should be rejected: %s", raw)
	}
}
