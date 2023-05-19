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

type RootNode struct {
	fs.Inode
	// TODO: Path -> UID
	*Filesystem
	Items        []putersdk.CloudItem
	PollDuration time.Duration
	LastPoll     time.Time
}

func (n *RootNode) Init() {
	n.Items = []putersdk.CloudItem{}
	n.PollDuration = 2 * time.Second
}

func (n *RootNode) syncItems() error {
	if time.Now().Compare(n.LastPoll.Add(n.PollDuration)) < 0 {
		return nil
	}
	n.LastPoll = time.Now()

	// TODO: Path -> UID
	items, err := n.Filesystem.SDK.Readdir("/")
	if err != nil {
		return err
	}

	n.Items = items
	return nil
}

func (n *RootNode) Lookup(
	ctx context.Context, name string, out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	fmt.Printf("root::lookup(%s)\n", name)
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

func (n *RootNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
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
