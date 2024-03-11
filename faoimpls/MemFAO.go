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
	"io"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/HeyPuter/puter-fuse-go/lang"
	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/google/uuid"
)

type node struct {
	fao.NodeInfo
	Nodes lang.IMap[string, *node]
	Data  []byte
}

func createNode() *node {
	return &node{
		Nodes: lang.CreateSyncMap[string, *node](nil),
		NodeInfo: fao.NodeInfo{
			CloudItem: putersdk.CloudItem{
				RemoteUID: uuid.NewString(),
			},
		},
	}
}

type MemFAO struct {
	fao.BaseFAO
	Tree *node
}

func CreateMemFAO() *MemFAO {
	root := createNode()

	root.RemoteUID = engine.ROOT_UUID
	root.Id = engine.ROOT_UUID
	root.IsDir = true

	fao := &MemFAO{
		BaseFAO: fao.BaseFAO{},
		Tree:    root,
	}

	return fao
}

func (f *MemFAO) resolvePath(path string) (*node, bool) {
	parts := lang.PathSplit(path)
	current := f.Tree
	// fmt.Println("starting with", current)
	// fmt.Println("travsing", parts)
	for _, part := range parts {
		if !current.IsDir {
			return nil, false
		}
		if next, ok := current.Nodes.Get(part); ok {
			current = next
		} else {
			return nil, false
		}
	}
	return current, true
}

func (f *MemFAO) Stat(path string) (fao.NodeInfo, bool, error) {
	fmt.Printf("statting %s\n", path)
	n, ok := f.resolvePath(path)
	if !ok {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->Stat",
		}, false, nil
	}
	return n.NodeInfo, true, nil
}

func (f *MemFAO) ReadDir(path string) ([]fao.NodeInfo, error) {
	n, ok := f.resolvePath(path)
	if !ok {
		return nil, nil
	}
	var nodes []fao.NodeInfo
	for _, node := range n.Nodes.Values() {
		nodes = append(nodes, node.NodeInfo)
	}
	return nodes, nil
}

func (f *MemFAO) Read(path string, dest []byte, off int64) (int, error) {
	n, ok := f.resolvePath(path)
	fmt.Println("file node?", n)
	if !ok {
		return 0, nil
	}
	if n.IsDir {
		return 0, nil
	}
	if off >= int64(len(n.Data)) {
		return 0, nil
	}
	nBytes := copy(dest, n.Data[off:])
	return nBytes, nil
}

func (f *MemFAO) Write(path string, src []byte, off int64) (int, error) {
	n, ok := f.resolvePath(path)
	fmt.Println("file node?", n)
	if !ok {
		return 0, nil
	}
	if n.IsDir {
		return 0, nil
	}
	if off > int64(len(n.Data)) {
		return 0, nil
	}
	if off+int64(len(src)) > int64(len(n.Data)) {
		n.Data = append(n.Data, make([]byte, off+int64(len(src))-int64(len(n.Data)))...)
	}
	nBytes := copy(n.Data[off:], src)
	n.Size = uint64(len(n.Data))
	return nBytes, nil
}

func (f *MemFAO) Create(path string, name string) (fao.NodeInfo, error) {
	n, ok := f.resolvePath(path)
	fmt.Println(n)

	// TODO: errors here need to map to filesystem error numbers
	if !ok {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->Create (path not found)",
		}, fao.Errorf(syscall.ENOENT, "path %s does not exist", path)
		//fmt.Errorf("path %s does not exist", path)
	}
	if !n.IsDir {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->Create (path not a directory)",
		}, fao.Errorf(syscall.ENOTDIR, "path %s is not a directory", path)
	}
	if _, ok := n.Nodes.Get(name); ok {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->Create (node already exists)",
		}, fao.Errorf(syscall.EEXIST, "node %s already exists", name)
	}

	newNode := createNode()
	newNode.Name = name
	newNode.Path = filepath.Join(path, name)
	newNode.Data = []byte{}
	n.Nodes.Set(name, newNode)
	return newNode.NodeInfo, nil
}

