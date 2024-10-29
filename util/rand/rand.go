package rand

import (
	mathRand "math/rand"
	"time"
)

const charSet string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func String(length int) string {
	mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	var result string
	for i := 0; i < length; i++ {
		randomIndex := mathRand.Intn(len(charSet))
		result += string(charSet[randomIndex])
	}
	return result
}
func Int(max int) int {
	return mathRand.Intn(max)
}
