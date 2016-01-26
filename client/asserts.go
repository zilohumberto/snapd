// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ubuntu-core/snappy/asserts" // for parsing
)

// Assert tries to add an assertion to the system assertion
// database. To succeed the assertion must be valid, its signature
// verified with a known public key and the assertion consistent with
// and its prerequisite in the database.
func (client *Client) Assert(b []byte) error {
	var rsp interface{}
	if err := client.doSync("POST", "/2.0/assertions", bytes.NewReader(b), &rsp); err != nil {
		return fmt.Errorf("cannot assert: %v", err)
	}

	return nil
}

// Asserts queries assertions with assertTypeName and matching headers.
func (client *Client) Asserts(assertTypeName string, headers map[string]string) ([]asserts.Assertion, error) {
	u := url.URL{Path: fmt.Sprintf("/2.0/assertions/%s", assertTypeName)}

	if len(headers) > 0 {
		q := u.Query()
		for k, v := range headers {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	response, err := client.raw("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query assertions: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		// TODO: distinguish kinds of not found ?
		return nil, nil
	}

	if response.StatusCode != http.StatusOK {
		return nil, parseError(response)
	}

	dec := asserts.NewDecoder(response.Body)

	asserts := []asserts.Assertion{}

	// TODO: make sure asserts can decode and deal with unknown types
	for {
		a, err := dec.Decode()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode assertions: %v", err)
		}
		asserts = append(asserts, a)
	}

	return asserts, nil
}
