package main

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"syscall/js"
	"time"
)

func init() {
	js.Global().Set("uglify", wrapUglify)
}

var wrapUglify = goPromise(func(args []js.Value) (js.Value, bool) {
	bufV := args[0]
	b := valueToBytes(bufV)

	b, err := uglify(b, 10*time.Second, 10)
	if err != nil {
		return js.ValueOf(err.Error()), false
	}
	return bytesToValue(b), true
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
