package cmdfinder

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	cmdcreatefolder "main.go/app/cli/cmd-create-folder"
	cmdcreatenote "main.go/app/cli/cmd-create-note"
	cmdeditfolder "main.go/app/cli/cmd-edit-folder"
	cmdeditnote "main.go/app/cli/cmd-edit-note"
	cmdreadfolder "main.go/app/cli/cmd-read-folder"
	cmdreadnote "main.go/app/cli/cmd-read-note"
	"main.go/app/shared/domain"
	"main.go/app/shared/services/huhform"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/deleteitems"
	"main.go/app/shared/services/process/listitems"
	"main.go/app/shared/services/process/moveitems"
	"main.go/app/shared/services/process/readitem"
	"main.go/app/shared/services/screen"
)

// Finder browses folders and notes from the given folder (the root by default).
// Forms and the note view run between two finder screens, which then reopens where it was.
func Finder(cobra *cobra.Command, args []string) error {
	ctx, id := cobra.Context(), ""
	if len(args) == 1 {
		id = args[0]
	}
	load := func(id string) (readitem.Result, error) { return loadFolder(ctx, id) }

	folder, err := load(id)
	if err != nil {
		return err
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		screen.Print(cmdreadfolder.Output(folder, folder.Children, -1, cmdreadfolder.EmptyFolder))
		return nil
	}

	cursor, status, failed := 0, "", false
	var clipboard *domain.Item
	for {
		final, err := tea.NewProgram(newModel(load, folder, cursor, clipboard, status, failed)).Run()
		if err != nil {
			return err
		}
		m := final.(*model)
		if m.err != nil || m.action.kind == actionQuit {
			return m.err
		}

		clipboard, status, failed = m.clipboard, "", false
		if m.action.kind == actionMove {
			// A refused move (into its own subfolder…) is shown in the finder, which stays open.
			if err := move(ctx, m.action.item, m.folder.Item.ID); err != nil {
				status, failed = err.Error(), true
			} else {
				clipboard, status = nil, fmt.Sprintf("Moved %q here", untitled(m.action.item.Title))
			}
		} else if err := do(ctx, m.action, m.folder.Item.ID); err != nil && !errors.Is(err, huh.ErrUserAborted) {
			// An aborted form comes back to the finder unchanged.
			return err
		}
		if folder, err = load(m.folder.Item.ID); err != nil {
			return err
		}
		cursor = m.cursor
	}
}

// move puts the item in the folder; moveitems refuses a folder going into itself or its subfolders.
func move(ctx context.Context, item domain.Item, folderID string) error {
	process := moveitems.Process(ctx, postgrey.GetDb(), moveitems.Input{IDs: []string{item.ID}, ParentID: folderID})
	process.Run()
	return process.GetErr()
}

func do(ctx context.Context, action action, folderID string) error {
	item := action.item
	switch action.kind {
	case actionOpen:
		return cmdreadnote.View(ctx, item.ID)
	case actionEdit:
		if item.Kind == domain.FOLDER {
			return cmdeditfolder.Interactive(ctx, item.ID)
		}
		return cmdeditnote.Interactive(ctx, item.ID)
	case actionCreate:
		return create(ctx, folderID)
	case actionDelete:
		return remove(ctx, item)
	}
	return nil
}

// create asks what to make, then opens its add form inside the current folder.
func create(ctx context.Context, folderID string) error {
	kind := domain.NOTE
	choice := huh.NewSelect[domain.Kind]().Title("New").Options(huh.NewOption("Note", domain.NOTE), huh.NewOption("Folder", domain.FOLDER)).Value(&kind)
	if err := huhform.New(choice).Run(); err != nil {
		return err
	}

	if kind == domain.FOLDER {
		return cmdcreatefolder.Interactive(ctx, folderID)
	}
	return cmdcreatenote.Interactive(ctx, folderID)
}

// remove moves the item to the trash once confirmed; a folder takes its content with it.
func remove(ctx context.Context, item domain.Item) error {
	title := item.Title
	if title == "" {
		title = "untitled"
	}
	question := fmt.Sprintf("Move %q to the trash?", title)
	if item.Kind == domain.FOLDER {
		question = fmt.Sprintf("Move %q and its content to the trash?", title)
	}

	confirmed := false
	if err := huhform.New(huh.NewConfirm().Title(question).Value(&confirmed)).Run(); err != nil || !confirmed {
		return err
	}

	process := deleteitems.Process(ctx, postgrey.GetDb(), item.Kind, []string{item.ID})
	process.Run()
	return process.GetErr()
}

// loadFolder reads a folder, or the root as a folder titled Notes at path /.
func loadFolder(ctx context.Context, id string) (readitem.Result, error) {
	if id != "" {
		process := readitem.Process(ctx, postgrey.GetDb(), domain.FOLDER, id)
		process.Run()
		return process.GetResult(), process.GetErr()
	}

	process := listitems.Process(ctx, postgrey.GetDb(), "")
	process.Run()
	root := readitem.Result{Item: domain.Item{Kind: domain.FOLDER, Title: "Notes"}, Path: "/", Children: process.GetResult()}
	return root, process.GetErr()
}
