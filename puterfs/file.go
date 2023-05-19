package puterfs

import (
	"context"
	"fmt"
	"syscall"

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
	data, err := n.Filesystem.SDK.Read(n.CloudItem.Path)
	if err != nil {
		return nil, syscall.EIO
	}

	fmt.Printf("this should be > 0 ??  [%d]\n", len(dest))
	fmt.Printf("size from the cloud :) [%d]\n", len(data))

	copy(dest, data[off:])

	return fuse.ReadResultData(dest), 0
}

func (n *FileNode) Write(
	ctx context.Context,
	f fs.FileHandle,
	data []byte, off int64,
) (uint32, syscall.Errno) {
	fileContents, err := n.Filesystem.SDK.Read(n.CloudItem.Path)
	if err != nil {
		return 0, syscall.EIO
	}
	if int64(len(fileContents)) < off+int64(len(data)) {
		newData := make([]byte, off+int64(len(data)))
		copy(newData, fileContents)
		fileContents = newData
	}
	copy(fileContents[off:], data)
	fmt.Println("data: " + string(fileContents))
	err = n.Filesystem.SDK.Write(n.CloudItem.Path, fileContents)
	if err != nil {
		panic(err)
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

	// TODO: load from configuration
	out.Uid = 1000
	out.Gid = 1000

	return 0
}
