package main

import (
	"syscall/js"
	"time"

	"github.com/earthboundkid/shitpic"
)

func init() {
	js.Global().Set("uglify", jsUglify)
}

var jsUglify = goPromise(func(args []js.Value) (js.Value, bool) {
	bufV, timeV, qualityV := args[0], args[1], args[2]
	b := valueToBytes(bufV)
	d := time.Duration(timeV.Int()) * time.Millisecond
	q := qualityV.Int()
	b, err := shitpic.Uglify(b, doFor(d), q)
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
