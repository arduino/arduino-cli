package global

import (
	"os"
	"path/filepath"
)

// appName is the command line name of the Arduino CLI executable on the user system (users may change it)
var appName = filepath.Base(os.Args[0])

var (
	application= "arduino-cli"
	version    = "missing"
	commit     = "missing"
	cvsRef     = "missing"
	buildDate  = "missing"
	repository = "missing"
)


func GetAppName() string {
	return appName
}

func GetApplication() string {
	return application
}

func GetVersion() string {
	return version
}

func GetCommit() string {
	return commit
}

func GetCvsRef() string {
	return cvsRef
}

func GetBuildDate() string {
	return buildDate
}

func GetRepository() string {
	return repository
}

type Info struct {
	Application string `json:"application"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	CvsRef      string `json:"cvsRef"`
	BuildDate   string `json:"buildDate"`
	Repository  string `json:"repository"`
}

