package main

import (
	"syscall/js"
)

func main() {
	// Prevent the function from returning, which is required in a wasm module
	select {}
}

var (
	fetch   = js.Global().Get("fetch")
	promise = js.Global().Get("Promise")
	array   = js.Global().Get("Uint8Array")
)

func newPromise() (p js.Value, set, throw func(js.Value)) {
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
	p = promise.New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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

func await(awaitable js.Value) chan []js.Value {
	ch := make(chan []js.Value)
	awaitable.
		Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ch <- args
			return nil
		})).
		Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			ch <- args
			return nil
		}))
	return ch
}
