package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"io"
	"math/rand"
	"os"

	_ "image/png"
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
		fmt.Fprintln(os.Stderr, "Usage of shitpic:")
		fmt.Fprintln(os.Stderr, "\tshitpic [options] input output")
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

	if *lowerBound < 0 || *lowerBound > 100 {
		flag.Usage()
		os.Exit(2)
	}

	var inf io.Reader = os.Stdin
	if flag.Arg(0) != "-" {
		f, err := os.Open(flag.Arg(0))
		die(err)
		defer f.Close()
		inf = f
	}

	if *reduceColor {
		var buf bytes.Buffer
		die(gifize(inf, &buf))
		inf = &buf
	}

	out, err := uglify(inf, int(*cycles), *lowerBound)
	die(err)

	var outf = os.Stdout
	if flag.Arg(1) != "-" {
		f, err := os.Create(flag.Arg(1))
		die(err)
		defer f.Close()
		outf = f
	}

	_, err = io.Copy(outf, out)
	die(err)
}
