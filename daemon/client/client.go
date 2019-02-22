package main

import (
	"context"
	"fmt"
	"log"

	"github.com/arduino/arduino-cli/rpc"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewArduinoCoreClient(conn)

	resp, err := client.Init(context.Background(), &rpc.InitReq{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.GetInstance())

	details, err := client.BoardDetails(context.Background(), &rpc.BoardDetailsReq{
		Instance: resp.GetInstance(),
		Fqbn:     "",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(details.GetName())
	/*
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
	*/
	fmt.Println("Done")
}
