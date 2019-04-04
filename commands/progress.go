package commands

import "github.com/arduino/arduino-cli/rpc"

// ProgressCB is a callback to get updates on download progress
type ProgressCB func(curr *rpc.DownloadProgress)
