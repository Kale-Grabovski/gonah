package helper

import (
	"math/rand"
	"regexp"
	"strings"
)

type StringHelper struct{}

func (s *StringHelper) RandString(n int) string {
	letterRunes := []rune("123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (s *StringHelper) RandStringLower(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var space = regexp.MustCompile(`\s+`)

func (s StringHelper) Trim(st string) string {
	return strings.ToLower(space.ReplaceAllString(strings.TrimSpace(st), " "))
}
