package observable

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/reactivex/rxgo/errors"
	"github.com/reactivex/rxgo/fx"
	"github.com/reactivex/rxgo/observer"
	"github.com/reactivex/rxgo/rx"
)

// Observable is a basic observable channel
type Observable <-chan interface{}

var (
	// DefaultObservable is a zero buffer observable
	// channel equivalent to New(0).
	DefaultObservable = make(Observable)
	none              = new(uint)
)

// New creates an Observable with the
func New(buffer uint) Observable {
	return make(Observable, int(buffer))
}

// Next returns the next item on the Observable.
func (o Observable) Next() (interface{}, error) {
	if next, ok := <-o; ok {
		return next, nil
	}
	return nil, errors.New(errors.EndOfIteratorError)
}

func subscribe(ctx context.Context, cancel context.CancelFunc, handler interface{}, o Observable) context.Context {
	ob := observer.New(handler)
	go func() {
		for item := range o {
			select {
			case <-ctx.Done():
				return
			default:
				switch item := item.(type) {
				case error:
					ob.OnError(item)
					cancel()
				default:
					ob.OnNext(item)
				}
			}
		}
		ob.OnDone()
		cancel()
	}()
	return ctx
}

// Subscribe subscribes an EventHandler and returns a Subscription channel.
func (o Observable) Subscribe(handler interface{}) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	return subscribe(ctx, cancel, handler, o)
}

// SubscribeFor is a shortcut for Subscribe with time.Duration provided as timeout. It creates
// a context.Context  with context.WithTimeout. See https://golang.org/pkg/context/#WithTimeout.
func (o Observable) SubscribeFor(handler interface{}, timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return subscribe(ctx, cancel, handler, o)
}

// SubscribeUntil takes a deadlinea and create a context.Context with context.WithDeadline
// see https://golang.org/pkg/context/#WithDeadline.
func (o Observable) SubscribeUntil(handler interface{}, deadline time.Time) context.Context {
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	return subscribe(ctx, cancel, handler, o)
}

// Map returns a mapping of each item on the source Observable as a new rx.Observable.
func (o Observable) Map(apply rx.MapFunc) rx.Observable {
	out := make(chan interface{})
	go func() {
		for item := range o {
			out <- apply(item)
		}
		close(out)
	}()
	return Observable(out)
}

// Take retrieves the first n items on the source
// Observable and emits them on a new returning Observable.
func (o Observable) Take(n uint) Observable {
	out := make(chan interface{})
	go func() {
		takeCount := 0
		for item := range o {
			if takeCount < int(n) {
				takeCount++
				out <- item
				continue
			}
			break
		}
		close(out)
	}()
	return Observable(out)
}

// TakeLast retrieves the last n items on the source Observable
// and emit them on a new returning Observable.
func (o Observable) TakeLast(n uint) Observable {
	out := make(chan interface{})
	go func() {
		buf := make([]interface{}, n)
		for item := range o {
			if len(buf) >= int(n) {
				buf = buf[1:]
			}
			buf = append(buf, item)
		}
		for _, takenItem := range buf {
			out <- takenItem
		}
		close(out)
	}()
	return Observable(out)
}

// Filter filters items on the source Observable and returns
// a new Observable with the filtered items.
func (o Observable) Filter(apply rx.FilterFunc) rx.Observable {
	out := make(chan interface{})
	go func() {
		for item := range o {
			if apply(item) {
				out <- item
			}
		}
		close(out)
	}()
	return Observable(out)
}

// First retrieves the first item on the source Observable
// and emits it on the new returning Observable.
func (o Observable) First() Observable {
	out := make(chan interface{})
	go func() {
		for item := range o {
			out <- item
			break
		}
		close(out)
	}()
	return Observable(out)
}

// Last retrieves the last item on the source Observable
// and emits it on the new returning Observable.
func (o Observable) Last() Observable {
	out := make(chan interface{})
	go func() {
		var last interface{}
		for item := range o {
			last = item
		}
		out <- last
		close(out)
	}()
	return Observable(out)
}

// Distinct merges duplicate items on the original Observable and
// only emits distinct items on a new returning Observable.
func (o Observable) Distinct(apply fx.KeySelectFunc) Observable {
	out := make(chan interface{})
	go func() {
		keysets := make(map[interface{}]struct{})
		for item := range o {
			key := apply(item)
			_, ok := keysets[key]
			if !ok {
				out <- item
			}
			keysets[key] = struct{}{}
		}
		close(out)
	}()
	return Observable(out)
}

// DistinctUntilChanged suppresses the subsequent duplicates on the
// source Observable and only emits distinct items on a new returning Observable.
func (o Observable) DistinctUntilChanged(apply fx.KeySelectFunc) Observable {
	out := make(chan interface{})
	go func() {
		var current interface{}
		for item := range o {
			key := apply(item)
			if current != key {
				out <- item
				current = key
			}
		}
		close(out)
	}()
	return Observable(out)
}

