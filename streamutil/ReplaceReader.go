package streamutil

import (
	"fmt"
	"io"
)

type ReplaceReader struct {
	source          io.ReadCloser // The original reader
	replacement     []byte        // The replacement to be inserted
	forLater        []byte        // Part of the source that needs to be inserted later
	forLaterPos     uint64        // The position in the forLater slice
	forLaterLen     uint64        // The length of the forLater slice
	offset          uint64        // The offset at which the replacement should be inserted
	readPos         uint64        // Keeps track of the current read position
	overreadAmount  uint64        // amount over-read from source before replacement
	skippingStarted bool
	skippingDone    bool
	skipped         chan error
	verbose         bool
}

func NewReplaceReader(original io.ReadCloser, source []byte, offset uint64) io.ReadCloser {
	return &ReplaceReader{
		source:      original,
		replacement: source,
		offset:      offset,
		forLater:    make([]byte, len(source)),
		skipped:     make(chan error),
	}
}

// Read implements the io.Reader interface for customReader.
func (replaceReader *ReplaceReader) Read(p []byte) (int, error) {
	amountWrittenSoFar := 0
	iters := 0
	for amountWrittenSoFar < len(p) || iters > 10 {
		iters++
		replaceEnd := replaceReader.offset + uint64(len(replaceReader.replacement))

		replaceReader.debug("iter")
		if replaceReader.verbose {
			fmt.Printf("amountWrittenSoFar: %d\n", amountWrittenSoFar)
			fmt.Printf("offset: %d\n", replaceReader.offset)
			fmt.Printf("replaceEnd: %d\n", replaceEnd)
		}
		if replaceReader.readPos >= replaceEnd && replaceReader.forLaterPos < replaceReader.forLaterLen {
			replaceReader.debug("writing from forLater")
			// Write the forLater slice to p
			copyLen := copy(p[amountWrittenSoFar:], replaceReader.forLater[replaceReader.forLaterPos:])
			amountWrittenSoFar += copyLen
			replaceReader.forLaterPos += uint64(copyLen)
			replaceReader.readPos += uint64(copyLen)
			continue
		}

		if replaceReader.skippingStarted && !replaceReader.skippingDone {
			replaceReader.debug("waiting for skipping to finish")
			err := <-replaceReader.skipped
			replaceReader.skippingDone = true
			if err != nil {
				return amountWrittenSoFar, err
			}
			replaceReader.debug("continuing")
		}

		if replaceReader.readPos < replaceReader.offset || replaceReader.readPos >= replaceEnd {
			replaceReader.debug("reading from source")
			// Read from the original reader until we reach the offset.
			n, err := replaceReader.source.Read(p[amountWrittenSoFar:])
			if err != nil && err != io.EOF {
				return amountWrittenSoFar, err
			}
			if n == 0 {
				replaceReader.debugPrint("EOF")
				return amountWrittenSoFar, io.EOF
			}
			replaceReader.debugPrint("read |" + string(p[:n]) + "|")
			if replaceReader.readPos+uint64(n) > replaceReader.offset+uint64(len(replaceReader.replacement)) && !replaceReader.skippingStarted {
				// We've read past the offset, so put the extra bytes in the forLater slice
				// extra := replaceReader.readPos + uint64(n) - replaceReader.offset
				extra := replaceReader.offset + uint64(len(replaceReader.replacement)) - replaceReader.readPos
				replaceReader.debugPrint("extra: |" + string(p[int(extra):n]) + "|")
				forLaterCopyLen := copy(replaceReader.forLater, p[int(extra):n])
				replaceReader.forLaterLen += uint64(forLaterCopyLen)
			}
			if replaceReader.readPos < replaceReader.offset && n > int(replaceReader.offset-replaceReader.readPos) {
				oldN := n
				n = int(replaceReader.offset - replaceReader.readPos)
				replaceReader.overreadAmount += uint64(oldN) - uint64(n)
			}
			replaceReader.readPos += uint64(n)
			amountWrittenSoFar += n
			if replaceReader.verbose {
				fmt.Println("amountWrittenSoFar: ", amountWrittenSoFar)
				fmt.Printf("p: %s\n", p)
			}
			continue
		}

		if replaceReader.readPos >= replaceReader.offset && replaceReader.readPos < replaceEnd {
			replaceReader.debug("writing replacement")
			if !replaceReader.skippingStarted {
				replaceReader.startSkipping()
				replaceReader.skippingStarted = true
			}
			// Insert the replacement into p
			copyLen := copy(p[amountWrittenSoFar:], replaceReader.replacement[replaceReader.readPos-replaceReader.offset:])
			if replaceReader.verbose {
				fmt.Printf("copyLen: %d\n", copyLen)
				fmt.Printf("replaceReader.readPos: %d\n", replaceReader.readPos)
				fmt.Printf("replaceReader.offset: %d\n", replaceReader.offset)
				fmt.Printf("replaceReader.readPos-replaceReader.offset: %d\n", replaceReader.readPos-replaceReader.offset)
				fmt.Printf("replaceReader.replacement: %s\n", replaceReader.replacement)
				fmt.Printf("p: %s\n", p)
				fmt.Printf("amountWrittenSoFar: %d\n", amountWrittenSoFar)
			}
			amountWrittenSoFar += copyLen
			replaceReader.readPos += uint64(copyLen)
			continue
		}
	}

	replaceReader.debugPrint("result: |" + string(p[:amountWrittenSoFar]) + "|")

	return amountWrittenSoFar, nil
}

func (replaceReader *ReplaceReader) startSkipping() {
	// amountToSkip := len(replaceReader.replacement) - int(replaceReader.forLaterLen)
	amountToSkip := len(replaceReader.replacement) - int(replaceReader.overreadAmount)
	go func() {
		if replaceReader.verbose {
			fmt.Printf("--> skipping %d bytes\n", amountToSkip)
		}
		_, err := io.CopyN(io.Discard, replaceReader.source, int64(amountToSkip))
		replaceReader.skipped <- err
	}()
}

func (replaceReader *ReplaceReader) printState() {
	if !replaceReader.verbose {
		return
	}

	fmt.Println("readPos:", replaceReader.readPos)
	fmt.Println("offset:", replaceReader.offset)
	fmt.Println("forLaterPos:", replaceReader.forLaterPos)
	fmt.Println("forLaterLen:", replaceReader.forLaterLen)
	fmt.Println("skippingStarted:", replaceReader.skippingStarted)
	fmt.Println("skippingDone:", replaceReader.skippingDone)
	fmt.Println("overreadAmount:", replaceReader.overreadAmount)
}

func (replaceReader *ReplaceReader) debug(s string) {
	if replaceReader.verbose {
		fmt.Println("===" + s + "===")
		replaceReader.printState()
	}
}

func (replaceReader *ReplaceReader) debugPrint(s string) {
	if replaceReader.verbose {
		fmt.Println(s)
	}
}

// Close implements the io.Closer interface, ensuring the original reader is closed.
func (replaceReader *ReplaceReader) Close() error {
	return replaceReader.source.Close()
}
