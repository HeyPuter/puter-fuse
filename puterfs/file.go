package puterfs

import (
	"context"
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/localutil"
	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type FileNode struct {
	fs.Inode
	CloudItemNode
}

func (n *FileNode) Init() {
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
	fmt.Printf("file::read(%s)\n", n.CloudItem.Path)

	// Try cache first
	svc_wfcache := n.Filesystem.Services.Get("wfcache").(*engine.WholeFileCacheService)
	data := svc_wfcache.GetFileData(n.CloudItem.Path)
	if data != nil {
		fmt.Printf("cache hit\n")
		copy(dest, data[off:])
		return fuse.ReadResultData(dest), 0
	}

	data, err := n.Filesystem.SDK.Read(n.CloudItem.Path)
	if err != nil {
		return nil, syscall.EIO
	}

	fmt.Printf("this should be > 0 ??  [%d]\n", len(dest))
	fmt.Printf("size from the cloud :) [%d]\n", len(data))
	fmt.Printf("off [%d] data [%s]\n", off, data[off:])

	copy(dest, data[off:])

	fmt.Printf("amount [%d] data [%s]\n", len(dest), string(dest))

	return fuse.ReadResultData(dest), 0
}

func (n *FileNode) Write(
	ctx context.Context,
	f fs.FileHandle,
	data []byte, off int64,
) (uint32, syscall.Errno) {
	svc_wfcache := n.Filesystem.Services.Get("wfcache").(*engine.WholeFileCacheService)

	fileContents := svc_wfcache.GetFileData(n.CloudItem.Path)
	var err error
	if fileContents == nil {
		fileContents, err = n.Filesystem.SDK.Read(n.CloudItem.Path)
	}
	if err != nil {
		return 0, syscall.EIO
	}
	if int64(len(fileContents)) < off+int64(len(data)) {
		newData := make([]byte, off+int64(len(data)))
		copy(newData, fileContents)
		fileContents = newData
	}
	copy(fileContents[off:], data)

	svc_operation := n.Filesystem.Services.Get("operation").(*engine.OperationService)

	dirname := filepath.Dir(n.CloudItem.Path)
	name := filepath.Base(n.CloudItem.Path)

	svc_wfcache.SetFileData(n.CloudItem.Path, fileContents)
	n.CloudItem.Size = uint64(len(fileContents))

	if false {
		go func() {
			resp := <-svc_operation.EnqueueOperationRequest(
				putersdk.Operation{
					"op":        "write",
					"path":      dirname,
					"name":      name,
					"overwrite": true,
				},
				fileContents,
			).Await

			cloudItem := &putersdk.CloudItem{}
			err = localutil.ReJSON(resp.Data, cloudItem)
			if err != nil {
				panic(err)
			}
			svc_wfcache.DeleteFileData(n.CloudItem.Path)
			n.CloudItem = *cloudItem
		}()
	}
	return uint32(len(data)), 0
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
		fileContents, err := n.Filesystem.SDK.Read(n.CloudItem.Path)
		if err != nil {
			return syscall.EIO
		}
		cloudItem, err := n.Filesystem.SDK.Write(n.CloudItem.Path, fileContents[:in.Size])
		if err != nil {
			panic(err)
		}
		n.CloudItem = *cloudItem
	}
	return 0
}

func (n *FileNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	if !n.CloudItem.IsSymlink {
		return nil, syscall.EINVAL
	}

	return []byte(n.CloudItem.SymlinkPath), 0
}
