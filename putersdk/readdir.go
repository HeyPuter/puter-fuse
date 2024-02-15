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
