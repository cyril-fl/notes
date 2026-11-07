// Package secret encrypts what the base keeps of the notes with the server's key NOTES_SECRET (AES-256-GCM): a stolen
// base, backup or disk is unreadable without the key. Not a compromised server: it holds the key.
// shortcut: one key for every account and no rotation, add a key per account wrapped by NOTES_SECRET to rotate it.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Prefix starts every sealed value: a value without it was written before encryption (notes encrypt).
const Prefix = "enc1:"

var (
	ErrNoKey   = errors.New("NOTES_SECRET must be set to 32 random bytes in base64: openssl rand -base64 32")
	ErrCorrupt = errors.New("a value of the base cannot be decrypted: is NOTES_SECRET the one it was written with?")
)

var aead = sync.OnceValues(func() (cipher.AEAD, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(os.Getenv("NOTES_SECRET")))
	if err != nil || len(key) != 32 {
		return nil, ErrNoKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
})

// Seal encrypts text with a random nonce: the same text never gives the same value twice.
func Seal(text string) (string, error) {
	gcm, err := aead()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	return Prefix + base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(text), nil)), nil
}

// Open decrypts what Seal gave. A value without Prefix comes back as is: written in clear, until notes encrypt.
func Open(value string) (string, error) {
	data, sealed := strings.CutPrefix(value, Prefix)
	if !sealed {
		return value, nil
	}
	gcm, err := aead()
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil || len(raw) < gcm.NonceSize() {
		return "", ErrCorrupt
	}
	text, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return "", ErrCorrupt
	}
	return string(text), nil
}

// Text is a string the base keeps sealed: bun seals it on write and opens it on scan.
type Text string

func (t Text) Value() (driver.Value, error) {
	return Seal(string(t))
}

func (t *Text) Scan(src any) error {
	var value string
	switch src := src.(type) {
	case nil:
	case string:
		value = src
	case []byte:
		value = string(src)
	default:
		return fmt.Errorf("secret.Text cannot scan %T", src)
	}
	text, err := Open(value)
	*t = Text(text)
	return err
}
