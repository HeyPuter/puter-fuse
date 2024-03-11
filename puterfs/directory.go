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
	"context"
	"fmt"
	"path/filepath"
	"syscall"
	"time"

	"github.com/HeyPuter/puter-fuse/debug"
	"github.com/HeyPuter/puter-fuse/fao"
	"github.com/HeyPuter/puter-fuse/kvdotgo"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type DirectoryNode struct {
	fs.Inode
	CloudItemNode
	Items        []fao.NodeInfo
	PollDuration time.Duration
	Logger       debug.ILogger
	PathLockMap  *kvdotgo.KVMap[string, struct{}]
}

func (n *DirectoryNode) Init() {
	// call super init
	n.CloudItemNode.Init()
	n.Items = []fao.NodeInfo{}
	n.PollDuration = 2 * time.Second
	svc_log := n.Filesystem.Services.Get("log").(*debug.LogService)
	n.Logger = svc_log.GetLogger("Inode:D " + n.CloudItem.Path)
	n.PathLockMap = kvdotgo.CreateKVMap[string, struct{}]()
}

func (n *DirectoryNode) syncItems() error {
	// TODO: Path -> UID
	var items []fao.NodeInfo
	var err error

	items, err = n.FAO.ReadDir(n.CloudItem.Path)
	if err != nil {
		return err
	}

	// check for duplicate item names
	seen := map[string]fao.NodeInfo{}
	for _, item := range items {
		fmt.Println("item", item)
		if item.Path == "" {
			panic("item is missing path")
		}
		if iseen, ok := seen[item.Name]; ok {
			// return fmt.Errorf("duplicate item name: %s", item.Name)
			fmt.Printf("item named %s has local uid %s\n", item.Name, item.LocalUID)
			fmt.Printf("item named %s has real uid %s\n", item.Name, item.RemoteUID)
			fmt.Printf("SEEN named %s has local uid %s\n", iseen.Name, iseen.LocalUID)
			fmt.Printf("SEEN named %s has real uid %s\n", iseen.Name, iseen.RemoteUID)
			// panic(fmt.Errorf("duplicate item name: %s", item.Name))
		}
		seen[item.Name] = item
	}

	n.Items = items
	return nil
}

func (n *DirectoryNode) lookupCloudItem(
	name string,
) (fao.NodeInfo, bool) {
	var foundItem fao.NodeInfo
	var found bool

	for _, item := range n.Items {
		if item.Name == name {
			foundItem = item
			found = true
			break
		}
	}

	return foundItem, found
}

func (n *DirectoryNode) addOrReplaceCloudItem(
	name string, item fao.NodeInfo,
) {
	for i, existingItem := range n.Items {
		if existingItem.Name == name {
			n.Items[i] = item
			return
		}
	}

	n.Items = append(n.Items, item)
}

func (n *DirectoryNode) Lookup(
	ctx context.Context, name string, out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	n.Logger.Log("lookup(%s)", name)
	n.syncItems()

	foundItem, found := n.lookupCloudItem(name)

	if !found {
		// TODO: return an error code?
		return nil, syscall.ENOENT
	}
	foundItemNode := n.Filesystem.GetNodeFromCloudItem(foundItem)

	iface := foundItemNode.(HasPuterNodeCapabilities)

	return n.NewInode(
		ctx,
		foundItemNode,
		fs.StableAttr{
			Mode: iface.GetStableAttrMode(),
			Ino:  iface.GetIno(),
		},
	), 0
}

func (n *DirectoryNode) Symlink(
	ctx context.Context,
	target string,
	name string,
	out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	fmt.Printf("dir::symlink(%s)\n", name)
	n.syncItems()

	node, err := n.FAO.Symlink(n.CloudItem.Path, name, target)
	if err != nil {
		return nil, syscall.EIO
	}

	cloudItemNode := n.Filesystem.GetNodeFromCloudItem(node)
	iface := cloudItemNode.(HasPuterNodeCapabilities)

	return n.NewInode(
		ctx,
		cloudItemNode,
		fs.StableAttr{
			Mode: iface.GetStableAttrMode(),
			Ino:  iface.GetIno(),
		},
	), 0
}

