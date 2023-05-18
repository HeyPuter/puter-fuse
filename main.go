package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/HeyPuter/puter-fuse-go/putersdk"
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
	args := os.Args[1:]
	fmt.Println(args)

	token, err := ioutil.ReadFile("token")
	if err != nil {
		panic(err)
	}

	fmt.Printf("token: |%s|\n", string(token))

	sdk := &putersdk.PuterSDK{
		PuterAuthToken: string(token),
	}
	sdk.Init()

	// items, err := sdk.Readdir("/")
	// if err != nil {
	// 	panic(err)
	// }

	// jsonBytes, err := json.Marshal(items)

	jsonBytes, err := sdk.Read("/ed/test.txt")
	fmt.Println(string(jsonBytes))
	/*
		server, err := fs.Mount("/tmp/mnt", &PuterFSDirectoryInode{}, &fs.Options{})
		if err != nil {
			panic(err)
		}
		// start serving the file system
		server.Wait()
	*/
}
