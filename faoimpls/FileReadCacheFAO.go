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
