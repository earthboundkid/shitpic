package shitpic

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand/v2"
)

func Recompress(in io.Reader, out io.Writer, quality int) error {
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}

	return jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
}

func Gifize(in io.Reader, out io.Writer) error {
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}

	return gif.Encode(out, img, nil)
}

func Pngerate(in io.Reader, out io.Writer) error {
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}

	return png.Encode(out, img)
}

func Uglify(in []byte, while func(func() bool), lowerBound int) ([]byte, error) {
	randRange := 100 - lowerBound

	var rbuf, wbuf bytes.Buffer

	_, err := rbuf.Write(in)
	if err != nil {
		return nil, err
	}

	for range while {
		quality := lowerBound + rand.N(randRange)
		if err = Recompress(&rbuf, &wbuf, quality); err != nil {
			return nil, err
		}
		// Prevent allocations by reseting and swapping
		rbuf.Reset()
		rbuf, wbuf = wbuf, rbuf
	}
	return rbuf.Bytes(), nil
}
