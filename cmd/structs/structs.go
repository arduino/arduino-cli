package structs

import (
	"fmt"
	"strings"
)

//LibProcessResults represent the result of a process on libraries.
type LibProcessResults struct {
	Libraries []libProcessResult `json:"libraries"`
}

// String returns a string representation of the object.
func (lpr LibProcessResults) String() string {
	ret := ""
	for _, lr := range lpr.Libraries {
		ret += fmt.Sprintln(lr)
	}
	return strings.TrimSpace(ret)
}

//libProcessResult contains info about a completed process.
type libProcessResult struct {
	LibraryName string `json:"libraryName"`
	Result      string `json:"result"`
}

// String returns a string representation of the object.
func (lr libProcessResult) String() string {
	return strings.TrimSpace(fmt.Sprintf("%s - %s", lr.LibraryName, lr.Result))
}
