package bitstream

import (
	"errors"
	"github.com/KitchenMishap/pudding-codec/bitcode"
	"strings"
)

// A FIFO queue
type StringSlice struct {
	entries []string // Index 0 is the first thing put in
}

// Check that implements
var _ IBitStream = (*StringSlice)(nil)

func NewStringSlice() *StringSlice {
	result := StringSlice{}
	return &result
}

func (bcs *StringSlice) PushBack(bitCode bitcode.IBitCode) error {
	sb := strings.Builder{}
	err := bitCode.WriteBytes(&sb)
	if err != nil {
		return err
	}
	bcs.entries = append(bcs.entries, sb.String())
	return nil
}

func (bcs *StringSlice) PopFront(bitCount int) (bitcode.IBitCode, error) {
	length := len(bcs.entries)
	if length == 0 {
		return nil, errors.New("no entries in StringSlice")
	}
	str := bcs.entries[0]
	bcs.entries = bcs.entries[1:]
	sr := strings.NewReader(str)
	bc := bitcode.NewBitCode64(0, 0)
	err := bc.ReadBytes(sr)
	if bc.Length() != bitCount {
		panic("found the wrong number of bits")
	}
	if err != nil {
		return nil, err
	}
	return bc, nil
}
