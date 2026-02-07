package codecs

import (
	"errors"
	"github.com/KitchenMishap/pudding-codec/bitcode"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

// CodecLut64 stores a 64 bit to 64 bit look up table in the metadata (params)
// for the numbers in knows about. So 64 bit indices rather than 64 bit values
// are stored when encoded
type CodecLut64 struct {
	values  []types.TData                // For turning an index into a value
	indices map[types.TData]types.TIndex // For turning a value into an index
}

// Check that implements
var _ ICodecClass = (*CodecLut64)(nil)

func NewCodecLut64() *CodecLut64 {
	result := CodecLut64{}
	result.values = make([]types.TData, 0)
	result.indices = make(map[types.TData]types.TIndex)
	return &result
}

func (cl *CodecLut64) WriteParams(paramsCodec ICodecClass,
	bitWriter bitstream.IBitWriter) error {

	// <count>
	count := types.TCount(len(cl.values))
	didntKnowHow, err := paramsCodec.Encode(types.TData(count), bitWriter)
	if err != nil {
		return err
	}
	if didntKnowHow {
		return errors.New("CodecLut64.WriteParams count: params codec didn't know how")
	}

	for _, v := range cl.values {
		// <value>
		didntKnowHow, err = paramsCodec.Encode(v, bitWriter)
		if err != nil {
			return err
		}
		if didntKnowHow {
			return errors.New("CodecLut64.WriteParams value: params codec didn't know how")
		}
	}
	return nil
}

func (cl *CodecLut64) ReadParams(paramsCodec ICodecClass,
	bitReader bitstream.IBitReader) error {

	// <count>
	count, err := paramsCodec.Decode(bitReader)
	if err != nil {
		return err
	}

	// Reset
	cl.values = make([]types.TData, count)
	cl.indices = make(map[types.TData]types.TIndex, count)

	for i := range count {
		// <value>
		v, err := paramsCodec.Decode(bitReader)
		if err != nil {
			return err
		}

		index := types.TIndex(i)
		cl.values[index] = v
		cl.indices[v] = index
	}
	return nil
}

func (cl *CodecLut64) Encode(data types.TData,
	bitWriter bitstream.IBitWriter) (didntKnowHow bool, err error) {
	index, ok := cl.indices[data]
	if !ok {
		return true, nil
	}

	bitCode := bitcode.NewBitCode64(uint64(index), 64)
	return false, bitWriter.WriteBits(bitCode.Bits(), bitCode.Length())
}

func (cl *CodecLut64) Decode(bitReader bitstream.IBitReader) (types.TData, error) {
	bits, err := bitReader.ReadBits(64)
	if err != nil {
		return 0, err
	}

	index := types.TIndex(bits)
	if index >= types.TIndex(len(cl.values)) {
		return 0, errors.New("CodecLut64: unrecognized index")
	}
	val := cl.values[index]
	return val, nil
}

func (cl *CodecLut64) teachValue(data types.TData) {
	cl.values = append(cl.values, data)
	newIndex := types.TIndex(len(cl.values) - 1)
	cl.indices[data] = newIndex
}
