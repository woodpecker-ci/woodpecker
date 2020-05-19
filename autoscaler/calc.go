package autoscaler

import "math"

func Max(x int, y int) int {
	return int(math.Max(float64(x), float64(y)))
}


func Min(x int, y int) int {
	return int(math.Min(float64(x), float64(y)))
}

func Abs(x int) int {
	return int(math.Abs(float64(x)))
}
