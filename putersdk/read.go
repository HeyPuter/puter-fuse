/*
 * Copyright (C) 2024  Puter Technologies Inc.
 *
 * This file is part of puter-fuse.
 *
 * puter-fuse is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
package putersdk

import (
	"io"
	"net/http"
	"net/url"
)

func (sdk *PuterSDK) Read(path string) (data []byte, err error) {
	u := sdk.GetEndpointURL("read")

	params := url.Values{}
	params.Add("path", path)

	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("Authorization", "Bearer "+sdk.PuterAuthToken)

	resp, err := sdk.Client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	data = body
	return
}

func (sdk *PuterSDK) ReadStream(path string) (reader io.ReadCloser, err error) {
	u := sdk.GetEndpointURL("read")

	params := url.Values{}
	params.Add("path", path)

	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+sdk.PuterAuthToken)

	resp, err := sdk.Client.Do(req)
	if err != nil {
		return
	}

	reader = resp.Body
	return
}
