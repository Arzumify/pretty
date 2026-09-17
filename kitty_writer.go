package pritty

import (
	"fmt"
	"io"
)

type KittyLocalWriter struct {
	writer io.Writer
}

func (writer KittyLocalWriter) Write(buffer []byte) (int, error) {
	return writer.writer.Write(buffer)
}

func (writer KittyLocalWriter) Close() error {
	_, err := fmt.Fprint(writer.writer, KITTY_IMG_FTR)
	return err
}

type KittyImageWriter struct {
	// the writer writer
	writer io.Writer
	// the size of each chunk to send
	chunkSize int
}

// Finishes the current stream. No more data may be written aften this.
func (writer KittyImageWriter) Close() error {
	_, err := fmt.Fprint(writer.writer, KITTY_IMG_HDR, "m=0;", KITTY_IMG_FTR)
	return err
}

func (writer KittyImageWriter) Write(buffer []byte) (int, error) {
	bufferRemaining := len(buffer)
	bufferWritten := 0
	chunkWritten := 0

	for bufferRemaining > 0 {
		// how many bytes to write on this chunk
		var toWrite int

		// if the
		if (chunkWritten + bufferRemaining) >= writer.chunkSize {
			toWrite = writer.chunkSize - chunkWritten
			chunkWritten = 0
		} else {
			toWrite = bufferRemaining
			chunkWritten += bufferRemaining
		}

		// write prefix
		_, err := fmt.Fprint(writer.writer, KITTY_IMG_HDR, "m=1;")
		if err != nil {
			return bufferWritten, err
		}

		// write data
		n, err := writer.writer.Write(buffer[:toWrite])
		if err != nil {
			return bufferWritten, err
		}

		// update values
		buffer = buffer[toWrite:]
		bufferRemaining -= toWrite
		bufferWritten += n

		// write suffix
		_, err = fmt.Fprint(writer.writer, KITTY_IMG_FTR)
		if err != nil {
			return bufferWritten, err
		}
	}

	// return n written
	return bufferWritten, nil
}
