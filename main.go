package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/HeyPuter/puter-fuse-go/debug"
	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/faoimpls"
	"github.com/HeyPuter/puter-fuse-go/puterfs"
	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/HeyPuter/puter-fuse-go/services"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
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

	// If it doesn't exist, add .config/puterfuse
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		panic(fmt.Errorf(("error getting user config directory: %s"), err))
	}

	puterfuseConfigDir := filepath.Join(userConfigDir, "puterfuse")
	err = os.MkdirAll(puterfuseConfigDir, 0755)
	if err != nil {
		panic(fmt.Errorf("error creating config directory: %s", err))
	}

	// If it doesn't exist, add config.yaml

	// TODO: this should go in ConfigService
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/puterfuse")
	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if !viper.IsSet("cacheDir") {
		var puterfuseCacheDir string
		if viper.GetBool("useUserCacheDir") {
			userCacheDir, err := os.UserCacheDir()
			if err != nil {
				panic(fmt.Errorf("error getting user cache directory: %s", err))
			}
			puterfuseCacheDir = filepath.Join(userCacheDir, "puterfuse")
		} else {
			puterfuseCacheDir = "/tmp/puterfuse"
		}

		err = os.MkdirAll(puterfuseCacheDir, 0755)
		if err != nil {
			panic(fmt.Errorf("error creating cache directory: %s", err))
		}

		viper.Set("cacheDir", puterfuseCacheDir)
	}

	sdk := &putersdk.PuterSDK{
		Url:            viper.GetString("url"),
		PuterAuthToken: viper.GetString("token"),
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
	svcc.Set("association", engine.CreateAssociationService())
	svcc.Set("blob-cache", engine.CreateBLOBCacheService(afero.NewOsFs()))

	for _, svc := range svcc.All() {
		svc.Init(svcc)
	}

	fao := faoimpls.CreatePuterFAO(
		faoimpls.P_PuterFAO{
			SDK: sdk,
		},
		faoimpls.D_PuterFAO{
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

	mountPoint := viper.GetString("mountPoint")
	if mountPoint == "" {
		mountPoint = "/tmp/mnt"
	}

	// Ensure /tmp/mnt exists
	err = os.MkdirAll(mountPoint, 0755)
	if err != nil {
		panic(err)
	}

	server, err := fs.Mount(mountPoint, rootNode, &fs.Options{})
	if err != nil {
		panic(err)
	}

	// Print debug info
	fmt.Println("Server started")
	fmt.Println("Configuration file:", viper.ConfigFileUsed())
	fmt.Println("Mountpoint:", mountPoint)
	fmt.Println("Cache directory:", viper.GetString("cacheDir"))

	// start serving the file system
	server.Wait()
}
