package main

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/disintegration/gift"
)

func getExt(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// LoadImage loads image from file
func LoadImage(filename string) (image.Image, error) {
	var decoder func(io.Reader) (image.Image, error)

	ext := getExt(filename)
	switch ext {
	case ".jpg", ".jpeg":
		decoder = jpeg.Decode
	case ".gif":
		decoder = gif.Decode
	case ".png":
		decoder = png.Decode
	}

	if decoder == nil {
		return nil, errors.New("Unsupported file format : " + ext)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		file.Close()
	}()

	img, err := decoder(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// SaveJpeg saves an image to jpeg file
func SaveJpeg(img image.Image, dir string, filename string, quality int) error {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	file, err := os.Create(path.Join(dir, filename))
	if err != nil {
		return err
	}

	defer func() {
		file.Close()
	}()

	return jpeg.Encode(file, img, &jpeg.Options{quality})
}

// CreateImage creates an image instance
func CreateImage(width, height int, bgColor color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, bgColor)
		}
	}
	return img
}

// FillRect draws a filled rect to image
func FillRect(img *image.RGBA, x1, y1, x2, y2 int, rectColor color.Color) {
	for x := x1; x < x2; x++ {
		for y := y1; y < y2; y++ {
			img.Set(x, y, rectColor)
		}
	}
}

// DrawLine draws a line to image
func DrawLine(img *image.RGBA, x1, y1, x2, y2 int, lineColor color.Color) {
	dx, dy := x2-x1, y2-y1
	if dx <= dy {
		incX := float32(dx) / float32(dy)
		x := float32(x1)
		for y := y1; y < y2; y++ {
			img.Set(int(x), y, lineColor)
			x += incX
		}
	} else {
		incY := float32(dy) / float32(dx)
		y := float32(y1)
		for x := x1; x < x2; x++ {
			img.Set(x, int(y), lineColor)
			y += incY
		}
	}
}

// CalcRotatedSize returns image width/height after rotating given angle.
func CalcRotatedSize(w, h int, angle float32) (int, int) {
	if w <= 0 || h <= 0 {
		return 0, 0
	}

	xoff := float32(w)/2 - 0.5
	yoff := float32(h)/2 - 0.5

	asin, acos := Sincosf32(angle)
	x1, y1 := RotatePoint(0-xoff, 0-yoff, asin, acos)
	x2, y2 := RotatePoint(float32(w-1)-xoff, 0-yoff, asin, acos)
	x3, y3 := RotatePoint(float32(w-1)-xoff, float32(h-1)-yoff, asin, acos)
	x4, y4 := RotatePoint(0-xoff, float32(h-1)-yoff, asin, acos)

	minx := Minf32(x1, Minf32(x2, Minf32(x3, x4)))
	maxx := Maxf32(x1, Maxf32(x2, Maxf32(x3, x4)))
	miny := Minf32(y1, Minf32(y2, Minf32(y3, y4)))
	maxy := Maxf32(y1, Maxf32(y2, Maxf32(y3, y4)))

	neww := maxx - minx + 1
	if neww-Floorf32(neww) > 0.01 {
		neww += 2
	}
	newh := maxy - miny + 1
	if newh-Floorf32(newh) > 0.01 {
		newh += 2
	}
	return int(neww), int(newh)
}

// RotatePoint rotates a given point
func RotatePoint(x, y, asin, acos float32) (float32, float32) {
	newx := x*acos - y*asin
	newy := x*asin + y*acos
	return newx, newy
}

// RotateImage rotates the image by given angle.
// empty area after rotation is filled with bgColor
func RotateImage(
	src image.Image,
	angle float32,
	bgColor color.Color) image.Image {
	bounds := src.Bounds()
	width, height := CalcRotatedSize(bounds.Dx(), bounds.Dy(), angle)
	dest := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dest, dest.Bounds(),
		&image.Uniform{color.White},
		image.ZP,
		draw.Src)
	rotateFilter := gift.Rotate(angle, bgColor, gift.CubicInterpolation)
	gift.New(rotateFilter).Draw(dest, src)
	return dest
}

// ResizeImage resizes image to given dimension while preserving aspect ratio.
func ResizeImage(src image.Image, width, height int) image.Image {
	g := gift.New(gift.ResizeToFit(width, height, gift.LanczosResampling))
	dest := image.NewRGBA(g.Bounds(src.Bounds()))
	g.Draw(dest, src)
	return dest
}

// ---------------------------------------------------------------------------
// 줄간격 변경
// ---------------------------------------------------------------------------
type lineRange struct {
	start        int
	end          int
	height       int
	targetHeight int
	emptyLine    bool
}

