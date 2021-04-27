// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	dbg "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/debug/v1"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/settings/v1"
	"google.golang.org/grpc"
)

var (
	dataDir string
)

// The main function implements an example workflow to show how to interact
// with the gRPC API exposed by arduino-cli when running in daemon mode.
func main() {

	// Establish a connection with the gRPC server, started with the command:
	// arduino-cli daemon
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal("error connecting to arduino-cli rpc server, you can start it by running `arduino-cli daemon`")
	}
	defer conn.Close()

	// To avoid polluting an existing arduino-cli installation, the example
	// client uses a temp folder to keep cores, libraries and the like.
	// You can point `dataDir` to a location that better fits your needs.
	dataDir, err = ioutil.TempDir("", "arduino-rpc-client")
	if err != nil {
		log.Fatal(err)
	}
	dataDir = filepath.ToSlash(dataDir)
	defer os.RemoveAll(dataDir)

	// Create an instance of the gRPC clients.
	client := rpc.NewArduinoCoreServiceClient(conn)

	settingsClient := settings.NewSettingsServiceClient(conn)

	// Now we can call various methods of the API...

	// `Version` can be called without any setup or init procedure.
	log.Println("calling Version")
	callVersion(client)

	log.Println("calling LoadSketch")
	callLoadSketch(client)

	// Use SetValue to configure the arduino-cli directories.
	log.Println("calling SetValue")
	callSetValue(settingsClient)

	// List all the settings.
	log.Println("calling GetAll()")
	callGetAll(settingsClient)

	// Merge applies multiple settings values at once.
	log.Println("calling Merge()")
	callMerge(settingsClient, `{"foo": {"value": "bar"}, "daemon":{"port":"422"}, "board_manager": {"additional_urls":["https://example.com"]}}`)

	log.Println("calling GetAll()")
	callGetAll(settingsClient)

	log.Println("calling Merge()")
	callMerge(settingsClient, `{"foo": {} }`)

	log.Println("calling GetAll()")
	callGetAll(settingsClient)

	log.Println("calling Merge()")
	callMerge(settingsClient, `{"foo": "bar" }`)

	// Get the value of the foo key.
	log.Println("calling GetValue(foo)")
	callGetValue(settingsClient)

	// List all the settings.
	log.Println("calling GetAll()")
	callGetAll(settingsClient)

	// Write settings to file.
	log.Println("calling Write()")
	callWrite(settingsClient)

	// Before we can do anything with the CLI, an "instance" must be created.
	// We keep a reference to the created instance because we will need it to
	// run subsequent commands.
	log.Println("calling Create")
	instance := createInstance(client)

	log.Println("calling Init")
	initInstance(client, instance)

	// We set up the proxy and then run the update to verify that the proxy settings are currently used
	log.Println("calling setProxy")
	callSetProxy(settingsClient)

	// With a brand new instance, the first operation should always be updating
	// the index.
	log.Println("calling UpdateIndex")
	callUpdateIndex(client, instance)

	// And we run update again
	log.Println("calling UpdateIndex")
	callUpdateIndex(client, instance)

	// Indexes are not implicitly detected after an update
	// so we must initialize again explicitly
	log.Println("calling Init")
	initInstance(client, instance)

	// Let's search for a platform (also known as 'core') called 'samd'.
	log.Println("calling PlatformSearch(samd)")
	callPlatformSearch(client, instance)

	// Install arduino:samd@1.6.19
	log.Println("calling PlatformInstall(arduino:samd@1.6.19)")
	callPlatformInstall(client, instance)

	// Now list the installed platforms to double check previous installation
	// went right.
	log.Println("calling PlatformList()")
	callPlatformList(client, instance)

	// Upgrade the installed platform to the latest version.
	log.Println("calling PlatformUpgrade(arduino:samd)")
	callPlatformUpgrade(client, instance)

	// Query board details for a mkr1000
	log.Println("calling BoardDetails(arduino:samd:mkr1000)")
	callBoardsDetails(client, instance)

	log.Println("calling BoardSearch()")
	callBoardSearch(client, instance)

	// Attach a board to a sketch.
	// Uncomment if you do have an actual board connected.
	// log.Println("calling BoardAttach(serial:///dev/ttyACM0)")
	// callBoardAttach(client, instance)

	// Compile a sketch
	log.Println("calling Compile(arduino:samd:mkr1000, VERBOSE, hello.ino)")
	callCompile(client, instance)

	// Upload a sketch
	// Uncomment if you do have an actual board connected.
	// log.Println("calling Upload(arduino:samd:mkr1000, /dev/ttyACM0, VERBOSE, hello.ino)")
	// callUpload(client, instance)

	// Debug a sketch on a board
	// Uncomment if you do have an actual board connected via debug port,
	// or a board connected to a debugger.
	// debugClient := dbg.NewDebugClient(conn)
	// debugStreamingClient, err := debugClient.Debug(context.Background())
	// if err != nil {
	// 	 log.Fatalf("debug streaming open error: %s\n", err)
	// }
	// log.Println("calling Debug(arduino:samd:mkr1000, hello.ino)")
	// callDebugger(debugStreamingClient, instance)

	// List all boards
	log.Println("calling BoardListAll(mkr)")
	callListAll(client, instance)

	// List connected boards
	log.Println("calling BoardList()")
	callBoardList(client, instance)

	// Watch for boards connection and disconnection
	log.Println("calling BoardListWatch()")
	callBoardListWatch(client, instance)

	// Uninstall a platform
	log.Println("calling PlatformUninstall(arduino:samd)")
	callPlatformUnInstall(client, instance)

	// Update the Library index
	log.Println("calling UpdateLibrariesIndex()")
	callUpdateLibraryIndex(client, instance)

	// Indexes are not implicitly detected after an update
	// so we must initialize again explicitly
	log.Println("calling Init")
	initInstance(client, instance)

	// Download a library
	log.Println("calling LibraryDownload(WiFi101@0.15.2)")
	callLibDownload(client, instance)

	// Install a library
	log.Println("calling LibraryInstall(WiFi101@0.15.1)")
	callLibInstall(client, instance, "0.15.1")

	// Replace the previous version
	log.Println("calling LibraryInstall(WiFi101@0.15.2)")
	callLibInstall(client, instance, "0.15.2")

	// Install a library skipping deps installation
	log.Println("calling LibraryInstall(Arduino_MKRIoTCarrier@0.9.9) skipping dependencies")
	callLibInstallNoDeps(client, instance, "0.9.9")

	// Upgrade all libs to latest
	log.Println("calling LibraryUpgradeAll()")
	callLibUpgradeAll(client, instance)

	// Search for a lib using the 'audio' keyword
	log.Println("calling LibrarySearch(audio)")
	callLibSearch(client, instance)

	// List the dependencies of the ArduinoIoTCloud library
	log.Println("calling LibraryResolveDependencies(ArduinoIoTCloud)")
	callLibraryResolveDependencies(client, instance)

	// List installed libraries
	log.Println("calling LibraryList")
	callLibList(client, instance)

	// Uninstall a library
	log.Println("calling LibraryUninstall(WiFi101)")
	callLibUninstall(client, instance)
}

