package compression

import (
	"github.com/KitchenMishap/pudding-codec/types"
	"io"
)

type IEngine interface {
	Encode(data []types.TData, writer io.Writer) error
	Decode(reader io.Reader) ([]types.TData, error)
}
