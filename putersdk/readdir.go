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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/HeyPuter/puter-fuse-go/debug"
)

func (sdk *PuterSDK) Readdir(logger debug.ILogger, path string) (
	items []CloudItem, err error,
) {
	logger.Log("readdir(%s)", path)
	payload := map[string]interface{}{}
	payload["path"] = path
	payload["no_thumbs"] = true
	payload["no_assocs"] = true

	u := sdk.GetEndpointURL("readdir")

	jsonStr, err := json.Marshal(payload)
	if err != nil {
		return
	}
	req, err := http.NewRequest(
		"POST",
		u.String(),
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+sdk.PuterAuthToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := sdk.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected status: %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &items)
	return
}
