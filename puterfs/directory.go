package puterfs

import (
	"context"
	"fmt"
	"path/filepath"
	"syscall"
	"time"

	"github.com/HeyPuter/puter-fuse-go/debug"
	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/localutil"
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
	Logger       debug.ILogger
}

func (n *DirectoryNode) Init() {
	// call super init
	n.CloudItemNode.Init()
	n.Items = []putersdk.CloudItem{}
	n.PollDuration = 2 * time.Second
	svc_log := n.Filesystem.Services.Get("log").(*debug.LogService)
	n.Logger = svc_log.GetLogger("Inode:D " + n.CloudItem.Path)
}

func (n *DirectoryNode) syncItems() error {
	if time.Now().Compare(n.LastPoll.Add(n.PollDuration)) < 0 {
		return nil
	}
	n.LastPoll = time.Now()

	// TODO: Path -> UID
	var items []putersdk.CloudItem
	var err error

	if !n.IsPending {
		items, err = n.CloudItemNode.Filesystem.SDK.Readdir(n.Logger.S("SDK"), n.CloudItem.Path)
		if err != nil {
			return err
		}
	} else {
		items = []putersdk.CloudItem{}
	}

	svc_pendingNode := n.CloudItemNode.Filesystem.Services.Get("pending-node").(*engine.PendingNodeService)
	pendingChildren := svc_pendingNode.GetChildren(n.CloudItem.Path)
	for _, pendingChild := range pendingChildren {
		var alreadyExists bool
		for _, item := range items {
			if item.Name == pendingChild.Name {
				alreadyExists = true
				break
			}
		}
		if alreadyExists {
			continue
		}

		// create artificial cloud item
		cloudItem := engine.NodeInfoToArtificialCloudItem(pendingChild)

		items = append(items, cloudItem)
	}

	n.Items = items
	return nil
}

func (n *DirectoryNode) lookupCloudItem(
	name string,
) (putersdk.CloudItem, bool) {
	var foundItem putersdk.CloudItem
	var found bool

	for _, item := range n.Items {
		if item.Name == name {
			foundItem = item
			found = true
			break
		}
	}

	return foundItem, found
}

func (n *DirectoryNode) addOrReplaceCloudItem(
	name string, item putersdk.CloudItem,
) {
	for i, existingItem := range n.Items {
		if existingItem.Name == name {
			n.Items[i] = item
			return
		}
	}

	n.Items = append(n.Items, item)
}

