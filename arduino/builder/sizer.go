package builder

import rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"

// ExecutableSectionSize represents a section of the executable output file
type ExecutableSectionSize struct {
	Name    string `json:"name"`
	Size    int    `json:"size"`
	MaxSize int    `json:"max_size"`
}

// ExecutablesFileSections is an array of ExecutablesFileSection
type ExecutablesFileSections []ExecutableSectionSize

// ToRPCExecutableSectionSizeArray transforms this array into a []*rpc.ExecutableSectionSize
func (s ExecutablesFileSections) ToRPCExecutableSectionSizeArray() []*rpc.ExecutableSectionSize {
	res := []*rpc.ExecutableSectionSize{}
	for _, section := range s {
		res = append(res, &rpc.ExecutableSectionSize{
			Name:    section.Name,
			Size:    int64(section.Size),
			MaxSize: int64(section.MaxSize),
		})
	}
	return res
}
