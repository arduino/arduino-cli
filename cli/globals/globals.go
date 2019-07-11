package globals

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/arduino-cli/version"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

var (
	// Debug determines whether to dump debug output to stderr or not
	Debug bool
	// OutputJSON is true output in JSON, false output as Text
	OutputJSON bool
	// HTTPClientHeader is the object that will be propagated to configure the clients inside the downloaders
	HTTPClientHeader = getHTTPClientHeader()
	// VersionInfo contains all info injected during build
	VersionInfo = version.NewInfo(filepath.Base(os.Args[0]))
	// Config FIXMEDOC
	Config *configs.Configuration
	// YAMLConfigFile contains the path to the config file
	YAMLConfigFile string
)

func getHTTPClientHeader() http.Header {
	userAgentValue := fmt.Sprintf("%s/%s (%s; %s; %s) Commit:%s/Build:%s", VersionInfo.Application,
		VersionInfo.VersionString, runtime.GOARCH, runtime.GOOS, runtime.Version(), VersionInfo.Commit, VersionInfo.BuildDate)
	downloaderHeaders := http.Header{"User-Agent": []string{userAgentValue}}
	return downloaderHeaders
}

// InitConfigs initializes the configuration from the specified file.
func InitConfigs() {
	// Start with default configuration
	if conf, err := configs.NewConfiguration(); err != nil {
		logrus.WithError(err).Error("Error creating default configuration")
		formatter.PrintError(err, "Error creating default configuration")
		os.Exit(errorcodes.ErrGeneric)
	} else {
		Config = conf
	}

	// Read configuration from global config file
	logrus.Info("Checking for config file in: " + Config.ConfigFile.String())
	if Config.ConfigFile.Exist() {
		readConfigFrom(Config.ConfigFile)
	}

	if Config.IsBundledInDesktopIDE() {
		logrus.Info("CLI is bundled into the IDE")
		err := Config.LoadFromDesktopIDEPreferences()
		if err != nil {
			logrus.WithError(err).Warn("Did not manage to get config file of IDE, using default configuration")
		}
	} else {
		logrus.Info("CLI is not bundled into the IDE")
	}

	// Read configuration from parent folders (project config)
	if pwd, err := paths.Getwd(); err != nil {
		logrus.WithError(err).Warn("Did not manage to find current path")
		if path := paths.New("arduino-yaml"); path.Exist() {
			readConfigFrom(path)
		}
	} else {
		Config.Navigate(pwd)
	}

	// Read configuration from old configuration file if found, but output a warning.
	if old := paths.New(".cli-config.yml"); old.Exist() {
		logrus.Errorf("Old configuration file detected: %s.", old)
		logrus.Info("The name of this file has been changed to `arduino-yaml`, please rename the file fix it.")
		formatter.PrintError(
			fmt.Errorf("WARNING: Old configuration file detected: %s", old),
			"The name of this file has been changed to `arduino-yaml`, in a future release we will not support"+
				"the old name `.cli-config.yml` anymore. Please rename the file to `arduino-yaml` to silence this warning.")
		readConfigFrom(old)
	}

	// Read configuration from environment vars
	Config.LoadFromEnv()

	// Read configuration from user specified file
	if YAMLConfigFile != "" {
		Config.ConfigFile = paths.New(YAMLConfigFile)
		readConfigFrom(Config.ConfigFile)
	}

	logrus.Info("Configuration set")
}

func readConfigFrom(path *paths.Path) {
	logrus.Infof("Reading configuration from %s", path)
	if err := Config.LoadFromYAML(path); err != nil {
		logrus.WithError(err).Warnf("Could not read configuration from %s", path)
	}
}
