package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"golang.org/x/crypto/bcrypt"
)

// CheckPassword verifies if the plaintext password matches the hashed password
// supports {MD5}, {CRYPT}, $6$ (SHA512-CRYPT), and Bcrypt ($2a$, $2y$, $2b$)
func CheckPassword(plain, hashed string) (bool, error) {
	if strings.HasPrefix(hashed, "{MD5}") {
		hash := md5.Sum([]byte(plain))
		hexHash := hex.EncodeToString(hash[:])
		return hexHash == hashed[5:], nil
	}

	hashed = strings.TrimPrefix(hashed, "{CRYPT}")

	// Check for Bcrypt ($2a$, $2y$, $2b$)
	// Check for Bcrypt ($2a$, $2y$, $2b$)
	if strings.HasPrefix(hashed, "$2") {
		// Go's bcrypt package expects $2a$ or $2b, but handles $2y$ by treating it as $2a$ if we replace it
		// standard lib support for $2y$ might be limited, so normalize for check
		normalizedHash := strings.Replace(hashed, "$2y$", "$2a$", 1)
		err := bcrypt.CompareHashAndPassword([]byte(normalizedHash), []byte(plain))
		return err == nil, nil
	}

	// For CRYPT and SHA512-CRYPT
	cryptScheme := crypt.SHA512.New()
	if strings.HasPrefix(hashed, "$1$") {
		cryptScheme = crypt.MD5.New()
	} else if strings.HasPrefix(hashed, "$6$") {
		cryptScheme = crypt.SHA512.New()
	} else {
		// Fallback or simple crypt? PostfixAdmin often uses salt
		// This is a simplified version.
		return plain == hashed, nil
	}

	err := cryptScheme.Verify(hashed, []byte(plain))
	return err == nil, nil
}

// HashPassword generates a Bcrypt hash ($2y$10$) compatible with PHP/PostfixAdmin
func HashPassword(plain string) (string, error) {
	// Generate bcrypt hash with default cost (10)
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	// Replace $2a$ with $2y$ for PHP compatibility (fixes high-bit bug in old PHP versions)
	// PostfixAdmin typically expects $2y$ prefix
	return strings.Replace(string(hash), "$2a$", "$2y$", 1), nil
}

// HashPasswordMD5Crypt generates a MD5-CRYPT hash ($1$) with a random salt
func HashPasswordMD5Crypt(plain string) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	saltStr := fmt.Sprintf("$1$%x", salt)

	cryptScheme := crypt.MD5.New()
	hash, err := cryptScheme.Generate([]byte(plain), []byte(saltStr))
	if err != nil {
		return "", err
	}
	return hash, nil
}

// HashPasswordBcrypt generates a Bcrypt hash ($2y$)
func HashPasswordBcrypt(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return strings.Replace(string(bytes), "$2a$", "$2y$", 1), nil
}
