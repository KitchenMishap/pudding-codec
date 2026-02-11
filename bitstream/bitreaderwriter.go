package bitstream

import (
	"io"
)

type BitReader struct {
	byteReader io.Reader
	buffer     uint64 // Holds bits read from the stream but not yet consumed
	bitsCount  int    // Number of valid bits currently in the buffer
}

func NewBitReader(byteReader io.Reader) *BitReader {
	return &BitReader{byteReader: byteReader}
}

func (br *BitReader) ReadBits(bitCount int) (uint64, error) {
	if bitCount == 0 {
		return 0, nil
	}

	var result uint64
	bitsRead := 0

	for bitsRead < bitCount {
		// If the buffer is empty, refill it
		if br.bitsCount == 0 {
			var b [1]byte
			_, err := br.byteReader.Read(b[:])
			if err != nil {
				return 0, err // May return io.EOF
			}
			br.buffer = uint64(b[0])
			br.bitsCount = 8
		}

		// Calculate how many bits we can take from the current buffer
		bitsAvailable := bitCount - bitsRead
		if bitsAvailable > br.bitsCount {
			bitsAvailable = br.bitsCount
		}

		// Extract the bits
		mask := uint64(0)
		if bitsAvailable == 64 {
			mask = 0xFFFFFFFFFFFFFFFF
		} else {
			mask = (uint64(1) << bitsAvailable) - 1
		}

		extracted := br.buffer & mask
		result |= (extracted << bitsRead)

		// Advance the buffer
		br.buffer >>= bitsAvailable
		br.bitsCount -= bitsAvailable
		bitsRead += bitsAvailable
	}

	return result, nil
}

type BitWriter struct {
	byteWriter io.Writer
	buffer     uint64 // Holds bits not yet written to the stream
	bitsCount  int    // Number of bits currently waiting in the buffer
}

func NewBitWriter(byteWriter io.Writer) *BitWriter {
	return &BitWriter{byteWriter: byteWriter}
}

func (bw *BitWriter) WriteBits(bits uint64, bitCount int) error {
	// Special mask for 64-bit to avoid 1 << 64 overflow
	if bitCount < 64 {
		bits &= (uint64(1) << bitCount) - 1
	}

	// We use a for-loop to pour bits in.
	// This handles the case where bits + bw.bitsCount > 64
	for bitCount > 0 {
		spaceInBuffer := 64 - bw.bitsCount
		if spaceInBuffer == 0 {
			// Buffer is full (should have been flushed by the loop below,
			// but safe to check)
			if err := bw.flushFullBytes(); err != nil {
				return err
			}
			spaceInBuffer = 64
		}

		bitsToWrite := bitCount
		if bitsToWrite > spaceInBuffer {
			bitsToWrite = spaceInBuffer
		}

		// Push what we can into the buffer
		mask := uint64(0)
		if bitsToWrite == 64 {
			mask = 0xFFFFFFFFFFFFFFFF
		} else {
			mask = (uint64(1) << bitsToWrite) - 1
		}

		bw.buffer |= (bits & mask) << bw.bitsCount
		bw.bitsCount += bitsToWrite

		// Move on to the remaining bits
		bits >>= bitsToWrite
		bitCount -= bitsToWrite

		if err := bw.flushFullBytes(); err != nil {
			return err
		}
	}
	return nil
}

func (bw *BitWriter) flushFullBytes() error {
	for bw.bitsCount >= 8 {
		b := [1]byte{byte(bw.buffer & 0xFF)}
		if _, err := bw.byteWriter.Write(b[:]); err != nil {
			return err
		}
		bw.buffer >>= 8
		bw.bitsCount -= 8
	}
	return nil
}

// FlushBits writes any remaining bits in the buffer as a final partial byte
func (bw *BitWriter) FlushBits() error {
	if bw.bitsCount > 0 {
		// We write the remaining bits as a full byte (padded with zeros)
		b := [1]byte{byte(bw.buffer & 0xFF)}
		_, err := bw.byteWriter.Write(b[:])
		if err != nil {
			return err
		}
		bw.buffer = 0
		bw.bitsCount = 0
	}
	return nil
}
