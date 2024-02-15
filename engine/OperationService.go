package engine

import (
	"fmt"
	"time"

	"github.com/HeyPuter/puter-fuse-go/putersdk"
	"github.com/HeyPuter/puter-fuse-go/services"
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

func (svc_op *OperationService) EnqueueOperationRequest(
	operation putersdk.Operation,
	blob []byte,
) OperationRequestPromise {
	resolve := make(chan OperationResponse)
	svc_op.OperationRequestQueue <- &OperationRequest{
		Operation: operation,
		blob:      blob,
		Resolve:   resolve,
	}
	return OperationRequestPromise{
		Await: resolve,
	}
}

func (svc_op *OperationService) Init(services services.IServiceContainer) {
	svc_op.services = services

	svc_op.OperationRequestQueue = make(chan *OperationRequest, 100)

	batchQueue := make(chan *OperationRequest, 100)

	go func() {
		for {
			fmt.Println("[GO] <== OP Req Queue -> Batch Queue")
			select {
			case req := <-svc_op.OperationRequestQueue:
				batchQueue <- req

				if len(batchQueue) == 100 {
					svc_op.QueueReadyQueue <- struct{}{}
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-svc_op.QueueReadyQueue:
			case <-time.After(800 * time.Millisecond):
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
			for i := 0; i < MAX_BATCH; i++ {
				var req *OperationRequest

				select {
				case req = <-batchQueue:
				default:
				}

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

			// clear the batchQueue
			batchQueue = make(chan *OperationRequest, 100)

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
					break
				}
				result := batchResponse.Results[i]
				resolves[i] <- OperationResponse{
					Data: result,
				}
			}
		}
	}()
}
