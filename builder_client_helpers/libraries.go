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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package builderClient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListLibrariesPath computes a request path to the list action of libraries.
func ListLibrariesPath() string {

	return fmt.Sprintf("/builder/v1/libraries")
}

// ListLibraries provides a list of all the latest versions of the libraries supported by Arduino Create. Doesn't require any authentication.
func (c *Client) ListLibraries(ctx context.Context, path string, maintainer *string, type_ *string, withoutType *string) (*http.Response, error) {
	req, err := c.NewListLibrariesRequest(ctx, path, maintainer, type_, withoutType)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListLibrariesRequest create the request corresponding to the list action endpoint of the libraries resource.
func (c *Client) NewListLibrariesRequest(ctx context.Context, path string, maintainer *string, type_ *string, withoutType *string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	values := u.Query()
	if maintainer != nil {
		values.Set("maintainer", *maintainer)
	}
	if type_ != nil {
		values.Set("type", *type_)
	}
	if withoutType != nil {
		values.Set("without_type", *withoutType)
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// ShowLibrariesPath computes a request path to the show action of libraries.
func ShowLibrariesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/builder/v1/libraries/%s", param0)
}

// Provides the library identified by the :id and :pid param. Doesn't require authentication. Also contains a list of other versions of the library
func (c *Client) ShowLibraries(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewShowLibrariesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewShowLibrariesRequest create the request corresponding to the show action endpoint of the libraries resource.
func (c *Client) NewShowLibrariesRequest(ctx context.Context, path string) (*http.Request, error) {
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
