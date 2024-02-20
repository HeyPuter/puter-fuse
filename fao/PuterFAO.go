package fao

import (
	"path/filepath"

	"github.com/HeyPuter/puter-fuse-go/debug"
	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/localutil"
	"github.com/HeyPuter/puter-fuse-go/putersdk"
)

type P_PuterFAO struct {
	SDK     *putersdk.PuterSDK
	ReadFAO FAO
}

type IC_PuterFAO interface {
	engine.I_Batcher_EnqueueOperationRequest
}

type D_PuterFAO struct {
	EnqueueOperationRequest func(
		operation putersdk.Operation,
		blob []byte,
	) engine.OperationRequestPromise
}

type PuterFAO struct {
	BaseFAO
	P_PuterFAO
	D_PuterFAO
}

func CreatePuterFAO(
	params P_PuterFAO,
	deps D_PuterFAO,
) *PuterFAO {
	fao := &PuterFAO{
		BaseFAO{},
		params,
		deps,
	}

	fao.BaseFAO.FAO = fao

	return fao
}

func (fao *PuterFAO) Stat(path string) (NodeInfo, bool, error) {
	item, err := fao.SDK.Stat(path)
	if err != nil {
		return NodeInfo{}, false, nil
	}

	return NodeInfo{item}, true, nil
}

func (fao *PuterFAO) ReadDir(path string) ([]NodeInfo, error) {
	items, err := fao.SDK.Readdir(debug.NewLogger("PuterFAO"), path)
	if err != nil {
		return nil, err
	}

	nodeInfos := make([]NodeInfo, len(items))
	for i, item := range items {
		nodeInfos[i] = NodeInfo{item}
	}

	return nodeInfos, nil
}

func (fao *PuterFAO) Read(path string, dest []byte, off int64) (int, error) {
	data, err := fao.SDK.Read(path)
	if err != nil {
		return 0, err
	}

	copy(dest, data[off:])
	return len(data) - int(off), nil
}

func (fao *PuterFAO) Write(path string, src []byte, off int64) (int, error) {
	parent := filepath.Dir(path)
	name := filepath.Base(path)

	fileContents, err := fao.ReadFAO.ReadAll(path)
	if err != nil {
		return 0, err
	}

	if int64(len(fileContents)) < off+int64(len(src)) {
		newData := make([]byte, off+int64(len(src)))
		copy(newData, fileContents)
		fileContents = newData
	}
	copy(fileContents[off:], src)

	<-fao.EnqueueOperationRequest(
		putersdk.Operation{
			"op":        "write",
			"path":      parent,
			"name":      name,
			"overwrite": true,
		},
		fileContents,
	).Await

	return len(src), nil
}

func (fao *PuterFAO) Create(path string, name string) (NodeInfo, error) {
	empty := make([]byte, 0)
	resp := <-fao.EnqueueOperationRequest(
		putersdk.Operation{
			"op":        "write",
			"path":      path,
			"name":      name,
			"overwrite": true,
		},
		empty,
	).Await

	cloudItem := &putersdk.CloudItem{}
	err := localutil.ReJSON(resp.Data, cloudItem)
	if err != nil {
		return NodeInfo{}, err
	}

	node := NodeInfo{*cloudItem}

	return node, nil
}

func (fao *PuterFAO) Truncate(path string, size uint64) error {
	fileContents, err := fao.ReadFAO.ReadAll(path)
	if err != nil {
		return err
	}
	if uint64(len(fileContents)) == size {
		return nil
	}

	newData := make([]byte, size)
	copy(newData, fileContents)
	fileContents = newData

	parent := filepath.Dir(path)
	name := filepath.Base(path)

	<-fao.EnqueueOperationRequest(
		putersdk.Operation{
			"op":        "write",
			"path":      parent,
			"name":      name,
			"overwrite": true,
		},
		fileContents,
	).Await

	return nil
}

func (fao *PuterFAO) MkDir(path string, name string) (NodeInfo, error) {
	resp := <-fao.EnqueueOperationRequest(
		putersdk.Operation{
			"op":   "mkdir",
			"path": path,
			"name": name,
		},
		nil,
	).Await

	cloudItem := &putersdk.CloudItem{}
	err := localutil.ReJSON(resp.Data, cloudItem)
	if err != nil {
		return NodeInfo{}, err
	}

	return NodeInfo{*cloudItem}, nil
}

func (fao *PuterFAO) Symlink(parent string, name string, target string) (NodeInfo, error) {
	cloudItem, err := fao.SDK.Symlink(filepath.Join(parent, name), target)
	if err != nil {
		return NodeInfo{}, err
	}

	nodeInfo := NodeInfo{*cloudItem}
	return nodeInfo, nil
}

func (fao *PuterFAO) Unlink(path string) error {
	return fao.SDK.Delete(path)
}

func (fao *PuterFAO) Move(source string, parent string, name string) error {
	_, err := fao.SDK.Move(source, parent, name)
	return err
}
