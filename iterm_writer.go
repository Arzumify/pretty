package pritty

import (
	"fmt"
	"io"
)

type ItermStandardWriter struct {
	writer io.Writer
}

func (writer ItermStandardWriter) Write(buffer []byte) (int, error) {
	return writer.writer.Write(buffer)
}

func (writer ItermStandardWriter) Close() error {
	_, err := fmt.Fprint(writer.writer, ITERM_IMG_FTR)
	return err
}

type ItermStreamingWriter struct {
	// the writer writer
	writer io.Writer
	// the size of each chunk to send
	chunkSize int
}

// Finishes the current stream. No more data may be written aften this.
func (writer ItermStreamingWriter) Close() error {
	_, err := fmt.Fprintf(writer.writer, "%sFileEnd%s", ITERM_IMG_HDR, ITERM_IMG_FTR)
	return err
}

func (writer ItermStreamingWriter) Write(buffer []byte) (int, error) {
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
		_, err := fmt.Fprintf(writer.writer, "%sFilepart=", ITERM_IMG_HDR)
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
		_, err = fmt.Fprint(writer.writer, ITERM_IMG_FTR)
		if err != nil {
			return bufferWritten, err
		}
	}

	// return n written
	return bufferWritten, nil
}
