package commands

import "github.com/arduino/arduino-cli/rpc"

// DownloadProgressCB is a callback to get updates on download progress
type DownloadProgressCB func(curr *rpc.DownloadProgress)

// TaskProgressCB is a callback to receive progress messages
type TaskProgressCB func(msg *rpc.TaskProgress)
