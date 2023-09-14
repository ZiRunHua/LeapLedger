package util

import (
	mathRand "math/rand"
	"time"
)

type rand struct{}

var Rand = &rand{}

const charSet string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func (rd *rand) GenerateRandomString(length int) string {
	mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	var result string
	for i := 0; i < length; i++ {
		randomIndex := mathRand.Intn(len(charSet))
		result += string(charSet[randomIndex])
	}
	return result
}
