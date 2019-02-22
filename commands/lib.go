//

package commands

/*
import (
	"fmt"
	"context"
)

func (s *Service) ListLibraries(ctx context.Context, in *ListLibrariesReq) (*ListLibrariesResp, error) {
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
*/
