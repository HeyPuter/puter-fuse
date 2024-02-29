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
