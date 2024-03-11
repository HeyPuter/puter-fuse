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

	"github.com/HeyPuter/puter-fuse/fao"
	"github.com/HeyPuter/puter-fuse/kvdotgo"
	"github.com/HeyPuter/puter-fuse/lang"
	"github.com/HeyPuter/puter-fuse/services"
	"github.com/google/uuid"
)

type AssociationService struct {
	LocalUIDToRemoteUID lang.IMap[string, string]
	RemoteUIDToLocalUID lang.IMap[string, string]
	LocalUIDToIno       lang.IMap[string, uint64]
	InoToLocalUID       lang.IMap[uint64, string]

	// LocalUIDToNodeInfo   lang.IMap[string, *fao.NodeInfo]
	LocalUIDToNodeInfo *kvdotgo.KVMap[string, fao.NodeInfo]
	// PathToLocalUID     *kvdotgo.KVMap[string, string]
	PathToLocalUID lang.IMap[string, string]

	LocalUIDToBaseHash lang.IMap[string, string]
	PathToBaseHash     lang.IMap[string, string]

	CacheStampedeMapLock sync.RWMutex
	CacheStampedeMap     map[string]*sync.Mutex
}

func (svc *AssociationService) Init(services services.IServiceContainer) {}

func CreateAssociationService() *AssociationService {
	ins := &AssociationService{}

	ins.LocalUIDToRemoteUID = lang.CreateSyncMap[string, string](nil)
	ins.RemoteUIDToLocalUID = lang.CreateSyncMap[string, string](nil)
	ins.LocalUIDToIno = lang.CreateSyncMap[string, uint64](nil)
	ins.InoToLocalUID = lang.CreateSyncMap[uint64, string](nil)

	// ins.LocalUIDToNodeInfo = lang.CreateSyncMap[string, *fao.NodeInfo](nil)
	ins.LocalUIDToNodeInfo = kvdotgo.CreateKVMap[string, fao.NodeInfo]()
	ins.PathToLocalUID = lang.CreateSyncMap[string, string](nil)

	ins.LocalUIDToBaseHash = lang.CreateSyncMap[string, string](nil)
	ins.PathToBaseHash = lang.CreateSyncMap[string, string](nil)

	ins.CacheStampedeMap = map[string]*sync.Mutex{}

	return ins
}

func (svc *AssociationService) GetLocalUIDFromRemote(remoteUID string) string {
	localUID, exists := svc.RemoteUIDToLocalUID.Get(remoteUID)
	if !exists {
		localUID = uuid.NewString()
		svc.RemoteUIDToLocalUID.Set(remoteUID, localUID)
		svc.LocalUIDToRemoteUID.Set(localUID, remoteUID)
	}
	return localUID
}
