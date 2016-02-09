package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"os"
	"strings"
)

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

func uglify(in io.Reader, cycles, lowerBound int) (io.Reader, error) {
	randRange := 100 - lowerBound

	var rbuf, wbuf bytes.Buffer

	_, err := io.Copy(&rbuf, in)
	if err != nil {
		return nil, err
	}

	for i := 0; i < cycles; i++ {
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

func die(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Usage = func() {
		usage := `Usage of shitpic:
	shitpic [options] input output

Shitpic accepts and can output JPEG, GIF, and PNG files.
`
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
	}

	cycles := flag.Uint("cycles", 100, "How many times to reprocess input")
	lowerBound := flag.Int("quality", 75, "Lower bound of quality (0–100)")
	reduceColor := flag.Bool("reduce-colors", false, "Reduce to 256 colors")

	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	infilename := flag.Arg(0)
	outfilename := flag.Arg(1)

	if *lowerBound < 0 || *lowerBound > 100 {
		flag.Usage()
		os.Exit(2)
	}

	var r io.Reader = os.Stdin
	if infilename != "-" {
		f, err := os.Open(infilename)
		die(err)
		defer f.Close()
		r = f
	}

	if *reduceColor {
		var buf bytes.Buffer
		die(gifize(r, &buf))
		r = &buf
	}

	var err error
	r, err = uglify(r, int(*cycles), *lowerBound)
	die(err)

	if strings.HasSuffix(outfilename, ".png") {
		var buf bytes.Buffer
		die(pngerate(r, &buf))
		r = &buf
	}

	if strings.HasSuffix(outfilename, ".gif") {
		var buf bytes.Buffer
		die(gifize(r, &buf))
		r = &buf
	}

	var outf = os.Stdout
	if outfilename != "-" {
		f, err := os.Create(outfilename)
		die(err)
		defer f.Close()
		outf = f
	}

	_, err = io.Copy(outf, r)
	die(err)
}
