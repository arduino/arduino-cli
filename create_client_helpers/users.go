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

package createClient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// EditUsersPath computes a request path to the edit action of users.
func EditUsersPath(id string) string {
	param0 := id

	return fmt.Sprintf("/create/v1/users/%s", param0)
}

// EditUsers creates or modifies the user identified by the :id param.
// Requires the ~create:users scope except for the special id "me" that provides the info about the authenticated user
func (c *Client) EditUsers(ctx context.Context, path string, payload *User, authorization string) (*http.Response, error) {
	req, err := c.NewEditUsersRequest(ctx, path, payload, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewEditUsersRequest create the request corresponding to the edit action endpoint of the users resource.
func (c *Client) NewEditUsersRequest(ctx context.Context, path string, payload *User, authorization string) (*http.Request, error) {
	var body bytes.Buffer
	err := c.Encoder.Encode(payload, &body, "*/*")
	if err != nil {
		return nil, fmt.Errorf("failed to encode body: %s", err)
	}
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("POST", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")

	header.Set("Authorization", authorization)

	return req, nil
}

// ShowUsersPath computes a request path to the show action of users.
func ShowUsersPath(id string) string {
	return fmt.Sprintf("/create/v1/users/%s", id)
}

// ShowUsers provides the user identified by the :id param.
// Requires the ~create:users scope except for the special id "me" that provides the info about the authenticated user
func (c *Client) ShowUsers(ctx context.Context, path string, authorization string) (*http.Response, error) {
	req, err := c.NewShowUsersRequest(ctx, path, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowUsersRequest create the request corresponding to the show action endpoint of the users resource.
func (c *Client) NewShowUsersRequest(ctx context.Context, path string, authorization string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	header := req.Header

	header.Set("Authorization", authorization)

	return req, nil
}

// StatsUsersPath computes a request path to the stats action of users.
func StatsUsersPath(id string) string {
	param0 := id

	return fmt.Sprintf("/create/v1/users/%s/stats", param0)
}

// StatsUsers provides the stats for the user identified by the :id param.
// Requires the ~create:users scope except for the special id "me" that provides the info about the authenticated user
func (c *Client) StatsUsers(ctx context.Context, path string, authorization string) (*http.Response, error) {
	req, err := c.NewStatsUsersRequest(ctx, path, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewStatsUsersRequest create the request corresponding to the stats action endpoint of the users resource.
func (c *Client) NewStatsUsersRequest(ctx context.Context, path string, authorization string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	header := req.Header

	header.Set("Authorization", authorization)

	return req, nil
}
