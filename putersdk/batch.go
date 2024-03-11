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
	"mime/multipart"
	"net/http"
	"strings"
)

type Operation map[string]interface{}

type BatchResoponse struct {
	Results []map[string]interface{}
}

func (sdk *PuterSDK) Batch(operations []Operation, blobs [][]byte) (*BatchResoponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fmt.Printf("batching %d operations with %d files\n", len(operations), len(blobs))

	for _, op := range operations {
		opJson, err := json.Marshal(op)
		if err != nil {
			return nil, err
		}

		fw, _ := writer.CreateFormField("operation")
		fw.Write(opJson)
	}

	if blobs != nil {
		for _, blob := range blobs {
			fileinfoJson, err := json.Marshal(map[string]interface{}{
				"name": "untitled",
				"size": len(blob),
			})
			if err != nil {
				panic(err)
			}
			fw, _ := writer.CreateFormField("fileinfo")
			fw.Write(fileinfoJson)
		}
		for _, blob := range blobs {
			fw, _ := writer.CreateFormFile("file", "untitled")
			fw.Write(blob)
		}
	}

	writer.Close()

	u := sdk.GetEndpointURL("batch")

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+sdk.PuterAuthToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := sdk.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 218 {
		respBytes, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf(
			"unexpected status: %d"+
				"\nbody: |%s|",
			resp.StatusCode,
			string(respBytes),
		)
		return nil, err
	}

	bgetbody := new(strings.Builder)
	rgetbody, _ := req.GetBody()
	io.Copy(bgetbody, rgetbody)

	respBytes, _ := io.ReadAll(resp.Body)

	fmt.Printf("batch response: %s\n", string(respBytes))

	batchResponse := &BatchResoponse{}
	err = json.Unmarshal(respBytes, batchResponse)
	if err != nil {
		return nil, err
	}

	fmt.Printf("status? [%d] when sending [%s]\n",
		resp.StatusCode,
		bgetbody.String(),
	)

	return batchResponse, nil
}
