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
