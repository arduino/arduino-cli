module github.com/arduino/arduino-cli/client_example

go 1.17

replace github.com/arduino/arduino-cli => ../

require (
	github.com/arduino/arduino-cli v0.0.0-20200109150215-ffa84fdaab21
	google.golang.org/grpc v1.38.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.0.0-20210505024714-0287a6fb4125 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)
