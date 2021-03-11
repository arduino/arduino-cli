module github.com/arduino/arduino-cli/term_example

go 1.16

replace github.com/arduino/arduino-cli => ../../..

require (
	github.com/arduino/arduino-cli v0.0.0-20200109150215-ffa84fdaab21
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.25.0
)
