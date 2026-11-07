// Package password hashes the passwords of accounts with argon2id, and makes the session tokens.
package password

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// The OWASP minimum for argon2id: 19 MiB, 2 passes, 1 thread.
const (
	memory  = 19 * 1024
	passes  = 2
	threads = 1
	keyLen  = 32
)

var encoding = base64.RawStdEncoding

// Hash gives the password's argon2id hash with its random salt, in the PHC format ($argon2id$v=19$m=…).
func Hash(password string) string {
	salt := make([]byte, 16)
	rand.Read(salt) // Never fails: crypto/rand panics rather than return an error.
	key := argon2.IDKey([]byte(password), salt, passes, memory, threads, keyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, passes, threads, encoding.EncodeToString(salt), encoding.EncodeToString(key))
}

// Verify tells whether password gives hash, in constant time; false for a hash it cannot read.
func Verify(password, hash string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var m uint32
	var t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false
	}
	salt, err1 := encoding.DecodeString(parts[4])
	want, err2 := encoding.DecodeString(parts[5])
	if err1 != nil || err2 != nil {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

// Token gives a new session token: 32 random bytes, for the client only.
func Token() string {
	token := make([]byte, 32)
	rand.Read(token)
	return base64.RawURLEncoding.EncodeToString(token)
}

// TokenHash is what the base keeps of a token: enough to recognize it, useless to log in.
func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
