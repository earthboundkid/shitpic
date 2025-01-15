package main

import (
	"syscall/js"
	"time"

	"github.com/earthboundkid/shitpic"
)

func init() {
	js.Global().Set("uglify", wrapUglify)
}

var wrapUglify = goPromise(func(args []js.Value) (js.Value, bool) {
	bufV := args[0]
	b := valueToBytes(bufV)
	b, err := shitpic.Uglify(b, doFor(10*time.Second), 10)
	if err != nil {
		return js.ValueOf(err.Error()), false
	}
	return bytesToValue(b), true
})

func doFor(d time.Duration) func(func() bool) {
	return func(yield func() bool) {
		for end := time.Now().Add(d); time.Now().Before(end) && yield(); {
		}
	}
}
