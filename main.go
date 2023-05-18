package main

import (
	"github.com/hanwen/go-fuse/v2/fs"
)

type PuterFSFile struct {
	UID  string
	Node *PuterFSFileInode
}

func (fh *PuterFSFile) GetSize() uint64 {
	return uint64(len(fh.Node.Contents))
}

func (fh *PuterFSFile) GetData() []byte {
	return fh.Node.Contents
}

func (fh *PuterFSFile) ReplaceData(newData []byte) {
	fh.Node.Contents = newData
}

func main() {
	server, err := fs.Mount("/tmp/mnt", &PuterFSDirectoryInode{}, &fs.Options{})
	if err != nil {
		panic(err)
	}
	// start serving the file system
	server.Wait()
}
