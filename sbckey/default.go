package sbckey

import (
	"fmt"

	sbc "github.com/Autodoc-Technology/streaming-based-config"
)

// DefaultKeyBuilder returns a KeyBuilder function that formats the type of the input value as a string.
func DefaultKeyBuilder[T any]() sbc.KeyBuilder[T] {
	return sbc.KeyBuilderFunc[T](func(t T) string {
		return fmt.Sprintf("%T", t)
	})
}
