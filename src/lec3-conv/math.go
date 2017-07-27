package main
import "math"

func Max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

func Min(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

func Minf32(x, y float32) float32 {
	if x < y {
		return x
	}
	return y
}

func Maxf32(x, y float32) float32 {
	if x > y {
		return x
	}
	return y
}

func Sincosf32(a float32) (float32, float32) {
	sin, cos := math.Sincos(math.Pi * float64(a) / 180)
	return float32(sin), float32(cos)
}

func Floorf32(x float32) float32 {
	return float32(math.Floor(float64(x)))
}

func InRangef32(value, rangeFrom, rangeTo float32) bool {
	epsilon := float32(0.00001)
	return rangeFrom <= (value - epsilon) && (value + epsilon) <= rangeTo
}

func InRange(value, rangeFrom, rangeTo int) bool {
	return rangeFrom <= value && value <= rangeTo
}


