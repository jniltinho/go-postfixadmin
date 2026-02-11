package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
)

// CheckPassword verifies if the plaintext password matches the hashed password
// supports {MD5}, {CRYPT}, and $6$ (SHA512-CRYPT)
func CheckPassword(plain, hashed string) (bool, error) {
	if strings.HasPrefix(hashed, "{MD5}") {
		hash := md5.Sum([]byte(plain))
		hexHash := hex.EncodeToString(hash[:])
		return hexHash == hashed[5:], nil
	}

	if strings.HasPrefix(hashed, "{CRYPT}") {
		hashed = hashed[7:]
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

// HashPassword generates a SHA512-CRYPT hash (default for new passwords)
func HashPassword(plain string) (string, error) {
	cryptScheme := crypt.SHA512.New()
	hash, err := cryptScheme.Generate([]byte(plain), []byte("$6$rounds=5000$salt"))
	if err != nil {
		return "", err
	}
	return hash, nil
}
