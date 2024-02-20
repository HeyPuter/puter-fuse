package puterfs

import (
	"context"
	"syscall"

	"github.com/HeyPuter/puter-fuse-go/debug"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type FileNode struct {
	fs.Inode
	CloudItemNode
	Logger debug.ILogger
}

func (n *FileNode) Init() {
	svc_log := n.Filesystem.Services.Get("log").(*debug.LogService)
	n.Logger = svc_log.GetLogger("Inode:R " + n.CloudItem.Path)
}

func (n *FileNode) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	fh := &FileHandler{
		Node: n,
	}
	return fh, 0, 0
}

func (n *FileNode) Read(
	ctx context.Context,
	f fs.FileHandle,
	dest []byte, off int64,
) (fuse.ReadResult, syscall.Errno) {
	n.Logger.Log("read(%s)", n.CloudItem.Path)

	_, err := n.FAO.Read(n.CloudItem.Path, dest, off)
	if err != nil {
		return nil, syscall.EIO
	}

	return fuse.ReadResultData(dest), 0
}

func (n *FileNode) Write(
	ctx context.Context,
	f fs.FileHandle,
	data []byte, off int64,
) (uint32, syscall.Errno) {
	amount, err := n.FAO.Write(n.CloudItem.Path, data, off)

	if err != nil {
		return 0, syscall.EIO
	}

	return uint32(amount), 0
}

// func (n *FileNode) Write(
// 	ctx context.Context,
// 	f fs.FileHandle,

// )

func (n *FileNode) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return 0
}

func (n *FileNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Size = n.CloudItem.Size

	// TODO: load from configuration
	// out.Mode = 0644
	out.Mode = 0644

	if n.CloudItem.IsSymlink {
		out.Mode = out.Mode | 0120000
	} else {
		out.Mode = out.Mode | 0100000
	}

	// TODO: load from configuration
	out.Uid = 1000
	out.Gid = 1000

	out.Mtime = uint64(n.CloudItem.Modified)
	out.Atime = uint64(n.CloudItem.Accessed)
	out.Ctime = uint64(n.CloudItem.Created)

	return 0
}

func (n *FileNode) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {

	// TODO: modify attributes
	// this NO-OP is here so commands like `touch` exit without error
	if in.Valid&fuse.FATTR_SIZE != 0 && in.Size != n.CloudItem.Size {
		n.FAO.Truncate(n.CloudItem.Path, in.Size)
	}
	return 0
}

func (n *FileNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	if !n.CloudItem.IsSymlink {
		return nil, syscall.EINVAL
	}

	return []byte(n.CloudItem.SymlinkPath), 0
}
