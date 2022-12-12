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

package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"google.golang.org/grpc"
)

var debugStdOut io.Writer
var debugSeq uint32

func log(isRequest bool, seq uint32, msg interface{}) {
	prefix := fmt.Sprint(seq, " |  ")
	j, _ := json.MarshalIndent(msg, prefix, "  ")
	inOut := "RESP: "
	if isRequest {
		inOut = "REQ:  "
	}
	fmt.Fprintln(debugStdOut, prefix+inOut+string(j))
}

func logError(seq uint32, err error) {
	if err != nil {
		fmt.Fprintln(debugStdOut, seq, "|  ERROR: ", err)
	}
}

func logSelector(method string) bool {
	if len(debugFilters) == 0 {
		return true
	}
	for _, filter := range debugFilters {
		if strings.Contains(method, filter) {
			return true
		}
	}
	return false
}

func unaryLoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !logSelector(info.FullMethod) {
		return handler(ctx, req)
	}
	seq := atomic.AddUint32(&debugSeq, 1)
	fmt.Fprintln(debugStdOut, seq, "CALLED:", info.FullMethod)
	log(true, seq, req)
	resp, err := handler(ctx, req)
	logError(seq, err)
	log(false, seq, resp)
	fmt.Fprintln(debugStdOut, seq, "CALL END")
	fmt.Fprintln(debugStdOut)
	return resp, err
}

func streamLoggerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !logSelector(info.FullMethod) {
		return handler(srv, stream)
	}
	seq := atomic.AddUint32(&debugSeq, 1)
	streamReq := ""
	if info.IsClientStream {
		streamReq = "STREAM_REQ "
	}
	if info.IsServerStream {
		streamReq += "STREAM_RESP"
	}
	fmt.Fprintln(debugStdOut, seq, "CALLED:", info.FullMethod, streamReq)
	err := handler(srv, &loggingServerStream{ServerStream: stream, seq: seq})
	logError(seq, err)
	fmt.Fprintln(debugStdOut, seq, "STREAM CLOSED")
	fmt.Fprintln(debugStdOut)
	return err
}

type loggingServerStream struct {
	grpc.ServerStream
	seq uint32
}

func (l *loggingServerStream) RecvMsg(m interface{}) error {
	err := l.ServerStream.RecvMsg(m)
	logError(l.seq, err)
	log(true, l.seq, m)
	return err
}

func (l *loggingServerStream) SendMsg(m interface{}) error {
	err := l.ServerStream.SendMsg(m)
	logError(l.seq, err)
	log(false, l.seq, m)
	return err
}
