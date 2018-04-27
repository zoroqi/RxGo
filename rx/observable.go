package rx

import (
	"context"
)

// Observable is an Iterator, Functor, and can Subscribe a handler.
type Observable interface {
	Iterator
	Functor
	Subscribe(interface{}) context.Context
}
