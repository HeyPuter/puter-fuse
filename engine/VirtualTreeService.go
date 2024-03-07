package engine

import (
	"fmt"
	"time"

	"github.com/HeyPuter/puter-fuse-go/lang"
	"github.com/HeyPuter/puter-fuse-go/services"
)

const (
	ROOT_UUID = "00000000-0000-0000-0000-000000000000"
)

type VirtualEntry struct {
}

type VirtualDirectoryEntry struct {
	// Directories lang.IMap[string, string]
	// Files       lang.IMap[string, string]
	VirtualEntry
	MemberUIDToName lang.IMap[string, string]
	MemberNameToUID lang.IMap[string, string]
	LastReaddir     time.Time
}

func CreateVirtualDirectoryEntry() *VirtualDirectoryEntry {
	ins := &VirtualDirectoryEntry{}

	ins.MemberUIDToName = lang.CreateSyncMap[string, string](nil)
	ins.MemberNameToUID = lang.CreateSyncMap[string, string](nil)

	return ins
}

type VirtualFileEntry struct {
	VirtualEntry
}

func CreateVirtualFileEntry() *VirtualFileEntry {
	ins := &VirtualFileEntry{}
	return ins
}

func (v *VirtualDirectoryEntry) GetUIDs() []string {
	return v.MemberUIDToName.Keys()
}

type VirtualTreeService struct {
	DirectoriesCacheLock *lang.CacheStampedeMap[string]
	Directories          lang.IMap[string, *VirtualDirectoryEntry]
	Files                lang.IMap[string, *VirtualFileEntry]
}

func (svc *VirtualTreeService) Init(services services.IServiceContainer) {
	svc.Directories.Set(ROOT_UUID, CreateVirtualDirectoryEntry())
}

func CreateVirtualTreeService() *VirtualTreeService {
	ins := &VirtualTreeService{}

	ins.Directories = lang.CreateSyncMap[string, *VirtualDirectoryEntry](nil)
	ins.Files = lang.CreateSyncMap[string, *VirtualFileEntry](nil)
	ins.DirectoriesCacheLock = lang.CreateCacheStampedeMap[string]()

	return ins
}

func (svc *VirtualTreeService) ResolvePath(parts []string) *VirtualDirectoryEntry {
	node, _ := svc.Directories.Get(ROOT_UUID)
	for _, part := range parts {
		dirUid, ok := node.MemberNameToUID.Get(part)
		if !ok {
			return nil
		}
		node, ok = svc.Directories.Get(dirUid)
		if !ok {
			return nil
		}
	}

	return node
}

func (svc *VirtualTreeService) RegisterDirectory(uid string) string {
	svc.Directories.Set(uid, CreateVirtualDirectoryEntry())
	return uid
}

func (svc *VirtualTreeService) RegisterFile(uid string) string {
	svc.Files.Set(uid, CreateVirtualFileEntry())
	return uid
}

func (svc *VirtualTreeService) Link(parentUID, childUID, name string) {
	fmt.Println("linking", parentUID, childUID, name)
	entry, _ := svc.Directories.Get(parentUID)
	entry.MemberUIDToName.Set(childUID, name)
	entry.MemberNameToUID.Set(name, childUID)
}

func (svc *VirtualTreeService) Unlink(parentUID, childUID string) {
	entry, _ := svc.Directories.Get(parentUID)
	name, _ := entry.MemberUIDToName.Get(childUID)
	entry.MemberUIDToName.Del(childUID)
	entry.MemberNameToUID.Del(name)
}

func (svc *VirtualTreeService) UpdateLastReaddir(uid string) {
	entry, _ := svc.Directories.Get(uid)
	entry.LastReaddir = time.Now()
}

// func (svc *VirtualTreeService) GetNodesFromEntry(entry *VirtualDirectoryEntry) []fao.NodeInfo {
// 	nodes := []fao.NodeInfo{}
// 	for _, fileUid := range entry.Files.Values() {
// 		file, _ := svc.Files.Get(fileUid)
// 		nodes = append(nodes, file.NodeInfo)
// 	}
// 	for _, dirUid := range entry.Directories.Values() {
// 		dir, _ := svc.Directories.Get(dirUid)
// 		nodes = append(nodes, dir.NodeInfo)
// 	}
// 	return nodes
// }