func callVersion(client rpc.ArduinoCoreServiceClient) {
	versionResp, err := client.Version(context.Background(), &rpc.VersionRequest{})
	if err != nil {
		log.Fatalf("Error getting version: %s", err)
	}

	log.Printf("arduino-cli version: %v", versionResp.GetVersion())
}

func callSetValue(client settings.SettingsServiceClient) {
	_, err := client.SetValue(context.Background(),
		&settings.SetValueRequest{
			Key:      "directories",
			JsonData: `{"data": "` + dataDir + `", "downloads": "` + path.Join(dataDir, "staging") + `", "user": "` + path.Join(dataDir, "sketchbook") + `"}`,
		})

	if err != nil {
		log.Fatalf("Error setting settings value: %s", err)

	}
}

func callSetProxy(client settings.SettingsServiceClient) {
	_, err := client.SetValue(context.Background(),
		&settings.SetValueRequest{
			Key:      "network.proxy",
			JsonData: `"http://localhost:3128"`,
		})

	if err != nil {
		log.Fatalf("Error setting settings value: %s", err)
	}
}

func callUnsetProxy(client settings.SettingsServiceClient) {
	_, err := client.SetValue(context.Background(),
		&settings.SetValueRequest{
			Key:      "network.proxy",
			JsonData: `""`,
		})

	if err != nil {
		log.Fatalf("Error setting settings value: %s", err)
	}
}

