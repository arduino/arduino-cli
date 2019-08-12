/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package board

import (
	"context"
	"errors"
	"strings"

	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// ListAll FIXMEDOC
func ListAll(ctx context.Context, req *rpc.BoardListAllReq) (*rpc.BoardListAllResp, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	args := req.GetSearchArgs()
	match := func(name string) bool {
		if len(args) == 0 {
			return true
		}
		name = strings.ToLower(name)
		for _, term := range args {
			if !strings.Contains(name, strings.ToLower(term)) {
				return false
			}
		}
		return true
	}

	list := &rpc.BoardListAllResp{Boards: []*rpc.BoardListItem{}}
	for _, targetPackage := range pm.Packages {
		for _, platform := range targetPackage.Platforms {
			platformRelease := pm.GetInstalledPlatformRelease(platform)
			if platformRelease == nil {
				continue
			}
			for _, board := range platformRelease.Boards {
				if !match(board.Name()) {
					continue
				}
				list.Boards = append(list.Boards, &rpc.BoardListItem{
					Name: board.Name(),
					FQBN: board.FQBN(),
				})
			}
		}
	}

	return list, nil
}
