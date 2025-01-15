package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"iter"
	"math"
	"os"

	"github.com/disintegration/gift"
	"github.com/disintegration/imaging"
)

func main() {
	flag.Usage = func() {
		usage := `Usage of DIthered Color PICture:
	dicpic [options]
`
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
	}

	nameIn := flag.String("in", "input.jpeg", "`path` to input file")
	nameOut := flag.String("out", "output.png", "`path` to output file")
	width := flag.Int("width", 160, "pixel width for dithering")
	linearize := flag.Bool("linearize", false, "convert to linear color space")
	mode := flag.String("mode", "atkinson", "dithering mode (atkinson, floyd, r2, r2-triangle)")
	pct := flag.Int("factor", 100, "error diffusion multiplication `percentage`")
	flag.Parse()

	factor := float64(*pct) / 100

	// open image
	f, err := os.Open(*nameIn)
	die(err)
	defer f.Close()
	original, _, err := image.Decode(f)
	die(err)

	// Resize down
	resizedImg := original.(draw.Image)
	if *width > 0 {
		resizedImg = imaging.Resize(original, *width, 0, imaging.Lanczos)
	}

	pallettedImg := image.NewPaletted(resizedImg.Bounds(), cgaPalette)
	// convert to linear color space
	if *linearize {
		pallettedImg.Palette = linearPalette
		toLinear.Draw(resizedImg, resizedImg, nil)
	}

	// Dither using a linear palette
	switch *mode {
	default:
		fmt.Fprintf(os.Stderr, "unknown -mode: %s", *mode)
		flag.Usage()
		os.Exit(1)
	case "atkinson":
		drawAtkinson(pallettedImg, resizedImg.Bounds(), resizedImg, factor)
	case "floyd":
		drawFloyd(pallettedImg, resizedImg.Bounds(), resizedImg, factor)
	case "r2":
		drawNoisy(pallettedImg, resizedImg.Bounds(), resizedImg, factor, false)
	case "r2-triangle":
		drawNoisy(pallettedImg, resizedImg.Bounds(), resizedImg, factor, true)
	}

	// Swap palette back to sRGB if necessary
	pallettedImg.Palette = cgaPalette

	// Resize to original
	var final image.Image = pallettedImg
	if *width > 0 {
		final = imaging.Resize(pallettedImg, original.Bounds().Dx(), original.Bounds().Dy(), imaging.NearestNeighbor)
	}
	// Save
	fout, err := os.Create(*nameOut)
	die(err)
	defer fout.Close()
	die(png.Encode(fout, final))
}

var cyan = color.RGBA{R: 0x55, G: 0xFF, B: 0xFF, A: 0xFF}
var magenta = color.RGBA{R: 0xFF, G: 0x55, B: 0xFF, A: 0xFF}
var cgaPalette = color.Palette{
	color.White,
	cyan,
	magenta,
	color.Black,
	color.Transparent,
}

var toLinear = gift.ColorspaceSRGBToLinear()

var linearPalette = func() color.Palette {
	img := image.NewPaletted(image.Rect(0, 0, len(cgaPalette), 1), cgaPalette)
	for x := range len(cgaPalette) {
		img.Set(x, 0, cgaPalette[x])
	}

	linearImg := image.NewRGBA(toLinear.Bounds(img.Bounds()))
	toLinear.Draw(linearImg, img, nil)

	lp := make(color.Palette, len(cgaPalette))
	for x := range len(lp) {
		lp[x] = linearImg.At(x, 0)
	}
	return lp
}()

