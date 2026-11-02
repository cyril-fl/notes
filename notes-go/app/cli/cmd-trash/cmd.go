package cmdtrash

import (
	"fmt"
	"math"
	"strings"
	"text/tabwriter"
	"time"

	"charm.land/lipgloss/v2"
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/listtrash"
	"main.go/app/shared/services/screen"
	"main.go/config"
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

	screen.Print(output(process.GetResult(), config.Current.Trash.RetentionDays, time.Now()))
	return nil
}

func output(trash []listtrash.Trashed, retention int, now time.Time) string {
	var out strings.Builder
	out.WriteString(lipgloss.NewStyle().Bold(true).Render("Trash") + "\n")
	out.WriteString(lipgloss.NewStyle().Faint(true).Render("notes restore <uuid> brings an item back") + "\n")
	out.WriteString(lipgloss.NewStyle().Faint(true).Render(info(len(trash), retention)) + "\n\n")

	if len(trash) == 0 {
		out.WriteString(lipgloss.NewStyle().Faint(true).Italic(true).Render("The trash is empty"))
		return out.String()
	}

	table := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	for _, item := range trash {
		title := item.Title
		if title == "" {
			title = "untitled"
		}
		if item.Inside > 0 {
			title += fmt.Sprintf(" (+%d inside)", item.Inside)
		}
		fmt.Fprintf(table, "%s\t%s\tremoved %s\t%s\t%s\n", item.Kind, title, item.DeletedAt.Format("2 Jan 2006"), purge(item.DeletedAt, retention, now), item.ID)
	}
	table.Flush() // writes to a strings.Builder: cannot fail
	return out.String()
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
