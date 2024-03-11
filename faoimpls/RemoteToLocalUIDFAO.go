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
	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/HeyPuter/puter-fuse-go/services"
)

type RemoteToLocalUIDFAO struct {
	fao.ProxyFAO
	associationService *engine.AssociationService
}

func CreateRemoteToLocalUIDFAO(
	delegate fao.FAO,
	services services.IServiceContainer,
) *RemoteToLocalUIDFAO {
	ins := &RemoteToLocalUIDFAO{}
	ins.associationService = services.Get("association").(*engine.AssociationService)
	ins.Delegate = delegate
	return ins
}

func (f *RemoteToLocalUIDFAO) Stat(path string) (fao.NodeInfo, bool, error) {
	nodeInfo, exists, err := f.Delegate.Stat(path)
	if err == nil && exists {
		localUID := f.associationService.GetLocalUIDFromRemote(nodeInfo.RemoteUID)
		nodeInfo.LocalUID = localUID
	}
	return nodeInfo, exists, err
}

func (f *RemoteToLocalUIDFAO) ReadDir(path string) ([]fao.NodeInfo, error) {
	nodeInfos, err := f.Delegate.ReadDir(path)
	if err == nil {
		for i, nodeInfo := range nodeInfos {
			localUID := f.associationService.GetLocalUIDFromRemote(nodeInfo.RemoteUID)
			nodeInfo.LocalUID = localUID
			nodeInfos[i] = nodeInfo
		}
	}
	return nodeInfos, err
}

func (f *RemoteToLocalUIDFAO) Create(path string, name string) (fao.NodeInfo, error) {
	nodeInfo, err := f.Delegate.Create(path, name)
	if err == nil {
		localUID := f.associationService.GetLocalUIDFromRemote(nodeInfo.RemoteUID)
		nodeInfo.LocalUID = localUID
	}
	return nodeInfo, err
}

func (f *RemoteToLocalUIDFAO) MkDir(parent, path string) (fao.NodeInfo, error) {
	nodeInfo, err := f.Delegate.MkDir(parent, path)
	if err == nil {
		localUID := f.associationService.GetLocalUIDFromRemote(nodeInfo.RemoteUID)
		nodeInfo.LocalUID = localUID
	}
	return nodeInfo, err
}
