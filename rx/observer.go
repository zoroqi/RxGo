package rx

// Observer represents a group of EventHandlers.
type Observer interface {
	OnNext(interface{})
	OnError(error)
	OnDone()
}

