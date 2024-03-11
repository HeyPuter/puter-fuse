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
	"errors"
	"net/http"
	"net/url"
	"path"
	"time"
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
	Path      string
	Name      string
	Metadata  interface{}
	Id        string
	LocalUID  string
	RemoteUID string `json:"uid"`
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
	LastStat    time.Time
}

type PuterSDKReaddirRequestPayload struct {
	Path string `json:"path"`
}
