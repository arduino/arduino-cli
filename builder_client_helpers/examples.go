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

// ListExamplesPath computes a request path to the list action of examples.
func ListExamplesPath() string {
	return fmt.Sprintf("/builder/v1/examples")
}

// ListExamples provides a list of all the builtin examples
func (c *Client) ListExamples(ctx context.Context, path string, maintainer *string, type_ *string) (*http.Response, error) {
	req, err := c.NewListExamplesRequest(ctx, path, maintainer, type_)
	if err != nil {
		return nil, err
	}
	return c.Client.Do(ctx, req)
}

// NewListExamplesRequest create the request corresponding to the list action endpoint of the examples resource.
func (c *Client) NewListExamplesRequest(ctx context.Context, path string, maintainer *string, type_ *string) (*http.Request, error) {
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
	u.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}
