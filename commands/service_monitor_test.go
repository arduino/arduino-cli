// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package commands

import (
	"testing"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

type fakeConfigurer struct{ calls [][2]string }

func (f *fakeConfigurer) Configure(k, v string) error {
	f.calls = append(f.calls, [2]string{k, v})
	return nil
}

func haveCall(calls [][2]string, k, v string) bool {
	for _, kv := range calls {
		if kv[0] == k && kv[1] == v {
			return true
		}
	}
	return false
}

// Test that we correctly read all buffer_config fields from the gRPC request
// and emit the expected CONFIGURE _buffer.* key/value pairs.
func Test_applyBufferConfig_AllFields(t *testing.T) {
	f := &fakeConfigurer{}
	cfg := &rpc.MonitorBufferConfig{
		HighWaterMarkBytes: 64,
		FlushIntervalMs:    16,
		LineBuffering:      true,
		FlushQueueCapacity: 256,
		OverflowStrategy:   rpc.BufferOverflowStrategy_BUFFER_OVERFLOW_STRATEGY_WAIT,
		OverflowWaitMs:     50,
	}
	applyBufferConfig(f, cfg)

	want := map[string]string{
		"_buffer.hwm":              "64",
		"_buffer.interval_ms":      "16",
		"_buffer.line":             "true",
		"_buffer.queue":            "256",
		"_buffer.overflow":         "wait",
		"_buffer.overflow_wait_ms": "50",
	}
	for k, v := range want {
		if !haveCall(f.calls, k, v) {
			t.Fatalf("missing or wrong CONFIGURE %s=%s; calls=%v", k, v, f.calls)
		}
	}
}

// Test that zeros/defaults are handled as intended: we still emit interval/line/overflow,
// default overflow to 'drop', and omit hwm/queue when zero.
func Test_applyBufferConfig_DefaultsAndZeros(t *testing.T) {
	f := &fakeConfigurer{}
	cfg := &rpc.MonitorBufferConfig{ // zeros/unset
		HighWaterMarkBytes: 0,
		FlushIntervalMs:    0,
		LineBuffering:      false,
		FlushQueueCapacity: 0,
		OverflowStrategy:   rpc.BufferOverflowStrategy_BUFFER_OVERFLOW_STRATEGY_UNSPECIFIED,
		OverflowWaitMs:     0,
	}
	applyBufferConfig(f, cfg)

	expects := map[string]string{
		"_buffer.interval_ms":      "0",
		"_buffer.line":             "false",
		"_buffer.overflow":         "drop",
		"_buffer.overflow_wait_ms": "0",
	}
	for k, v := range expects {
		if !haveCall(f.calls, k, v) {
			t.Fatalf("expected CONFIGURE %s=%s not found; calls=%v", k, v, f.calls)
		}
	}
	for _, dis := range []string{"_buffer.hwm", "_buffer.queue"} {
		for _, kv := range f.calls {
			if kv[0] == dis {
				t.Fatalf("did not expect CONFIGURE for %s when value is zero", dis)
			}
		}
	}
}
