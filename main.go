package main

import (
	"fmt"
	"github.com/KitchenMishap/pudding-codec/bitstream"
	"github.com/KitchenMishap/pudding-codec/codecs"
	"github.com/KitchenMishap/pudding-codec/types"
)

func main() {
	// Prepare
	codeclist := codecs.NewCodecSlice()
	codeclist.AddCodec(codecs.NewCodecRaw64())
	rawIndex := 0
	stream := bitstream.StringSlice{}

	// Write
	num := types.TData(123456789)
	fmt.Printf("Number In: %d\n", num)
	bitCode := codeclist.GetCodec(rawIndex).Encode(num)
	err := stream.PushBack(bitCode)
	if err != nil {
		panic(err)
	}

	// Read
	newBitCode, err := stream.Pop()
	if err != nil {
		panic(err)
	}
	newNum := codeclist.GetCodec(rawIndex).Decode(newBitCode)
	fmt.Printf("Number Out: %d\n", newNum)
}
