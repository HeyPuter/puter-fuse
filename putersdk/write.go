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
	"path/filepath"
	"strings"
)

func (sdk *PuterSDK) Write(path string, data []byte) (*CloudItem, error) {
	cloudItem, err := sdk.write(path, data, "")
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	return cloudItem, err
}

func (sdk *PuterSDK) Symlink(path, target string) (*CloudItem, error) {
	parent := filepath.Dir(path)
	name := filepath.Base(path)
	batchResponse, err := sdk.Batch([]Operation{
		{
			"op":     "symlink",
			"path":   parent,
			"name":   name,
			"target": target,
		},
	}, nil)

	if err != nil {
		return nil, err
	}

	if len(batchResponse.Results) != 1 {
		return nil, fmt.Errorf("unexpected batch response length: %d", len(batchResponse.Results))
	}

	// Get CloudItem from batch response at index 0
	cloudItem := &CloudItem{}
	err = unmarshalIntoStruct(batchResponse.Results[0], cloudItem)
	if err != nil {
		return nil, err
	}

	return cloudItem, nil
}

func (sdk *PuterSDK) write(path string, data []byte, target string) (*CloudItem, error) {
	fmt.Printf("write(%s)\n", path)
	filename := filepath.Base(path)
	path = filepath.Dir(path)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		fw, _ := writer.CreateFormField("path")
		io.Copy(fw, strings.NewReader(path))
	}
	{
		fw, _ := writer.CreateFormField("overwrite")
		io.Copy(fw, strings.NewReader("true"))
	}
	{
		fw, _ := writer.CreateFormField("size")
		io.Copy(fw, strings.NewReader(fmt.Sprintf("%d", len(data))))
	}
	if target != "" {
		fw, _ := writer.CreateFormField("symlink_path")
		io.Copy(fw, strings.NewReader(target))
	}
	{
		fw, _ := writer.CreateFormFile("file", filename)
		fw.Write(data)
	}
	writer.Close()

	// fmt.Printf("Did it work? [%s]\n", body.String())

	u := sdk.GetEndpointURL("write")

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+sdk.PuterAuthToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Printf("Content-Type: %s\n", writer.FormDataContentType())

	resp, err := sdk.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
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

	cloudItem := &CloudItem{}
	err = json.Unmarshal(respBytes, cloudItem)
	if err != nil {
		return nil, err
	}

	fmt.Printf("status? [%d] when sending [%s]\n",
		resp.StatusCode,
		bgetbody.String(),
	)

	return cloudItem, nil
}
