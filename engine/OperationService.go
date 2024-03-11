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
	"fmt"
	"time"

	"github.com/HeyPuter/puter-fuse/putersdk"
	"github.com/HeyPuter/puter-fuse/services"
	"github.com/google/uuid"
)

type OperationResponse struct {
	Data map[string]interface{}
}

type OperationRequest struct {
	Operation putersdk.Operation
	Resolve   chan<- OperationResponse
	blob      []byte
}

type OperationRequestPromise struct {
	Await <-chan OperationResponse
}

type OperationService struct {
	SDK                   *putersdk.PuterSDK
	OperationRequestQueue chan *OperationRequest
	QueueReadyQueue       chan struct{}

	services services.IServiceContainer
}

type I_Batcher_EnqueueOperationRequest interface {
	EnqueueOperationRequest(
		operation putersdk.Operation,
		blob []byte,
	) OperationRequestPromise
}

func (svc_op *OperationService) EnqueueOperationRequest(
	operation putersdk.Operation,
	blob []byte,
) OperationRequestPromise {
	resolve := make(chan OperationResponse)
	await := make(chan OperationResponse)
	svc_op.OperationRequestQueue <- &OperationRequest{
		Operation: operation,
		blob:      blob,
		Resolve:   resolve,
	}
	go func() {
		// make a uuid for this timeout
		uuid := uuid.New().String()
		// log operation so the debugger can find it
		fmt.Printf("Operation: %s %s\n", uuid, operation)
		select {
		case res := <-resolve:
			fmt.Printf("RESOLVED uuid: %s\n", uuid)
			await <- res
		case <-time.After(20 * time.Second):
			// Print the uuid
			fmt.Printf("TIMEOUT uuid: %s\n", uuid)
			panic("oof time")
			await <- OperationResponse{
				Data: map[string]interface{}{
					"error": "internal timeout",
				},
			}
		}
	}()
	return OperationRequestPromise{
		Await: await,
	}
}

func (svc_op *OperationService) Init(services services.IServiceContainer) {
	svc_op.services = services

	svc_op.OperationRequestQueue = make(chan *OperationRequest, 100)

	batchQueue := make(chan *OperationRequest, 100)

	go func() {
		for val := range svc_op.OperationRequestQueue {
			fmt.Println("[GO] <== OP Req Queue -> Batch Queue")
			batchQueue <- val
			if len(batchQueue) == 100 {
				svc_op.QueueReadyQueue <- struct{}{}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-svc_op.QueueReadyQueue:
			case <-time.After(200 * time.Millisecond):
			}

			if len(batchQueue) == 0 {
				continue
			}

			fmt.Printf("len(batchQueue): %d\n", len(batchQueue))

			operations := []putersdk.Operation{}
			blobs := [][]byte{}
			resolves := []chan<- OperationResponse{}

			fmt.Println("Umm")

			MAX_BATCH := 100
			amountToGet := min(MAX_BATCH, len(batchQueue))

			for i := 0; i < amountToGet; i++ {
				var req *OperationRequest

				req = <-batchQueue

				if req == nil {
					break
				}

				operations = append(operations, req.Operation)
				resolves = append(resolves, req.Resolve)
				if req.blob != nil {
					blobs = append(blobs, req.blob)
				}
			}
			fmt.Println("Why?")

			// The commented-out line below was a mistake!
			// This was force of habit from dealing with queues
			// in javascript, but clearing the "queue" makes
			// absolutely no sense when working with channels.
			// I left it here as a warning; the debugger can't
			// help in situations like this, and it cost hours.

			// batchQueue = make(chan *OperationRequest, 100)

			// send the batch to the server
			fmt.Println("BATCH")
			batchResponse, err := svc_op.SDK.Batch(operations, blobs)

			if err != nil {
				// TODO: batch error handling
				fmt.Printf("error: %s\n", err)
				continue
			}

			// Print the batch response
			fmt.Printf("batchResponse: %s\n", batchResponse)

			for i := 0; i < len(resolves); i++ {
				if i >= len(batchResponse.Results) {
					panic(fmt.Errorf("batch response length mismatch"))
				}
				result := batchResponse.Results[i]
				resolves[i] <- OperationResponse{
					Data: result,
				}
			}
		}
	}()
}