func die(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func drawAtkinson(dst *image.Paletted, r image.Rectangle, src image.Image, diffusionFactor float64) {
	type diffusion struct {
		r, g, b, a int32
	}
	palette := make([]diffusion, len(dst.Palette))
	for i, col := range dst.Palette {
		r, g, b, a := col.RGBA()
		palette[i].r = int32(r)
		palette[i].g = int32(g)
		palette[i].b = int32(b)
		palette[i].a = int32(a)
	}
	pix, stride := dst.Pix[dst.PixOffset(r.Min.X, r.Min.Y):], dst.Stride

	// quantErrorCurr and quantErrorNext are the quantization
	// errors that have been propagated to the pixels in a
	// three row window. The +3 simplifies calculation at the right edge.
	quantErrorCurr := make([]diffusion, r.Dx()+3)
	quantErrorNext1 := make([]diffusion, r.Dx()+3)
	quantErrorNext2 := make([]diffusion, r.Dx()+3)

	pxRGBA := func(x, y int) (r, g, b, a uint32) { return src.At(x, y).RGBA() }
	// Fast paths for special cases to avoid excessive use of the color.Color
	// interface which escapes to the heap but need to be discovered for
	// each pixel on r. See also https://golang.org/issues/15759.
	switch src0 := src.(type) {
	case *image.RGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.RGBAAt(x, y).RGBA() }
	case *image.NRGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.NRGBAAt(x, y).RGBA() }
	case *image.YCbCr:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.YCbCrAt(x, y).RGBA() }
	}

	for x, y := range tblr(r) {
		if x == r.Min.X && y != r.Min.Y {
			// Recycle the quantization error buffers.
			clear(quantErrorCurr)
			quantErrorCurr, quantErrorNext1, quantErrorNext2 = quantErrorNext1, quantErrorNext2, quantErrorCurr
		}

		// e.r, e.g and e.b are the pixel's R,G,B values plus the
		// error diffusion.
		sr, sg, sb, sa := pxRGBA(x, y)
		var e diffusion
		e.r, e.g, e.b, e.a = int32(sr), int32(sg), int32(sb), int32(sa)

		e.r = clamp(e.r + int32(float64(quantErrorCurr[x+1].r)*(diffusionFactor/8)))
		e.g = clamp(e.g + int32(float64(quantErrorCurr[x+1].g)*(diffusionFactor/8)))
		e.b = clamp(e.b + int32(float64(quantErrorCurr[x+1].b)*(diffusionFactor/8)))
		e.a = clamp(e.a + int32(float64(quantErrorCurr[x+1].a)*(diffusionFactor/8)))

		// Find the closest palette color in Euclidean R,G,B,A space:
		// the one that minimizes sum-squared-difference.
		bestIndex, bestSum := 0, uint32(1<<32-1)
		for index, p := range palette {
			sum := sqDiff(e.r, p.r) + sqDiff(e.g, p.g) + sqDiff(e.b, p.b) + sqDiff(e.a, p.a)
			if sum < bestSum {
				bestIndex, bestSum = index, sum
				if sum == 0 {
					break
				}
			}
		}
		pix[y*stride+x] = byte(bestIndex)

		e.r -= palette[bestIndex].r
		e.g -= palette[bestIndex].g
		e.b -= palette[bestIndex].b
		e.a -= palette[bestIndex].a

		// Propagate the Atkinson quantization error.
		quantErrorCurr[x+2].r += e.r
		quantErrorCurr[x+2].g += e.g
		quantErrorCurr[x+2].b += e.b
		quantErrorCurr[x+2].a += e.a
		quantErrorCurr[x+3].r += e.r
		quantErrorCurr[x+3].g += e.g
		quantErrorCurr[x+3].b += e.b
		quantErrorCurr[x+3].a += e.a

		quantErrorNext1[x+0].r += e.r
		quantErrorNext1[x+0].g += e.g
		quantErrorNext1[x+0].b += e.b
		quantErrorNext1[x+0].a += e.a
		quantErrorNext1[x+1].r += e.r
		quantErrorNext1[x+1].g += e.g
		quantErrorNext1[x+1].b += e.b
		quantErrorNext1[x+1].a += e.a
		quantErrorNext1[x+2].r += e.r
		quantErrorNext1[x+2].g += e.g
		quantErrorNext1[x+2].b += e.b
		quantErrorNext1[x+2].a += e.a

		quantErrorNext2[x+1].r += e.r
		quantErrorNext2[x+1].g += e.g
		quantErrorNext2[x+1].b += e.b
		quantErrorNext2[x+1].a += e.a

	}
}

