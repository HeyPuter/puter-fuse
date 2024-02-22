package lang

import "io"

type SignalReader struct {
	Delegate io.Reader
	Done     chan struct{}
}

func CreateSignalReader(delegate io.Reader) *SignalReader {
	return &SignalReader{
		Delegate: delegate,
		Done:     make(chan struct{}),
	}
}

func (r *SignalReader) Read(p []byte) (n int, err error) {
	n, err = r.Delegate.Read(p)
	if err != nil {
		close(r.Done)
	}
	return
}
