package engine

import (
	"sync"

	"github.com/HeyPuter/puter-fuse-go/services"
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
