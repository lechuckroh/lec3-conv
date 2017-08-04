package main

import (
	"image"
	"image/draw"
)

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
					emptyLine = false
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
	opt ChangeLineSpaceOption) int {

	targetHeight := 0
	for i := 0; i < len(ranges); i++ {
		r := &ranges[i]
		r.calc(opt.lineSpaceScale, opt.minSpace, opt.maxRemove)
		targetHeight += r.targetHeight
	}

	minTargetHeight := int(opt.heightRatio * float32(width) / opt.widthRatio)

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

// ChangeLineSpaceOption contains options for changeLineSpace filter
type ChangeLineSpaceOption struct {
	widthRatio         float32
	heightRatio        float32
	lineSpaceScale     float32
	minSpace           int
	maxRemove          int
	threshold          uint32
	emptyLineThreshold float64
}

// ChangeLineSpace removes/adds spaces between lines
func ChangeLineSpace(
	src image.Image,
	option ChangeLineSpaceOption) image.Image {

	ranges := getLineRanges(src, option.threshold, option.emptyLineThreshold)

	width := src.Bounds().Dx()
	targetHeight := processLineRanges(ranges, width, option)

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
