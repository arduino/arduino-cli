package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"

	pb "github.com/arduino/arduino-cli/daemon/arduino"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := pb.NewArduinoCoreClient(conn)
	compile, err := client.Compile(context.Background(), &pb.CompileReq{})
	if err != nil {
		log.Fatal(err)
	}
	for {
		resp, err := compile.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("%+v\n", err)
			log.Fatal(err)
		}
		fmt.Println(resp)
	}
	fmt.Println("Done")
}
