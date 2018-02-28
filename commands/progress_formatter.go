package commands

import (
	"gopkg.in/cheggaaa/pb.v1"
	"fmt"

	"github.com/sirupsen/logrus"
)

// ProgressBarFormatter implements the visualization of the progress bars
// to display the progress of a ParallelDownload task set (i.e. one bar per file)
// WARNING: The implementation library is experimental and unstable; Do not print to terminal while the bars are active.
type ProgressBarFormatter struct {
	progressBars     map[string]*pb.ProgressBar
	progressBarsPool *pb.Pool
}

// Implement interface releases.ParallelDownloadProgressHandler

func (pbf *ProgressBarFormatter) OnNewDownloadTask(fileName string, fileSize int64) {
	// Initialize progress bars and a new one for each the new task
	if pbf.progressBars == nil {
		pbf.progressBars = map[string]*pb.ProgressBar{}
	}

	logrus.Debug(fmt.Sprintf("Initializing progress bar for file '%s'", fileName))

	// Initialization is in bytes, to display full information about the file (not only the percentage)
	progressBar := pb.New64(fileSize).SetUnits(pb.U_BYTES).Prefix(fmt.Sprintf("%-20s", fileName)).Start()
	pbf.progressBars[fileName] = progressBar
}

func (pbf *ProgressBarFormatter) OnProgressChanged(fileName string, fileSize int64, downloadedSoFar int64) {
	// Update a specific file's progress bar
	progressBar := pbf.progressBars[fileName]

	if progressBar != nil {
		progressBar.Set(int(downloadedSoFar))
	} else {
		logrus.Debug(fmt.Sprintf("Progress bar for file '%s' not found", fileName))
	}
}

func (pbf *ProgressBarFormatter) OnDownloadStarted() {
	// WARNING!!
	// (experimental and unstable) Do not print to terminal while pool is active.
	// See https://github.com/cheggaaa/pb#multiple-progress-bars-experimental-and-unstable

	// Start the progress bar pool
	progressBarsAsSlice := []*pb.ProgressBar{}
	for _, value := range pbf.progressBars {
		progressBarsAsSlice = append(progressBarsAsSlice, value)
	}
	pbf.progressBarsPool, _ = pb.StartPool(progressBarsAsSlice...)
}

func (pbf *ProgressBarFormatter) OnDownloadFinished() {
	// Stop the progress bar pool
	if pbf.progressBarsPool != nil {
		pbf.progressBarsPool.Stop()
	}
}

// END -- Implement interface releases.ParallelDownloadProgressHandler
