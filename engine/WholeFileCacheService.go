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
package engine

import (
	"sync"

	"github.com/HeyPuter/puter-fuse/services"
)

type WholeFileCacheService struct {
	FileVer  map[string]int
	FileData map[string][]byte
	rwmutex  sync.RWMutex

	services services.IServiceContainer
}

func (svc_cache *WholeFileCacheService) Init(services services.IServiceContainer) {
	svc_cache.FileVer = map[string]int{}
	svc_cache.FileData = map[string][]byte{}

	svc_cache.services = services
}

func (svc_cache *WholeFileCacheService) GetFileData(path string) []byte {
	svc_cache.rwmutex.RLock()
	v, _ := svc_cache.FileData[path]
	svc_cache.rwmutex.RUnlock()
	return v
}

func (svc_cache *WholeFileCacheService) SetFileData(path string, data []byte) int {
	var ver int
	var exists bool

	svc_cache.rwmutex.Lock()
	if ver, exists = svc_cache.FileVer[path]; exists {
		svc_cache.FileVer[path] = ver + 1
	} else {
		svc_cache.FileVer[path] = 1
	}
	svc_cache.FileData[path] = data
	svc_cache.rwmutex.Unlock()
	return ver
}

func (svc_cache *WholeFileCacheService) DeleteFileData(path string, ver int) {
	svc_cache.rwmutex.Lock()
	if svc_cache.FileVer[path] != ver {
		svc_cache.rwmutex.Unlock()
		return
	}
	delete(svc_cache.FileData, path)
	svc_cache.rwmutex.Unlock()
}
