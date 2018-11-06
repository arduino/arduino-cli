//
// Copyright 2018 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package cleanup

import (
	"context"
	"os"
	"os/signal"
)

// InterruptableContext adds to a context the capability to be interrupted by the os.Interrupt signal.
func InterruptableContext(inCtx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(inCtx)
	go func() {
		ctrlC := make(chan os.Signal, 1)
		signal.Notify(ctrlC, os.Interrupt)
		select {
		case <-ctx.Done():
			break // used to show test coverage
		case <-ctrlC:
			cancel()
		}
		signal.Stop(ctrlC)
		close(ctrlC)
	}()
	return ctx, cancel
}
