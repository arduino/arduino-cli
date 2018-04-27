/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
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
