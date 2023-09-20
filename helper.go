package main

import "strings"

func EscapeText(text string) string {
	replacements := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, replacement := range replacements {
		text = strings.ReplaceAll(text, replacement, "\\"+replacement)
	}
	return text
}
