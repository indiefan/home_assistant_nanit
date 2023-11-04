package utils

import "strings"

// AnonymizeToken - replaces chunk of a string with '*'
func AnonymizeToken(token string, clearLen int) string {
	if clearLen != 0 && (len(token)-2*clearLen) > 6 {
		runes := []rune(token)
		return string(runes[0:clearLen]) + strings.Repeat("*", len(token)-2*clearLen) + string(runes[len(token)-clearLen:])
	}

	return strings.Repeat("*", len(token))
}
