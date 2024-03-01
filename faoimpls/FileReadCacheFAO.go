package faoimpls

import (
	"fmt"
	"time"

	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/HeyPuter/puter-fuse-go/services"
)

type P_FileReadCacheFAO struct {
	TTL time.Duration
}

type FileReadCacheFAO struct {
	fao.ProxyFAO
	associationService *engine.AssociationService
	blobCacheService   *engine.BLOBCacheService
	P_FileReadCacheFAO
}

func CreateFileReadCacheFAO(delegate fao.FAO, services services.IServiceContainer, params P_FileReadCacheFAO) *FileReadCacheFAO {
	ins := &FileReadCacheFAO{}
	ins.associationService = services.Get("association").(*engine.AssociationService)
	ins.blobCacheService = services.Get("blob-cache").(*engine.BLOBCacheService)
	ins.Delegate = delegate
	ins.P_FileReadCacheFAO = params
	return ins
}

func (f *FileReadCacheFAO) tryGetCache(path string, dest []byte, offset int64) (int, bool, error) {
	// localUID, exists := f.associationService.PathToLocalUID.Get(path)
	// if !exists {
	// 	fmt.Println("No localUID for path", path)
	// 	return false, nil
	// }
	// baseHash, exists := f.associationService.LocalUIDToBaseHash.Get(localUID)
	// if !exists {
	// 	return false, nil
	// }

	baseHash, exists := f.associationService.PathToBaseHash.Get(path)
	if !exists {
		return 0, false, nil
	}

	n, exists, err := f.blobCacheService.GetBytes(baseHash, offset, dest)
	if err != nil {
		return 0, false, err
	}
	return n, exists, nil
}

func (f *FileReadCacheFAO) Read(path string, dest []byte, offset int64) (int, error) {
	fmt.Println("READ CACHE FAO ACCESSED")
	n, cacheHit, err := f.tryGetCache(path, dest, offset)
	if err != nil {
		return 0, err
	}
	if cacheHit {
		fmt.Println("Read file cache hit")
		return n, nil
	}

	fmt.Println("Read file cache miss")

	reader, err := f.Delegate.ReadAll(path)
	if err != nil {
		return 0, err
	}
	cacheRef := f.blobCacheService.Store(reader)
	f.associationService.PathToBaseHash.Set(path, cacheRef.GetHash())

	// For now, a naive TTL eviction policy
	// go func() {
	// 	<-time.After(f.TTL)
	// 	cacheRef.Release()
	// }()

	f.blobCacheService.GetBytes(cacheRef.GetHash(), offset, dest)
	return len(dest), nil
}
