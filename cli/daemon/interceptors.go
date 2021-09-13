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
	"strings"

	"google.golang.org/grpc"
)

func log(isRequest bool, msg interface{}) {
	j, _ := json.MarshalIndent(msg, "|  ", "  ")
	inOut := map[bool]string{true: "REQ:  ", false: "RESP: "}
	fmt.Println("|  " + inOut[isRequest] + string(j))
}

func logError(err error) {
	if err != nil {
		fmt.Println("|  ERROR: ", err)
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
	fmt.Println("CALLED:", info.FullMethod)
	log(true, req)
	resp, err := handler(ctx, req)
	logError(err)
	log(false, resp)
	fmt.Println()
	return resp, err
}

func streamLoggerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !logSelector(info.FullMethod) {
		return handler(srv, stream)
	}
	streamReq := ""
	if info.IsClientStream {
		streamReq = "STREAM_REQ "
	}
	if info.IsServerStream {
		streamReq += "STREAM_RESP"
	}
	fmt.Println("CALLED:", info.FullMethod, streamReq)
	err := handler(srv, &loggingServerStream{ServerStream: stream})
	logError(err)
	fmt.Println()
	return err
}

type loggingServerStream struct {
	grpc.ServerStream
}

func (l *loggingServerStream) RecvMsg(m interface{}) error {
	err := l.ServerStream.RecvMsg(m)
	logError(err)
	log(true, m)
	return err
}

func (l *loggingServerStream) SendMsg(m interface{}) error {
	err := l.ServerStream.SendMsg(m)
	logError(err)
	log(false, m)
	return err
}
