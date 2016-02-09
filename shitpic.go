package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math/rand"
	"os"

	_ "image/gif" // register extra decoders for flexible input
	_ "image/png"
)

func recompress(in io.Reader, out io.Writer, quality int) error {
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}

	return jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
}

func uglify(in io.Reader, cycles, lowerBound int) (io.Reader, error) {
	randRange := 100 - lowerBound

	var buf, temp bytes.Buffer

	_, err := io.Copy(&buf, in)
	if err != nil {
		return nil, err
	}

	for i := 0; i < cycles; i++ {
		fmt.Fprint(os.Stderr, ".")

		quality := lowerBound + rand.Intn(randRange)
		if err = recompress(&buf, &temp, quality); err != nil {
			return nil, err
		}
		// Prevent allocations by reseting and swapping
		buf.Reset()
		buf, temp = temp, buf
	}
	fmt.Fprintln(os.Stderr)
	return &buf, nil
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

	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	if *lowerBound < 0 || *lowerBound > 100 {
		flag.Usage()
		os.Exit(2)
	}

	var inf = os.Stdin
	if flag.Arg(0) != "-" {
		f, err := os.Open(flag.Arg(0))
		die(err)
		defer f.Close()
		inf = f
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
