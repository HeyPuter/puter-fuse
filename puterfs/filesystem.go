package puterfs

import (
	"fmt"
	"log"
	"sync"

	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/hanwen/go-fuse/v2/fs"
)

type Filesystem struct {
	// map Puter UIDs to inode numbers
	UidInoMap  map[string]uint64
	InoCounter uint64

	Nodes map[uint64]fs.InodeEmbedder

	SDK *putersdk.PuterSDK

	NodesMutex     sync.RWMutex
	UidInoMapMutex sync.RWMutex
}

func (pfs *Filesystem) Init() {
	pfs.UidInoMap = map[string]uint64{}
	pfs.Nodes = map[uint64]fs.InodeEmbedder{}
}

func (fs *Filesystem) GetNodeFromCloudItem(cloudItem putersdk.CloudItem) fs.InodeEmbedder {
	ino := fs.GetInoFromUID(cloudItem.Uid)
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
	return node
}

func (fs *Filesystem) GetInoFromUID(uid string) uint64 {
	fs.UidInoMapMutex.RLock()
	ino, exists := fs.UidInoMap[uid]
	fs.UidInoMapMutex.RUnlock()
	if !exists {
		fs.UidInoMapMutex.Lock()
		// check again in case another thread just did this
		ino, exists = fs.UidInoMap[uid]
		if !exists {
			fs.InoCounter++
			ino = fs.InoCounter
			fs.UidInoMap[uid] = ino
		}
		fs.UidInoMapMutex.Unlock()
	}
	return ino
}

func (fs *Filesystem) CreateNodeFromCloudItem(cloudItem putersdk.CloudItem) fs.InodeEmbedder {
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

func (pfs *Filesystem) CreateDirNodeFromCloudItem(cloudItem putersdk.CloudItem) fs.InodeEmbedder {
	dirNode := &DirectoryNode{}
	dirNode.CloudItem = cloudItem
	dirNode.Filesystem = pfs

	dirNode.Init()

	return dirNode
}

func (pfs *Filesystem) CreateFileNodeFromCloudItem(cloudItem putersdk.CloudItem) fs.InodeEmbedder {
	fileNode := &FileNode{}
	fileNode.CloudItem = cloudItem
	fileNode.Filesystem = pfs

	fileNode.Init()

	return fileNode
}

// end :: redundant
