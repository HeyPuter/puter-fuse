package putersdk

import (
	"errors"
	"net/http"
	"net/url"
	"path"
)

type PuterSDK struct {
	PuterAuthToken string
	Client         *http.Client
	Url            string
}

func (sdk *PuterSDK) Init() {
	sdk.Client = &http.Client{}
	if sdk.Url == "" {
		sdk.Url = "https://api.puter.local"
		// sdk.Url = "https://api.puter.com"
	}
}

func (sdk *PuterSDK) GetEndpointURL(name string) *url.URL {
	u, err := url.Parse(sdk.Url)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, name)
	return u
}

type PuterIntBool bool

func (b *PuterIntBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "true":
		fallthrough
	case "1":
		*b = true
	case "false":
		fallthrough
	case "0":
		*b = false
	default:
		return errors.New("invalid bool or integer for bool")
	}
	return nil
}

type CloudItem struct {
	Path     string
	Name     string
	Metadata interface{}
	Id       string
	Uid      string
	// IsShortcut     PuterIntBool `json:"is_shortcut"`
	IsSymlink PuterIntBool `json:"is_symlink"`
	Immutable PuterIntBool
	IsDir     PuterIntBool `json:"is_dir"`
	// ShortcutTo     string       `json:"shortcut_to"`
	// ShortcutToPath string       `json:"shortcut_to_path"`
	SymlinkPath string `json:"symlink_path"`
	Modified    float64
	Created     float64
	Accessed    float64
	Size        uint64
	Type        string
	IsPending   bool
}

type PuterSDKReaddirRequestPayload struct {
	Path string `json:"path"`
}