// Adapted from image/draw
func drawFloyd(dst *image.Paletted, r image.Rectangle, src image.Image, diffusionFactor float64) {
	// If dst is an *image.Paletted, we have a fast path for dst.Set and
	// dst.At. The dst.Set equivalent is a batch version of the algorithm
	// used by color.Palette's Index method in image/color/color.go, plus
	// optional Floyd-Steinberg error diffusion.

	type diffusion struct {
		r, g, b, a int32
	}
	palette := make([]diffusion, len(dst.Palette))
	for i, col := range dst.Palette {
		r, g, b, a := col.RGBA()
		palette[i].r = int32(r)
		palette[i].g = int32(g)
		palette[i].b = int32(b)
		palette[i].a = int32(a)
	}
	pix, stride := dst.Pix[dst.PixOffset(r.Min.X, r.Min.Y):], dst.Stride

	// quantErrorCurr and quantErrorNext are the Floyd-Steinberg quantization
	// errors that have been propagated to the pixels in the current and next
	// rows. The +2 simplifies calculation near the edges.
	// var quantErrorCurr, quantErrorNext [][4]int32

	quantErrorCurr := make([]diffusion, r.Dx()+2)
	quantErrorNext := make([]diffusion, r.Dx()+2)

	pxRGBA := func(x, y int) (r, g, b, a uint32) { return src.At(x, y).RGBA() }
	// Fast paths for special cases to avoid excessive use of the color.Color
	// interface which escapes to the heap but need to be discovered for
	// each pixel on r. See also https://golang.org/issues/15759.
	switch src0 := src.(type) {
	case *image.RGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.RGBAAt(x, y).RGBA() }
	case *image.NRGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.NRGBAAt(x, y).RGBA() }
	case *image.YCbCr:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.YCbCrAt(x, y).RGBA() }
	}

	// Loop over each source pixel.
	// out := color.RGBA64{A: 0xffff}
	for y := 0; y != r.Dy(); y++ {
		for x := 0; x != r.Dx(); x++ {
			// er, eg and eb are the pixel's R,G,B values plus the
			// optional Floyd-Steinberg error.
			sr, sg, sb, sa := pxRGBA(x, y)
			er, eg, eb, ea := int32(sr), int32(sg), int32(sb), int32(sa)

			er = clamp(er + int32(float64(quantErrorCurr[x+1].r)*(diffusionFactor/16)))
			eg = clamp(eg + int32(float64(quantErrorCurr[x+1].g)*(diffusionFactor/16)))
			eb = clamp(eb + int32(float64(quantErrorCurr[x+1].b)*(diffusionFactor/16)))
			ea = clamp(ea + int32(float64(quantErrorCurr[x+1].a)*(diffusionFactor/16)))

			// Find the closest palette color in Euclidean R,G,B,A space:
			// the one that minimizes sum-squared-difference.
			// TODO(nigeltao): consider smarter algorithms.
			bestIndex, bestSum := 0, uint32(1<<32-1)
			for index, p := range palette {
				sum := sqDiff(er, p.r) + sqDiff(eg, p.g) + sqDiff(eb, p.b) + sqDiff(ea, p.a)
				if sum < bestSum {
					bestIndex, bestSum = index, sum
					if sum == 0 {
						break
					}
				}
			}
			pix[y*stride+x] = byte(bestIndex)

			er -= palette[bestIndex].r
			eg -= palette[bestIndex].g
			eb -= palette[bestIndex].b
			ea -= palette[bestIndex].a

			// Propagate the Floyd-Steinberg quantization error.
			quantErrorNext[x+0].r += er * 3
			quantErrorNext[x+0].g += eg * 3
			quantErrorNext[x+0].b += eb * 3
			quantErrorNext[x+0].a += ea * 3
			quantErrorNext[x+1].r += er * 5
			quantErrorNext[x+1].g += eg * 5
			quantErrorNext[x+1].b += eb * 5
			quantErrorNext[x+1].a += ea * 5
			quantErrorNext[x+2].r += er * 1
			quantErrorNext[x+2].g += eg * 1
			quantErrorNext[x+2].b += eb * 1
			quantErrorNext[x+2].a += ea * 1
			quantErrorCurr[x+2].r += er * 7
			quantErrorCurr[x+2].g += eg * 7
			quantErrorCurr[x+2].b += eb * 7
			quantErrorCurr[x+2].a += ea * 7
		}

		// Recycle the quantization error buffers.
		quantErrorCurr, quantErrorNext = quantErrorNext, quantErrorCurr
		clear(quantErrorNext)
	}
}

