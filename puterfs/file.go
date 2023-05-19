package puterfs

import (
	"github.com/hanwen/go-fuse/v2/fs"
)

type FileNode struct {
	fs.Inode
	CloudItemNode
}

func (n *FileNode) Init() {
}
