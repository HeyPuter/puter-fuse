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
