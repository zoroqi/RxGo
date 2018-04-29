// Package connectable provides a Connectable and its methods.
package connectable

import (
	"context"
	"sync"
	"time"

	"github.com/reactivex/rxgo/fx"
	"github.com/reactivex/rxgo/observable"
	"github.com/reactivex/rxgo/observer"
	"github.com/reactivex/rxgo/rx"
)

// Connectable is an Observable which can subscribe several
// EventHandlers before starting processing with Connect.
type Connectable struct {
	observable.Observable
	// observers []rx.Observer
	observers []rx.Observer
}

// New creates a Connectable with optional observer(s) as parameters.
// func New(buffer uint, observers ...rx.Observer) Connectable {
func New(buffer uint, handlers ...interface{}) Connectable {
	observers := []rx.Observer{}
	for _, handler := range handlers {
		ob := observer.New(handler)
		observers = append(observers, ob)
	}
	return Connectable{
		// Observable: make(observable.Observable, int(buffer)),
		Observable: observable.New(buffer),
		observers:  observers,
	}
}

// From creates a Connectable from an Iterator.
func From(it rx.Iterator) Connectable {
	return Connectable{Observable: observable.From(it)}
	// source := make(chan interface{})
	// go func() {
	// 	for {
	// 		val, err := it.Next()
	// 		if err != nil {
	// 			break
	// 		}
	// 		source <- val
	// 	}
	// 	close(source)
	// }()
	// return Connectable{Observable: source}
}

// Empty creates a Connectable with no item and terminate immediately.
func Empty() Connectable {
	// source := make(chan interface{})
	// go func() {
	// 	close(source)
	// }()
	// return Connectable{Observable: source}
	return Connectable{Observable: observable.Empty()}
}

// Interval creates a Connectable emitting incremental integers infinitely between
// each given time interval.
// func Interval(term chan struct{}, timeout time.Duration) Connectable {
func Interval(ctx context.Context, interval time.Duration) Connectable {
	// source := make(chan interface{})
	// go func(term chan struct{}) {
	// 	i := 0
	// OuterLoop:
	// 	for {
	// 		select {
	// 		case <-term:
	// 			break OuterLoop
	// 		case <-time.After(timeout):
	// 			source <- i
	// 		}
	// 		i++
	// 	}
	// 	close(source)
	// }(term)
	// return Connectable{Observable: source}
	return Connectable{Observable: observable.Interval(ctx, interval)}
}

// Range creates an Observable that emits a particular range of sequential integers.
func Range(start, end int) Connectable {
	// source := make(chan interface{})
	// go func() {
	// 	i := start
	// 	for i < end {
	// 		source <- i
	// 		i++
	// 	}
	// 	close(source)
	// }()
	// return Connectable{Observable: source}
	return Connectable{Observable: observable.Range(start, end)}
}

// Just creates an Connectable with the provided item(s).
func Just(item interface{}, items ...interface{}) Connectable {
	// source := make(chan interface{})
	// if len(items) > 0 {
	// 	items = append([]interface{}{item}, items...)
	// } else {
	// 	items = []interface{}{item}
	// }

	// go func() {
	// 	for _, item := range items {
	// 		source <- item
	// 	}
	// 	close(source)
	// }()

	// return Connectable{Observable: source}
	return Connectable{Observable: observable.Just(item, items...)}
}

// Start creates a Connectable from one or more directive-like EmittableFunc
// and emits the result of each operation asynchronously on a new Connectable.
func Start(f fx.EmitFunc, fs ...fx.EmitFunc) Connectable {

	// if len(fs) > 0 {
	// 	fs = append([]fx.EmittableFunc{f}, fs...)
	// } else {
	// 	fs = []fx.EmittableFunc{f}
	// }

	// source := make(chan interface{})

	// var wg sync.WaitGroup
	// for _, f := range fs {
	// 	wg.Add(1)
	// 	go func(f fx.EmittableFunc) {
	// 		source <- f()
	// 		wg.Done()
	// 	}(f)
	// }

	// go func() {
	// 	wg.Wait()
	// 	close(source)
	// }()

	// return Connectable{Observable: source}
	return Connectable{Observable: observable.Start(f, fs...)}
}

// Subscribe subscribes an EventHandler and returns a Connectable.
func (co Connectable) Subscribe(handler interface{}, handlers ...interface{}) Connectable {
	// observers := []rx.Observer{observer.New(handler)}
	co.observers = append(co.observers, observer.New(handler))
	if len(handlers) > 0 {
		for _, handler := range handlers {
			// observers = append(observers, observer.New(handler))
			co.observers = append(co.observers, observer.New(handler))
		}
	}
	// co.observers = append(co.observers, ob)
	return co
}

// Do is like Subscribe but subscribes a func(interface{}) as a NextHandler
func (co Connectable) Do(nextf func(interface{})) Connectable {
	ob := observer.New(nextf)
	co.observers = append(co.observers, ob)
	return co
}

func subscribe(ctx context.Context, cancel context.CancelFunc, ob rx.Observer, copy []interface{}, wg *sync.WaitGroup) {
	go func(ob rx.Observer) {
		// defer wg.Done()
		for _, item := range copy {
			select {
			case <-ctx.Done():
				return
			default:
				switch item := item.(type) {
				case error:
					ob.OnError(item)
					// cancel()
					wg.Done()
					return
				default:
					ob.OnNext(item)
				}
			}
		}
		ob.OnDone()
		// cancel()
		wg.Done()
	}(ob)
}

