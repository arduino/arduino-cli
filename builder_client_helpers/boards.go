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
)

// ListBoardsPath computes a request path to the list action of boards.
func ListBoardsPath() string {
	return fmt.Sprintf("/builder/v1/boards")
}

// ListBoards provides a list of all the boards supported by Arduino Create. Doesn't require any authentication.
func (c *Client) ListBoards(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewListBoardsRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListBoardsRequest create the request corresponding to the list action endpoint of the boards resource.
func (c *Client) NewListBoardsRequest(ctx context.Context, path string) (*http.Request, error) {
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

// ShowBoardsPath computes a request path to the show action of boards.
func ShowBoardsPath(vid string, pid string) string {
	param0 := vid
	param1 := pid

	return fmt.Sprintf("/builder/v1/boards/%s/%s", param0, param1)
}

// ShowBoards provides the board identified by the :vid and :pid params. Doesn't require authentication.
func (c *Client) ShowBoards(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowBoardsRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowBoardsRequest create the request corresponding to the show action endpoint of the boards resource.
func (c *Client) NewShowBoardsRequest(ctx context.Context, path string) (*http.Request, error) {
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
