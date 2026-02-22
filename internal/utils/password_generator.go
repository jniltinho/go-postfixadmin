package utils

import (
	"crypto/rand"
	"math/big"
	"time"
)

// GenerateComplexPassword gera uma senha complexa aleatória.
func GenerateComplexPassword() string {
	const (
		length      = 16
		upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowerChars  = "abcdefghijklmnopqrstuvwxyz"
		digitChars  = "0123456789"
		symbolChars = "!@#$%^&*()-_=+[]{}|;:,.<>?"
		allChars    = upperChars + lowerChars + digitChars + symbolChars
	)

	password := make([]byte, length)

	// Ensure at least one character from each category
	password[0] = upperChars[RandomInt(len(upperChars))]
	password[1] = lowerChars[RandomInt(len(lowerChars))]
	password[2] = digitChars[RandomInt(len(digitChars))]
	password[3] = symbolChars[RandomInt(len(symbolChars))]

	// Fill the rest with random characters
	for i := 4; i < length; i++ {
		password[i] = allChars[RandomInt(len(allChars))]
	}

	// Shuffle the password to avoid predictable patterns
	for i := range password {
		j := RandomInt(len(password))
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

// RandomInt gera um número inteiro aleatório entre 0 e max-1.
func RandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// Fallback to timestamp-based randomness if crypto/rand fails
		return int(time.Now().UnixNano()) % max
	}
	return int(n.Int64())
}
