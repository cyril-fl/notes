package rotatesecret

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createschema"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/testaccount"
)

// Rewrap itself is tested in services/secret: here, keys already sealed by NOTES_SECRET stay, and one no secret
// opens stops everything.
func TestRotateSecret(t *testing.T) {
	ctx := context.Background()
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	defer db.Close()

	schema := createschema.Process(ctx, db)
	schema.Run()
	if schema.HasErr() {
		t.Fatalf("create schema: %v", schema.GetErr())
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	ctx, tx = testaccount.New(t, ctx, tx)

	const old = "b2xkIHNlY3JldCBvZiB0aGUgc2VydmVyLCAzMiBieXQ="
	process := Process(ctx, tx, old)
	process.Run()
	if process.HasErr() || process.GetResult() != 0 {
		t.Fatalf("rewrapped = %d, err = %v; want none, the key is sealed by NOTES_SECRET", process.GetResult(), process.GetErr())
	}

	if _, err := tx.NewUpdate().Model((*domain.User)(nil)).Set("data_key = 'bm90IGEga2V5'").Where("id = ?", domain.UserID(ctx)).Exec(ctx); err != nil {
		t.Fatal(err)
	}
	broken := Process(ctx, tx, old)
	broken.Run()
	if !broken.HasErr() {
		t.Error("err = nil, want a key no secret opens to stop the rotation")
	}
}

// An account whose key the old secret sealed: a rotation stopped by a broken key changes nothing, a full one keeps
// what was written before readable. The values are made here, by hand, as the base kept them before the rotation.
func TestRotateSecretKeepsNotes(t *testing.T) {
	ctx := context.Background()
	env := postgrey.NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	defer db.Close()

	schema := createschema.Process(ctx, db)
	schema.Run()
	if schema.HasErr() {
		t.Fatalf("create schema: %v", schema.GetErr())
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	const old = "b2xkIHNlY3JldCBvZiB0aGUgc2VydmVyLCAzMiBieXQ="
	oldKey, _ := base64.StdEncoding.DecodeString(old)
	key := make([]byte, 32)
	rand.Read(key)
	sealedKey := base64.StdEncoding.EncodeToString(seal(t, oldKey, key))

	user := domain.User{Name: fmt.Sprintf("test-%d", time.Now().UnixNano()), PasswordHash: "-", DataKey: sealedKey}
	broken := domain.User{Name: user.Name + "-broken", PasswordHash: "-", DataKey: "bm90IGEga2V5"}
	for _, u := range []*domain.User{&user, &broken} {
		if err := tx.NewInsert().Model(u).Returning("id").Scan(ctx); err != nil {
			t.Fatal(err)
		}
	}

	stopped := Process(ctx, tx, old)
	stopped.Run()
	var kept string
	if err := tx.NewRaw("SELECT data_key FROM users WHERE id = ?", user.ID).Scan(ctx, &kept); err != nil {
		t.Fatal(err)
	}
	if !stopped.HasErr() || kept != sealedKey {
		t.Fatalf("err = %v, key changed = %v; want the broken key to stop the rotation and leave every key as it was", stopped.GetErr(), kept != sealedKey)
	}

	if _, err := tx.NewDelete().Model(&broken).WherePK().Exec(ctx); err != nil {
		t.Fatal(err)
	}
	process := Process(ctx, tx, old)
	process.Run()
	if process.HasErr() || process.GetResult() != 1 {
		t.Fatalf("rewrapped = %d, err = %v; want the key of the account", process.GetResult(), process.GetErr())
	}

	// Notes, titles and image names are secret.Text, tags are found by Blind, images are SealBytes.
	if err := secret.Load(ctx, tx, user.ID); err != nil {
		t.Fatalf("load the key after the rotation: %v", err)
	}
	text, err := secret.Open(secret.Prefix + user.ID + ":" + base64.StdEncoding.EncodeToString(seal(t, key, []byte("a note"))))
	if err != nil || text != "a note" {
		t.Errorf("open = %q, %v; want the note written before the rotation", text, err)
	}
	image, err := secret.OpenBytes(user.ID, seal(t, key, []byte("an image")))
	if err != nil || string(image) != "an image" {
		t.Errorf("open bytes = %q, %v; want the image put before the rotation", image, err)
	}
	blind, err := secret.Blind(user.ID, "Tag")
	if want := hex.EncodeToString(mac(mac(key, "blind"), "tag")); err != nil || blind != want {
		t.Errorf("blind = %q, %v; want %q, the one kept before the rotation", blind, err, want)
	}
}

// seal is AES-GCM, nonce then ciphertext, as services/secret writes it.
func seal(t *testing.T, key, data []byte) []byte {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)
	return gcm.Seal(nonce, nonce, data, nil)
}

func mac(key []byte, text string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(text))
	return h.Sum(nil)
}
