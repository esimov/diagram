package canvas

import "math/rand"

// min returns the smallest number between two numbers.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// max returns the biggest number between two numbers.
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// random generate a random number.
func random() float64 {
	return rand.Float64()
}
