package cmdtrash

import (
	"fmt"
	"math"
	"time"

	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/listtrash"
	"main.go/app/shared/services/screen"
	"main.go/app/shared/services/style"
)

// Trash lists what was removed, each removal once with how many items went with it, and when it is purged.
func Trash(cobra *cobra.Command, args []string) error {
	var asJSON bool
	if err := plumcliutilsargs.GetBoolFlags(cobra, map[string]*bool{"json": &asJSON}); err != nil {
		return err
	}

	process := listtrash.Process(cobra.Context(), postgrey.GetDb())
	process.Run()
	if process.HasErr() {
		return process.GetErr()
	}

	if asJSON {
		screen.PrintJSON(process.GetResult())
		return nil
	}

	var user domain.User
	if err := postgrey.GetDb().NewSelect().Model(&user).Where("id = ?", domain.UserID(cobra.Context())).Scan(cobra.Context()); err != nil {
		return err
	}
	screen.Print(output(process.GetResult(), user.TrashRetentionDays, time.Now()))
	return nil
}

func output(trash []listtrash.Trashed, retention int, now time.Time) string {
	header := style.Header("Trash", "notes restore <uuid> brings an item back", info(len(trash), retention)) + "\n"
	if len(trash) == 0 {
		return header + style.Empty().Render("The trash is empty")
	}

	cells := make([][]string, len(trash))
	for i, item := range trash {
		title := string(item.Title)
		if title == "" {
			title = "untitled"
		}
		if item.Inside > 0 {
			title += fmt.Sprintf(" (+%d inside)", item.Inside)
		}
		cells[i] = []string{string(item.Kind), title, "removed " + item.DeletedAt.Format("2 Jan 2006"), purge(item.DeletedAt, retention, now), item.ID}
	}
	return header + style.Table(cells, -1, 1)
}

func info(removals, retention int) string {
	count := fmt.Sprintf("%d removals", removals)
	if removals == 1 {
		count = "1 removal"
	}
	if retention <= 0 {
		return count + " · never purged"
	}
	return fmt.Sprintf("%s · purged after %d days", count, retention)
}

// purge says when purgeitems (run at every start) deletes the item for good.
func purge(deletedAt time.Time, retention int, now time.Time) string {
	if retention <= 0 {
		return "kept"
	}
	days := int(math.Ceil(deletedAt.AddDate(0, 0, retention).Sub(now).Hours() / 24))
	switch {
	case days <= 0:
		return "purged at next start"
	case days == 1:
		return "purged in 1 day"
	default:
		return fmt.Sprintf("purged in %d days", days)
	}
}
