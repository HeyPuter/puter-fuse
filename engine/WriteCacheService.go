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
package engine

import (
	"io"

	"github.com/HeyPuter/puter-fuse-go/lang"
	"github.com/HeyPuter/puter-fuse-go/services"
	"github.com/HeyPuter/puter-fuse-go/streamutil"
)

type Mutation interface {
	Apply(inStream io.ReadCloser) (outStream io.ReadCloser, err error)
	ApplyToBuffer(buffer []byte, offset int64)
}

type MutationChain struct {
	Releasables []Releasable
	Mutations   []interface{}
}

func (chain *MutationChain) ApplyToBuffer(buffer []byte, offset int64) {
	for _, mut := range chain.Mutations {
		switch mut := mut.(type) {
		case Mutation:
			mut.ApplyToBuffer(buffer, offset)
		}
	}
}

type WriteMutation struct {
	Data   []byte
	Offset int64
}

type MutationReference struct{}

func (ref *MutationReference) Release() {
}

// Returns a reader which will emit the contents of inStream, replacing the
// bytes at 'Offset' with the buffer 'Data'.
func (mut *WriteMutation) Apply(inStream io.ReadCloser) (outStream io.ReadCloser, err error) {
	// I thought Golang was going to have this in the `io` package,
	// but instead I had to spend hours developing it...
	return streamutil.NewReplaceReader(
		inStream, mut.Data, uint64(mut.Offset),
	), nil
}

func (mut *WriteMutation) ApplyToBuffer(buffer []byte, offset int64) {
	rightEdgeOfMutationInRead := (mut.Offset + int64(len(mut.Data))) > offset
	leftEdgeOfMutationInRead := mut.Offset < (offset + int64(len(buffer)))

	if !(rightEdgeOfMutationInRead && leftEdgeOfMutationInRead) {
		return
	}

	var iBufferStart, iBufferEnd int
	var iDataStart, iDataEnd int

	// You know what, this code below - as simple as it is - is actually
	// annoying complicated to think about.

	// Instead of trying to document it, I've opted to add diagrams to
	// the doc directory instead, and I might write a devlog entry later.
	// (doc/read_write_mut_D.drawio.png)

	if mut.Offset < offset {
		iBufferStart = 0
		iDataStart = int(offset - mut.Offset)
	} else {
		iBufferStart = int(mut.Offset - offset)
		iDataStart = 0
	}

	if (mut.Offset + int64(len(mut.Data))) > (offset + int64(len(buffer))) {
		iBufferEnd = len(buffer)
		iDataEnd = int(offset + int64(len(buffer)) - mut.Offset)
	} else {
		iBufferEnd = int(mut.Offset + int64(len(mut.Data)) - offset)
		iDataEnd = len(mut.Data)
	}

	copy(buffer[iBufferStart:iBufferEnd], mut.Data[iDataStart:iDataEnd])
}

type TruncateMutation struct {
	Size uint64
}

type RemoveMutation struct{}

type WriteCacheService struct {
	CachedOperations lang.IMap[string, *MutationChain]
}

func CreateWriteCacheService() *WriteCacheService {
	return &WriteCacheService{
		CachedOperations: lang.CreateSyncMap[string, *MutationChain](nil),
	}
}

func (svc *WriteCacheService) Init(services services.IServiceContainer) {
}

func (svc *WriteCacheService) ApplyToBuffer(localUID string, buffer []byte, offset int64) {
	chain, _, _ := svc.CachedOperations.
		GetWithFactory(localUID, func() (*MutationChain, bool, error) {
			chain := &MutationChain{}
			return chain, false, nil
		})

	chain.ApplyToBuffer(buffer, offset)
}

func (svc *WriteCacheService) ApplyMutation(localUID string, mut Mutation) *MutationReference {
	chain, _, _ := svc.CachedOperations.
		GetWithFactory(localUID, func() (*MutationChain, bool, error) {
			chain := &MutationChain{}
			return chain, false, nil
		})

	chain.Mutations = append(chain.Mutations, mut)
	return &MutationReference{}
}
