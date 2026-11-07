// Package secret encrypts what the base keeps of the notes (AES-256-GCM): a stolen base, backup or disk is unreadable
// without the server's key NOTES_SECRET. Not a compromised server: it holds the key.
// Each account has its own random key, kept in users.data_key sealed by NOTES_SECRET: changing NOTES_SECRET only seals
// those again (Rewrap), not every note.
package secret

import (
	"context"
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

	"github.com/uptrace/bun"
)

// Prefix starts every sealed value, followed by the id of its account.
const Prefix = "enc1:"

var (
	ErrNoKey     = errors.New("NOTES_SECRET must be set to 32 random bytes in base64: openssl rand -base64 32")
	ErrNotLoaded = errors.New("the key of the account is not loaded")
	ErrCorrupt   = errors.New("a value of the base cannot be decrypted: is NOTES_SECRET the one it was written with?")
)

var master = sync.OnceValues(func() (cipher.AEAD, error) {
	return newAEAD(os.Getenv("NOTES_SECRET"))
})

// keys are the keys of the accounts loaded, by account id: they never change, Rewrap only seals them again.
var keys sync.Map

// Load reads the key of an account into memory, made on its first use: Seal and Open need it.
func Load(ctx context.Context, db bun.IDB, userID string) error {
	if _, ok := keys.Load(userID); ok {
		return nil
	}
	gcm, err := master()
	if err != nil {
		return err
	}

	var wrapped string
	if err := db.NewRaw("SELECT coalesce(data_key, '') FROM users WHERE id = ?", userID).Scan(ctx, &wrapped); err != nil {
		return err
	}
	if wrapped == "" {
		key := make([]byte, 32)
		rand.Read(key)
		// Two first uses at once: the key saved first wins, the other reads it back.
		err := db.NewRaw("UPDATE users SET data_key = coalesce(data_key, ?) WHERE id = ? RETURNING data_key",
			seal(gcm, key), userID).Scan(ctx, &wrapped)
		if err != nil {
			return err
		}
	}

	key, err := open(gcm, wrapped)
	if err != nil {
		return err
	}
	account, err := newAEAD(base64.StdEncoding.EncodeToString(key))
	if err != nil {
		return err
	}
	keys.Store(userID, account)
	return nil
}

// Rewrap gives back the key of an account sealed by oldSecret, sealed by NOTES_SECRET instead; unchanged when
// NOTES_SECRET opens it already.
func Rewrap(wrapped, oldSecret string) (string, error) {
	gcm, err := master()
	if err != nil {
		return "", err
	}
	if _, err := open(gcm, wrapped); err == nil {
		return wrapped, nil
	}
	old, err := newAEAD(oldSecret)
	if err != nil {
		return "", err
	}
	key, err := open(old, wrapped)
	if err != nil {
		return "", err
	}
	return seal(gcm, key), nil
}

// Seal encrypts text with the key of the account, loaded: the same text never gives the same value twice.
func Seal(userID, text string) (Text, error) {
	gcm, err := account(userID)
	if err != nil {
		return "", err
	}
	return Text(Prefix + userID + ":" + seal(gcm, []byte(text))), nil
}

// SealBytes encrypts the bytes of a file with the key of the account, loaded: nonce then ciphertext, no base64.
func SealBytes(userID string, data []byte) ([]byte, error) {
	gcm, err := account(userID)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	return gcm.Seal(nonce, nonce, data, nil), nil
}

// OpenBytes decrypts what SealBytes gave.
func OpenBytes(userID string, data []byte) ([]byte, error) {
	gcm, err := account(userID)
	if err != nil {
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, ErrCorrupt
	}
	text, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	if err != nil {
		return nil, ErrCorrupt
	}
	return text, nil
}

// Open decrypts what Seal gave, its account's key loaded.
func Open(value string) (string, error) {
	data, sealed := strings.CutPrefix(value, Prefix)
	if !sealed {
		return "", ErrCorrupt
	}
	userID, data, _ := strings.Cut(data, ":")
	gcm, err := account(userID)
	if err != nil {
		return "", err
	}
	text, err := open(gcm, data)
	return string(text), err
}

// Text is a string the base keeps sealed: bun opens it on scan. To write one, Seal it: bun refuses it in clear.
type Text string

func (t Text) Value() (driver.Value, error) {
	if !strings.HasPrefix(string(t), Prefix) {
		return nil, errors.New("secret.Text written in clear: Seal it first")
	}
	return string(t), nil
}

func (t *Text) Scan(src any) error {
	var value string
	switch src := src.(type) {
	case nil:
		*t = ""
		return nil
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

// Helpers

func account(userID string) (cipher.AEAD, error) {
	gcm, ok := keys.Load(userID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotLoaded, userID)
	}
	return gcm.(cipher.AEAD), nil
}

func newAEAD(secret string) (cipher.AEAD, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(secret))
	if err != nil || len(key) != 32 {
		return nil, ErrNoKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func seal(gcm cipher.AEAD, data []byte) string {
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, data, nil))
}

func open(gcm cipher.AEAD, value string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil || len(raw) < gcm.NonceSize() {
		return nil, ErrCorrupt
	}
	data, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return nil, ErrCorrupt
	}
	return data, nil
}
