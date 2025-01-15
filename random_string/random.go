package random_string

import (
	"math/rand"
	"time"
)

const alphaNumUpper = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const alphaNumLower = "abcdefghijklmnopqrstuvwxyz0123456789"
const alphaLower    = "abcdefghijklmnopqrstuvwxyz"

func AlphaNumString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = alphaNumUpper[seededRand.Intn(len(alphaNumUpper))]
	}
	return string(result)
}

func AlphaNumStringLower(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = alphaNumLower[seededRand.Intn(len(alphaNumLower))]
	}
	return string(result)
}

func AlphaStringLower(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = alphaLower[seededRand.Intn(len(alphaLower))]
	}
	return string(result)
}