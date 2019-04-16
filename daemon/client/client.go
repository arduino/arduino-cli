//
// This file is part of arduino-cli.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

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
				os.Exit(1)
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
		if resp := compResp.GetOutStream(); resp != nil {
			fmt.Printf(">> STDOUT: %s", resp)
		}
		if resperr := compResp.GetErrStream(); resperr != nil {
			fmt.Printf(">> STDERR: %s", resperr)
		}
		if compResp.GetDownloadProgress() != nil {
			fmt.Printf(">> DOWNLOAD: %s\n", compResp.GetDownloadProgress())
		}
		if compResp.GetTaskProgress() != nil {
			fmt.Printf(">> TASK: %s\n", compResp.GetTaskProgress())
		}
	}

	uplRespStream, err := client.Upload(context.Background(), &rpc.UploadReq{
		Instance:   instance,
		Fqbn:       "arduino:samd:mkr1000",
		SketchPath: os.Args[2],
		Port:       "/dev/ttyACM0",
		Verbose:    true,
	})
	if err != nil {
		fmt.Printf("Upload error: %s\n", err)
		os.Exit(1)
	}
	for {
		uplResp, err := uplRespStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Upload error: %s\n", err)
			os.Exit(1)
		}
		if resp := uplResp.GetOutStream(); resp != nil {
			fmt.Printf(">> STDOUT: %s", resp)
		}
		if resperr := uplResp.GetErrStream(); resperr != nil {
			fmt.Printf(">> STDERR: %s", resperr)
		}
	}

	uiRespStream, err := client.UpdateIndex(context.Background(), &rpc.UpdateIndexReq{
		Instance: instance,
	})
	if err != nil {
		fmt.Printf("Error Upgrade platform: %s\n", err)
		os.Exit(1)
	}
	for {
		uiResp, err := uiRespStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Compile error: %s\n", err)
			os.Exit(1)
		}
		if uiResp.GetDownloadProgress() != nil {
			fmt.Printf(">> DOWNLOAD: %s\n", uiResp.GetDownloadProgress())
		}
	}

	searchResp, err := client.PlatformSearch(context.Background(), &rpc.PlatformSearchReq{
		Instance:   instance,
		SearchArgs: "uno",
	})
	if err != nil {
		fmt.Printf("Search error: %s\n", err)
		os.Exit(1)
	}
	serchedOutput := searchResp.GetSearchOutput()
	for _, outsearch := range serchedOutput {
		fmt.Printf(">> SEARCH: %+v\n", outsearch)
	}

	listResp, err := client.PlatformList(context.Background(), &rpc.PlatformListReq{
		Instance: instance,
	})
	if err != nil {
		fmt.Printf("List error: %s\n", err)
		os.Exit(1)
	}
	Installedplatforms := listResp.GetInstalledPlatform()
	for _, listpfm := range Installedplatforms {
		fmt.Printf(">> LIST: %+v\n", listpfm)
	}

	uninstallRespStream, err := client.PlatformUninstall(context.Background(), &rpc.PlatformUninstallReq{
		Instance:        instance,
		PlatformPackage: "arduino",
		Architecture:    "samd",
		Version:         "1.6.21",
	})
	if err != nil {
		fmt.Printf("uninstall error: %s\n", err)
		os.Exit(1)
	}
	for {
		uninstallResp, err := uninstallRespStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("uninstall error: %s\n", err)
			os.Exit(1)
		}
		if uninstallResp.GetTaskProgress() != nil {
			fmt.Printf(">> TASK: %s\n", uninstallResp.GetTaskProgress())
		}
	}

	downloadRespStream, err := client.LibraryDownload(context.Background(), &rpc.LibraryDownloadReq{
		Instance: instance,
		Name:     "WiFi101",
		Version:  "0.15.2",
	})
	if err != nil {
		fmt.Printf("Error Upgrade platform: %s\n", err)
		os.Exit(1)
	}
	for {
		downloadResp, err := downloadRespStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Download error: %s\n", err)
			os.Exit(1)
		}
		if downloadResp.GetProgress() != nil {
			fmt.Printf(">> DOWNLOAD: %s\n", downloadResp.GetProgress())
		}
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
