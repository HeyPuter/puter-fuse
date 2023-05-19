package puterfs

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type DirectoryNode struct {
	fs.Inode
	CloudItemNode
	Items        []putersdk.CloudItem
	PollDuration time.Duration
	LastPoll     time.Time
}

func (n *DirectoryNode) Init() {
	n.Items = []putersdk.CloudItem{}
	n.PollDuration = 2 * time.Second
}

func (n *DirectoryNode) syncItems() error {
	if time.Now().Compare(n.LastPoll.Add(n.PollDuration)) < 0 {
		return nil
	}
	n.LastPoll = time.Now()

	// TODO: Path -> UID
	items, err := n.CloudItemNode.Filesystem.SDK.Readdir(n.CloudItem.Path)
	if err != nil {
		return err
	}

	n.Items = items
	return nil
}

func (n *DirectoryNode) Lookup(
	ctx context.Context, name string, out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	fmt.Printf("dir::lookup(%s)\n", name)
	n.syncItems()

	var foundItem putersdk.CloudItem
	var found bool

	for _, item := range n.Items {
		if item.Name == name {
			foundItem = item
			found = true
			break
		}
	}

	if !found {
		// TODO: return an error code?
		return nil, syscall.ENOENT
	}
	foundItemNode := n.Filesystem.GetNodeFromCloudItem(foundItem)

	iface := foundItemNode.(HasPuterNodeCapabilities)

	return n.NewInode(
		ctx,
		foundItemNode,
		fs.StableAttr{
			Mode: iface.GetStableAttrMode(),
			Ino:  iface.GetIno(),
		},
	), 0
}

func (n *DirectoryNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	n.syncItems()

	entries := []fuse.DirEntry{}
	for _, item := range n.Items {
		node := n.Filesystem.GetNodeFromCloudItem(item)
		iface := node.(HasPuterNodeCapabilities)
		entry := fuse.DirEntry{
			Mode: iface.GetStableAttrMode(),
			Name: item.Name,
			Ino:  iface.GetIno(),
		}
		entries = append(entries, entry)
	}
	return fs.NewListDirStream(entries), 0
}

func (n *DirectoryNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Size = 4096

	// TODO: load from configuration
	// out.Mode = 0644
	out.Mode = 0755

	// TODO: load from configuration
	out.Uid = 1000
	out.Gid = 1000

	return 0
}
