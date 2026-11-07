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

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ErrNotConfigured = errors.New("S3_ENDPOINT, S3_ACCESS_KEY, S3_SECRET_KEY and S3_BUCKET must be set")

// Store is Silo, or Memory in the tests.
type Store interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Remove(ctx context.Context, key string) error
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
	return &s3{client: client, bucket: bucket}, nil
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

// Memory is a Store in a map, for the tests.
type Memory struct {
	mu      sync.Mutex
	Objects map[string][]byte
}

func NewMemory() *Memory {
	return &Memory{Objects: map[string][]byte{}}
}

func (m *Memory) Put(_ context.Context, key string, data []byte, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Objects[key] = data
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
	return nil
}
