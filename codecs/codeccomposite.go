package codecs

import (
	"errors"
	"github.com/KitchenMishap/pudding-codec/bitstream"
)

type CodecCompositeBase struct {
	childCodecNames []string
	childCodecs     map[string]ICodecClass
}

// Check that implements
var _ INamedCodecCollection = (*CodecCompositeBase)(nil)

// CodecCompositeBase doesn't have enough code to implement
// ICodecClass on its own, you must embed it in your own class that
// adds the required functions

func (cc *CodecCompositeBase) Init() {
	cc.childCodecNames = make([]string, 0, 10)
	cc.childCodecs = make(map[string]ICodecClass, 10)
}

func (cc *CodecCompositeBase) AddCodec(codec ICodecClass, name string) error {
	if cc.childCodecs == nil {
		cc.Init()
	}
	_, alreadyExists := cc.childCodecs[name]
	if alreadyExists {
		return errors.New("CodecComposite.AddCodec: name already exists")
	}
	cc.childCodecNames = append(cc.childCodecNames, name)
	cc.childCodecs[name] = codec
	return nil
}

func (cc *CodecCompositeBase) GetCodec(name string) (ICodecClass, error) {
	codec, exists := cc.childCodecs[name]
	if !exists {
		return nil, errors.New("CodecComposite.GetCodec: codec does not exist")
	}
	return codec, nil
}

func (cc *CodecCompositeBase) WriteParams(paramsCodec ICodecClass, stream bitstream.IBitWriter) error {
	for _, name := range cc.childCodecNames {
		codec := cc.childCodecs[name]
		err := codec.WriteParams(paramsCodec, stream)
		if err != nil {
			return err
		}
	}
	return nil
}
func (cc *CodecCompositeBase) ReadParams(paramsCodec ICodecClass, stream bitstream.IBitReader) error {
	for _, name := range cc.childCodecNames {
		codec := cc.childCodecs[name]
		err := codec.ReadParams(paramsCodec, stream)
		if err != nil {
			return err
		}
	}
	return nil
}
