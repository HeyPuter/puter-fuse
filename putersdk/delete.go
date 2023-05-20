package putersdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (sdk *PuterSDK) Delete(path string) (err error) {
	fmt.Printf("delete(%s)\n", path)
	payload := map[string]interface{}{}
	payload["paths"] = []string{path}

	u := sdk.GetEndpointURL("delete")

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
	}

	return
}
