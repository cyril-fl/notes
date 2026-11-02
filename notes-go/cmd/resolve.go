package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/uptrace/bun"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/resolveitem"
)

// stdinContent is the --content value meaning "read the content from stdin".
const stdinContent = "-"

var errEmptyStdin = errors.New("no content on stdin")

// resolved runs a command with paths turned into uuids, in its args (of this kind, any when empty)
// and in --parent, and --content - replaced by stdin. Commands and mills only ever see uuids and values.
func resolved(kind domain.Kind, run func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cobra *cobra.Command, args []string) error {
		ids, err := resolveRefs(cobra.Context(), postgrey.GetDb(), kind, args, cobra.Flags())
		if err != nil {
			return err
		}
		if err := readContent(cobra.Flags(), os.Stdin); err != nil {
			return err
		}
		return run(cobra, ids)
	}
}

func resolveRefs(ctx context.Context, db bun.IDB, kind domain.Kind, args []string, flags *pflag.FlagSet) ([]string, error) {
	process := resolveitem.Process(ctx, db, kind, args)
	process.Run()
	if process.HasErr() {
		return nil, process.GetErr()
	}

	parent := flags.Lookup("parent")
	if parent == nil || !parent.Changed {
		return process.GetResult(), nil
	}

	folder := resolveitem.Process(ctx, db, domain.FOLDER, []string{parent.Value.String()})
	folder.Run()
	if folder.HasErr() {
		return nil, folder.GetErr()
	}
	return process.GetResult(), flags.Set("parent", folder.GetResult()[0])
}

// readContent fails on a blank stdin: without content, add note would open a form instead.
func readContent(flags *pflag.FlagSet, stdin io.Reader) error {
	content := flags.Lookup("content")
	if content == nil || !content.Changed || content.Value.String() != stdinContent {
		return nil
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(data)) == "" {
		return errEmptyStdin
	}
	return flags.Set("content", string(data))
}
