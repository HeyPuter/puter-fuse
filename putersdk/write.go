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
	fmt.Printf("write(%s)\n", path)
	filename := filepath.Base(path)
	path = filepath.Dir(path)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	{
		fw, _ := writer.CreateFormFile("file", filename)
		fw.Write(data)
	}
	{
		fw, _ := writer.CreateFormField("path")
		io.Copy(fw, strings.NewReader(path))
	}
	{
		fw, _ := writer.CreateFormField("overwrite")
		io.Copy(fw, strings.NewReader("true"))
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
