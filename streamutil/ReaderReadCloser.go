package streamutil

import "io"

// ReaderReadCloser is a wrapper around an io.Reader that also implements the io.ReadCloser interface.
type ReaderReadCloser struct {
	Delegate io.Reader
	Closer   io.Closer
}

func NewReaderReadCloser(delegate io.Reader, closer io.Closer) *ReaderReadCloser {
	return &ReaderReadCloser{
		Delegate: delegate,
		Closer:   closer,
	}
}

// Read implements the io.Reader interface for ReaderReadCloser.
func (r *ReaderReadCloser) Read(p []byte) (n int, err error) {
	return r.Delegate.Read(p)
}

// Close implements the io.Closer interface for ReaderReadCloser.
func (r *ReaderReadCloser) Close() error {
	if r.Closer != nil {
		return r.Closer.Close()
	}
	if closer, ok := r.Delegate.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
