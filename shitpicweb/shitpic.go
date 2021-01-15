package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"os"
	"syscall/js"
	"time"
)

func init() {
	js.Global().Set("uglify", wrapUglify)
}

var wrapUglify = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	bufV := args[0]
	size := bufV.Length()
	b := make([]byte, size)
	if n := js.CopyBytesToGo(b, bufV); n != size {
		panic("bad read")
	}

	p, set, reject := newPromise()
	go func() {
		b, err := uglify(b, 10*time.Second, 10)
		if err != nil {
			reject(js.ValueOf(err.Error()))
			return
		}
		set(bytesToValue(b))
	}()
	return p
})

func recompress(in io.Reader, out io.Writer, quality int) error {
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}

	return jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
}

func gifize(in io.Reader, out io.Writer) error {
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}

	return gif.Encode(out, img, nil)
}

func pngerate(in io.Reader, out io.Writer) error {
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}

	return png.Encode(out, img)
}

func uglify(in []byte, d time.Duration, lowerBound int) ([]byte, error) {
	randRange := 100 - lowerBound

	var rbuf, wbuf bytes.Buffer
	_, err := rbuf.Write(in)
	if err != nil {
		return nil, err
	}

	for done := time.Now().Add(d); time.Now().Before(done); {
		fmt.Fprint(os.Stderr, ".")

		quality := lowerBound + rand.Intn(randRange)
		if err = recompress(&rbuf, &wbuf, quality); err != nil {
			return nil, err
		}
		// Prevent allocations by reseting and swapping
		rbuf.Reset()
		rbuf, wbuf = wbuf, rbuf
	}
	return rbuf.Bytes(), nil
}
