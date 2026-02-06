package codecs

type CodecSlice struct {
	Codecs []ICodecClass
}

// Check that implements
var _ ICodecCollection = (*CodecSlice)(nil)

func NewCodecSlice() ICodecCollection {
	result := CodecSlice{}
	result.Codecs = make([]ICodecClass, 0)
	return &result
}

func (cs *CodecSlice) Count() int {
	return len(cs.Codecs)
}

func (cs *CodecSlice) AddCodec(codec ICodecClass) {
	cs.Codecs = append(cs.Codecs, codec)
}

func (cs *CodecSlice) GetCodec(idx int) ICodecClass {
	if idx < 0 || idx >= len(cs.Codecs) {
		return nil
	}
	return cs.Codecs[idx]
}
