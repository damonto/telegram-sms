package util

import (
	"fmt"
	"strings"
)

func EscapeText(text string) string {
	replacements := []string{"#", "-", ".", "!"}
	// replacements := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, replacement := range replacements {
		text = strings.ReplaceAll(text, replacement, "\\"+replacement)
	}
	fmt.Println(text)
	return text
}
