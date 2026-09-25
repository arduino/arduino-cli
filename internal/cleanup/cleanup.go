// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package cleanup

import (
	"context"
	"os"
	"os/signal"
)

// InterruptableContext adds to a context the capability to be interrupted by
// the os.Interrupt signal. The signal handler is removed as soon as the
// context is done, so a repeated CTRL-C terminates the program.
func InterruptableContext(inCtx context.Context) (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(inCtx, os.Interrupt)
	go func() {
		<-ctx.Done()
		stop()
	}()
	return ctx, stop
}
