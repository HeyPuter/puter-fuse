package faoimpls

import (
	"fmt"
	"time"

	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/HeyPuter/puter-fuse-go/lang"
	"github.com/HeyPuter/puter-fuse-go/putersdk"
)

type P_TreeCacheFAO struct {
	TTL time.Duration
}

type D_TreeCacheFAO struct {
	*engine.VirtualTreeService
	*engine.AssociationService
}

type TreeCacheFAO struct {
	fao.ProxyFAO
	P_TreeCacheFAO
	D_TreeCacheFAO
}

func CreateTreeCacheFAO(
	delegate fao.FAO,
	params P_TreeCacheFAO,
	deps D_TreeCacheFAO,
) *TreeCacheFAO {
	fao := &TreeCacheFAO{
		ProxyFAO: fao.ProxyFAO{
			P_CreateProxyFAO: fao.P_CreateProxyFAO{
				Delegate: delegate,
			},
		},
		P_TreeCacheFAO: params,
		D_TreeCacheFAO: deps,
	}

	return fao
}

func (f *TreeCacheFAO) ReadDir(path string) ([]fao.NodeInfo, error) {
	parts := lang.PathSplit(path)
	entry := f.VirtualTreeService.ResolvePath(parts)

	// fmt.Println("The entry in question", entry)

	if entry == nil || entry.LastReaddir.Add(f.TTL).Before(time.Now()) {
		l := f.VirtualTreeService.DirectoriesCacheLock.Lock(path)
		entry = f.VirtualTreeService.ResolvePath(parts)
		if entry == nil || entry.LastReaddir.Add(f.TTL).Before(time.Now()) {
			defer l.Unlock()
			return f.readDirAndUpdateCache(path)
		}
		l.Unlock()
	}

	var nodeInfos []fao.NodeInfo
	populate_nodeinfos := func() bool {
		nodeInfos = []fao.NodeInfo{}
		allExist := true
		// fmt.Println("entry.GetUIDs()", entry.GetUIDs())
		for _, localUID := range entry.GetUIDs() {
			// fmt.Println("localUID", localUID)
			nodeInfo := f.AssociationService.LocalUIDToNodeInfo.Get(localUID)
			// fmt.Println("nodeInfo", *nodeInfo)
			if nodeInfo == nil {
				allExist = false
				break
			}
			nodeInfos = append(nodeInfos, *nodeInfo)
		}
		return allExist
	}

	if !populate_nodeinfos() {
		l := f.VirtualTreeService.DirectoriesCacheLock.Lock(path)
		if !populate_nodeinfos() {
			defer l.Unlock()
			return f.readDirAndUpdateCache(path)
		}
		l.Unlock()
	}

	// fmt.Println("readdir cache hit", path)
	// fmt.Println("readdir cache hit", nodeInfos)

	return nodeInfos, nil
}

func (f *TreeCacheFAO) Stat(path string) (fao.NodeInfo, bool, error) {
	localUID, exists := f.AssociationService.PathToLocalUID.Get(path)
	if exists {
		nodeInfo, ok, err := f.AssociationService.LocalUIDToNodeInfo.GetOrSet(
			localUID,
			f.TTL,
			func() (fao.NodeInfo, bool, error) {
				stat, exists, err := f.Delegate.Stat(path)
				if err != nil {
					return fao.NodeInfo{}, false, err
				}
				if !exists {
					return fao.NodeInfo{}, false, nil
				}
				stat.LastStat = time.Now()
				f.AssociationService.PathToLocalUID.Set(path, stat.LocalUID)
				return stat, true, nil
			},
		)
		if err != nil {
			return fao.NodeInfo{}, false, err
		}
		if !ok {
			return fao.NodeInfo{}, false, nil
		}

		return nodeInfo, true, nil
	}

	stat, exists, err := f.Delegate.Stat(path)
	if err != nil {
		return fao.NodeInfo{}, false, err
	}

	if !exists {
		return fao.NodeInfo{}, false, nil
	}

	stat.LastStat = time.Now()
	m := f.AssociationService.LocalUIDToNodeInfo.SetAndLock(
		stat.LocalUID, stat, f.TTL)
	f.AssociationService.PathToLocalUID.Set(path, stat.LocalUID)
	m.Unlock()
	return stat, true, nil
}

func (f *TreeCacheFAO) readDirAndUpdateCache(path string) ([]fao.NodeInfo, error) {
	fmt.Println("readdir cache miss", path)

	// Stat the directory (prerequisite to cache the path association)
	var stat fao.NodeInfo
	var exists bool
	var err error

	if path == "/" {
		exists = true
		stat = fao.NodeInfo{
			CloudItem: putersdk.CloudItem{
				LocalUID: engine.ROOT_UUID,
				IsDir:    true,
			},
		}
	} else {
		stat, exists, err = f.Stat(path)
	}

	if err != nil {
		fmt.Printf("error statting %s: %s\n", path, err)
		return nil, err
	}

	if !exists {
		fmt.Printf("does not exist: %s\n", path)
		return nil, &fao.ErrDoesNotExist{}
	}

	if !stat.IsDir {
		fmt.Printf("not a directory: %s\n", path)
		return nil, &fao.ErrNotDirectory{}
	}

	nodeInfos, err := f.Delegate.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Cache the nodeInfos
	for _, nodeInfo := range nodeInfos {
		if nodeInfo.IsDir {
			f.VirtualTreeService.RegisterDirectory(nodeInfo.LocalUID)
		}
		f.AssociationService.LocalUIDToNodeInfo.Set(nodeInfo.LocalUID, nodeInfo, f.TTL)
		f.VirtualTreeService.Link(stat.LocalUID, nodeInfo.LocalUID, nodeInfo.Name)
	}

	f.VirtualTreeService.UpdateLastReaddir(stat.LocalUID)
	// fmt.Println("result", nodeInfos)

	return nodeInfos, nil
}