func callMerge(client settings.SettingsServiceClient, jsonData string) {
	_, err := client.Merge(context.Background(),
		&settings.MergeRequest{
			JsonData: jsonData,
		})

	if err != nil {
		log.Fatalf("Error merging settings: %s", err)
	}
}

func callGetValue(client settings.SettingsServiceClient) {
	getValueResp, err := client.GetValue(context.Background(),
		&settings.GetValueRequest{
			Key: "foo",
		})

	if err != nil {
		log.Fatalf("Error getting settings value: %s", err)
	}

	log.Printf("Value: %s: %s", getValueResp.GetKey(), getValueResp.GetJsonData())
}

func callGetAll(client settings.SettingsServiceClient) {
	getAllResp, err := client.GetAll(context.Background(), &settings.GetAllRequest{})

	if err != nil {
		log.Fatalf("Error getting settings: %s", err)
	}

	log.Printf("Settings: %s", getAllResp.GetJsonData())
}

func callWrite(client settings.SettingsServiceClient) {
	_, err := client.Write(context.Background(),
		&settings.WriteRequest{
			FilePath: path.Join(dataDir, "written-settings.yml"),
		})

	if err != nil {
		log.Fatalf("Error writing settings: %s", err)
	}
}

func createInstance(client rpc.ArduinoCoreServiceClient) *rpc.Instance {
	res, err := client.Create(context.Background(), &rpc.CreateRequest{})
	if err != nil {
		log.Fatalf("Error creating server instance: %s", err)
	}
	return res.Instance
}

func initInstance(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	stream, err := client.Init(context.Background(), &rpc.InitRequest{
		Instance: instance,
	})
	if err != nil {
		log.Fatalf("Error initializing server instance: %s", err)
	}

	for {
		res, err := stream.Recv()
		// Server has finished sending
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Init error: %s", err)
		}

		if status := res.GetError(); status != nil {
			log.Printf("Init error %s", status.String())
		}

		if progress := res.GetInitProgress(); progress != nil {
			if downloadProgress := progress.GetDownloadProgress(); downloadProgress != nil {
				log.Printf("DOWNLOAD: %s", downloadProgress)
			}
			if taskProgress := progress.GetTaskProgress(); taskProgress != nil {
				log.Printf("TASK: %s", taskProgress)
			}
		}
	}
}

func callUpdateIndex(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	uiRespStream, err := client.UpdateIndex(context.Background(), &rpc.UpdateIndexRequest{
		Instance: instance,
	})
	if err != nil {
		log.Fatalf("Error updating index: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		uiResp, err := uiRespStream.Recv()

		// the server is done
		if err == io.EOF {
			log.Print("Update index done")
			break
		}

		// there was an error
		if err != nil {
			log.Fatalf("Update error: %s", err)
		}

		// operations in progress
		if uiResp.GetDownloadProgress() != nil {
			log.Printf("DOWNLOAD: %s", uiResp.GetDownloadProgress())
		}
	}
}

func callPlatformSearch(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	searchResp, err := client.PlatformSearch(context.Background(), &rpc.PlatformSearchRequest{
		Instance:   instance,
		SearchArgs: "samd",
	})

	if err != nil {
		log.Fatalf("Search error: %s", err)
	}

	platforms := searchResp.GetSearchOutput()
	for _, plat := range platforms {
		// We only print ID and version of the platforms found but you can look
		// at the definition for the rpc.Platform struct for more fields.
		log.Printf("Search result: %+v - %+v", plat.GetId(), plat.GetLatest())
	}
}

func callPlatformInstall(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	installRespStream, err := client.PlatformInstall(context.Background(),
		&rpc.PlatformInstallRequest{
			Instance:        instance,
			PlatformPackage: "arduino",
			Architecture:    "samd",
			Version:         "1.6.19",
		})

	if err != nil {
		log.Fatalf("Error installing platform: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		installResp, err := installRespStream.Recv()

		// The server is done.
		if err == io.EOF {
			log.Printf("Install done")
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Install error: %s", err)
		}

		// When a download is ongoing, log the progress
		if installResp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s", installResp.GetProgress())
		}

		// When an overall task is ongoing, log the progress
		if installResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s", installResp.GetTaskProgress())
		}
	}
}

