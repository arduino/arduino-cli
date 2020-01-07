// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

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
