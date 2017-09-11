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
)

// AddPinnedLibrariesPath computes a request path to the add action of pinnedLibraries.
func AddPinnedLibrariesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/builder/pinned/%s", param0)
}

//AddPinnedLibraries adds a new library to the list of libraries pinned by the user
func (c *Client) AddPinnedLibraries(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewAddPinnedLibrariesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewAddPinnedLibrariesRequest create the request corresponding to the add action endpoint of the pinnedLibraries resource.
func (c *Client) NewAddPinnedLibrariesRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("PUT", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.Oauth2Signer != nil {
		c.Oauth2Signer.Sign(req)
	}
	return req, nil
}

// ListPinnedLibrariesPath computes a request path to the list action of pinnedLibraries.
func ListPinnedLibrariesPath() string {
	return fmt.Sprintf("/builder/pinned")
}

//ListPinnedLibraries provides a list of all the libraries pinned by the user
func (c *Client) ListPinnedLibraries(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewListPinnedLibrariesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListPinnedLibrariesRequest create the request corresponding to the list action endpoint of the pinnedLibraries resource.
func (c *Client) NewListPinnedLibrariesRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.Oauth2Signer != nil {
		c.Oauth2Signer.Sign(req)
	}
	return req, nil
}

// RemovePinnedLibrariesPath computes a request path to the remove action of pinnedLibraries.
func RemovePinnedLibrariesPath(id string) string {
	param0 := id

	return fmt.Sprintf("/builder/pinned/%s", param0)
}

//RemovePinnedLibraries removes a library to the list of libraries pinned by the user
func (c *Client) RemovePinnedLibraries(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewRemovePinnedLibrariesRequest(ctx, path)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewRemovePinnedLibrariesRequest create the request corresponding to the remove action endpoint of the pinnedLibraries resource.
func (c *Client) NewRemovePinnedLibrariesRequest(ctx context.Context, path string) (*http.Request, error) {
	scheme := c.Scheme
	if scheme == "" {
		scheme = "http"
	}
	u := url.URL{Host: c.Host, Scheme: scheme, Path: path}
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.Oauth2Signer != nil {
		c.Oauth2Signer.Sign(req)
	}
	return req, nil
}
