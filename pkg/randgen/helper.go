package randgen

import (
	"math"
	"math/rand"
)

func RandomNumber(digits int) int {
	low := int(math.Pow10(digits - 1))
	high := int(math.Pow10(digits)) - 1
	return low + rand.Intn(high-low+1)
}
