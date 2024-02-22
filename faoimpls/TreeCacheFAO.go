package faoimpls

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/fao"
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

func (f *TreeCacheFAO) ReadDir(path string) ([]fao.NodeInfo, error) {
	parts := filepath.SplitList(path)
	entry := f.VirtualTreeService.ResolvePath(parts)

	if entry == nil && entry.LastReaddir.Add(f.TTL).Before(time.Now()) {
		return f.readDirAndUpdateCache(path)
	}

	nodeInfos := []fao.NodeInfo{}
	allExist := true
	for _, localUID := range entry.GetUIDs() {
		nodeInfo := f.AssociationService.LocalUIDToNodeInfo.Get(localUID)
		if nodeInfo == nil {
			allExist = false
			break
		}
		nodeInfos = append(nodeInfos, **nodeInfo)
	}

	if !allExist {
		return f.readDirAndUpdateCache(path)
	}

	fmt.Println("readdir cache hit", path)

	return nodeInfos, nil
}

func (f *TreeCacheFAO) Stat(path string) (fao.NodeInfo, bool, error) {
	localUID, exists := f.AssociationService.PathToLocalUID.Get(path)
	if exists {
		nodeInfo, err := f.AssociationService.LocalUIDToNodeInfo.GetOrSet(
			localUID,
			f.TTL,
			func() (*fao.NodeInfo, error) {
				stat, exists, err := f.Delegate.Stat(path)
				if err != nil {
					return nil, err
				}
				if !exists {
					return nil, nil
				}
				stat.LastStat = time.Now()
				f.AssociationService.PathToLocalUID.Set(path, stat.LocalUID)
				return &stat, nil
			},
		)
		if err != nil {
			return fao.NodeInfo{}, false, err
		}
		if nodeInfo == nil {
			return fao.NodeInfo{}, false, nil
		}

		return *nodeInfo, true, nil
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
		stat.LocalUID, &stat, f.TTL)
	f.AssociationService.PathToLocalUID.Set(path, stat.LocalUID)
	m.Unlock()
	return stat, true, nil
}

func (f *TreeCacheFAO) readDirAndUpdateCache(path string) ([]fao.NodeInfo, error) {
	fmt.Println("readdir cache miss", path)

	// Stat the directory (prerequisite to cache the path association)
	stat, exists, err := f.Stat(path)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, &fao.ErrDoesNotExist{}
	}

	if !stat.IsDir {
		return nil, &fao.ErrNotDirectory{}
	}

	nodeInfos, err := f.Delegate.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Cache the nodeInfos
	for _, nodeInfo := range nodeInfos {
		f.AssociationService.LocalUIDToNodeInfo.Set(nodeInfo.LocalUID, &nodeInfo, f.TTL)
		f.VirtualTreeService.Link(stat.LocalUID, nodeInfo.LocalUID, nodeInfo.Name)
	}

	return nodeInfos, nil
}
