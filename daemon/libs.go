//

package daemon

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/commands/lib"

	pb "github.com/arduino/arduino-cli/daemon/arduino"
)

func (s *daemon) ListLibraries(ctx context.Context, in *pb.ListLibrariesReq) (*pb.ListLibrariesResp, error) {
	if in.Instance == nil {
		return nil, fmt.Errorf("invalid request")
	}
	instance, ok := instances[in.Instance.Id]
	if !ok {
		return nil, fmt.Errorf("instance not found")
	}
	libs := lib.ListLibraries(instance.lm, in.Updatable)

	result := []*pb.Library{}
	for _, lib := range libs.Libraries {
		result = append(result, &pb.Library{
			Name:        lib.Library.Name,
			Paragraph:   lib.Library.Paragraph,
			Precompiled: lib.Library.Precompiled,
		})
	}
	return &pb.ListLibrariesResp{
		Libraries: result,
	}, nil
}