func (f *MemFAO) MkDir(parent, path string) (fao.NodeInfo, error) {
	n, ok := f.resolvePath(parent)
	if !ok {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->MkDir (parent not found)",
		}, fao.Errorf(syscall.ENOENT, "parent %s does not exist", parent)
	}
	if !n.IsDir {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->MkDir (parent not a directory)",
		}, fao.Errorf(syscall.ENOTDIR, "parent %s is not a directory", parent)
	}
	if _, ok := n.Nodes.Get(path); ok {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->MkDir (node already exists)",
		}, fao.Errorf(syscall.EEXIST, "node %s already exists", path)
	}
	newNode := createNode()
	newNode.Name = path
	newNode.Path = filepath.Join(parent, path)
	newNode.IsDir = true
	n.Nodes.Set(path, newNode)
	return newNode.NodeInfo, nil
}

func (f *MemFAO) Truncate(path string, size uint64) error {
	n, ok := f.resolvePath(path)
	if !ok {
		return nil
	}
	if n.IsDir {
		return nil
	}
	if size < uint64(len(n.Data)) {
		n.Data = n.Data[:size]
	}
	return nil
}

func (f *MemFAO) Symlink(parent, name, target string) (fao.NodeInfo, error) {
	n, ok := f.resolvePath(parent)
	if !ok {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->Symlink (parent not found)",
		}, fao.Errorf(syscall.ENOENT, "parent %s does not exist", parent)
	}
	if !n.IsDir {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->Symlink (parent not a directory)",
		}, fao.Errorf(syscall.ENOTDIR, "parent %s is not a directory", parent)
	}
	if _, ok := n.Nodes.Get(name); ok {
		return fao.NodeInfo{
			DebugName: "empty from MemFAO->Symlink (node already exists)",
		}, fao.Errorf(syscall.EEXIST, "node %s already exists", name)
	}
	newNode := createNode()
	newNode.IsSymlink = true
	newNode.SymlinkPath = target
	n.Nodes.Set(name, newNode)
	return newNode.NodeInfo, nil
}

func (f *MemFAO) Unlink(path string) error {
	parent := filepath.Dir(path)
	name := filepath.Base(path)

	n, ok := f.resolvePath(parent)
	if !ok {
		return fao.Errorf(syscall.ENOENT, "parent %s does not exist", parent)
	}
	if _, ok := n.Nodes.Get(name); !ok {
		return fao.Errorf(syscall.ENOENT, "node %s does not exist", name)
	}
	fmt.Printf("deleting %s from %s\n", name, parent)
	n.Nodes.Del(name)
	return nil
}

func (f *MemFAO) Move(source, parent, name string) error {
	sourceParent := filepath.Dir(source)
	sourceParentNode, ok := f.resolvePath(sourceParent)
	if !ok {
		return fao.Errorf(syscall.ENOENT, "parent %s does not exist", sourceParent)
	}
	sourceNode, ok := f.resolvePath(source)
	if !ok {
		return fao.Errorf(syscall.ENOENT, "node %s does not exist", source)
	}
	sourceParentNode.Nodes.Del(filepath.Base(source))

	newParentNode, ok := f.resolvePath(parent)
	if !ok {
		return fao.Errorf(syscall.ENOENT, "parent %s does not exist", parent)
	}
	newParentNode.Nodes.Set(name, sourceNode)
	return nil
}

func (f *MemFAO) ReadAll(path string) (io.ReadCloser, error) {
	n, ok := f.resolvePath(path)
	if !ok {
		return nil, fao.Errorf(syscall.ENOENT, "node %s does not exist", path)
	}
	if n.IsDir {
		return nil, fao.Errorf(syscall.EISDIR, "node %s is a directory", path)
	}
	return io.NopCloser(strings.NewReader(string(n.Data))), nil
}