func callPlatformList(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	listResp, err := client.PlatformList(context.Background(),
		&rpc.PlatformListRequest{Instance: instance})

	if err != nil {
		log.Fatalf("List error: %s", err)
	}

	for _, plat := range listResp.GetInstalledPlatforms() {
		// We only print ID and version of the installed platforms but you can look
		// at the definition for the rpc.Platform struct for more fields.
		log.Printf("Installed platform: %s - %s", plat.GetId(), plat.GetInstalled())
	}
}

func callPlatformUpgrade(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	upgradeRespStream, err := client.PlatformUpgrade(context.Background(),
		&rpc.PlatformUpgradeRequest{
			Instance:        instance,
			PlatformPackage: "arduino",
			Architecture:    "samd",
		})

	if err != nil {
		log.Fatalf("Error upgrading platform: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		upgradeResp, err := upgradeRespStream.Recv()

		// The server is done.
		if err == io.EOF {
			log.Printf("Upgrade done")
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Upgrade error: %s", err)
		}

		// When a download is ongoing, log the progress
		if upgradeResp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s", upgradeResp.GetProgress())
		}

		// When an overall task is ongoing, log the progress
		if upgradeResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s", upgradeResp.GetTaskProgress())
		}
	}
}

func callBoardsDetails(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	details, err := client.BoardDetails(context.Background(),
		&rpc.BoardDetailsRequest{
			Instance: instance,
			Fqbn:     "arduino:samd:mkr1000",
		})

	if err != nil {
		log.Fatalf("Error getting board data: %s\n", err)
	}

	log.Printf("Board details for %s", details.GetName())
	log.Printf("Required tools: %s", details.GetToolsDependencies())
	log.Printf("Config options: %s", details.GetConfigOptions())
}

func callBoardSearch(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	res, err := client.BoardSearch(context.Background(),
		&rpc.BoardSearchRequest{
			Instance:   instance,
			SearchArgs: "",
		})

	if err != nil {
		log.Fatalf("Error getting board data: %s\n", err)
	}

	for _, board := range res.Boards {
		log.Printf("Board Name: %s, Board Platform: %s\n", board.Name, board.Platform.Id)
	}
}

func callBoardAttach(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	currDir, _ := os.Getwd()
	boardattachresp, err := client.BoardAttach(context.Background(),
		&rpc.BoardAttachRequest{
			Instance:   instance,
			BoardUri:   "/dev/ttyACM0",
			SketchPath: filepath.Join(currDir, "hello"),
		})

	if err != nil {
		log.Fatalf("Attach error: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		attachResp, err := boardattachresp.Recv()

		// The server is done.
		if err == io.EOF {
			log.Print("Attach done")
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Attach error: %s\n", err)
		}

		// When an overall task is ongoing, log the progress
		if attachResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s", attachResp.GetTaskProgress())
		}
	}
}