func (n *DirectoryNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	n.syncItems()

	entries := []fuse.DirEntry{}
	for _, item := range n.Items {
		node := n.Filesystem.GetNodeFromCloudItem(item)
		iface := node.(HasPuterNodeCapabilities)
		entry := fuse.DirEntry{
			Mode: iface.GetStableAttrMode(),
			Name: item.Name,
			Ino:  iface.GetIno(),
		}
		entries = append(entries, entry)
	}
	return fs.NewListDirStream(entries), 0
}

func (n *DirectoryNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Size = 4096

	// TODO: load from configuration
	// out.Mode = 0644
	out.Mode = 0755
	out.Mode = out.Mode | 040000

	// TODO: load from configuration
	out.Uid = 1000
	out.Gid = 1000

	return 0
}

func (n *DirectoryNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	n.Logger.Log("create(%s)", name)
	// check if directory already exists
	_, exists, err := n.FAO.Stat(filepath.Join(n.CloudItem.Path, name))
	if err != nil {
		return nil, nil, 0, syscall.EIO
	}
	if exists {
		return nil, nil, 0, syscall.EEXIST
	}

	nodeInfo, err := n.FAO.Create(n.CloudItem.Path, name)
	if err != nil {
		panic(err)
		// return nil, nil, 0, syscall.EIO
	}

	// log the node info
	n.Logger.Log("nodeInfo: %+v", nodeInfo)
	// log "is dir"
	n.Logger.Log("is dir: %v", nodeInfo.IsDir)

	cloudItemNode := n.Filesystem.GetNodeFromCloudItem(nodeInfo)
	iface := cloudItemNode.(HasPuterNodeCapabilities)

	return n.NewInode(
		ctx,
		cloudItemNode,
		fs.StableAttr{
			Mode: iface.GetStableAttrMode(),
			Ino:  iface.GetIno(),
		},
	), &FileHandler{Node: cloudItemNode.(*FileNode)}, 0, 0
}

func (n *DirectoryNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	mutex := n.PathLockMap.SetAndLock(name, struct{}{}, 0)
	defer mutex.Unlock()

	// check if directory already exists
	_, exists, err := n.FAO.Stat(filepath.Join(n.CloudItem.Path, name))
	if err != nil {
		return nil, syscall.EIO
	}
	if exists {
		return nil, syscall.EEXIST
	}

	nodeInfo, err := n.FAO.MkDir(n.CloudItem.Path, name)
	if err != nil {
		return nil, syscall.EIO
	}

	cloudItemNode := n.Filesystem.GetNodeFromCloudItem(nodeInfo)
	iface := cloudItemNode.(HasPuterNodeCapabilities)

	return n.NewInode(
		ctx,
		cloudItemNode,
		fs.StableAttr{
			Mode: iface.GetStableAttrMode(),
			Ino:  iface.GetIno(),
		},
	), 0
}

func (n *DirectoryNode) Unlink(ctx context.Context, name string) syscall.Errno {
	path := filepath.Join(n.CloudItem.Path, name)
	stat, exists, err := n.FAO.Stat(path)
	if err != nil {
		return syscall.EIO
	}
	if !exists {
		return syscall.ENOENT
	}
	if stat.IsDir {
		return syscall.EISDIR
	}

	err = n.FAO.Unlink(path)
	if err != nil {
		return syscall.EIO
	}

	return 0
}

func (n *DirectoryNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	path := filepath.Join(n.CloudItem.Path, name)
	stat, exists, err := n.FAO.Stat(path)
	if err != nil {
		return syscall.EIO
	}
	if !exists {
		return syscall.ENOENT
	}
	if !stat.IsDir {
		return syscall.ENOTDIR
	}

	err = n.FAO.Unlink(path)
	if err != nil {
		return syscall.EIO
	}

	return 0
}

func (n *DirectoryNode) Rename(
	ctx context.Context,
	name string,
	newParent fs.InodeEmbedder,
	newName string,
	flags uint32,
) syscall.Errno {
	sourcePath := filepath.Join(n.CloudItem.Path, name)
	parentNode := newParent.(*DirectoryNode)
	err := n.FAO.Move(sourcePath, parentNode.CloudItem.Path, newName)
	if err != nil {
		n.Logger.Log("rename error: %v", err)
		return syscall.EIO
	}
	return 0
}
