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

package discovery

import (
	"io"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/executils"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryStdioHandling(t *testing.T) {
	// Build `cat` helper inside testdata/cat
	builder, err := executils.NewProcess("go", "build")
	require.NoError(t, err)
	builder.SetDir("testdata/cat")
	require.NoError(t, builder.Run())

	// Run cat and test if streaming json works as expected
	disc, err := New("test", "testdata/cat/cat") // copy stdin to stdout
	require.NoError(t, err)

	err = disc.runProcess()
	require.NoError(t, err)

	_, err = disc.outgoingCommandsPipe.Write([]byte(`{ "eventType":`)) // send partial JSON
	require.NoError(t, err)
	msg, err := disc.waitMessage(time.Millisecond * 100)
	require.Error(t, err)
	require.Nil(t, msg)

	_, err = disc.outgoingCommandsPipe.Write([]byte(`"ev1" }{ `)) // complete previous json and start another one
	require.NoError(t, err)

	msg, err = disc.waitMessage(time.Millisecond * 100)
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "ev1", msg.EventType)

	msg, err = disc.waitMessage(time.Millisecond * 100)
	require.Error(t, err)
	require.Nil(t, msg)

	_, err = disc.outgoingCommandsPipe.Write([]byte(`"eventType":"ev2" }`)) // complete previous json
	require.NoError(t, err)

	msg, err = disc.waitMessage(time.Millisecond * 100)
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.Equal(t, "ev2", msg.EventType)

	require.Equal(t, disc.State(), Alive)

	err = disc.outgoingCommandsPipe.(io.ReadCloser).Close()
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 100)

	require.Equal(t, disc.State(), Dead)
}
