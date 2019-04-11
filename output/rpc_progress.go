//

package output

import (
	"fmt"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/rpc"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// DownloadProgressBar returns a progress bar callback
func DownloadProgressBar() func(*rpc.DownloadProgress) {
	var bar *pb.ProgressBar
	var prefix string
	return func(curr *rpc.DownloadProgress) {
		// fmt.Printf(">>> %v\n", curr)
		if filename := curr.GetFile(); filename != "" {
			prefix = filename
			bar = pb.StartNew(int(curr.GetTotalSize()))
			bar.Prefix(prefix)
			bar.SetUnits(pb.U_BYTES)
		}
		if curr.GetDownloaded() != 0 {
			bar.Set(int(curr.GetDownloaded()))
		}
		if curr.GetCompleted() {
			bar.FinishPrintOver(prefix + " downloaded")
		}
	}
}

// NewTaskProgressCB returns a commands.TaskProgressCB progress listener
// that outputs to terminal
func NewTaskProgressCB() commands.TaskProgressCB {
	var name string
	return func(curr *rpc.TaskProgress) {
		// fmt.Printf(">>> %v\n", curr)
		msg := curr.GetMessage()
		if curr.GetName() != "" {
			name = curr.GetName()
			if msg == "" {
				msg = name
			}
		}
		if msg != "" {
			fmt.Print(msg)
			if curr.GetCompleted() {
				fmt.Println()
			} else {
				fmt.Println("...")
			}
		}
	}
}
