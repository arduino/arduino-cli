package daemon_test

import (
	"context"
	"io"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/arduino/arduino-cli/arduino/monitors"
	"github.com/arduino/arduino-cli/commands/daemon"
	"github.com/arduino/arduino-cli/rpc/monitor"
	st "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

var (
	recvCounter        int
	resPortName        string
	resBaudRate        int
	resWrittenToSerial []byte
	resReadFromSerial  []byte
)

type TestStreamingOpenServer struct{}

func (s *TestStreamingOpenServer) Send(mon *monitor.StreamingOpenResp) error {
	// if we're here, the Monitor read something from the target and
	// is sending it back to the stream client
	resReadFromSerial = mon.GetData()
	return nil
}

func (s *TestStreamingOpenServer) Recv() (*monitor.StreamingOpenReq, error) {
	// if we're here, the monitor is reading the stream client
	// we only send 3 messages, one for config, another with data and a final
	// one with EOF so the monitor will gracefully exit
	recvCounter++
	if recvCounter == 1 {
		// send the first message containing the configuration
		additionalFields := make(map[string]*st.Value, 1)
		additionalFields["BaudRate"] = &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(42),
			},
		}
		return &monitor.StreamingOpenReq{
			Content: &monitor.StreamingOpenReq_MonitorConfig{
				MonitorConfig: &monitor.MonitorConfig{
					Target: "/dev/tty42",
					Type:   monitor.MonitorConfig_SERIAL,
					AdditionalConfig: &st.Struct{
						Fields: additionalFields,
					},
				},
			},
		}, nil
	} else if recvCounter == 2 {
		return &monitor.StreamingOpenReq{
			Content: &monitor.StreamingOpenReq_Data{
				Data: []byte("Hello Serial, this if for you!"),
			},
		}, nil
	}

	return nil, io.EOF
}

func (s *TestStreamingOpenServer) SetHeader(metadata.MD) error  { return nil }
func (s *TestStreamingOpenServer) SendHeader(metadata.MD) error { return nil }
func (s *TestStreamingOpenServer) SetTrailer(metadata.MD)       {}
func (s *TestStreamingOpenServer) Context() context.Context     { return context.Background() }
func (s *TestStreamingOpenServer) SendMsg(m interface{}) error  { return nil }
func (s *TestStreamingOpenServer) RecvMsg(m interface{}) error  { return nil }

func mockOpenSerialMonitor(portName string, baudRate int) (*monitors.SerialMonitor, error) {
	// this function will be called by the Monitor as soon as it receives the
	// first message from the stream client

	// save parameters so the Test function can assert on the values passed to the monitor
	// by the client
	resPortName = portName
	resBaudRate = baudRate

	mon := &monitors.SerialMonitor{}
	monkey.PatchInstanceMethod(reflect.TypeOf(mon), "Close", func(_ *monitors.SerialMonitor) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(mon), "Read", func(_ *monitors.SerialMonitor, bytes []byte) (int, error) {
		copy(bytes, "I am Serial")
		return len(bytes), nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(mon), "Write", func(_ *monitors.SerialMonitor, bytes []byte) (int, error) {
		resWrittenToSerial = bytes
		return len(bytes), nil
	})

	return mon, nil
}

func TestFoo(t *testing.T) {
	monkey.Patch(monitors.OpenSerialMonitor, mockOpenSerialMonitor)

	svc := daemon.MonitorService{}
	stream := &TestStreamingOpenServer{}

	// let the monitor go, this will return when the monitor receives
	// the EOF from the stream client
	assert.Nil(t, svc.StreamingOpen(stream))

	// ensure port setup was correct
	assert.Equal(t, "/dev/tty42", resPortName)
	assert.Equal(t, 42, resBaudRate)

	// ensure the serial received the message
	assert.Equal(t, []byte("Hello Serial, this if for you!"), resWrittenToSerial)

	// ensure the monitor read from the serial, output is truncated because the test
	// doesn't consume the whole buffer
	assert.Equal(t, []byte("I am Ser"), resReadFromSerial)
}
