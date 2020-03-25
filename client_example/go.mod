module github.com/arduino/arduino-cli/client_example

go 1.14

replace github.com/arduino/arduino-cli => ../

require (
	github.com/arduino/arduino-cli v0.0.0-20200109150215-ffa84fdaab21
	google.golang.org/grpc v1.27.0
)
