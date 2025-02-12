package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/earthboundkid/shitpic"
)

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
		die(shitpic.Gifize(r, &buf))
		r = &buf
	}

	b, err := io.ReadAll(r)
	die(err)

	b, err = shitpic.Uglify(b, nTimes(int(*cycles)), *lowerBound)
	die(err)
	fmt.Fprintln(os.Stderr)

	r = bytes.NewBuffer(b)
	if strings.HasSuffix(outfilename, ".png") {
		var buf bytes.Buffer
		die(shitpic.Pngerate(r, &buf))
		r = &buf
	}

	if strings.HasSuffix(outfilename, ".gif") {
		var buf bytes.Buffer
		die(shitpic.Gifize(r, &buf))
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

func nTimes(n int) func(func() bool) {
	return func(yield func() bool) {
		for i := 0; i < n && yield(); i++ {
			fmt.Fprint(os.Stderr, ".")
		}
	}
}