func tblr(r image.Rectangle) iter.Seq2[int, int] {
	r = r.Canon()
	return func(yield func(int, int) bool) {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x != r.Max.X; x++ {
				if !yield(x, y) {
					return
				}
			}
		}
	}
}

// sqDiff returns the squared-difference of x and y, shifted by 2 so that
// adding four of those won't overflow a uint32.
//
// x and y are both assumed to be in the range [0, 0xffff].
func sqDiff(x, y int32) uint32 {
	// This is an optimized code relying on the overflow/wrap around
	// properties of unsigned integers operations guaranteed by the language
	// spec. See sqDiff from the image/color package for more details.
	d := uint32(x - y)
	return (d * d) >> 2
}

// clamp clamps i to the interval [0, 0xffff].
func clamp(i int32) int32 {
	if i < 0 {
		return 0
	}
	if i > 0xffff {
		return 0xffff
	}
	return i
}

// See https://extremelearning.com.au/unreasonable-effectiveness-of-quasirandom-sequences/
var sqrt3 = math.Sqrt(3)
var pr = math.Cosh(math.Acosh(sqrt3*1.5)/3) * 2 / sqrt3
var a1 = 1.0 / pr
var a2 = 1.0 / (pr * pr)

func r2(x, y int) (float64, float64) {
	return float64(x+1) * a1, float64(y+1) * a2
}

func r2intensity(x, y int) float64 {
	v1, v2 := r2(x, y)
	return math.Mod(v1+v2, 1)
}

// triangleWave given a number between 0 and 1, biases it toward 1
func triangleWave(z float64) float64 {
	if z < .5 {
		return 2 * z
	}
	return 2 - 2*z
}

func r2noise(x, y int, errorFactor float64, triangle bool) int32 {
	z := r2intensity(x, y)
	if triangle {
		z = triangleWave(z)
	}
	z = 0xffff*z - 0x7FFF
	return int32(z * errorFactor)
}

func drawNoisy(dst *image.Paletted, r image.Rectangle, src image.Image, diffusionFactor float64, triangle bool) {
	type diffusion struct {
		r, g, b, a int32
	}
	palette := make([]diffusion, len(dst.Palette))
	for i, col := range dst.Palette {
		r, g, b, a := col.RGBA()
		palette[i].r = int32(r)
		palette[i].g = int32(g)
		palette[i].b = int32(b)
		palette[i].a = int32(a)
	}
	pix, stride := dst.Pix[dst.PixOffset(r.Min.X, r.Min.Y):], dst.Stride

	pxRGBA := func(x, y int) (r, g, b, a uint32) { return src.At(x, y).RGBA() }
	// Fast paths for special cases to avoid excessive use of the color.Color
	// interface which escapes to the heap but need to be discovered for
	// each pixel on r. See also https://golang.org/issues/15759.
	switch src0 := src.(type) {
	case *image.RGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.RGBAAt(x, y).RGBA() }
	case *image.NRGBA:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.NRGBAAt(x, y).RGBA() }
	case *image.YCbCr:
		pxRGBA = func(x, y int) (r, g, b, a uint32) { return src0.YCbCrAt(x, y).RGBA() }
	}

	for x, y := range tblr(r) {
		sr, sg, sb, sa := pxRGBA(x, y)
		var e diffusion
		e.r, e.g, e.b, e.a = int32(sr), int32(sg), int32(sb), int32(sa)

		e.r = clamp(e.r + r2noise(3*x+0, y, diffusionFactor, triangle))
		e.g = clamp(e.g + r2noise(3*x+1, y, diffusionFactor, triangle))
		e.b = clamp(e.b + r2noise(3*x+2, y, diffusionFactor, triangle))
		e.a = clamp(e.a)

		// Find the closest palette color in Euclidean R,G,B,A space:
		// the one that minimizes sum-squared-difference.
		bestIndex, bestSum := 0, uint32(1<<32-1)
		for index, p := range palette {
			sum := sqDiff(e.r, p.r) + sqDiff(e.g, p.g) + sqDiff(e.b, p.b) + sqDiff(e.a, p.a)
			if sum < bestSum {
				bestIndex, bestSum = index, sum
				if sum == 0 {
					break
				}
			}
		}
		pix[y*stride+x] = byte(bestIndex)
	}
}
