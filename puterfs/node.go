package puterfs

import (
	"syscall"

	"github.com/HeyPuter/puter-fuse-go/fao"
)

type HasPuterNodeCapabilities interface {
	GetStableAttrMode() uint32
	GetIno() uint64
}

type CloudItemNode struct {
	*Filesystem
	CloudItem fao.NodeInfo
}

func (n *CloudItemNode) GetStableAttrMode() uint32 {
	if n.CloudItem.IsDir {
		return syscall.S_IFDIR
	}

	if n.CloudItem.IsSymlink {
		return syscall.S_IFLNK
	}

	return syscall.S_IFREG
}

func (n *CloudItemNode) GetIno() uint64 {
	return n.Filesystem.GetInoFromUID(n.CloudItem.Uid)
}
