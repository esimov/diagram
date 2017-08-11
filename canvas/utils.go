package canvas

import "math/rand"

// Returns the smallest number between two numbers.
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Returns the biggest number between two numbers.
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Generate a random number.
func random() float64 {
	return rand.Float64()
}