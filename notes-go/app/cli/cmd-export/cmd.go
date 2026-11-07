package cmdexport

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/exportnotes"
	"main.go/app/shared/services/process/resolveitem"
	"main.go/app/shared/services/silo"
)

var ErrNotEmpty = errors.New("the directory is not empty")

// Export writes the notes of a folder (args[1], everything without it) as .md files into the directory args[0].
func Export(cobra *cobra.Command, args []string) error {
	dir, folderID := args[0], ""
	if len(args) > 1 {
		resolve := resolveitem.Process(cobra.Context(), postgrey.GetDb(), domain.FOLDER, args[1:])
		resolve.Run()
		if resolve.HasErr() {
			return resolve.GetErr()
		}
		folderID = resolve.GetResult()[0]
	}

	// Never write over files: an export goes into a new or empty directory.
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		return fmt.Errorf("%w: %s", ErrNotEmpty, dir)
	}

	store, err := silo.New(cobra.Context())
	if err != nil {
		return err
	}
	process := exportnotes.Process(cobra.Context(), postgrey.GetDb(), folderID)
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	notes := 0
	for _, file := range process.GetResult() {
		target := filepath.Join(dir, filepath.FromSlash(file.Path))
		if file.Dir {
			if err := os.Mkdir(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if file.Key != "" {
			if err := copyImage(cobra, store, file.Key, target); err != nil {
				return err
			}
			continue
		}
		if err := os.WriteFile(target, []byte(file.Content), 0o644); err != nil {
			return err
		}
		notes++
	}

	fmt.Printf("Successfully exported %d note(s) to %s\n", notes, dir)
	return nil
}

// copyImage writes the bytes of an image from Silo into target.
func copyImage(cobra *cobra.Command, store silo.Store, key, target string) error {
	body, err := store.Get(cobra.Context(), key)
	if err != nil {
		return err
	}
	defer body.Close()
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, body); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
