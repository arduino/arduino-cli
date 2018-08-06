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

package formatter

import (
	"fmt"
	"time"

	"github.com/cavaliercoder/grab"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// TextFormatter represents a Formatter for a text console
type TextFormatter struct{}

// Format implements Formatter interface
func (tp *TextFormatter) Format(msg interface{}) (string, error) {
	return fmt.Sprintf("%s", msg), nil
}

// DownloadProgressBar implements Formatter interface
func (tp *TextFormatter) DownloadProgressBar(resp *grab.Response, prefix string) {
	t := time.NewTicker(250 * time.Millisecond)
	defer t.Stop()

	bar := pb.StartNew(int(resp.Size))
	bar.SetUnits(pb.U_BYTES)
	bar.Prefix(prefix)
	for {
		select {
		case <-t.C:
			bar.Set(int(resp.BytesComplete()))
		case <-resp.Done:
			bar.ShowCounters = false
			bar.ShowPercent = false
			bar.ShowFinalTime = false
			bar.ShowBar = false
			bar.Prefix(prefix + " downloaded")
			bar.Set(int(resp.BytesComplete()))
			bar.Finish()
			return
		}
	}
}
