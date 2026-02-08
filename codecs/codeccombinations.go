package codecs

func NewSelectorRawLutCombo(paramsCodec ICodecClass) (ICodecClass, error) {
	raw := NewCodecRaw64()
	lut, err := NewCodecLut64(8)
	if err != nil {
		return nil, err
	}
	forSelection := NewCodecUintNBits(8)
	selection := NewCodecSelector(paramsCodec)
	err = selection.AddCodec(forSelection, "SelectionChoice")
	if err != nil {
		return nil, err
	}
	err = selection.AddCodec(raw, "Raw")
	if err != nil {
		return nil, err
	}
	err = selection.AddCodec(lut, "Lut")
	if err != nil {
		return nil, err
	}
	return selection, nil
}
