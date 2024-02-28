package puterfs

import (
	"context"
	"syscall"
	"time"

	"github.com/HeyPuter/puter-fuse-go/debug"
	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type RootNode struct {
	fs.Inode
	// TODO: Path -> UID
	*Filesystem
	Items        []fao.NodeInfo
	PollDuration time.Duration
	LastPoll     time.Time
	Logger       debug.ILogger
}

func (n *RootNode) Init() {
	n.Items = []fao.NodeInfo{}
	n.PollDuration = 2 * time.Second
	svc_log := n.Filesystem.Services.Get("log").(*debug.LogService)
	n.Logger = svc_log.GetLogger("ROOT")
}

func (n *RootNode) syncItems() error {
	if time.Now().Compare(n.LastPoll.Add(n.PollDuration)) < 0 {
		return nil
	}
	n.LastPoll = time.Now()

	// TODO: Path -> UID
	items, err := n.FAO.ReadDir("/")
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Path == "" {
			panic("item is missing path")
		}
	}

	n.Items = items
	return nil
}

func (n *RootNode) Lookup(
	ctx context.Context, name string, out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	n.Logger.Log("lookup(%s)", name)
	n.syncItems()

	var foundItem fao.NodeInfo
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
		if item.Path == "" {
			panic("item is missing path")
		}
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
