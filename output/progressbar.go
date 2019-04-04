//

package output

import (
	"github.com/arduino/arduino-cli/rpc"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// DownloadProgressBar returns a progress bar callback
func DownloadProgressBar() func(*rpc.DownloadProgress) {
	var bar *pb.ProgressBar
	var prefix string
	return func(curr *rpc.DownloadProgress) {
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
