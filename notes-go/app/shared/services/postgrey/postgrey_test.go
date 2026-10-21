package postgrey

import (
	"context"
	"net/url"
	"testing"
	"time"
)

func TestGetEnvURIEscapesCredentials(t *testing.T) {
	env := &PostgresEnvironment{username: "app", password: "p@ss/:w", dbname: "notes"}

	uri, err := url.Parse(env.GetEnvURI())
	if err != nil {
		t.Fatalf("parse %q: %v", env.GetEnvURI(), err)
	}
	if got, _ := uri.User.Password(); got != "p@ss/:w" {
		t.Errorf("password = %q, want %q", got, "p@ss/:w")
	}
}

func TestGetEnvURIDefaults(t *testing.T) {
	env := &PostgresEnvironment{dbname: "notes"}

	uri, err := url.Parse(env.GetEnvURI())
	if err != nil {
		t.Fatal(err)
	}
	if uri.Host != HOST+":"+PORT {
		t.Errorf("host = %q, want %q", uri.Host, HOST+":"+PORT)
	}
	if uri.User != nil {
		t.Errorf("user = %v, want none", uri.User)
	}
}

func TestConnect(t *testing.T) {
	env := NewEnvironment()
	env.ReadEnv()
	db := env.Connect()
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
}
