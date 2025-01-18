package main

import (
	"bytes"
	"syscall/js"
	"time"

	"github.com/disintegration/imaging"
	"github.com/earthboundkid/shitpic"
	"github.com/earthboundkid/shitpic/jsutil"
)

func main() {
	// Prevent the function from returning, which is required in a wasm module
	select {}
}

func init() {
	js.Global().Set("uglify", jsUglify)
	js.Global().Set("resize", jsResize)
}

var jsUglify = jsutil.AsyncFunc(func(args []js.Value) (js.Value, bool) {
	bufV, timeV, qualityV := args[0], args[1], args[2]
	b := jsutil.ValueToBytes(bufV)
	d := time.Duration(timeV.Int()) * time.Millisecond
	q := qualityV.Int()
	b, err := shitpic.Uglify(b, doFor(d), q)
	if err != nil {
		return js.ValueOf(err.Error()), false
	}
	return jsutil.BytesToValue(b), true
})

func doFor(d time.Duration) func(func() bool) {
	return func(yield func() bool) {
		for end := time.Now().Add(d); time.Now().Before(end) && yield(); {
		}
	}
}

var jsResize = jsutil.AsyncFunc(func(args []js.Value) (js.Value, bool) {
	bufV, widthV, smoothV := args[0], args[1], args[2]
	b := jsutil.ValueToBytes(bufV)
	width := widthV.Int()
	smooth := smoothV.Bool()

	buf := bytes.NewBuffer(b)
	original, err := imaging.Decode(buf, imaging.AutoOrientation(true))
	if err != nil {
		return js.ValueOf(err.Error()), false
	}

	mode := imaging.NearestNeighbor
	if smooth {
		mode = imaging.Lanczos
	}
	resizedImg := imaging.Resize(original, width, 0, mode)

	buf.Reset()
	if err = imaging.Encode(buf, resizedImg, imaging.PNG); err != nil {
		return js.ValueOf(err.Error()), false
	}
	return jsutil.BytesToValue(buf.Bytes()), true
})
