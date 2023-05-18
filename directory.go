package main

import (
	"context"
	"fmt"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type PuterFSDirectoryInode struct {
	fs.Inode
}

func (n *PuterFSDirectoryInode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fmt.Printf("lookup(%s)\n", name)
	inode := &PuterFSFileInode{
		Contents: []byte("hello\nworld\n"),
	}
	return n.NewInode(ctx,
		inode,
		fs.StableAttr{Mode: syscall.S_IFREG, Ino: 1},
	), 0
}

func (n *PuterFSDirectoryInode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	fmt.Println("readdir")
	return fs.NewListDirStream([]fuse.DirEntry{
		{
			Name: "test.txt",
		},
	}), 0
}
