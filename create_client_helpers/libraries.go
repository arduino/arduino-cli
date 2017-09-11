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
	"strconv"
)

// CreateLibrariesPath computes a request path to the create action of libraries.
func CreateLibrariesPath() string {
	return fmt.Sprintf("/create/v1/libraries")
}

// CreateLibraries adds a new library.
func (c *Client) CreateLibraries(ctx context.Context, path string, payload *Library, force *bool, authorization string) (*http.Response, error) {
	req, err := c.NewCreateLibrariesRequest(ctx, path, payload, force, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewCreateLibrariesRequest create the request corresponding to the create action endpoint of the libraries resource.
func (c *Client) NewCreateLibrariesRequest(ctx context.Context, path string, payload *Library, force *bool, authorization string) (*http.Request, error) {
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
	values := u.Query()
	if force != nil {
		tmp17 := strconv.FormatBool(*force)
		values.Set("force", tmp17)
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("PUT", u.String(), &body)
	if err != nil {
		return nil, err
	}
	header := req.Header
	header.Set("Content-Type", "application/json")

	header.Set("Authorization", authorization)

	return req, nil
}

// DeleteLibrariesPath computes a request path to the delete action of libraries.
func DeleteLibrariesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/create/v1/libraries/%s", param0)
}

// DeleteLibraries removes the library identified by the :id param.
func (c *Client) DeleteLibraries(ctx context.Context, path string, authorization string) (*http.Response, error) {
	req, err := c.NewDeleteLibrariesRequest(ctx, path, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewDeleteLibrariesRequest create the request corresponding to the delete action endpoint of the libraries resource.
func (c *Client) NewDeleteLibrariesRequest(ctx context.Context, path string, authorization string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, err
	}
	header := req.Header

	header.Set("Authorization", authorization)

	return req, nil
}

// EditLibrariesPath computes a request path to the edit action of libraries.
func EditLibrariesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/create/v1/libraries/%s", param0)
}

// EditLibraries modifies the library identified by the :id param.
// If a file has a valid data field, it will be modified too.
func (c *Client) EditLibraries(ctx context.Context, path string, payload *Library, authorization string) (*http.Response, error) {
	req, err := c.NewEditLibrariesRequest(ctx, path, payload, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewEditLibrariesRequest create the request corresponding to the edit action endpoint of the libraries resource.
func (c *Client) NewEditLibrariesRequest(ctx context.Context, path string, payload *Library, authorization string) (*http.Request, error) {
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

// SearchLibrariesPath computes a request path to the search action of libraries.
func SearchLibrariesPath() string {

	return fmt.Sprintf("/create/v1/libraries")
}

// SearchLibraries provides a paginated list of libraries filtered according to the params. The page size is 100 items.
func (c *Client) SearchLibraries(ctx context.Context, path string, offset *string, owner *string, authorization *string) (*http.Response, error) {
	req, err := c.NewSearchLibrariesRequest(ctx, path, offset, owner, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewSearchLibrariesRequest create the request corresponding to the search action endpoint of the libraries resource.
func (c *Client) NewSearchLibrariesRequest(ctx context.Context, path string, offset *string, owner *string, authorization *string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	if offset != nil {
		values.Set("offset", *offset)
	}
	if owner != nil {
		values.Set("owner", *owner)
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	header := req.Header
	if authorization != nil {

		header.Set("Authorization", *authorization)
	}
	return req, nil
}

// ShowLibrariesPath computes a request path to the show action of libraries.
func ShowLibrariesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/create/v1/libraries/%s", param0)
}

// ShowLibraries provides the library identified by the :id param.
func (c *Client) ShowLibraries(ctx context.Context, path string, authorization *string) (*http.Response, error) {
	req, err := c.NewShowLibrariesRequest(ctx, path, authorization)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowLibrariesRequest create the request corresponding to the show action endpoint of the libraries resource.
func (c *Client) NewShowLibrariesRequest(ctx context.Context, path string, authorization *string) (*http.Request, error) {
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
	if authorization != nil {

		header.Set("Authorization", *authorization)
	}
	return req, nil
}
