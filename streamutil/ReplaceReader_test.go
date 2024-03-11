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
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestReplaceReader(t *testing.T) {
	t.Run("ReplaceReader replaces bytes at offset", func(t *testing.T) {
		// Create a reader with the string "hello, world"
		reader := NewReaderReadCloser(
			strings.NewReader("hello, world!"),
			nil,
		)

		// Create a ReplaceReader that replaces the bytes at offset 7 with "there"
		replaceReader := NewReplaceReader(reader, []byte("there"), uint64(len("hello, ")))
		replaceReader.(*ReplaceReader).verbose = true

		// Read from the ReplaceReader
		buf := make([]byte, 100)
		n, err := replaceReader.Read(buf)
		if err != nil && err != io.EOF {
			t.Errorf("unexpected error: %v", err)
		}

		// Check that the bytes were replaced
		if string(buf[:len([]byte("hello, there!"))]) != "hello, there!" {
			asHex := hex.EncodeToString(buf[:n])
			t.Errorf("expected 'hello, there!', got '%s' (%s)", buf[:n], asHex)
		}
		if buf[len([]byte("hello, there!"))] != 0 {
			t.Errorf("expected EOF, got %v", buf[len([]byte("hello, there!"))])
		}
	})

	// Testing with all these different buffer sizes is very important!
	// There are a surprising number of edge-cases.
	bufferSizesToTest := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	for _, bufSize := range bufferSizesToTest {
		t.Run(fmt.Sprintf("ReplaceReader %d char buffer test", bufSize), func(t *testing.T) {
			// Create a reader with the string "hello, world"
			reader := NewReaderReadCloser(
				strings.NewReader("hello, world!"),
				nil,
			)

			// Create a ReplaceReader that replaces the bytes at offset 7 with "there"
			replaceReader := NewReplaceReader(reader, []byte("there"), uint64(len("hello, ")))
			replaceReader.(*ReplaceReader).verbose = true

			smallBufferReader := NewSmallBufferReader(replaceReader, bufSize)

			// Read from the ReplaceReader
			buf := make([]byte, 100)
			n, err := smallBufferReader.Read(buf)
			if err != nil && err != io.EOF {
				t.Errorf("unexpected error: %v", err)
			}

			// Check that the bytes were replaced
			if string(buf[:len([]byte("hello, there!"))]) != "hello, there!" {
				asHex := hex.EncodeToString(buf[:n])
				t.Errorf("expected 'hello, there!', got '%s' (%s)", buf[:n], asHex)
			}
			if buf[len([]byte("hello, there!"))] != 0 {
				t.Errorf("expected EOF, got %v", buf[len([]byte("hello, there!"))])
			}
		})
	}
}
