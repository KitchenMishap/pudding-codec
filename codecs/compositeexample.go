package codecs

import (
	"errors"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

// Composite example is just an example of a class embedding CodecCompositeBase.
// If the data is < 65536 it uses CodecRaw64, else CodecLut64.
// There is no real use for such a scheme!
type CompositeExample struct {
	CodecCompositeBase
}

// Check that implements
var _ ICodecClass = (*CompositeExample)(nil)

func NewCompositeExample() *CompositeExample {
	result := CompositeExample{}
	codecRaw := NewCodecRaw64()
	codecLut := NewCodecLut64()
	err := result.AddCodec(codecRaw, "Raw")
	if err != nil {
		panic("couldn't add raw codec")
	}
	err = result.AddCodec(codecLut, "Lut")
	if err != nil {
		panic("couldn't add lut codec")
	}
	return &result
}

func (ce *CompositeExample) Encode(data types.TData,
	writer bitstream.IBitWriter) (didntKnowHow bool, err error) {
	raw, err := ce.GetCodec("Raw")
	if err != nil {
		return false, err
	}
	lut, err := ce.GetCodec("Lut")
	if err != nil {
		return false, err
	}

	var selector types.TData
	if data < 65536 {
		selector = 0
	} else {
		selector = 1
	}

	didntKnowHow, err = raw.Encode(selector, writer)
	if err != nil {
		return false, err
	}
	if didntKnowHow {
		return false, errors.New("Raw encoder didn't know how to encode selector")
	}

	if selector == 0 {
		didntKnowHow, err = raw.Encode(data, writer)
		if err != nil {
			return false, err
		}
		if didntKnowHow {
			return false, errors.New("Raw encoder didn't know how to encode data")
		}
	} else if selector == 1 {
		didntKnowHow, err = lut.Encode(data, writer)
		if err != nil {
			return false, err
		}
		if didntKnowHow {
			return false, errors.New("Lut encoder didn't know how to encode data")
		}
	}
	return false, nil
}

func (ce *CompositeExample) Decode(reader bitstream.IBitReader) (types.TData, error) {
	raw, err := ce.GetCodec("Raw")
	if err != nil {
		return 0, err
	}
	lut, err := ce.GetCodec("Lut")
	if err != nil {
		return 0, err
	}

	var data types.TData
	selector, err := raw.Decode(reader)
	if err != nil {
		return 0, err
	}
	if selector == 0 {
		data, err = raw.Decode(reader)
		if err != nil {
			return 0, err
		}
	} else if selector == 1 {
		data, err = lut.Decode(reader)
		if err != nil {
			return 0, err
		}
	} else {
		return 0, errors.New("Raw decoder didn't know how to decode selector")
	}
	return data, nil
}
