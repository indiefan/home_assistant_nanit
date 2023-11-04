package utils

// MinInt - returns smaller of 2 integers
// TODO: Go seriously doesn't have a Min function??
func MinInt(a int, b int) int {
	if a > b {
		return b
	}

	return a
}