func callCompile(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	currDir, _ := os.Getwd()
	compRespStream, err := client.Compile(context.Background(),
		&rpc.CompileRequest{
			Instance:   instance,
			Fqbn:       "arduino:samd:mkr1000",
			SketchPath: filepath.Join(currDir, "hello"),
			Verbose:    true,
		})

	if err != nil {
		log.Fatalf("Compile error: %s\n", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		compResp, err := compRespStream.Recv()

		// The server is done.
		if err == io.EOF {
			log.Print("Compilation done")
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Compile error: %s\n", err)
		}

		// When an operation is ongoing you can get its output
		if resp := compResp.GetOutStream(); resp != nil {
			log.Printf("STDOUT: %s", resp)
		}
		if resperr := compResp.GetErrStream(); resperr != nil {
			log.Printf("STDERR: %s", resperr)
		}
	}
}

func callUpload(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	currDir, _ := os.Getwd()
	uplRespStream, err := client.Upload(context.Background(),
		&rpc.UploadRequest{
			Instance:   instance,
			Fqbn:       "arduino:samd:mkr1000",
			SketchPath: filepath.Join(currDir, "hello"),
			Port:       "/dev/ttyACM0",
			Verbose:    true,
		})

	if err != nil {
		log.Fatalf("Upload error: %s\n", err)
	}

	for {
		uplResp, err := uplRespStream.Recv()
		if err == io.EOF {
			log.Printf("Upload done")
			break
		}

		if err != nil {
			log.Fatalf("Upload error: %s", err)
			break
		}

		// When an operation is ongoing you can get its output
		if resp := uplResp.GetOutStream(); resp != nil {
			log.Printf("STDOUT: %s", resp)
		}
		if resperr := uplResp.GetErrStream(); resperr != nil {
			log.Printf("STDERR: %s", resperr)
		}
	}
}

func callListAll(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	boardListAllResp, err := client.BoardListAll(context.Background(),
		&rpc.BoardListAllRequest{
			Instance:   instance,
			SearchArgs: []string{"mkr"},
		})

	if err != nil {
		log.Fatalf("Board list-all error: %s", err)
	}

	for _, board := range boardListAllResp.GetBoards() {
		log.Printf("%s: %s", board.GetName(), board.GetFqbn())
	}
}

func callBoardList(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	boardListResp, err := client.BoardList(context.Background(),
		&rpc.BoardListRequest{Instance: instance})

	if err != nil {
		log.Fatalf("Board list error: %s\n", err)
	}

	for _, port := range boardListResp.GetPorts() {
		log.Printf("port: %s, boards: %+v\n", port.GetAddress(), port.GetBoards())
	}
}

func callBoardListWatch(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	watchClient, err := client.BoardListWatch(context.Background())
	if err != nil {
		log.Fatalf("Board list watch error: %s\n", err)
	}

	// Start the watcher
	watchClient.Send(&rpc.BoardListWatchRequest{
		Instance: instance,
	})

	go func() {
		for {
			res, err := watchClient.Recv()
			if err != nil {
				log.Fatalf("Board list watch error: %s\n", err)
			}

			log.Printf("event: %s, address: %s\n", res.EventType, res.Port.Address)
			if res.EventType == "add" {
				log.Printf("protocol: %s, ", res.Port.Protocol)
				log.Printf("protocolLabel: %s, ", res.Port.ProtocolLabel)
				log.Printf("boards: %s\n\n", res.Port.Boards)
			}
		}
	}()

	// Watch for 10 seconds and then interrupts
	timer := time.NewTicker(time.Duration(10 * time.Second))
	<-timer.C
	watchClient.Send(&rpc.BoardListWatchRequest{
		Interrupt: true,
	})
}

func callPlatformUnInstall(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	uninstallRespStream, err := client.PlatformUninstall(context.Background(),
		&rpc.PlatformUninstallRequest{
			Instance:        instance,
			PlatformPackage: "arduino",
			Architecture:    "samd",
		})

	if err != nil {
		log.Fatalf("Uninstall error: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		uninstallResp, err := uninstallRespStream.Recv()
		if err == io.EOF {
			log.Print("Uninstall done")
			break
		}

		if err != nil {
			log.Fatalf("Uninstall error: %s\n", err)
		}

		if uninstallResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s\n", uninstallResp.GetTaskProgress())
		}
	}
}

func callUpdateLibraryIndex(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	libIdxUpdateStream, err := client.UpdateLibrariesIndex(context.Background(),
		&rpc.UpdateLibrariesIndexRequest{Instance: instance})

	if err != nil {
		log.Fatalf("Error updating libraries index: %s\n", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		resp, err := libIdxUpdateStream.Recv()
		if err == io.EOF {
			log.Print("Library index update done")
			break
		}

		if err != nil {
			log.Fatalf("Error updating libraries index: %s", err)
		}

		if resp.GetDownloadProgress() != nil {
			log.Printf("DOWNLOAD: %s", resp.GetDownloadProgress())
		}
	}
}

func callLibDownload(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	downloadRespStream, err := client.LibraryDownload(context.Background(),
		&rpc.LibraryDownloadRequest{
			Instance: instance,
			Name:     "WiFi101",
			Version:  "0.15.2",
		})

	if err != nil {
		log.Fatalf("Error downloading library: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		downloadResp, err := downloadRespStream.Recv()
		if err == io.EOF {
			log.Print("Lib download done")
			break
		}

		if err != nil {
			log.Fatalf("Download error: %s", err)
		}

		if downloadResp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s", downloadResp.GetProgress())
		}
	}
}

func callLibInstall(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance, version string) {
	installRespStream, err := client.LibraryInstall(context.Background(),
		&rpc.LibraryInstallRequest{
			Instance: instance,
			Name:     "WiFi101",
			Version:  version,
		})

	if err != nil {
		log.Fatalf("Error installing library: %s", err)
	}

	for {
		installResp, err := installRespStream.Recv()
		if err == io.EOF {
			log.Print("Lib install done")
			break
		}

		if err != nil {
			log.Fatalf("Install error: %s", err)
		}

		if installResp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s\n", installResp.GetProgress())
		}
		if installResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s\n", installResp.GetTaskProgress())
		}
	}
}

