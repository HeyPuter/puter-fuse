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
