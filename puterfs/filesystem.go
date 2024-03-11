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
	"fmt"
	"log"
	"sync"

	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/HeyPuter/puter-fuse-go/services"
	"github.com/hanwen/go-fuse/v2/fs"
)

type Filesystem struct {
	// map Puter UIDs to inode numbers
	UidInoMap  map[string]uint64
	InoCounter uint64

	Nodes map[uint64]fs.InodeEmbedder

	SDK *putersdk.PuterSDK
	fao.FAO
	Services *services.ServicesContainer

	NodesMutex     sync.RWMutex
	UidInoMapMutex sync.RWMutex
}

func (pfs *Filesystem) Init() {
	if pfs.UidInoMap != nil {
		panic("Filesystem already initialized")
	}
	pfs.UidInoMap = map[string]uint64{}
	pfs.Nodes = map[uint64]fs.InodeEmbedder{}
}

func (fs *Filesystem) GetNodeFromCloudItem(cloudItem fao.NodeInfo) fs.InodeEmbedder {
	// TODO: use LocalUID instead of RemoteUID
	ino := fs.GetInoFromUID(cloudItem.RemoteUID)
	fs.NodesMutex.RLock()
	node, exists := fs.Nodes[ino]
	fs.NodesMutex.RUnlock()
	if !exists {
		fs.NodesMutex.Lock()
		// check again in case another thread just did this
		node, exists = fs.Nodes[ino]
		// util.Printvar(node, "before")
		if !exists {
			node = fs.CreateNodeFromCloudItem(cloudItem)
			// util.Printvar(node, "after")
		}
		fs.Nodes[ino] = node
		fs.NodesMutex.Unlock()
	}

	iface := node.(HasPuterNodeCapabilities)
	iface.SetCloudItem(cloudItem)

	return node
}

func (fs *Filesystem) GetInoFromUID(uid string) uint64 {
	fs.UidInoMapMutex.RLock()
	ino, exists := fs.UidInoMap[uid]
	fmt.Printf("Filesystem mem address: %p\n", fs)
	fmt.Println("GetInoFromUID (A)", uid, ino, exists)
	fs.UidInoMapMutex.RUnlock()
	if !exists {
		fs.UidInoMapMutex.Lock()
		// check again in case another thread just did this
		ino, exists = fs.UidInoMap[uid]
		if !exists {
			fs.InoCounter++
			ino = fs.InoCounter
			fmt.Println("new ino", ino, "for uid", uid)
			fs.UidInoMap[uid] = ino
		}
		fs.UidInoMapMutex.Unlock()
	}
	fmt.Println("GetInoFromUID (B)", uid, ino, exists)
	return ino
}

func (fs *Filesystem) CreateNodeFromCloudItem(cloudItem fao.NodeInfo) fs.InodeEmbedder {
	// log info about clouditem
	log.Printf("cloudItem: %+v\n", cloudItem)

	if cloudItem.IsDir {
		fmt.Println("creating dir node")
		return fs.CreateDirNodeFromCloudItem(cloudItem)
	}

	fmt.Println("creating file node")
	return fs.CreateFileNodeFromCloudItem(cloudItem)
}

// start :: redundant (file,file;dir,directory)

func (pfs *Filesystem) CreateDirNodeFromCloudItem(cloudItem fao.NodeInfo) fs.InodeEmbedder {
	dirNode := &DirectoryNode{}
	dirNode.CloudItem = cloudItem
	dirNode.Filesystem = pfs

	dirNode.Init()

	return dirNode
}

func (pfs *Filesystem) CreateFileNodeFromCloudItem(cloudItem fao.NodeInfo) fs.InodeEmbedder {
	fileNode := &FileNode{}
	fileNode.CloudItem = cloudItem
	fileNode.Filesystem = pfs

	fileNode.Init()

	return fileNode
}

// end :: redundant
