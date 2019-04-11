package main

import (
	"context"
	"fmt"
	"io"
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
		fmt.Printf("Error connecting to server: %s\n", err)
		os.Exit(1)
	}
	client := rpc.NewArduinoCoreClient(conn)

	resp, err := client.Init(context.Background(), &rpc.InitReq{
		Configuration: &rpc.Configuration{
			DataDir: datadir,
		},
	})
	if err != nil {
		fmt.Printf("Error creating server instance: %s\n", err)
		os.Exit(1)
	}
	instance := resp.GetInstance()
	fmt.Println("Created new server instance:", instance)

	install := func() {
		installRespStream, err := client.PlatformInstall(context.Background(), &rpc.PlatformInstallReq{
			Instance:        instance,
			PlatformPackage: "arduino",
			Architecture:    "samd",
			Version:         "1.6.19",
		})
		if err != nil {
			fmt.Printf("Error installing platform: %s\n", err)
			os.Exit(1)
		}
		for {
			installResp, err := installRespStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("Install error: %s\n", err)
				return
			}
			if installResp.GetProgress() != nil {
				fmt.Printf(">> DOWNLOAD: %s\n", installResp.GetProgress())
			}
			if installResp.GetTaskProgress() != nil {
				fmt.Printf(">> TASK: %s\n", installResp.GetTaskProgress())
			}
		}
		fmt.Println("Installation completed!")
	}

	install()
	install()

	upgradeRespStream, err := client.PlatformUpgrade(context.Background(), &rpc.PlatformUpgradeReq{
		Instance:        instance,
		PlatformPackage: "arduino",
		Architecture:    "samd",
	})
	if err != nil {
		fmt.Printf("Error Upgrade platform: %s\n", err)
		os.Exit(1)
	}
	for {
		upgradeResp, err := upgradeRespStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Upgrade error: %s\n", err)
			os.Exit(1)
		}
		if upgradeResp.GetProgress() != nil {
			fmt.Printf(">> DOWNLOAD: %s\n", upgradeResp.GetProgress())
		}
		if upgradeResp.GetTaskProgress() != nil {
			fmt.Printf(">> TASK: %s\n", upgradeResp.GetTaskProgress())
		}
	}
	fmt.Println("Upgrade completed!")

	details, err := client.BoardDetails(context.Background(), &rpc.BoardDetailsReq{
		Instance: instance,
		Fqbn:     "arduino:samd:mkr1000",
	})
	if err != nil {
		fmt.Printf("Error getting board data: %s\n", err)
	} else {
		fmt.Printf("Board name: %s\n", details.GetName())
	}

	compRespStream, err := client.Compile(context.Background(), &rpc.CompileReq{
		Instance:   instance,
		Fqbn:       "arduino:samd:mkr1000",
		SketchPath: os.Args[2],
		Verbose:    true,
	})
	if err != nil {
		fmt.Printf("Compile error: %s\n", err)
		os.Exit(1)
	}
	for {
		compResp, err := compRespStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Compile error: %s\n", err)
			os.Exit(1)
		}
		if compResp.GetOutput() != nil {
			fmt.Printf("%s", compResp.GetOutput())
		}
		if compResp.GetTaskProgress() != nil {
			fmt.Printf(">> TASK: %s\n", compResp.GetTaskProgress())
		}
	}

	_, err = client.PlatformUninstall(context.Background(), &rpc.PlatformUninstallReq{
		Instance:        instance,
		PlatformPackage: "arduino",
		Architecture:    "samd",
		Version:         "1.6.21",
	})
	if err != nil {
		fmt.Printf("uninstall error: %s\n", err)
		// os.Exit(1)
	}

	_, err = client.Destroy(context.Background(), &rpc.DestroyReq{
		Instance: instance,
	})
	if err != nil {
		fmt.Printf("Error closing server instance: %s\n", err)
	} else {
		fmt.Println("Successfully closed server instance")
	}
	fmt.Println("Done")
}
