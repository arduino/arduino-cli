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

	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Please specify Arduino DATA_DIR as first argument")
		os.Exit(1)
	}
	datadir := os.Args[1]

	fmt.Println("=== Connecting to GRPC server")
	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Error connecting to server: %s\n", err)
		os.Exit(1)
	}
	client := rpc.NewArduinoCoreClient(conn)
	fmt.Println()

	// INIT
	fmt.Println("=== calling Init")
	initResp, err := client.Init(context.Background(), &rpc.InitReq{
		Configuration: &rpc.Configuration{
			DataDir: datadir,
		},
	})
	if err != nil {
		fmt.Printf("Error creating server instance: %s\n", err)
		os.Exit(1)
	}
	instance := initResp.GetInstance()
	fmt.Printf("---> %+v\n", initResp)
	fmt.Println()

	// UPDATE-INDEX
	fmt.Println("=== calling UpdateIndex")
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
			fmt.Printf("---> %+v\n", uiResp)
			fmt.Println()
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

	// PLATFORM SEARCH
	fmt.Println("=== calling PlatformSearch(uno)")
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
	fmt.Printf("---> %+v\n", searchResp)
	fmt.Println()

	// PLATFORM INSTALL
	install := func() {
		fmt.Println("=== calling PlatformInstall(arduino:samd@1.6.19)")
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
				fmt.Printf("---> %+v\n", installResp)
				fmt.Println()
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
	}

	install()
	install()

	// PLATFORM LIST
	fmt.Println("=== calling PlatformList()")
	listResp, err := client.PlatformList(context.Background(), &rpc.PlatformListReq{
		Instance: instance,
	})
	if err != nil {
		fmt.Printf("List error: %s\n", err)
		os.Exit(1)
	}
	Installedplatforms := listResp.GetInstalledPlatform()
	for _, listpfm := range Installedplatforms {
		fmt.Printf("---> LIST: %+v\n", listpfm)
	}
	fmt.Println()

	// PLATFORM UPGRADE
	fmt.Println("=== calling PlatformUpgrade(arduino:samd)")
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
			fmt.Printf("---> %+v\n", upgradeResp)
			fmt.Println()
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

	// BOARDS DETAILS
	fmt.Println("=== calling BoardDetails(arduino:samd:mkr1000)")
	details, err := client.BoardDetails(context.Background(), &rpc.BoardDetailsReq{
		Instance: instance,
		Fqbn:     "arduino:samd:mkr1000",
	})
	if err != nil {
		fmt.Printf("Error getting board data: %s\n", err)
	}
	fmt.Printf("---> %+v\n", details)
	fmt.Println()

	// COMPILE
	fmt.Println("=== calling Compile(arduino:samd:mkr1000, VERBOSE, " + os.Args[2] + ")")
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
			fmt.Printf("---> %+v\n", compResp)
			fmt.Println()
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

	// UPLOAD
	fmt.Println("=== calling Upload(arduino:samd:mkr1000, /dev/ttyACM0, VERBOSE, " + os.Args[2] + ")")
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
			fmt.Printf("---> %+v\n", uplResp)
			fmt.Println()
			break
		}
		if err != nil {
			fmt.Printf("Upload error: %s\n", err)
			// os.Exit(1)
			break
		}
		if resp := uplResp.GetOutStream(); resp != nil {
			fmt.Printf(">> STDOUT: %s", resp)
		}
		if resperr := uplResp.GetErrStream(); resperr != nil {
			fmt.Printf(">> STDERR: %s", resperr)
		}
	}

	// PLATFORM UNINSTALL
	fmt.Println("=== calling PlatformUninstall(arduino:samd@1.6.21)")
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
			fmt.Printf("---> %+v\n", uninstallResp)
			fmt.Println()
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

	// LIB DOWNLOAD
	fmt.Println("=== calling LibraryDownload(WiFi101@0.15.2)")
	downloadRespStream, err := client.LibraryDownload(context.Background(), &rpc.LibraryDownloadReq{
		Instance: instance,
		Name:     "WiFi101",
		Version:  "0.15.2",
	})
	if err != nil {
		fmt.Printf("Error downloading library: %s\n", err)
		os.Exit(1)
	}
	for {
		downloadResp, err := downloadRespStream.Recv()
		if err == io.EOF {
			fmt.Printf("---> %+v\n", downloadResp)
			fmt.Println()
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

	libInstall := func(version string) {
		// LIB INSTALL
		fmt.Println("=== calling LibraryInstall(WiFi101@" + version + ")")
		installRespStream, err := client.LibraryInstall(context.Background(), &rpc.LibraryInstallReq{
			Instance: instance,
			Name:     "WiFi101",
			Version:  version,
		})
		if err != nil {
			fmt.Printf("Error installing library: %s\n", err)
			os.Exit(1)
		}
		for {
			installResp, err := installRespStream.Recv()
			if err == io.EOF {
				fmt.Printf("---> %+v\n", installResp)
				fmt.Println()
				break
			}
			if err != nil {
				fmt.Printf("install error: %s\n", err)
				os.Exit(1)
			}
			if installResp.GetProgress() != nil {
				fmt.Printf(">> DOWNLOAD: %s\n", installResp.GetProgress())
			}
			if installResp.GetTaskProgress() != nil {
				fmt.Printf(">> TASK: %s\n", installResp.GetTaskProgress())
			}
		}
	}

	libInstall("0.15.1") // Install
	libInstall("0.15.2") // Replace

	// LIB UPGRADE
	fmt.Println("=== calling LibraryUpgradeAll()")
	libUpgradeAllRespStream, err := client.LibraryUpgradeAll(context.Background(), &rpc.LibraryUpgradeAllReq{
		Instance: instance,
	})
	if err != nil {
		fmt.Printf("Error upgrading all: %s\n", err)
		os.Exit(1)
	}
	for {
		resp, err := libUpgradeAllRespStream.Recv()
		if err == io.EOF {
			fmt.Printf("---> %+v\n", resp)
			fmt.Println()
			break
		}
		if err != nil {
			fmt.Printf("upgrading error: %s\n", err)
			os.Exit(1)
		}
		if resp.GetProgress() != nil {
			fmt.Printf(">> DOWNLOAD: %s\n", resp.GetProgress())
		}
		if resp.GetTaskProgress() != nil {
			fmt.Printf(">> TASK: %s\n", resp.GetTaskProgress())
		}
	}

	// LIB SEARCH
	fmt.Println("=== calling LibrarySearch(audio)")
	outputsrc, err := client.LibrarySearch(context.Background(), &rpc.LibrarySearchReq{
		Instance: instance,
		Names:    true,
		Query:    "audio",
	})
	if err != nil {
		formatter.PrintError(err, "Error saerching for Library")
		os.Exit(1)
	}

	fmt.Println(outputsrc)

	// LIB UNINSTALL
	fmt.Println("=== calling LibraryUninstall(WiFi101)")
	libUninstallRespStream, err := client.LibraryUninstall(context.Background(), &rpc.LibraryUninstallReq{
		Instance: instance,
		Name:     "WiFi101",
	})
	if err != nil {
		fmt.Printf("Error uninstalling: %s\n", err)
		os.Exit(1)
	}
	for {
		uninstallResp, err := libUninstallRespStream.Recv()
		if err == io.EOF {
			fmt.Printf("---> %+v\n", uninstallResp)
			fmt.Println()
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

	// DESTROY
	fmt.Println("=== calling Destroy()")
	_, err = client.Destroy(context.Background(), &rpc.DestroyReq{
		Instance: instance,
	})
	if err != nil {
		fmt.Printf("Error closing server instance: %s\n", err)
	} else {
		fmt.Println("Successfully closed server instance")
	}
}
