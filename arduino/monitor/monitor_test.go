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

package monitor

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestDummyMonitor(t *testing.T) {
	// Build `dummy-monitor` helper inside testdata/dummy-monitor
	testDataDir, err := paths.New("testdata").Abs()
	require.NoError(t, err)
	builder, err := executils.NewProcess("go", "install", "github.com/arduino/pluggable-monitor-protocol-handler/dummy-monitor@main")
	fmt.Println(testDataDir.String())
	env := os.Environ()
	env = append(env, "GOBIN="+testDataDir.String())
	builder.SetEnvironment(env)
	require.NoError(t, err)
	builder.SetDir(".")
	require.NoError(t, builder.Run())

	// Run dummy-monitor and test if everything is working as expected
	mon, err := New("dummy", "testdata/dummy-monitor")
	require.NoError(t, err)
	err = mon.Run()
	require.NoError(t, err)

	res, err := mon.Describe()
	require.NoError(t, err)
	fmt.Println(res)

	err = mon.Configure("sped", "38400")
	require.Error(t, err)
	err = mon.Configure("speed", "384")
	require.Error(t, err)
	err = mon.Configure("speed", "38400")
	require.NoError(t, err)

	port := &discovery.Port{Address: "/dev/ttyACM0", Protocol: "test"}
	rw, err := mon.Open(port.ToRPC())
	require.NoError(t, err)

	// Double open -> error: port already opened
	_, err = mon.Open(port.ToRPC())
	require.Error(t, err)

	// Write "TEST"
	n, err := rw.Write([]byte("TEST"))
	require.NoError(t, err)
	require.Equal(t, 4, n)

	completed := int32(0)
	go func() {
		buff := [1024]byte{}
		// Receive "TEST" echoed back
		n, err = rw.Read(buff[:])
		require.NoError(t, err)
		require.Equal(t, 4, n)
		require.Equal(t, "TEST", string(buff[:4]))

		// Block on read until the port is closed
		n, err = rw.Read(buff[:])
		require.ErrorIs(t, err, io.EOF)
		atomic.StoreInt32(&completed, 1) // notify completion
	}()

	time.Sleep(100 * time.Millisecond)
	err = mon.Close()
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, int32(1), atomic.LoadInt32(&completed))

	rw, err = mon.Open(port.ToRPC())
	require.NoError(t, err)
	n, err = rw.Write([]byte("TEST"))
	require.NoError(t, err)
	require.Equal(t, 4, n)

	// Close TCP connection
	err = rw.(net.Conn).Close()
	require.NoError(t, err)
	time.Sleep((100 * time.Millisecond))

	err = mon.Close()
	require.Error(t, err) // should be port already closed
}
