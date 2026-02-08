package codecs

type SelectorExample struct {
	CodecSelectorBase
}

func NewSelectorExample(paramsCodec ICodecClass) *SelectorExample {
	result := SelectorExample{}
	result.paramsCodec = paramsCodec
	return &result
}
