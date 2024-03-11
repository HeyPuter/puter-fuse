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
package faoimpls

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/HeyPuter/puter-fuse/debug"
	"github.com/HeyPuter/puter-fuse/engine"
	"github.com/HeyPuter/puter-fuse/fao"
	"github.com/HeyPuter/puter-fuse/localutil"
	"github.com/HeyPuter/puter-fuse/putersdk"
	"github.com/google/uuid"
)

type P_PuterFAO struct {
	SDK     *putersdk.PuterSDK
	ReadFAO fao.FAO
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
	fao.BaseFAO
	P_PuterFAO
	D_PuterFAO
}

func CreatePuterFAO(
	params P_PuterFAO,
	deps D_PuterFAO,
) *PuterFAO {
	fao := &PuterFAO{
		fao.BaseFAO{},
		params,
		deps,
	}

	fao.BaseFAO.FAO = fao

	return fao
}

func (f *PuterFAO) Stat(path string) (fao.NodeInfo, bool, error) {
	item, err := f.SDK.Stat(path)
	if err != nil {
		return fao.NodeInfo{}, false, nil
	}

	return fao.NodeInfo{CloudItem: item}, true, nil
}

func (f *PuterFAO) ReadDir(path string) ([]fao.NodeInfo, error) {
	items, err := f.SDK.Readdir(debug.NewLogger("PuterFAO"), path)
	if err != nil {
		return nil, err
	}

	nodeInfos := make([]fao.NodeInfo, len(items))
	for i, item := range items {
		nodeInfos[i] = fao.NodeInfo{CloudItem: item}
	}

	return nodeInfos, nil
}

func (f *PuterFAO) Read(path string, dest []byte, off int64) (int, error) {
	data, err := f.SDK.Read(path)
	if err != nil {
		return 0, err
	}

	copy(dest, data[off:])
	return len(data) - int(off), nil
}

func (f *PuterFAO) Write(path string, src []byte, off int64) (int, error) {
	parent := filepath.Dir(path)
	name := filepath.Base(path)

	fileContentsReader, err := f.ReadFAO.ReadAll(path)
	if err != nil {
		return 0, err
	}

	fileContents, err := io.ReadAll(fileContentsReader)
	if err != nil {
		return 0, nil
	}
	fileContentsReader.Close()

	if int64(len(fileContents)) < off+int64(len(src)) {
		newData := make([]byte, off+int64(len(src)))
		copy(newData, fileContents)
		fileContents = newData
	}
	copy(fileContents[off:], src)

	<-f.EnqueueOperationRequest(
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

func (f *PuterFAO) Create(path string, name string) (fao.NodeInfo, error) {
	empty := make([]byte, 0)
	resp := <-f.EnqueueOperationRequest(
		putersdk.Operation{
			"op":        "write",
			"path":      path,
			"name":      name,
			"overwrite": true,
		},
		empty,
	).Await

	// random trace uuid
	uid := uuid.New().String()

	{
		jsonBytes, _ := json.Marshal(resp)
		fmt.Println("resp is: ", uid, string(jsonBytes))
	}

	cloudItem := &putersdk.CloudItem{}
	err := localutil.ReJSON(resp.Data, cloudItem)
	if err != nil {
		return fao.NodeInfo{}, err
	}

	node := fao.NodeInfo{CloudItem: *cloudItem}

	// assert that node is not a directory
	if node.IsDir {
		fmt.Println("Path is: ", uid, path)
		fmt.Println("Name is: ", uid, name)
		fmt.Printf("Node is: %s %+v\n", uid, node)
		panic("created node is a directory")
	}

	if node.Path == "" {
		fmt.Println("Path is: ", uid, path)
		fmt.Println("Name is: ", uid, name)
		fmt.Printf("Node is: %s %+v\n", uid, node)
		panic("created node is missing path")
	}

	return node, nil
}

func (f *PuterFAO) Truncate(path string, size uint64) error {
	fileContentsReader, err := f.ReadFAO.ReadAll(path)
	if err != nil {
		return err
	}

	fileContents, err := io.ReadAll(fileContentsReader)
	if err != nil {
		return err
	}
	fileContentsReader.Close()

	if uint64(len(fileContents)) == size {
		return nil
	}

	newData := make([]byte, size)
	copy(newData, fileContents)
	fileContents = newData

	parent := filepath.Dir(path)
	name := filepath.Base(path)

	<-f.EnqueueOperationRequest(
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

func (f *PuterFAO) MkDir(parent string, path string) (fao.NodeInfo, error) {
	resp := <-f.EnqueueOperationRequest(
		putersdk.Operation{
			"op":     "mkdir",
			"parent": parent,
			"path":   path,
		},
		nil,
	).Await

	cloudItem := &putersdk.CloudItem{}
	err := localutil.ReJSON(resp.Data, cloudItem)
	if err != nil {
		return fao.NodeInfo{}, err
	}

	return fao.NodeInfo{CloudItem: *cloudItem}, nil
}

func (f *PuterFAO) Symlink(parent string, name string, target string) (fao.NodeInfo, error) {
	cloudItem, err := f.SDK.Symlink(filepath.Join(parent, name), target)
	if err != nil {
		return fao.NodeInfo{}, err
	}

	nodeInfo := fao.NodeInfo{CloudItem: *cloudItem}
	return nodeInfo, nil
}

func (f *PuterFAO) Unlink(path string) error {
	return f.SDK.Delete(path)
}

func (f *PuterFAO) Move(source string, parent string, name string) error {
	fmt.Println("performing a move operation")
	_, err := f.SDK.Move(source, parent, name)
	return err
}

func (f *PuterFAO) ReadAll(path string) (io.ReadCloser, error) {
	return f.SDK.ReadStream(path)
}
