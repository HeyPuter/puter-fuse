package faoimpls

import (
	"fmt"

	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/HeyPuter/puter-fuse-go/services"
	"github.com/google/uuid"
)

type FileWriteCacheFAO struct {
	fao.ProxyFAO
	associationService *engine.AssociationService
	blobCacheService   *engine.BLOBCacheService
	writeCacheService  *engine.WriteCacheService
}

func CreateFileWriteCacheFAO(delegate fao.FAO, services services.IServiceContainer) *FileWriteCacheFAO {
	ins := &FileWriteCacheFAO{}
	ins.associationService = services.Get("association").(*engine.AssociationService)
	ins.blobCacheService = services.Get("blob-cache").(*engine.BLOBCacheService)
	ins.writeCacheService = services.Get("write-cache").(*engine.WriteCacheService)
	ins.Delegate = delegate
	return ins
}

func (f *FileWriteCacheFAO) getOrCreateCachedRead(path string) (string, error) {
	// Determine if we have a cached read to write against
	baseHash, exists := f.associationService.PathToBaseHash.Get(path)
	if exists {
		return baseHash, nil
	}

	// If not, create a new one
	reader, err := f.Delegate.ReadAll(path)
	if err != nil {
		return "", err
	}
	cacheRef := f.blobCacheService.Store(reader)
	f.associationService.PathToBaseHash.Set(path, cacheRef.GetHash())
	return cacheRef.GetHash(), nil
}

func (f *FileWriteCacheFAO) Read(path string, dest []byte, offset int64) (int, error) {
	f.Delegate.Read(path, dest, offset)

	localUID, _, _ := f.associationService.PathToLocalUID.
		GetWithFactory(path, func() (string, bool, error) {
			return uuid.NewString(), true, nil
		})

	fmt.Println(localUID)

	f.writeCacheService.ApplyToBuffer(localUID, dest, offset)

	return len(dest), nil
}

func (f *FileWriteCacheFAO) Write(path string, data []byte, offset int64) (int, error) {
	// Get a cached read to write against
	// baseHash, err := f.getOrCreateCachedRead(path)
	// if err != nil {
	// 	return 0, err
	// }

	// Create a write mutation
	mut := &engine.WriteMutation{
		Data:   data,
		Offset: offset,
	}

	// it's okay to ignore 'err' here since only the factory can
	// return an error (and it invariably returns nil)
	localUID, _, _ := f.associationService.PathToLocalUID.
		GetWithFactory(path, func() (string, bool, error) {
			return uuid.NewString(), true, nil
		})

	// Apply the mutation
	ref := f.writeCacheService.ApplyMutation(localUID, mut)

	go func() {
		f.Delegate.Write(path, data, offset)
		ref.Release()
	}()

	return len(data), nil
}
