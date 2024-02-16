package fao

import (
	"path/filepath"

	"github.com/HeyPuter/puter-fuse-go/debug"
	"github.com/HeyPuter/puter-fuse-go/engine"
	"github.com/HeyPuter/puter-fuse-go/putersdk"
)

type P_PuterFAO struct {
	SDK     putersdk.PuterSDK
	Batcher *engine.OperationService
	ReadFAO FAO
}

type PuterFAO struct {
	P_PuterFAO
}

func CreatePuterFAO(params P_PuterFAO) *PuterFAO {
	return &PuterFAO{params}
}

func (fao *PuterFAO) Stat(path string) (NodeInfo, error) {
	item, err := fao.SDK.Stat(path)
	if err != nil {
		return NodeInfo{}, err
	}

	return NodeInfo{item}, nil
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

	<-fao.Batcher.EnqueueOperationRequest(
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

func (fao *PuterFAO) Truncate(path string, size int64) error {
	fileContents, err := fao.ReadFAO.ReadAll(path)
	if err != nil {
		return err
	}
	if int64(len(fileContents)) == size {
		return nil
	}

	newData := make([]byte, size)
	copy(newData, fileContents)
	fileContents = newData

	parent := filepath.Dir(path)
	name := filepath.Base(path)

	<-fao.Batcher.EnqueueOperationRequest(
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
