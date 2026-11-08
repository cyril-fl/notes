package cmdedittag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/huhform"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/edittag"
	"github.com/spf13/cobra"
)

var errRequired = errors.New("required")

func Edit(cobra *cobra.Command, args []string) error {
	_, err := rename(cobra.Context(), args[0], args[1])
	return err
}

// Interactive renames the tag through a form pre-filled with its value, and returns the new value.
func Interactive(ctx context.Context, value string) (string, error) {
	to := value
	if err := huhform.New(huh.NewInput().Title("Tag").Value(&to).Validate(notBlank)).Run(); err != nil {
		return value, err
	}
	return rename(ctx, value, to)
}

func rename(ctx context.Context, from string, to string) (string, error) {
	process := edittag.Process(ctx, postgrey.GetDb(), from, to)
	process.Run()

	if process.HasErr() {
		return from, process.GetErr()
	}

	fmt.Printf("Successfully renamed %q to %q\n", from, process.GetResult().Value)
	return string(process.GetResult().Value), nil
}

func notBlank(value string) error {
	if strings.TrimSpace(value) == "" {
		return errRequired
	}

	return nil
}
