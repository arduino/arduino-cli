// This file is part of arduino-cli.
//
// Copyright 2021 ARDUINO SA (http://www.arduino.cc/)
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

package executils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProcessWithinContext(t *testing.T) {
	// Build `delay` helper inside testdata/delay
	builder, err := NewProcess(nil, "go", "build")
	require.NoError(t, err)
	builder.SetDir("testdata/delay")
	require.NoError(t, builder.Run())

	// Run delay and test if the process is terminated correctly due to context
	process, err := NewProcess(nil, "testdata/delay/delay")
	require.NoError(t, err)
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	err = process.RunWithinContext(ctx)
	require.Error(t, err)
	require.Less(t, time.Since(start), 500*time.Millisecond)
	cancel()
}
