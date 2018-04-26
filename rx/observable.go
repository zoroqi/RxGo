package rx

import "context"

type Observable interface {
	Iterator
	Functor
	Subscribe(interface{}) (context.Context, context.CancelFunc)
}
