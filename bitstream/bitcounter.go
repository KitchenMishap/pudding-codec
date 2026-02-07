package bitstream

import "fmt"

type BitCounterPassThrough struct {
	writer   IBitWriter
	bitCount uint64
}

type BitCounter struct {
	bitCount uint64
}

func NewBitCounterPassThrough(bitWriter IBitWriter) *BitCounterPassThrough {
	result := BitCounterPassThrough{}
	result.writer = bitWriter
	result.bitCount = 0
	return &result
}

func NewBitCounter() *BitCounter {
	result := BitCounter{}
	result.bitCount = 0
	return &result
}

// Check that implements
var _ IBitCounter = (*BitCounterPassThrough)(nil)

func (bc *BitCounterPassThrough) WriteBits(bits uint64, bitCount int) error {
	bc.bitCount += uint64(bitCount)
	return bc.writer.WriteBits(bits, bitCount)
}
func (bc *BitCounterPassThrough) FlushBits() error {
	spareBits := bc.bitCount % 8
	if spareBits != 0 {
		bc.bitCount += 8 - spareBits
	}
	return bc.writer.FlushBits()
}
func (bc *BitCounterPassThrough) CountBits() uint64 {
	return bc.bitCount
}

func (bc *BitCounter) WriteBits(bits uint64, bitCount int) error {
	bc.bitCount += uint64(bitCount)
	return nil
}
func (bc *BitCounter) FlushBits() error {
	spareBits := bc.bitCount % 8
	if spareBits != 0 {
		bc.bitCount += 8 - spareBits
	}
	return nil
}
func (bc *BitCounter) CountBits() uint64 {
	return bc.bitCount
}

func FormatBits(bitCount uint64) string {
	result := ""
	if bitCount < 8 {
		result = fmt.Sprintf("%d bits", bitCount)
	} else if bitCount < 1024*8 {
		result = fmt.Sprintf("%.1f bytes", float64(bitCount)/8)
	} else if bitCount < 1024*1024*8 {
		result = fmt.Sprintf("%.3f kB", float64(bitCount)/1024/8)
	} else if bitCount < 1024*1024*1024*8 {
		result = fmt.Sprintf("%.3f MB", float64(bitCount)/1024/1024/8)
	} else {
		result = fmt.Sprintf("%.3f GB", float64(bitCount)/1024/1024/1024/8)
	}
	return result
}
