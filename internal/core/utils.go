package core

import (
	"encoding/hex"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandBytes() []byte {
	token := make([]byte, 16)
	for i := range token {
		//nolint
		token[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return token
}

func MustHexDecode(h string) []byte {
	ret, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	return ret
}