// Scan applies rx.ScanFunc to each item in the original Observable sequentially
// and emits each successive value (accumulator) on a new returning Observable.
// It is similar to fold left without an initial value.
func (o Observable) Scan(apply rx.ScanFunc) rx.Observable {
	out := make(chan interface{})
	go func() {
		var current interface{}
		for item := range o {
			out <- apply(current, item)
			current = apply(current, item)
		}
		close(out)
	}()
	return Observable(out)
}

// From creates a new Observable from an Iterator, which can be created
// from any sequence of items such as slice and array.
func From(it rx.Iterator) Observable {
	source := make(chan interface{})
	go func() {
		for {
			val, err := it.Next()
			if err != nil {
				break
			}
			source <- val
		}
		close(source)
	}()
	return Observable(source)
}

// Empty creates an Observable with no item and terminate immediately.
func Empty() Observable {
	source := make(chan interface{})
	go func() {
		close(source)
	}()
	return Observable(source)
}

// Interval creates an Observable emitting incremental integers infinitely between
// each given time interval.
// func Interval(term chan struct{}, interval time.Duration) Observable {
func Interval(ctx context.Context, interval time.Duration) Observable {
	source := make(chan interface{})
	// go func(term chan struct{}) {
	// 	i := 0
	// OuterLoop:
	// 	for {
	// 		select {
	// 		case <-term:
	// 			break OuterLoop
	// 		case <-time.After(interval):
	// 			source <- i
	// 		}
	// 		i++
	// 	}
	// 	close(source)
	// }(term)
	go func() {
		i := 0
	OuterLoop:
		for {
			select {
			case <-ctx.Done():
				break OuterLoop
			case <-time.After(interval):
				source <- i
			}
			i++
		}
		close(source)
	}()
	return Observable(source)
}

// Range creates an Observable that emits a particular range of sequential integers.
func Range(start, end int) Observable {
	source := make(chan interface{})
	go func() {
		i := start
		for i < end {
			source <- i
			i++
		}
		close(source)
	}()
	return Observable(source)
}

// Just creates an Observable with the provided item(s) as-is.
func Just(item interface{}, items ...interface{}) Observable {
	source := make(chan interface{})
	if len(items) > 0 {
		items = append([]interface{}{item}, items...)
	} else {
		items = []interface{}{item}
	}

	go func() {
		for _, item := range items {
			source <- item
		}
		close(source)
	}()

	return Observable(source)
}

// Start creates an Observable from one or more directive EmitFunc
// and emits the result of each operation asynchronously on a new
// returning Observable.
func Start(f fx.EmitFunc, fs ...fx.EmitFunc) Observable {
	if len(fs) > 0 {
		fs = append([]fx.EmitFunc{f}, fs...)
	} else {
		fs = []fx.EmitFunc{f}
	}

	source := make(chan interface{})

	var wg sync.WaitGroup
	for _, f := range fs {
		wg.Add(1)
		go func(f fx.EmitFunc) {
			source <- f()
			wg.Done()
		}(f)
	}

	// Wait in another goroutine to not block
	go func() {
		wg.Wait()
		close(source)
	}()

	return Observable(source)
}

// Merge combines multiple Observables into one by merging their emissions.
func Merge(o1 rx.Observable, o2 rx.Observable, on ...rx.Observable) Observable {
	out := make(chan interface{})
	go func() {
		chans := append([]rx.Observable{o1, o2}, on...)
		count := len(chans)
		cases := make([]reflect.SelectCase, count)
		for i := range cases {
			cases[i].Dir = reflect.SelectRecv
			cases[i].Chan = reflect.ValueOf(chans[i])
		}
		for count > 0 {
			chosen, recv, recvOk := reflect.Select(cases)
			if recvOk {
				out <- recv.Interface()
			} else {
				cases[chosen].Chan = reflect.ValueOf(nil)
				count--
			}
		}
		close(out)
	}()
	return Observable(out)
}

// CombineLatest emits an item whenever any of the source Observables emits an item
func CombineLatest(o []Observable, apply fx.CombineFunc) Observable {
	out := make(chan interface{})
	go func() {
		chans := o
		count := len(chans)
		left := len(chans)
		is := make([]interface{}, len(chans))
		for i := 0; i < len(is); i++ {
			is[i] = none
		}
		cases := make([]reflect.SelectCase, count)
		for i := range cases {
			cases[i].Dir = reflect.SelectRecv
			cases[i].Chan = reflect.ValueOf(chans[i])
		}
		for count > 0 {
			chosen, recv, recvOk := reflect.Select(cases)
			if recvOk {
				if is[chosen] == none {
					left--
				}
				is[chosen] = recv.Interface()
				if left == 0 {
					out <- apply(is)
				}
			} else {
				cases[chosen].Chan = reflect.ValueOf(nil)
				count--
			}
		}
		close(out)
	}()
	return Observable(out)
}
