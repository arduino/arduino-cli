//

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/output"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
)

// OutputJSONOrElse outputs the JSON encoding of v if the JSON output format has been
// selected by the user and returns false. Otherwise no output is produced and the
// function returns true.
func OutputJSONOrElse(v interface{}) bool {
	if !GlobalFlags.OutputJSON {
		return true
	}
	d, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		formatter.PrintError(err, "Error during JSON encoding of the output")
		os.Exit(ErrGeneric)
	}
	fmt.Print(string(d))
	return false
}

// OutputProgressBar returns a DownloadProgressCB that prints a progress bar.
// If JSON output format has been selected, the callback outputs nothing.
func OutputProgressBar() commands.DownloadProgressCB {
	if !GlobalFlags.OutputJSON {
		return output.NewDownloadProgressBarCB()
	}
	return func(curr *rpc.DownloadProgress) {
		// XXX: Output progress in JSON?
	}
}

// OutputTaskProgress returns a TaskProgressCB that prints the task progress.
// If JSON output format has been selected, the callback outputs nothing.
func OutputTaskProgress() commands.TaskProgressCB {
	if !GlobalFlags.OutputJSON {
		return output.NewTaskProgressCB()
	}
	return func(curr *rpc.TaskProgress) {
		// XXX: Output progress in JSON?
	}
}