func (n *DirectoryNode) Lookup(
	ctx context.Context, name string, out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	n.Logger.Log("lookup(%s)", name)
	n.syncItems()

	foundItem, found := n.lookupCloudItem(name)

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

func (n *DirectoryNode) Symlink(
	ctx context.Context,
	target string,
	name string,
	out *fuse.EntryOut,
) (*fs.Inode, syscall.Errno) {
	fmt.Printf("dir::symlink(%s)\n", name)
	n.syncItems()

	newFilePath := filepath.Join(n.CloudItem.Path, name)
	cloudItem, err := n.Filesystem.SDK.Symlink(newFilePath, target)
	if err != nil {
		fmt.Println("THIS IS WHERE THE ERROR IS")
		fmt.Println(err.Error())
		fmt.Println("AND THIS IS AFTER THAT")
		return nil, syscall.EIO
	}

	n.Items = append(n.Items, *cloudItem)

	cloudItemNode := n.Filesystem.GetNodeFromCloudItem(*cloudItem)

	iface := cloudItemNode.(HasPuterNodeCapabilities)

	return n.NewInode(
		ctx,
		cloudItemNode,
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
	out.Mode = out.Mode | 040000

	// TODO: load from configuration
	out.Uid = 1000
	out.Gid = 1000

	return 0
}

func (n *DirectoryNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	n.Logger.Log("create(%s)", name)
	// check if file already exists
	{
		_, found := n.lookupCloudItem(name)
		if found {
			n.Logger.Log("create(%s) -> EEXIST", name)
			return nil, nil, 0, syscall.EEXIST
		}
	}

	newFilePath := filepath.Join(n.CloudItem.Path, name)
	svc_operation := n.Filesystem.Services.Get("operation").(*engine.OperationService)
	svc_wfcache := n.Filesystem.Services.Get("wfcache").(*engine.WholeFileCacheService)
	ver := svc_wfcache.SetFileData(newFilePath, []byte{})

	svc_pendingNode := n.Filesystem.Services.Get("pending-node").(*engine.PendingNodeService)
	nodeInfo := svc_pendingNode.Link(n.CloudItem.Path, name, engine.File)

	pendingItem := engine.NodeInfoToArtificialCloudItem(nodeInfo)
	n.addOrReplaceCloudItem(name, pendingItem)

	go func() {
		resp := <-svc_operation.EnqueueOperationRequest(
			putersdk.Operation{
				"op":        "write",
				"path":      n.CloudItem.Path,
				"name":      name,
				"overwrite": false,
			},
			[]byte{},
		).Await

		cloudItem := &putersdk.CloudItem{}
		err := localutil.ReJSON(resp.Data, cloudItem)
		if err != nil {
			panic(err)
		}
		svc_wfcache.DeleteFileData(newFilePath, ver)
		n.addOrReplaceCloudItem(name, *cloudItem)
	}()

	// cloudItem, err := n.Filesystem.SDK.Write(newFilePath, []byte{})
	// if err != nil {
	// 	return nil, nil, 0, syscall.EIO
	// }

	// n.Items = append(n.Items, *cloudItem)

	// cloudItemNode := n.Filesystem.GetNodeFromCloudItem(*cloudItem)

	// iface := cloudItemNode.(HasPuterNodeCapabilities)

	cloudItemNode := n.Filesystem.GetNodeFromCloudItem(pendingItem)
	iface := cloudItemNode.(HasPuterNodeCapabilities)

	return n.NewInode(
		ctx,
		cloudItemNode,
		fs.StableAttr{
			Mode: iface.GetStableAttrMode(),
			Ino:  iface.GetIno(),
		},
	), &FileHandler{Node: cloudItemNode.(*FileNode)}, 0, 0
}

func (n *DirectoryNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	svc_pendingNode := n.Filesystem.Services.Get("pending-node").(*engine.PendingNodeService)
	nodeInfo := svc_pendingNode.Link(n.CloudItem.Path, name, engine.Dir)

	pendingItem := engine.NodeInfoToArtificialCloudItem(nodeInfo)
	n.addOrReplaceCloudItem(name, pendingItem)

	localutil.Printvar(nodeInfo, "nodeInfo")
	localutil.Printvar(pendingItem, "pendingItem")

	svc_operation := n.Filesystem.Services.Get("operation").(*engine.OperationService)

	go func() {
		resp := <-svc_operation.EnqueueOperationRequest(
			putersdk.Operation{
				"op":     "mkdir",
				"parent": n.CloudItem.Path,
				"path":   name,
			},
			nil,
		).Await

		cloudItem := &putersdk.CloudItem{}
		err := localutil.ReJSON(resp.Data, cloudItem)
		if err != nil {
			panic(err)
		}
		n.addOrReplaceCloudItem(name, *cloudItem)
	}()

	// dirPath := filepath.Join(n.CloudItem.Path, name)
	// cloudItem, err := n.Filesystem.SDK.Mkdir(dirPath)
	// if err != nil {
	// 	return nil, syscall.EIO
	// }

	// n.Items = append(n.Items, cloudItem)

	// cloudItemNode := n.Filesystem.GetNodeFromCloudItem(cloudItem)

	// iface := cloudItemNode.(HasPuterNodeCapabilities)

	cloudItemNode := n.Filesystem.GetNodeFromCloudItem(pendingItem)
	iface := cloudItemNode.(HasPuterNodeCapabilities)

	return n.NewInode(
		ctx,
		cloudItemNode,
		fs.StableAttr{
			Mode: iface.GetStableAttrMode(),
			Ino:  iface.GetIno(),
		},
	), 0
}

func (n *DirectoryNode) Unlink(ctx context.Context, name string) syscall.Errno {
	{
		cloudItem, found := n.lookupCloudItem(name)
		if !found {
			return syscall.ENOENT
		}
		if cloudItem.IsDir {
			return syscall.EISDIR
		}
	}

	filePath := filepath.Join(n.CloudItem.Path, name)
	err := n.Filesystem.SDK.Delete(filePath)
	if err != nil {
		return syscall.EIO
	}
	// TODO: remove node properly
	n.LastPoll = time.Now().Add(-2 * n.PollDuration)
	return 0
}

func (n *DirectoryNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	{
		cloudItem, found := n.lookupCloudItem(name)
		if !found {
			return syscall.ENOENT
		}
		if !cloudItem.IsDir {
			return syscall.ENOTDIR
		}
	}

	filePath := filepath.Join(n.CloudItem.Path, name)
	err := n.Filesystem.SDK.Delete(filePath)
	if err != nil {
		return syscall.EIO
	}
	// TODO: remove node properly
	n.LastPoll = time.Now().Add(-2 * n.PollDuration)
	return 0
}

func (n *DirectoryNode) Rename(
	ctx context.Context,
	name string,
	newParent fs.InodeEmbedder,
	newName string,
	flags uint32,
) syscall.Errno {
	sourcePath := filepath.Join(n.CloudItem.Path, name)
	parentNode := newParent.(*DirectoryNode)
	_, err := n.SDK.Move(sourcePath, parentNode.CloudItem.Path, newName)
	if err != nil {
		return syscall.EIO
	}
	return 0
}