func callLibInstallNoDeps(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance, version string) {
	installRespStream, err := client.LibraryInstall(context.Background(),
		&rpc.LibraryInstallRequest{
			Instance: instance,
			Name:     "Arduino_MKRIoTCarrier",
			Version:  version,
			NoDeps:   true,
		})

	if err != nil {
		log.Fatalf("Error installing library: %s", err)
	}

	for {
		installResp, err := installRespStream.Recv()
		if err == io.EOF {
			log.Print("Lib install done")
			break
		}

		if err != nil {
			log.Fatalf("Install error: %s", err)
		}

		if installResp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s\n", installResp.GetProgress())
		}
		if installResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s\n", installResp.GetTaskProgress())
		}
	}
}
func callLibUpgradeAll(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	libUpgradeAllRespStream, err := client.LibraryUpgradeAll(context.Background(),
		&rpc.LibraryUpgradeAllRequest{
			Instance: instance,
		})

	if err != nil {
		log.Fatalf("Error upgrading all: %s\n", err)
	}

	for {
		resp, err := libUpgradeAllRespStream.Recv()
		if err == io.EOF {
			log.Printf("Lib upgrade all done")
			break
		}

		if err != nil {
			log.Fatalf("Upgrading error: %s", err)
		}

		if resp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s\n", resp.GetProgress())
		}
		if resp.GetTaskProgress() != nil {
			log.Printf("TASK: %s\n", resp.GetTaskProgress())
		}
	}
}

func callLibSearch(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	libSearchResp, err := client.LibrarySearch(context.Background(),
		&rpc.LibrarySearchRequest{
			Instance: instance,
			Query:    "audio",
		})

	if err != nil {
		log.Fatalf("Error searching for library: %s", err)
	}

	for _, res := range libSearchResp.GetLibraries() {
		log.Printf("Result: %s - %s", res.GetName(), res.GetLatest().GetVersion())
	}
}

func callLibraryResolveDependencies(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	libraryResolveDependenciesResp, err := client.LibraryResolveDependencies(context.Background(),
		&rpc.LibraryResolveDependenciesRequest{
			Instance: instance,
			Name:     "ArduinoIoTCloud",
		})

	if err != nil {
		log.Fatalf("Error listing library dependencies: %s", err)
	}

	for _, resp := range libraryResolveDependenciesResp.GetDependencies() {
		log.Printf("Dependency Name: %s", resp.GetName())
		log.Printf("Version Required: %s", resp.GetVersionRequired())
		if resp.GetVersionInstalled() != "" {
			log.Printf("Version Installed: %s\n", resp.GetVersionInstalled())
		}
	}
}

