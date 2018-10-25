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
	"errors"
	"fmt"
	"time"

	"go.bug.st/downloader"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// TextFormatter represents a Formatter for a text console
type TextFormatter struct{}

// Format implements Formatter interface
func (tp *TextFormatter) Format(msg interface{}) (string, error) {
	if msg == nil {
		return "<nil>", nil
	}
	if str, ok := msg.(string); ok {
		return str, nil
	}
	str, ok := msg.(fmt.Stringer)
	if !ok {
		return "", errors.New("object can't be formatted as text")
	}
	return str.String(), nil
}

// DownloadProgressBar implements Formatter interface
func (tp *TextFormatter) DownloadProgressBar(d *downloader.Downloader, prefix string) {
	t := time.NewTicker(250 * time.Millisecond)
	defer t.Stop()

	bar := pb.StartNew(int(d.Size()))
	bar.SetUnits(pb.U_BYTES)
	bar.Prefix(prefix)
	update := func(curr int64) {
		bar.Set(int(curr))
	}
	d.RunAndPoll(update, 250*time.Millisecond)
	bar.FinishPrintOver(prefix + " downloaded")
}
