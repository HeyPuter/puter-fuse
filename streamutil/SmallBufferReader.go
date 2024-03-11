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