// Connect activates the Observable stream and returns a channel of Subscription channel.
// func (co Connectable) Connect() <-chan (chan subscription.Subscription) {
func (co Connectable) Connect() context.Context {
	// done := make(chan (chan subscription.Subscription), 1)
	source := []interface{}{}

	for item := range co.Observable {
		source = append(source, item)
	}

	var wg sync.WaitGroup
	wg.Add(len(co.observers))

	ctx, cancel := context.WithCancel(context.Background())

	for _, ob := range co.observers {
		local := make([]interface{}, len(source))
		copy(local, source)

		// fin := make(chan struct{})
		// sub := subscription.New().Subscribe()

		go subscribe(ctx, cancel, ob, local, &wg)
		/*
			go func(ob rx.Observer) {
			OuterLoop:
				for _, item := range local {
					switch item := item.(type) {
					case error:
						ob.OnError(item)

						// Record error
						// sub.Error = item
						break OuterLoop
					default:
						ob.OnNext(item)
					}
				}
				// fin <- struct{}{}
				cancel()
			}(ob)
		*/

		// temp := make(chan subscription.Subscription)

		// go func(ob observer.Observer) {
		// 	<-fin
		// 	if sub.Error == nil {
		// 		ob.OnDone()
		// 		sub.Unsubscribe()
		// 	}

		// 	go func() {
		// 		temp <- sub
		// 		done <- temp
		// 	}()
		// 	wg.Done()
		// }(ob)
	}

	go func() {
		wg.Wait()
		cancel()
		// close(done)
	}()

	// return done
	return ctx
}

// Map maps a MappableFunc predicate to each item in Connectable and
// returns a new Connectable with applied items.
func (co Connectable) Map(apply rx.MapFunc) Connectable {
	// source := make(chan interface{}, len(co.Observable))
	// go func() {
	// 	for item := range co.Observable {
	// 		source <- fn(item)
	// 	}
	// 	close(source)
	// }()
	// return Connectable{Observable: Observable(source)}
	o := co.Observable.Map(apply)
	if ob, ok := o.(observable.Observable); ok {
		return Connectable{Observable: ob}
	}
	return Connectable{}
}

// Filter filters items in the original Connectable and returns
// a new Connectable with the filtered items.
func (co Connectable) Filter(apply rx.FilterFunc) Connectable {
	// source := make(chan interface{}, len(co.Observable))
	// go func() {
	// 	for item := range co.Observable {
	// 		if fn(item) {
	// 			source <- item
	// 		}
	// 	}
	// 	close(source)
	// }()
	// return Connectable{Observable: Observable(source)}
	o := co.Observable.Filter(apply)
	if ob, ok := o.(observable.Observable); ok {
		return Connectable{Observable: ob}
	}
	return Connectable{}
}

// Scan applies ScannableFunc predicate to each item in the original
// Connectable sequentially and emits each successive value on a new Connectable.
func (co Connectable) Scan(apply rx.ScanFunc) Connectable {
	// out := make(chan interface{})
	// go func() {
	// 	var current interface{}
	// 	for item := range co.Observable {
	// 		out <- apply(current, item)
	// 		current = apply(current, item)
	// 	}
	// 	close(out)
	// }()
	// return Connectable{Observable: Observable(out)}
	o := co.Observable.Scan(apply)
	if ob, ok := o.(observable.Observable); ok {
		return Connectable{Observable: ob}
	}
	return Connectable{}
}

// First returns new Connectable which emits only first item.
func (co Connectable) First() Connectable {
	// out := make(chan interface{})
	// go func() {
	// 	for item := range co.Observable {
	// 		out <- item
	// 		break
	// 	}
	// 	close(out)
	// }()
	// return Connectable{Observable: out}
	return Connectable{Observable: co.Observable.First()}
}

// Last returns a new Connectable which emits only last item.
func (co Connectable) Last() Connectable {
	out := make(chan interface{})
	go func() {
		var last interface{}
		for item := range co.Observable {
			last = item
		}
		out <- last
		close(out)
	}()
	return Connectable{Observable: out}
}

//Distinct suppress duplicate items in the original Connectable and
//returns a new Connectable.
func (co Connectable) Distinct(apply fx.KeySelectFunc) Connectable {
	out := make(chan interface{})
	go func() {
		keysets := make(map[interface{}]struct{})
		for item := range co.Observable {
			key := apply(item)
			_, ok := keysets[key]
			if !ok {
				out <- item
			}
			keysets[key] = struct{}{}
		}
		close(out)
	}()
	return Connectable{Observable: out}
}

//DistinctUntilChanged suppress duplicate items in the original Connectable only
// if they are successive to one another and returns a new Connectable.
func (co Connectable) DistinctUntilChanged(apply fx.KeySelectFunc) Connectable {
	out := make(chan interface{})
	go func() {
		var current interface{}
		for item := range co.Observable {
			key := apply(item)
			if current != key {
				out <- item
				current = key
			}
		}
		close(out)
	}()
	return Connectable{Observable: out}
}
