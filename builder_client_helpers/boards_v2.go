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

package builderclient

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
