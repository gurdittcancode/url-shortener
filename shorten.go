package main

import (
	"math/rand"
)

func EncodeUrl(url string) string {
	const length = 7
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// Create a byte slice to hold the result
	b := make([]byte, length)

	// Fill the byte slice with random characters
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}
