package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
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

	p, set, _ := newPromise()
	go func() {
		time.Sleep(1 * time.Second)
		buf := bytes.NewReader(b)
		r, _ := uglify(buf, 10*time.Second, 10)
		b, _ := ioutil.ReadAll(r)
		_ = b
		v := array.New(js.ValueOf(len(b)))
		js.CopyBytesToJS(v, b)
		set(v)
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

func uglify(in io.Reader, d time.Duration, lowerBound int) (io.Reader, error) {
	randRange := 100 - lowerBound

	var rbuf, wbuf bytes.Buffer

	_, err := io.Copy(&rbuf, in)
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
	fmt.Fprintln(os.Stderr)
	return &rbuf, nil
}
