package streamutil

import (
	"io"
)

// SmallBufferReader is an io.Reader that reads from an underlying reader
// in smaller, fixed-size chunks. (written by
type SmallBufferReader struct {
	r       io.Reader // underlying reader
	bufSize int       // buffer size for each read
}

// NewSmallBufferReader creates a new SmallBufferReader with the given buffer size.
func NewSmallBufferReader(r io.Reader, bufSize int) *SmallBufferReader {
	return &SmallBufferReader{
		r:       r,
		bufSize: bufSize,
	}
}

// Read reads up to len(p) bytes into p from the underlying reader, but limited
// to bufSize bytes per read operation.
func (sbr *SmallBufferReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Limit the read size to the smaller of bufSize or len(p)
	readSize := len(p)
	if sbr.bufSize < readSize {
		readSize = sbr.bufSize
	}

	amountRead := 0
	for amountRead < len(p) {
		tmpBuf := make([]byte, readSize)
		n, err = sbr.r.Read(tmpBuf)
		if n > 0 {
			copy(p[amountRead:], tmpBuf[:n])
			amountRead += n
		}
		if err != nil {
			return amountRead, err
		}
	}

	return amountRead, nil
}
