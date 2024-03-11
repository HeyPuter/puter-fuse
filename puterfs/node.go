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
package puterfs

import (
	"syscall"

	"github.com/HeyPuter/puter-fuse/fao"
)

type HasPuterNodeCapabilities interface {
	GetStableAttrMode() uint32
	GetIno() uint64
	SetCloudItem(cloudItem fao.NodeInfo)
}

type CloudItemNode struct {
	*Filesystem
	CloudItem fao.NodeInfo
}

func (n *CloudItemNode) Init() {
	// NO-OP to prevent Filesystem.Init from being called
}

func (n *CloudItemNode) GetStableAttrMode() uint32 {
	if n.CloudItem.IsDir {
		return syscall.S_IFDIR
	}

	if n.CloudItem.IsSymlink {
		return syscall.S_IFLNK
	}

	return syscall.S_IFREG
}

func (n *CloudItemNode) GetIno() uint64 {
	// TODO: use LocalUID instead of RemoteUID
	return n.Filesystem.GetInoFromUID(n.CloudItem.RemoteUID)
}

func (n *CloudItemNode) SetCloudItem(cloudItem fao.NodeInfo) {
	n.CloudItem = cloudItem
}
