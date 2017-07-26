package output

import (
	"fmt"
	"strings"
)

//ProcessResult contains info about a completed process.
type ProcessResult struct {
	ItemName string `json:"name,required"`
	Status   string `json:"status,omitempty"`
	Error    string `json:"error,omitempty"`
	Path     string `json:"path,omitempty"`
}

// String returns a string representation of the object.
//   EXAMPLE:
//   ToolName - ErrorText: Error explaining why failed
//   ToolName - StatusText: PATH = /path/to/result/folder
func (lr ProcessResult) String() string {
	ret := fmt.Sprintf("%s - %s", lr.ItemName, lr.Status)
	if lr.Error != "" {
		ret += fmt.Sprint(": ", lr.Error)
	} else if lr.Path != "" {
		ret += fmt.Sprint(": PATH = ", lr.Path)
	}
	return strings.TrimSpace(ret)
}
