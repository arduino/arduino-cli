/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package builderClient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ByVidPidBoardsV2Path computes a request path to the byVidPid action of boards_v2.
func ByVidPidBoardsV2Path(vid string, pid string) string {
	param0 := vid
	param1 := pid

	return fmt.Sprintf("/builder/v2/boards/byVidPid/%s/%s", param0, param1)
}

// Provides the board identified by the :vid and :pid params. Doesn't require authentication.
func (c *Client) ByVidPidBoardsV2(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewByVidPidBoardsV2Request(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewByVidPidBoardsV2Request create the request corresponding to the byVidPid action endpoint of the boards_v2 resource.
func (c *Client) NewByVidPidBoardsV2Request(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// ListBoardsV2Path computes a request path to the list action of boards_v2.
func ListBoardsV2Path() string {

	return fmt.Sprintf("/builder/v2/boards")
}

// Provides a list of all the boards supported by Arduino Create. Doesn't require any authentication.
func (c *Client) ListBoardsV2(ctx context.Context, path string, limit *int, offset *int) (*http.Response, error) {
	req, err := c.NewListBoardsV2Request(ctx, path, limit, offset)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListBoardsV2Request create the request corresponding to the list action endpoint of the boards_v2 resource.
func (c *Client) NewListBoardsV2Request(ctx context.Context, path string, limit *int, offset *int) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	if limit != nil {
		tmp15 := strconv.Itoa(*limit)
		values.Set("limit", tmp15)
	}
	if offset != nil {
		tmp16 := strconv.Itoa(*offset)
		values.Set("offset", tmp16)
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// ShowBoardsV2Path computes a request path to the show action of boards_v2.
func ShowBoardsV2Path(fqbn string) string {
	param0 := fqbn

	return fmt.Sprintf("/builder/v2/boards/%s", param0)
}

// Provides the board identified by an fqbn. Doesn't require authentication.
func (c *Client) ShowBoardsV2(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowBoardsV2Request(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowBoardsV2Request create the request corresponding to the show action endpoint of the boards_v2 resource.
func (c *Client) NewShowBoardsV2Request(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}
