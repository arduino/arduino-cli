module github.com/arduino/arduino-cli/client_example

go 1.19

replace github.com/arduino/arduino-cli => ../

require (
	github.com/arduino/arduino-cli v0.0.0-20200109150215-ffa84fdaab21
	google.golang.org/grpc v1.38.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)
