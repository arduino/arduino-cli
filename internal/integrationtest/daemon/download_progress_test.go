// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package daemon_test

import (
	"testing"

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/stretchr/testify/require"
)

// DownloadProgressAnalyzer analyzes DownloadProgress messages for consistency
type DownloadProgressAnalyzer struct {
	t               *testing.T
	ongoingDownload string
	Results         map[string]*commands.DownloadProgressEnd
}

// NewDownloadProgressAnalyzer creates a new DownloadProgressAnalyzer
func NewDownloadProgressAnalyzer(t *testing.T) *DownloadProgressAnalyzer {
	return &DownloadProgressAnalyzer{
		t:       t,
		Results: map[string]*commands.DownloadProgressEnd{},
	}
}

// Process the given DownloadProgress message. If inconsistencies are detected the
// test will fail.
func (a *DownloadProgressAnalyzer) Process(progress *commands.DownloadProgress) {
	if progress == nil {
		return
	}
	if start := progress.GetStart(); start != nil {
		require.Empty(a.t, a.ongoingDownload, "DownloadProgressStart: started a download without 'completing' the previous one")
		a.ongoingDownload = start.GetUrl()
	} else if update := progress.GetUpdate(); update != nil {
		require.NotEmpty(a.t, a.ongoingDownload, "DownloadProgressUpdate: received update, but the download is not yet started...")
	} else if end := progress.GetEnd(); end != nil {
		require.NotEmpty(a.t, a.ongoingDownload, "DownloadProgress: received a 'completed' notification but never initiated a download")
		a.Results[a.ongoingDownload] = end
		a.ongoingDownload = ""
	} else {
		require.FailNow(a.t, "DownloadProgress: received an empty DownloadProgress (without Start, Update or End)")
	}
}
