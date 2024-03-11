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
