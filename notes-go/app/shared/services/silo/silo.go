// Package silo keeps the bytes of the files in Silo, the S3 storage of compose.yaml.
package silo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/secret"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ErrNotConfigured = errors.New("S3_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY and S3_BUCKET must be set")

// Store is Silo, or Memory in the tests.
type Store interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Remove(ctx context.Context, key string) error
	List(ctx context.Context) ([]Object, error)
}

// Object is what the store keeps under a key, its bytes left out: purgefiles reads Written to leave alone an
// upload that may still be running.
type Object struct {
	Key     string
	Written time.Time
}

type s3 struct {
	client *minio.Client
	bucket string
}

// New connects to the Silo of the S3_* variables, and makes its bucket on first use.
// S3_SECURE=true for TLS: Silo is reached on the loopback or the Docker network, in clear.
func New(ctx context.Context) (Store, error) {
	endpoint, bucket := strings.TrimSpace(os.Getenv("S3_ENDPOINT")), strings.TrimSpace(os.Getenv("S3_BUCKET"))
	access, secret := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY")), strings.TrimSpace(os.Getenv("S3_SECRET_KEY"))
	if endpoint == "" || bucket == "" || access == "" || secret == "" {
		return nil, ErrNotConfigured
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, secret, ""),
		Secure: os.Getenv("S3_SECURE") == "true",
	})
	if err != nil {
		return nil, err
	}
	exists, err := client.BucketExists(ctx, bucket)
	if err == nil && !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("silo %s: %w", endpoint, err)
	}
	return Sealed(&s3{client: client, bucket: bucket}), nil
}

// Sealed encrypts the bytes put in store with the key of their account, the first part of the key, and decrypts
// them on Get: the account's key must be loaded.
func Sealed(store Store) Store {
	return sealed{store}
}

type sealed struct {
	Store
}

func (s sealed) Put(ctx context.Context, key string, data []byte, contentType string) error {
	userID, _, _ := strings.Cut(key, "/")
	data, err := secret.SealBytes(userID, data)
	if err != nil {
		return err
	}
	return s.Store.Put(ctx, key, data, contentType)
}

// Get reads the whole file to decrypt it: GCM checks the bytes only at their end. The API caps a file at 20 MB.
func (s sealed) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	body, err := s.Store.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	userID, _, _ := strings.Cut(key, "/")
	data, err = secret.OpenBytes(userID, data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *s3) Put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: contentType})
	return err
}

// Get fails on a missing key at its first read, not here: minio-go asks Silo only then.
func (s *s3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
}

// Remove succeeds on a missing key.
func (s *s3) Remove(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

// List walks the whole bucket, every account: the keys only, never the bytes.
func (s *s3) List(ctx context.Context) ([]Object, error) {
	var objects []Object
	for info := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Recursive: true}) {
		if info.Err != nil {
			return nil, info.Err
		}
		objects = append(objects, Object{Key: info.Key, Written: info.LastModified})
	}
	return objects, nil
}

// Memory is a Store in a map, for the tests.
type Memory struct {
	mu      sync.Mutex
	Objects map[string][]byte
	Written map[string]time.Time // when each key was put; a test moves it back to age an object
}

func NewMemory() *Memory {
	return &Memory{Objects: map[string][]byte{}, Written: map[string]time.Time{}}
}

func (m *Memory) Put(_ context.Context, key string, data []byte, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Objects[key] = data
	m.Written[key] = time.Now()
	return nil
}

func (m *Memory) Get(_ context.Context, key string) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.Objects[key]
	if !ok {
		return nil, fmt.Errorf("no object %q", key)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *Memory) Remove(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Objects, key)
	delete(m.Written, key)
	return nil
}

func (m *Memory) List(_ context.Context) ([]Object, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var objects []Object
	for key := range m.Objects {
		objects = append(objects, Object{Key: key, Written: m.Written[key]})
	}
	return objects, nil
}
