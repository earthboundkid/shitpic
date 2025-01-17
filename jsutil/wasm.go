package jsutil

import (
	"syscall/js"
)

var (
	promise = js.Global().Get("Promise")
	array   = js.Global().Get("Uint8Array")
)

func NewPromise() (p js.Value, set, throw func(js.Value)) {
	type resultT struct {
		v  js.Value
		ok bool
	}
	type resolveT [2]js.Value
	resultCh := make(chan resultT)
	resolveCh := make(chan resolveT, 1)
	go func() {
		result := <-resultCh
		resolvers := <-resolveCh
		if result.ok {
			resolve := resolvers[0]
			resolve.Invoke(result.v)
		} else {
			reject := resolvers[1]
			reject.Invoke(result.v)
		}
	}()
	p = promise.New(js.FuncOf(func(this js.Value, args []js.Value) any {
		resolveCh <- resolveT{args[0], args[1]}
		return nil
	}))
	set = func(v js.Value) {
		resultCh <- resultT{v, true}
	}
	throw = func(v js.Value) {
		resultCh <- resultT{v, false}
	}
	return
}

func AsyncFunc(cb func(args []js.Value) (ret js.Value, ok bool)) js.Value {
	f := js.FuncOf(func(this js.Value, args []js.Value) any {
		p, set, reject := NewPromise()
		go func() {
			if ret, ok := cb(args); ok {
				set(ret)
			} else {
				reject(ret)
			}
		}()
		return p
	})
	return f.Value
}

func Await(awaitable js.Value) (ret js.Value, ok bool) {
	ch := make(chan struct{})
	go func() {
		awaitable.
			Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				ret = args[0]
				ok = true
				close(ch)
				return nil
			})).
			Call("catch", js.FuncOf(func(this js.Value, args []js.Value) any {
				ret = args[0]
				ok = false
				close(ch)
				return nil
			}))
	}()
	<-ch
	return
}

func BytesToValue(b []byte) js.Value {
	v := array.New(js.ValueOf(len(b)))
	js.CopyBytesToJS(v, b)
	return v
}

func ValueToBytes(v js.Value) []byte {
	size := v.Length()
	b := make([]byte, size)
	if n := js.CopyBytesToGo(b, v); n != size {
		panic("bad read")
	}
	return b
}
