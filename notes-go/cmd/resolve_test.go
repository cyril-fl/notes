package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/spf13/pflag"
)

func contentFlags(t *testing.T, args ...string) *pflag.FlagSet {
	t.Helper()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("content", "", "")
	if err := flags.Parse(args); err != nil {
		t.Fatal(err)
	}
	return flags
}

func TestReadContent(t *testing.T) {
	tests := map[string]struct {
		args    []string
		stdin   string
		want    string
		wantErr error
	}{
		"from stdin":  {[]string{"--content", "-"}, "# hello\n", "# hello\n", nil},
		"blank stdin": {[]string{"--content", "-"}, " \n", "", errEmptyStdin},
		"plain value": {[]string{"--content", "text"}, "ignored", "text", nil},
		"no flag":     {nil, "ignored", "", nil},
	}

	for name, test := range tests {
		flags := contentFlags(t, test.args...)
		err := readContent(flags, strings.NewReader(test.stdin))
		if !errors.Is(err, test.wantErr) {
			t.Errorf("%s: err = %v, want %v", name, err, test.wantErr)
			continue
		}
		if test.wantErr != nil {
			continue
		}
		if got, _ := flags.GetString("content"); got != test.want {
			t.Errorf("%s: content = %q, want %q", name, got, test.want)
		}
	}
}

// A command without --content leaves stdin alone.
func TestReadContentNoFlag(t *testing.T) {
	if err := readContent(pflag.NewFlagSet("test", pflag.ContinueOnError), strings.NewReader("")); err != nil {
		t.Fatal(err)
	}
}

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
