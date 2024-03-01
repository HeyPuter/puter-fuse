package engine

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/HeyPuter/puter-fuse-go/lang"
	"github.com/HeyPuter/puter-fuse-go/services"
	"github.com/google/uuid"
	"github.com/spf13/afero"
)

type BLOBCacheReference struct {
	entry *BLOBCacheEntry
}

func (ref *BLOBCacheReference) GetHash() string {
	return ref.entry.Hash
}

func (ref *BLOBCacheReference) Release() {
	ref.entry.ReferencesLock.Lock()
	defer ref.entry.ReferencesLock.Unlock()

	fmt.Println("looking for ref in", []*BLOBCacheReference{ref}, ref.entry.References)
	for i, r := range ref.entry.References {
		if r == ref {
			fmt.Println("found ref in", ref, r, ref.entry.References)
			ref.entry.References = append(
				ref.entry.References[:i],
				ref.entry.References[i+1:]...,
			)
			break
		}
	}

	fmt.Println("ref count is", len(ref.entry.References))

	if len(ref.entry.References) == 0 {
		fmt.Println("ref not found in", ref.entry.References)
		close(ref.entry.AwaitRelease)
	}
}

func (ref *BLOBCacheReference) AwaitForgotten() <-chan struct{} {
	return ref.entry.AwaitForgotten
}

func (ref *BLOBCacheReference) AwaitRemovedFromFS() <-chan struct{} {
	return ref.entry.AwaitRemovedFromFS
}

type BLOBCacheEntry struct {
	Uid                string
	Hash               string
	ReferencesLock     sync.RWMutex
	References         []*BLOBCacheReference
	AwaitRelease       chan struct{}
	AwaitForgotten     chan struct{}
	AwaitRemovedFromFS chan struct{}
}

type BLOBCacheService struct {
	ConfigService IConfig
	KnownBlobs    lang.IMap[string, *BLOBCacheEntry]
	Filesystem    afero.Fs
}

func (svc *BLOBCacheService) Init(services services.IServiceContainer) {
	svc.ConfigService = services.Get("config").(*ConfigService)
}

func CreateBLOBCacheService(fs afero.Fs) *BLOBCacheService {
	return &BLOBCacheService{
		Filesystem: fs,
		KnownBlobs: lang.CreateSyncMap[string, *BLOBCacheEntry](nil),
	}
}

func (svc *BLOBCacheService) Store(
	reader io.Reader,
) *BLOBCacheReference {
	// {
	// 	maybeRef := svc.Hold(hash)
	// 	if maybeRef != nil {
	// 		return maybeRef
	// 	}
	// }

	tmpid := uuid.New().String()

	ref := &BLOBCacheReference{}

	entry := &BLOBCacheEntry{
		ReferencesLock:     sync.RWMutex{},
		References:         []*BLOBCacheReference{ref},
		AwaitRelease:       make(chan struct{}),
		AwaitForgotten:     make(chan struct{}),
		AwaitRemovedFromFS: make(chan struct{}),
	}

	ref.entry = entry

	hasher := sha1.New()
	reader = io.TeeReader(reader, hasher)
	svc.storeFile(tmpid, reader)

	// TODO: see if we can remove encode to hex (i.e. is []byte "comparable"?)
	hash := hex.EncodeToString(hasher.Sum(nil))
	entry.Hash = hash

	svc.Filesystem.Rename(
		filepath.Join(
			svc.ConfigService.GetString("cacheDir"),
			tmpid,
		),
		filepath.Join(
			svc.ConfigService.GetString("cacheDir"),
			hash,
		),
	)

	svc.KnownBlobs.Set(hash, entry)

	go func() {
		<-entry.AwaitRelease
		svc.KnownBlobs.Del(hash)
		close(entry.AwaitForgotten)
		svc.deleteFile(hash)
		close(entry.AwaitRemovedFromFS)
	}()

	return ref
}

func (svc *BLOBCacheService) GetBytes(
	hash string, offset int64,
	buffer []byte,
) (int, bool, error) {
	fmt.Printf("GETTING BYTES FOR %s; OFFSET %d; SIZE %d\n", hash, offset, len(buffer))
	reader := svc.Get(hash, offset, int64(len(buffer)))
	if reader == nil {
		return 0, false, nil
	}

	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return 0, false, err
	}

	return n, true, nil
}

func (svc *BLOBCacheService) Get(
	hash string, offset, size int64,
) io.Reader {
	maybeRef := svc.Hold(hash)
	if maybeRef == nil {
		return nil
	}

	atReader := svc.getFile(hash)

	var reader io.Reader
	reader = io.NewSectionReader(atReader, offset, size)
	reader = lang.CreateSignalReader(reader)

	go func() {
		<-reader.(*lang.SignalReader).Done
		fmt.Println("DONE SIGNAL IS WORKING")
		maybeRef.Release()
	}()

	return reader
}

func (svc *BLOBCacheService) Hold(
	hash string,
) *BLOBCacheReference {
	entry, ok := svc.KnownBlobs.Get(hash)
	if !ok {
		return nil
	}

	entry.ReferencesLock.Lock()
	defer entry.ReferencesLock.Unlock()

	// if the entry is already being released, we can't hold it
	if len(entry.References) == 0 {
		return nil
	}

	ref := &BLOBCacheReference{}
	ref.entry = entry

	entry.References = append(entry.References, ref)

	return ref
}

func (svc *BLOBCacheService) storeFile(
	hash string,
	reader io.Reader,
) error {
	path := filepath.Join(
		svc.ConfigService.GetString("cacheDir"),
		hash,
	)

	file, err := svc.Filesystem.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()
	_, err = io.Copy(file, reader)
	return err
}

func (svc *BLOBCacheService) deleteFile(
	hash string,
) error {
	path := filepath.Join(
		svc.ConfigService.GetString("cacheDir"),
		hash,
	)

	return svc.Filesystem.Remove(path)
}

func (svc *BLOBCacheService) getFile(
	hash string,
) io.ReaderAt {
	path := filepath.Join(
		svc.ConfigService.GetString("cacheDir"),
		hash,
	)

	file, err := svc.Filesystem.Open(path)
	if err != nil {
		return nil
	}

	return file
}
