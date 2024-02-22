package engine

import (
	"sync"

	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/HeyPuter/puter-fuse-go/kvdotgo"
	"github.com/HeyPuter/puter-fuse-go/lang"
	"github.com/HeyPuter/puter-fuse-go/services"
)

type AssociationService struct {
	LocalUIDToRemoteUID lang.IMap[string, string]
	RemoteUIDToLocalUID lang.IMap[string, string]
	LocalUIDToIno       lang.IMap[string, uint64]
	InoToLocalUID       lang.IMap[uint64, string]

	// LocalUIDToNodeInfo   lang.IMap[string, *fao.NodeInfo]
	LocalUIDToNodeInfo *kvdotgo.KVMap[string, *fao.NodeInfo]
	// PathToLocalUID     *kvdotgo.KVMap[string, string]
	PathToLocalUID lang.IMap[string, string]

	CacheStampedeMapLock sync.RWMutex
	CacheStampedeMap     map[string]*sync.Mutex
}

func (svc *AssociationService) Init(services services.IServiceContainer) {}

func CreateAssociationService() *AssociationService {
	ins := &AssociationService{}

	ins.LocalUIDToRemoteUID = lang.CreateSyncMap[string, string](nil)
	ins.RemoteUIDToLocalUID = lang.CreateSyncMap[string, string](nil)
	ins.LocalUIDToIno = lang.CreateSyncMap[string, uint64](nil)
	ins.InoToLocalUID = lang.CreateSyncMap[uint64, string](nil)

	// ins.LocalUIDToNodeInfo = lang.CreateSyncMap[string, *fao.NodeInfo](nil)
	ins.LocalUIDToNodeInfo = kvdotgo.CreateKVMap[string, *fao.NodeInfo]()

	return ins
}
