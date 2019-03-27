package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/arduino/arduino-cli/rpc"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Please specify Arduino DATA_DIR as first argument")
		os.Exit(1)
	}
	datadir := os.Args[1]
	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewArduinoCoreClient(conn)

	resp, err := client.Init(context.Background(), &rpc.InitReq{
		Configuration: &rpc.Configuration{
			DataDir: datadir,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.GetResult().GetFailed() {
		fmt.Println("Error opening server instance:", resp.GetResult().GetMessage())
		os.Exit(1)
	}
	instance := resp.GetInstance()
	fmt.Println("Opened new server instance:", instance)

	details, err := client.BoardDetails(context.Background(), &rpc.BoardDetailsReq{
		Instance: instance,
		Fqbn:     "arduino:samd:mkr1000",
	})
	if err != nil {
		log.Fatal(err)
	}
	if details.GetResult().GetFailed() {
		fmt.Println("Error getting board data:", details.GetResult().GetMessage())
	} else {
		fmt.Println("Board name: ", details.GetName())
	}

	compResp, err := client.Compile(context.Background(), &rpc.CompileReq{
		Instance:   instance,
		Fqbn:       "arduino:samd:mkr1000",
		SketchPath: os.Args[2],
	})

	if err != nil {
		fmt.Println(compResp.GetResult().Message, err)
		os.Exit(1)
	}

	destroyResp, err := client.Destroy(context.Background(), &rpc.DestroyReq{
		Instance: instance,
	})
	if err != nil {
		log.Fatal(err)
	}
	if destroyResp.GetResult().GetFailed() {
		fmt.Println("Error closing instance:", destroyResp.GetResult().GetMessage())
	} else {
		fmt.Println("Successfully closed server instance")
	}
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