func callLibList(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	libLstResp, err := client.LibraryList(context.Background(),
		&rpc.LibraryListRequest{
			Instance:  instance,
			All:       false,
			Updatable: false,
		})

	if err != nil {
		log.Fatalf("Error List Library: %s", err)
	}

	for _, res := range libLstResp.GetInstalledLibraries() {
		log.Printf("%s - %s", res.GetLibrary().GetName(), res.GetLibrary().GetVersion())
	}
}

func callLibUninstall(client rpc.ArduinoCoreServiceClient, instance *rpc.Instance) {
	libUninstallRespStream, err := client.LibraryUninstall(context.Background(),
		&rpc.LibraryUninstallRequest{
			Instance: instance,
			Name:     "WiFi101",
		})

	if err != nil {
		log.Fatalf("Error uninstalling: %s", err)
	}

	for {
		uninstallResp, err := libUninstallRespStream.Recv()
		if err == io.EOF {
			log.Printf("Lib uninstall done")
			break
		}

		if err != nil {
			log.Fatalf("Uninstall error: %s", err)
		}

		if uninstallResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s", uninstallResp.GetTaskProgress())
		}
	}
}

func callDebugger(debugStreamingOpenClient dbg.DebugService_DebugClient, instance *rpc.Instance) {
	currDir, _ := os.Getwd()
	log.Printf("Send debug request")
	err := debugStreamingOpenClient.Send(&dbg.DebugRequest{
		DebugRequest: &dbg.DebugConfigRequest{
			Instance:   &rpc.Instance{Id: instance.GetId()},
			Fqbn:       "arduino:samd:mkr1000",
			SketchPath: filepath.Join(currDir, "hello"),
			Port:       "none",
		}})
	if err != nil {
		log.Fatalf("Send error: %s\n", err)
	}
	// Loop and consume the server stream until all the operations are done.
	waitForPrompt(debugStreamingOpenClient, "(gdb)")
	// Wait for gdb to init and show the prompt
	log.Printf("Send 'info registers' rcommand")
	err = debugStreamingOpenClient.Send(&dbg.DebugRequest{Data: []byte("info registers\n")})
	if err != nil {
		log.Fatalf("Send error: %s\n", err)
	}

	// Loop and consume the server stream until all the operations are done.
	waitForPrompt(debugStreamingOpenClient, "(gdb)")

	// Send quit command to gdb
	log.Printf("Send 'quit' command")
	err = debugStreamingOpenClient.Send(&dbg.DebugRequest{Data: []byte("quit\n")})
	if err != nil {
		log.Fatalf("Send error: %s\n", err)
	}

	// Close connection with the debug server
	log.Printf("Close session")
	err = debugStreamingOpenClient.CloseSend()
	if err != nil {
		log.Fatalf("Send error: %s\n", err)
	}
}

func waitForPrompt(debugStreamingOpenClient dbg.DebugService_DebugClient, prompt string) {
	var buffer bytes.Buffer
	for {
		compResp, err := debugStreamingOpenClient.Recv()

		// There was an error.
		if err != nil {
			log.Fatalf("debug error: %s\n", err)
		}

		// Consume output and search for the gdb prompt to exit the loop
		if resp := compResp.GetData(); resp != nil {
			fmt.Printf("%s", resp)
			buffer.Write(resp)
			if strings.Contains(buffer.String(), prompt) {
				break
			}
		}
	}
}

func callLoadSketch(client rpc.ArduinoCoreServiceClient) {
	currDir, _ := os.Getwd()
	sketch, err := client.LoadSketch(context.Background(), &rpc.LoadSketchRequest{
		SketchPath: filepath.Join(currDir, "hello"),
	})
	if err != nil {
		log.Fatalf("Error getting version: %s", err)
	}

	log.Printf("Sketch main file: %s", sketch.GetMainFile())
	log.Printf("Sketch location: %s", sketch.GetLocationPath())
	log.Printf("Other sketch files: %v", sketch.GetOtherSketchFiles())
	log.Printf("Sketch additional files: %v", sketch.GetAdditionalFiles())
}
