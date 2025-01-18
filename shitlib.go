package shitpic

import (
	"bytes"
	"image/gif"
	"image/png"
	"io"
	"math/rand/v2"

	"github.com/disintegration/imaging"
)

func Recompress(in io.Reader, out io.Writer, quality int) error {
	img, err := imaging.Decode(in, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}
	return imaging.Encode(out, img, imaging.JPEG, imaging.JPEGQuality(quality))
}

func Gifize(in io.Reader, out io.Writer) error {
	img, err := imaging.Decode(in, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	return gif.Encode(out, img, nil)
}

func Pngerate(in io.Reader, out io.Writer) error {
	img, err := imaging.Decode(in, imaging.AutoOrientation(true))
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
