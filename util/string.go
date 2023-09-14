package util

import (
	"bytes"
	"regexp"
)

type str struct{}

var Str str

func (s *str) MaskEmail(email string) string {
	re := regexp.MustCompile(`([^@]+)@(.+)`)
	matches := re.FindStringSubmatch(email)

	if len(matches) == 3 {
		username := matches[1]
		domain := matches[2]

		maskedUsername := s.MaskString(username, 3)

		maskedEmail := maskedUsername + "@" + domain
		return maskedEmail
	}

	return email
}

func (s *str) MaskString(input string, visibleChars int) string {
	if len(input) <= visibleChars {
		return input
	}

	maskedChars := len(input) - visibleChars
	maskedString := input[:visibleChars] + s.RepeatChar('*', maskedChars)
	return maskedString
}

func (s *str) RepeatChar(char byte, count int) string {
	return s.StringOf(char, count)
}

func (s *str) StringOf(char byte, count int) string {
	return string(bytes.Repeat([]byte{char}, count))
}
