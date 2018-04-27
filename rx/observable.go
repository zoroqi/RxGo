package rx

import (
	"context"
	"time"
)

type Observable interface {
	Iterator
	Functor
	Subscribe(interface{}, ...time.Duration) context.Context
}
