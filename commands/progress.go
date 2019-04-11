package commands

import "github.com/arduino/arduino-cli/rpc"

// ProgressCB is a callback to get updates on download progress
type ProgressCB func(curr *rpc.DownloadProgress)

// TaskProgressCB is a callback to receive progress messages
type TaskProgressCB func(msg *rpc.TaskProgress)
