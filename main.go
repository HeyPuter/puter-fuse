package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/HeyPuter/puter-fuse-go/debug"
	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/fao"
	"github.com/HeyPuter/puter-fuse-go/puterfs"
	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/HeyPuter/puter-fuse-go/services"
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
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))

	svcc := &services.ServicesContainer{}
	svcc.Init()

	svcc.Set("operation", &engine.OperationService{
		SDK: sdk,
	})
	svcc.Set("pending-node", &engine.PendingNodeService{})
	svcc.Set("wfcache", &engine.WholeFileCacheService{})
	svcc.Set("log", &debug.LogService{})

	for _, svc := range svcc.All() {
		svc.Init(svcc)
	}

	fao := fao.CreatePuterFAO(
		fao.P_PuterFAO{
			SDK: sdk,
		},
		fao.D_PuterFAO{
			EnqueueOperationRequest: svcc.Get("operation").(*engine.OperationService).EnqueueOperationRequest,
		},
	)

	fao.ReadFAO = fao

	puterFS := &puterfs.Filesystem{
		SDK:      sdk,
		FAO:      fao,
		Services: svcc,
	}
	puterFS.Init()

	rootNode := &puterfs.RootNode{}
	rootNode.Filesystem = puterFS
	rootNode.Init()

	// Ensure /tmp/mnt exists
	err = os.MkdirAll("/tmp/mnt", 0755)
	if err != nil {
		panic(err)
	}

	server, err := fs.Mount("/tmp/mnt", rootNode, &fs.Options{})
	if err != nil {
		panic(err)
	}
	// start serving the file system
	server.Wait()
}
