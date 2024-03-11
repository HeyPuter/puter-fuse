/*
 * Copyright (C) 2024  Puter Technologies Inc.
 *
 * This file is part of puter-fuse.
 *
 * puter-fuse is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
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
