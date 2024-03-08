package faoimpls

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/HeyPuter/puter-fuse-go/lang"
	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/google/uuid"
)

type node struct {
	fao.NodeInfo
	Nodes map[string]*node
	Data  []byte
}

func createNode() *node {
	return &node{
		Nodes: make(map[string]*node),
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
		if next, ok := current.Nodes[part]; ok {
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
		return fao.NodeInfo{}, false, nil
	}
	return n.NodeInfo, true, nil
}

func (f *MemFAO) ReadDir(path string) ([]fao.NodeInfo, error) {
	n, ok := f.resolvePath(path)
	if !ok {
		return nil, nil
	}
	var nodes []fao.NodeInfo
	for _, node := range n.Nodes {
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
	if !ok {
		return fao.NodeInfo{}, nil
	}
	if !n.IsDir {
		return fao.NodeInfo{}, nil
	}
	if _, ok := n.Nodes[name]; ok {
		return fao.NodeInfo{}, nil
	}
	newNode := createNode()
	newNode.Name = name
	newNode.Path = filepath.Join(path, name)
	newNode.Data = []byte{}
	n.Nodes[name] = newNode
	return n.Nodes[name].NodeInfo, nil
}

func (f *MemFAO) MkDir(parent, path string) (fao.NodeInfo, error) {
	n, ok := f.resolvePath(parent)
	if !ok {
		return fao.NodeInfo{}, nil
	}
	if !n.IsDir {
		return fao.NodeInfo{}, nil
	}
	if _, ok := n.Nodes[path]; ok {
		return fao.NodeInfo{}, nil
	}
	newNode := createNode()
	newNode.Name = path
	newNode.Path = filepath.Join(parent, path)
	newNode.IsDir = true
	n.Nodes[path] = newNode
	return n.Nodes[path].NodeInfo, nil
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
		return fao.NodeInfo{}, nil
	}
	if !n.IsDir {
		return fao.NodeInfo{}, nil
	}
	if _, ok := n.Nodes[name]; ok {
		return fao.NodeInfo{}, nil
	}
	newNode := createNode()
	newNode.IsSymlink = true
	newNode.SymlinkPath = target
	n.Nodes[name] = newNode
	return n.Nodes[name].NodeInfo, nil
}

func (f *MemFAO) Unlink(path string) error {
	parent := filepath.Dir(path)
	name := filepath.Base(path)

	n, ok := f.resolvePath(parent)
	if !ok {
		return nil
	}
	if _, ok := n.Nodes[name]; !ok {
		return nil
	}
	fmt.Printf("deleting %s from %s\n", name, parent)
	delete(n.Nodes, name)
	return nil
}

func (f *MemFAO) Move(source, parent, name string) error {
	sourceParent := filepath.Dir(source)
	sourceParentNode, ok := f.resolvePath(sourceParent)
	if !ok {
		return nil
	}
	sourceNode, ok := f.resolvePath(source)
	if !ok {
		return nil
	}
	delete(sourceParentNode.Nodes, filepath.Base(source))

	newParentNode, ok := f.resolvePath(parent)
	if !ok {
		return nil
	}
	newParentNode.Nodes[name] = sourceNode
	return nil
}

func (f *MemFAO) ReadAll(path string) (io.ReadCloser, error) {
	n, ok := f.resolvePath(path)
	if !ok {
		return nil, nil
	}
	if n.IsDir {
		return nil, nil
	}
	return io.NopCloser(strings.NewReader(string(n.Data))), nil
}
