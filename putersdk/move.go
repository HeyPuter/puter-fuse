package putersdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (sdk *PuterSDK) Move(sourcePath, dstPath, newName string) (cloudItem CloudItem, err error) {
	fmt.Printf("move(%s,%s,%s)\n", sourcePath, dstPath, newName)
	payload := map[string]interface{}{}
	payload["source_path"] = sourcePath
	payload["dest_path"] = dstPath
	payload["new_name"] = newName

	u := sdk.GetEndpointURL("move")

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

	err = json.Unmarshal(body, &cloudItem)
	return
}
