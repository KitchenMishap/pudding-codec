package scribenode

import (
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/types"
)

// LiteralScribe represents a single, fixed symbol.
type LiteralScribe struct {
	symbol types.TSymbol
}

func NewLiteralScribe(s types.TSymbol) *LiteralScribe {
	return &LiteralScribe{symbol: s}
}

// BidBits returns 0 bits if the symbol matches, otherwise it refuses.
func (ls *LiteralScribe) BidBits(sequence []types.TSymbol) (types.TBitCount, bool, error) {
	if len(sequence) == 1 && sequence[0] == ls.symbol {
		return 0, false, nil
	}
	return 0, true, nil // "I don't know how to encode anything else"
}

func (ls *LiteralScribe) Encode(sequence []types.TSymbol, _ bitstream.IBitWriter) (bool, error) {
	if len(sequence) == 1 && sequence[0] == ls.symbol {
		return false, nil // Success! 0 bits written.
	}
	return true, nil // Refused
}

func (ls *LiteralScribe) Decode(_ bitstream.IBitReader) ([]types.TSymbol, error) {
	return []types.TSymbol{ls.symbol}, nil
}

func (ls *LiteralScribe) GetSymbol() types.TSymbol { return ls.symbol }
