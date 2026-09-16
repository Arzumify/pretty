package rasterm

import (
	"fmt"
	"io"
)

type KittyImageWriter struct {
	// the inner writer
	inner io.Writer
	// the size of each chunk to send
	chunkSize int
}

// Finishes the current stream. No more data may be written aften this.
func (self KittyImageWriter) Close() error {
	_, err := fmt.Fprint(self.inner, KITTY_IMG_HDR, "m=0;", KITTY_IMG_FTR)
	return err
}

func (self KittyImageWriter) Write(buffer []byte) (int, error) {
	bufferRemaining := len(buffer)
	bufferWritten := 0
	chunkWritten := 0

	for bufferRemaining > 0 {
		// how many bytes to write on this chunk
		var toWrite int

		// if the
		if (chunkWritten + bufferRemaining) >= self.chunkSize {
			toWrite = self.chunkSize - chunkWritten
			chunkWritten = 0
		} else {
			toWrite = bufferRemaining
			chunkWritten += bufferRemaining
		}

		// write prefix
		_, err := fmt.Fprint(self.inner, KITTY_IMG_HDR, "m=1;")
		if err != nil {
			return bufferWritten, err
		}

		// write data
		n, err := self.inner.Write(buffer[:toWrite])
		if err != nil {
			return bufferWritten, err
		}

		// update values
		buffer = buffer[toWrite:]
		bufferRemaining -= toWrite
		bufferWritten += n

		// write suffix
		_, err = fmt.Fprint(self.inner, KITTY_IMG_FTR)
		if err != nil {
			return bufferWritten, err
		}
	}

	// return n written
	return bufferWritten, nil
}
