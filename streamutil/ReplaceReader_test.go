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
