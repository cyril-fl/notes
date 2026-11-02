package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"main.go/app/shared/domain"
)

// No DB needed: uuids are not looked up, so a nil db is never touched.
func TestResolveRefsKeepsUuids(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("parent", "", "")
	if err := flags.Parse([]string{"--parent", "parent-uuid"}); err != nil {
		t.Fatal(err)
	}

	ids, err := resolveRefs(context.Background(), nil, domain.NOTE, []string{"a", "b"}, flags)
	if err != nil || strings.Join(ids, ",") != "a,b" {
		t.Fatalf("ids = %v, %v", ids, err)
	}
	if got, _ := flags.GetString("parent"); got != "parent-uuid" {
		t.Errorf("parent = %q", got)
	}
}
