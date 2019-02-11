package board

//go:generate protoc -I . -I .. --go_out=plugins=grpc:../../../../.. board.proto
