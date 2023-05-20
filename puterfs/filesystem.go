package puterfs

import (
	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/hanwen/go-fuse/v2/fs"
)

type Filesystem struct {
	// map Puter UIDs to inode numbers
	UidInoMap  map[string]uint64
	InoCounter uint64

	Nodes map[uint64]fs.InodeEmbedder

	SDK *putersdk.PuterSDK
}

func (pfs *Filesystem) Init() {
	pfs.UidInoMap = map[string]uint64{}
	pfs.Nodes = map[uint64]fs.InodeEmbedder{}
}

func (fs *Filesystem) GetNodeFromCloudItem(cloudItem putersdk.CloudItem) fs.InodeEmbedder {
	ino := fs.GetInoFromUID(cloudItem.Uid)
	node, exists := fs.Nodes[ino]
	if !exists {
		node = fs.CreateNodeFromCloudItem(cloudItem)
	}
	fs.Nodes[ino] = node
	return node
}

func (fs *Filesystem) GetInoFromUID(uid string) uint64 {
	ino, exists := fs.UidInoMap[uid]
	if !exists {
		fs.InoCounter++
		ino = fs.InoCounter
	}
	fs.UidInoMap[uid] = ino
	return ino
}

func (fs *Filesystem) CreateNodeFromCloudItem(cloudItem putersdk.CloudItem) fs.InodeEmbedder {
	if cloudItem.IsDir {
		return fs.CreateDirNodeFromCloudItem(cloudItem)
	}

	return fs.CreateFileNodeFromCloudItem(cloudItem)
}

// start :: redundant (file,file;dir;directory)

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