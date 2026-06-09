package istio_bridge

import (
	"math/rand"
	"time"
)

var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
var letterName = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString(n int, t string) string {
	b := make([]rune, n)
	for i := range b {
		if t == "name" {
			b[i] = letterName[rand.Intn(len(letterName))]
		} else {
			b[i] = letter[rand.Intn(len(letter))]
		}
	}
	return string(b)
}
