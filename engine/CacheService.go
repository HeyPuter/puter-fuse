package engine

import "sync"

type WholeFileCacheService struct {
	FileData map[string][]byte
	rwmutex  sync.RWMutex
}

func (svc_cache *WholeFileCacheService) Init() {
	svc_cache.FileData = map[string][]byte{}
}

func (svc_cache *WholeFileCacheService) GetFileData(path string) []byte {
	svc_cache.rwmutex.RLock()
	v, _ := svc_cache.FileData[path]
	svc_cache.rwmutex.RUnlock()
	return v
}

func (svc_cache *WholeFileCacheService) SetFileData(path string, data []byte) {
	svc_cache.rwmutex.Lock()
	svc_cache.FileData[path] = data
	svc_cache.rwmutex.Unlock()
}

func (svc_cache *WholeFileCacheService) DeleteFileData(path string) {
	svc_cache.rwmutex.Lock()
	delete(svc_cache.FileData, path)
	svc_cache.rwmutex.Unlock()
}
