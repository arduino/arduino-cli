//

package daemon

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/configs"
	pb "github.com/arduino/arduino-cli/daemon/arduino"
	paths "github.com/arduino/go-paths-helper"
)

// this map contains all the running Arduino Core Services instances
// referenced by an int32 handle
var instances = map[int32]*Instance{}
var instancesCount int32 = 1

// Instance is an instance of the Arduino Core Services. The daemon can
// instantiate as many as needed by providing a different configuration
// for each one.
type Instance struct {
	config *configs.Configuration
	pm     *packagemanager.PackageManager
	lm     *librariesmanager.LibrariesManager
}

func (s *daemon) Init(ctx context.Context, req *pb.InitReq) (*pb.InitResp, error) {
	inConfig := req.Configuration
	if inConfig == nil {
		return nil, fmt.Errorf("invalid request")
	}
	log.Printf("Received: %v", inConfig)
	config := &configs.Configuration{
		DataDir:       paths.New(inConfig.DataDir),
		SketchbookDir: paths.New(inConfig.SketchbookDir),
	}
	if inConfig.DownloadsDir != "" {
		config.ArduinoDownloadsDir = paths.New(inConfig.DownloadsDir)
	}
	urls := []*url.URL{}
	for _, rawurl := range inConfig.BoardManagerAdditionalUrls {
		if u, err := url.Parse(rawurl); err == nil {
			urls = append(urls, u)
		} else {
			return nil, fmt.Errorf("parsing url %s: %s", rawurl, err)
		}
	}

	pm := packagemanager.NewPackageManager(
		config.IndexesDir(),
		config.PackagesDir(),
		config.DownloadsDir(),
		config.DataDir.Join("tmp"))

	for _, URL := range config.BoardManagerAdditionalUrls {
		if err := pm.LoadPackageIndex(URL); err != nil {
			return nil, fmt.Errorf("loading "+URL.String()+" package index: %s", err)
		}
	}

	if err := pm.LoadHardware(config); err != nil {
		return nil, fmt.Errorf("loading hardware packages: %s", err)
	}

	lm := cli.InitLibraryManager(config, nil)

	instance := &Instance{config: config, pm: pm, lm: lm}
	handle := instancesCount
	instancesCount++
	instances[handle] = instance

	return &pb.InitResp{
		Instance: &pb.Instance{Id: handle},
	}, nil
}

func (s *daemon) Destroy(ctx context.Context, req *pb.DestroyReq) (*pb.DestroyResp, error) {
	id := req.Instance.Id
	_, ok := instances[id]
	if !ok {
		return nil, fmt.Errorf("invalid handle")
	}
	delete(instances, id)
	return &pb.DestroyResp{}, nil
}
