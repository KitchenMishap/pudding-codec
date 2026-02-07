package codecs

import (
	"errors"
	"github.com/KitchenMishap/pudding-codec/bitstream"
)

type CodecCompositeBase struct {
	names  []string
	codecs map[string]ICodecClass
}

// Check that implements
var _ INamedCodecCollection = (*CodecCompositeBase)(nil)

// CodecCompositeBase doesn't have enough code to implement
// ICodecClass on its own, you must embed it in your own class that
// adds the required functions

func (cc *CodecCompositeBase) Init() {
	cc.names = make([]string, 0, 10)
	cc.codecs = make(map[string]ICodecClass, 10)
}

func (cc *CodecCompositeBase) AddCodec(codec ICodecClass, name string) error {
	if cc.codecs == nil {
		cc.Init()
	}
	_, alreadyExists := cc.codecs[name]
	if alreadyExists {
		return errors.New("CodecComposite.AddCodec: name already exists")
	}
	cc.names = append(cc.names, name)
	cc.codecs[name] = codec
	return nil
}

func (cc *CodecCompositeBase) GetCodec(name string) (ICodecClass, error) {
	codec, exists := cc.codecs[name]
	if !exists {
		return nil, errors.New("CodecComposite.GetCodec: codec does not exist")
	}
	return codec, nil
}

func (cc *CodecCompositeBase) WriteParams(paramsCodec ICodecClass, stream bitstream.IBitWriter) error {
	for _, name := range cc.names {
		codec := cc.codecs[name]
		err := codec.WriteParams(paramsCodec, stream)
		if err != nil {
			return err
		}
	}
	return nil
}
func (cc *CodecCompositeBase) ReadParams(paramsCodec ICodecClass, stream bitstream.IBitReader) error {
	for _, name := range cc.names {
		codec := cc.codecs[name]
		err := codec.ReadParams(paramsCodec, stream)
		if err != nil {
			return err
		}
	}
	return nil
}