func (r *lineRange) calc(scale float32, minHeight, maxRemove int) {
	r.height = r.end - r.start + 1
	if !r.emptyLine || r.height <= minHeight {
		r.targetHeight = r.height
	} else {
		if maxRemove > 0 {
			r.targetHeight = Max(minHeight, int(float32(r.height)*scale+0.5))
			if removed := r.height - r.targetHeight; removed > maxRemove {
				r.targetHeight = r.height - maxRemove
			}
		}
	}
}

func (r lineRange) getReducedHeight() int {
	return r.height - r.targetHeight
}

func getBrightness(r, g, b uint32) uint32 {
	return (r + g + b) / 3
}

// getLineRanges returns list of text and empty lines
func getLineRanges(src image.Image,
	threshold uint32,
	emptyLineThreshold float64) []lineRange {
	bounds := src.Bounds()
	srcWidth, srcHeight := bounds.Dx(), bounds.Dy()
	threshold16 := threshold * 256

	var ranges []lineRange
	var r lineRange

	maxDotCount := int(emptyLineThreshold)
	if emptyLineThreshold < 1 {
		maxDotCount = int(float64(srcWidth) * emptyLineThreshold)
	}
	for y := 0; y < srcHeight; y++ {
		emptyLine := true
		dotCount := 0
		for x := 0; x < srcWidth; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			brightness := getBrightness(r, g, b)
			if brightness < threshold16 {
				dotCount++
				if dotCount >= maxDotCount {
					emptyLine = true
					break
				}
			}
		}

		if emptyLine {
			if y == 0 {
				r = lineRange{start: y, end: y, emptyLine: true}
			} else {
				if r.emptyLine {
					r.end = y
				} else {
					ranges = append(ranges, r)
					r = lineRange{start: y, end: y, emptyLine: true}
				}
			}
		} else {
			if y == 0 {
				r = lineRange{start: y, end: y, emptyLine: false}
			} else {
				if r.emptyLine {
					ranges = append(ranges, r)
					r = lineRange{start: y, end: y, emptyLine: false}
				} else {
					r.end = y
				}
			}
		}

	}
	ranges = append(ranges, r)
	return ranges
}

func processLineRanges(
	ranges []lineRange,
	width int,
	widthRatio float32,
	heightRatio float32,
	lineSpaceScale float32,
	minSpace int,
	maxRemove int) int {

	targetHeight := 0
	for i := 0; i < len(ranges); i++ {
		r := &ranges[i]
		r.calc(lineSpaceScale, minSpace, maxRemove)
		targetHeight += r.targetHeight
	}

	minTargetHeight := int(heightRatio * float32(width) / widthRatio)

	loop := 0
	maxLoopCount := 5
	for targetHeight < minTargetHeight && loop < maxLoopCount {
		totalReducedHeight := 0
		for i := 0; i < len(ranges); i++ {
			r := ranges[i]
			if r.emptyLine {
				totalReducedHeight += r.getReducedHeight()
			}
		}

		totalInc := 0
		if totalReducedHeight > 0 {
			totalEmptyLinesToInc := minTargetHeight - targetHeight
			for i := 0; i < len(ranges); i++ {
				r := &ranges[i]
				if r.emptyLine {
					reducedHeight := r.getReducedHeight()
					heightToInc := Min(reducedHeight,
						reducedHeight*totalEmptyLinesToInc/totalReducedHeight)
					inc := int(float32(heightToInc) + 0.5)
					r.targetHeight = r.targetHeight + inc
					totalInc += inc
				}
			}
		}

		targetHeight = 0
		for i := 0; i < len(ranges); i++ {
			r := &ranges[i]
			targetHeight += r.targetHeight
		}

		loop++

		if totalInc <= int(float32(targetHeight)/100) {
			break
		}
	}

	return targetHeight
}

// ChangeLineSpace removes/adds spaces between lines
func ChangeLineSpace(
	src image.Image,
	widthRatio float32,
	heightRatio float32,
	lineSpaceScale float32,
	minSpace int,
	maxRemove int,
	threshold uint32) image.Image {

	ranges := getLineRanges(src, threshold)

	width := src.Bounds().Dx()
	targetHeight := processLineRanges(ranges,
		width,
		widthRatio,
		heightRatio,
		lineSpaceScale,
		minSpace,
		maxRemove)

	dest := image.NewRGBA(image.Rect(0, 0, width, targetHeight))
	destY := 0
	for i := 0; i < len(ranges); i++ {
		r := &ranges[i]
		rangeHeight := r.height
		rangeTargetHeight := r.targetHeight

		if rangeHeight > 0 && rangeTargetHeight > 0 {
			srcRect := image.Rect(0, r.start, width, r.start+rangeTargetHeight)
			subImage := src.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(srcRect)

			destRect := image.Rect(0, destY, width, targetHeight)
			draw.Draw(dest, destRect, subImage, image.ZP, draw.Src)
			destY -= (rangeHeight - rangeTargetHeight)
		}
	}
	return dest
}
