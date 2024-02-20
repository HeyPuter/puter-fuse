package engine

import (
	"path/filepath"
	"sync"

	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/HeyPuter/puter-fuse-go/services"
)

type NodeType int

const (
	Dir NodeType = iota
	File
	Symlink
)

type NodeInfo struct {
	Path string
	Name string
	Type NodeType
	Size uint64
}

type PendingNodeService struct {
	LookupTablePath   map[string]*NodeInfo
	LookupTableParent map[string][]*NodeInfo

	// locks
	LookupTablePathLock   sync.RWMutex
	LookupTableParentLock sync.RWMutex

	// services
	services services.IServiceContainer
}

func (svc *PendingNodeService) Init(services services.IServiceContainer) {
	svc.LookupTablePath = map[string]*NodeInfo{}
	svc.LookupTableParent = map[string][]*NodeInfo{}
	svc.services = services
}

func (svc *PendingNodeService) Link(parent, name string, typ NodeType) *NodeInfo {
	// normalize parent path
	if parent[len(parent)-1] == '/' {
		parent = parent[:len(parent)-1]
	}

	nodeInfo := &NodeInfo{
		Path: filepath.Join(parent, name),
		Name: name,
		Type: typ,
	}

	// add to path lookup table
	path := filepath.Join(parent, name)
	svc.LookupTablePathLock.Lock()
	svc.LookupTablePath[path] = nodeInfo
	svc.LookupTablePathLock.Unlock()

	// add to parent lookup table
	svc.LookupTableParentLock.Lock()
	svc.LookupTableParent[parent] = append(svc.LookupTableParent[parent], nodeInfo)
	svc.LookupTableParentLock.Unlock()

	return nodeInfo
}

func (svc *PendingNodeService) GetNodeInfo(path string) *NodeInfo {
	svc.LookupTablePathLock.RLock()
	nodeInfo, _ := svc.LookupTablePath[path]
	svc.LookupTablePathLock.RUnlock()
	return nodeInfo
}

// The caller of this method must aquire a lock
func (svc *PendingNodeService) SetFileData(path string, data []byte) int {
	svc_wfcache := svc.services.Get("wfcache").(*WholeFileCacheService)
	svc.LookupTablePathLock.Lock()
	ver := svc_wfcache.SetFileData(path, data)
	nodeInfo := svc.LookupTablePath[path]
	nodeInfo.Size = uint64(len(data))
	svc.LookupTablePath[path] = nodeInfo
	svc.LookupTablePathLock.Unlock()
	return ver
}

func (svc *PendingNodeService) Forget(parent, name string) {
	// normalize parent path
	if parent[len(parent)-1] == '/' {
		parent = parent[:len(parent)-1]
	}

	// remove from path lookup table
	path := filepath.Join(parent, name)
	svc.LookupTablePathLock.Lock()
	delete(svc.LookupTablePath, path)
	svc.LookupTablePathLock.Unlock()

	// remove from parent lookup table
	svc.LookupTableParentLock.Lock()
	for i, node := range svc.LookupTableParent[parent] {
		if node.Path == path {
			svc.LookupTableParent[parent] = append(svc.LookupTableParent[parent][:i], svc.LookupTableParent[parent][i+1:]...)
			break
		}
	}
	svc.LookupTableParentLock.Unlock()
}

func (svc *PendingNodeService) GetChildren(parent string) []*NodeInfo {
	svc.LookupTableParentLock.RLock()
	children := svc.LookupTableParent[parent]
	svc.LookupTableParentLock.RUnlock()
	return children
}

func NodeInfoToArtificialCloudItem(nodeInfo *NodeInfo) putersdk.CloudItem {
	return putersdk.CloudItem{
		IsPending: true,
		Name:      nodeInfo.Name,
		Path:      nodeInfo.Path,
		IsDir:     nodeInfo.Type == Dir,
		Size:      nodeInfo.Size,
		// TODO: both of these won't be used once Local UIDs are used
		RemoteUID: "pending://" + nodeInfo.Path,
		Id:        "pending://" + nodeInfo.Path,
	}
}
