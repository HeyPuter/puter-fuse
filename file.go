package main

import (
	"context"
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type PuterFSFileInode struct {
	fs.Inode
	Contents []byte
}

func (n *PuterFSFileInode) GetSize() uint64 {
	return uint64(len(n.Contents))
}

func (n *PuterFSFileInode) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	fmt.Printf("open was called\n")
	return &PuterFSFile{
		Node: n,
	}, 0, 0
}

func (n *PuterFSFileInode) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	fmt.Printf("Read was called\n")
	puterFile := f.(*PuterFSFile)
	copy(dest, puterFile.GetData())
	return fuse.ReadResultData(dest), 0
}

func (n *PuterFSFileInode) Write(ctx context.Context, f fs.FileHandle, data []byte, off int64) (uint32, syscall.Errno) {
	fh := f.(*PuterFSFile)
	if int64(len(data))+off > int64(fh.GetSize()) {
		newData := make([]byte, off+int64(len(data)))
		copy(newData, fh.GetData())
		fh.ReplaceData(newData)
	}

	copy(fh.GetData()[off:], data)

	return uint32(len(data)), 0
}

func (n *PuterFSFileInode) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	return 0
}

func (n *PuterFSFileInode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	fmt.Println("getattr: puter file handle")
	out.Mode = 0644
	out.Size = n.GetSize()
	fmt.Printf("reporting the size as: %d\n", out.Size)
	out.Uid = 1000
	out.Gid = 1000
	fmt.Printf("Whats ino? [%d]\n", out.Ino)
	return 0
}

var _ = (fs.InodeEmbedder)((*PuterFSFileInode)(nil))
