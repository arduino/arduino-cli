package rpc

//go:generate protoc -I . -I .. --go_out=plugins=grpc:../../../.. install.proto
