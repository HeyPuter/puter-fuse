package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/HeyPuter/puter-fuse-go/debug"
	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/fao"
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

	// viper defaults
	viper.SetDefault("treeCacheTTL", "2h")
	// viper.SetDefault("treeCacheTTL", "2s")

	// TODO: change this default before release
	fmt.Printf("\x1B[33;1mWARNING: fileReadCacheTTL DEFAULTS TO 30s\x1B[0m\n")
	viper.SetDefault("fileReadCacheTTL", "2h")

	if viper.GetBool("testMode") {
		viper.SetDefault("treeCacheTTL", "5s")

		viper.SetDefault("testDelay", "200ms")
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
	svcc.Set("virtual-tree", engine.CreateVirtualTreeService())
	svcc.Set("config", engine.CreateConfigService())
	svcc.Set("blob-cache", engine.CreateBLOBCacheService(afero.NewOsFs()))
	svcc.Set("write-cache", engine.CreateWriteCacheService())

	for _, svc := range svcc.All() {
		svc.Init(svcc)
	}

	var fao fao.FAO
	if viper.GetBool("testMode") {
		memFAO := faoimpls.CreateMemFAO()
		fao = memFAO
		// Populate with test data
		{
			fao.MkDir("/", "user")
			fao.MkDir("/user", "one-file")
			fao.Create("/user/one-file", "file")
			fao.Write("/user/one-file/file", []byte("file"), 0)
			fao.MkDir("/user", "three-files")
			for i := 0; i < 3; i++ {
				fao.Create("/user/three-files", fmt.Sprintf("file-%d", i))
				fao.Write(fmt.Sprintf("/user/three-files/file-%d", i),
					[]byte(fmt.Sprintf("file-%d", i)), 0)
			}
			fao.MkDir("/user", "fifty-files")
			for i := 0; i < 50; i++ {
				fao.Create("/user/fifty-files", fmt.Sprintf("file-%d", i))
				fao.Write(fmt.Sprintf("/user/fifty-files/file-%d", i),
					[]byte(fmt.Sprintf("file-%d", i)), 0)
			}
		}
		fao = faoimpls.CreateSlowFAO(fao, viper.GetDuration("testDelay"))
		fao = faoimpls.CreateLogFAO(
			fao,
			svcc.Get("log").(*debug.LogService).GetLogger("test-storage"),
		)
	} else {
		fao = faoimpls.CreatePuterFAO(
			faoimpls.P_PuterFAO{
				SDK: sdk,
			},
			faoimpls.D_PuterFAO{
				EnqueueOperationRequest: svcc.Get("operation").(*engine.OperationService).EnqueueOperationRequest,
			},
		)
		fao.(*faoimpls.PuterFAO).ReadFAO = fao
	}

	fao = faoimpls.CreateRemoteToLocalUIDFAO(fao, svcc)

	if viper.GetBool("experimental_cache") {
		fao = faoimpls.CreateFileReadCacheFAO(fao, svcc, faoimpls.P_FileReadCacheFAO{
			TTL: viper.GetDuration("fileReadCacheTTL"),
		})
	}

	treeCacheFAOTTL, err := time.ParseDuration(viper.GetString("treeCacheTTL"))
	if err != nil {
		panic(err)
	}

	if viper.GetBool("experimental_cache") {
		fao = faoimpls.CreateTreeCacheFAO(
			fao,
			faoimpls.P_TreeCacheFAO{
				TTL: treeCacheFAOTTL,
			},
			faoimpls.D_TreeCacheFAO{
				VirtualTreeService: svcc.Get("virtual-tree").(*engine.VirtualTreeService),
				AssociationService: svcc.Get("association").(*engine.AssociationService),
			},
		)
	}

	if viper.GetBool("experimental_cache") {
		fao = faoimpls.CreateFileWriteCacheFAO(fao, svcc)
	}

	fao = faoimpls.CreateLogFAO(
		fao,
		svcc.Get("log").(*debug.LogService).GetLogger("top"),
	)

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

	if viper.GetBool("panik") {
		fmt.Printf("\n\x1B[31;1m=== Panik mode is enabled ===\x1B[0m\n\n")
	}

	// start serving the file system
	server.Wait()
}
