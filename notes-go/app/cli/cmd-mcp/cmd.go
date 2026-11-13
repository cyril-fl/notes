// Package cmdmcp serves the account's notes to an agent over MCP, a third front end on the
// shared processes next to the CLI and the API.
package cmdmcp

import (
	"context"
	"errors"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/counttags"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listrecent"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listtrash"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/moveitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/resolveitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/restoreitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/searchnotes"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var errNotFound = errors.New("no note or folder of this name")

// item is a note or a folder as an agent needs it: no timestamps, no model, nothing to read twice.
type item struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

// note is what the read and the writes answer, the content left out when the agent did not ask for it.
type note struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Path     string   `json:"path"`
	Content  string   `json:"content,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Readonly bool     `json:"readonly,omitempty"`
}

type searchArgs struct {
	Text string `json:"text" jsonschema:"the text to find in the title or the content of the notes"`
}

type noteArgs struct {
	Note string `json:"note" jsonschema:"the note: its uuid, or its path from the root (/Folder/Note)"`
}

type folderArgs struct {
	Folder string `json:"folder,omitempty" jsonschema:"the folder: its uuid, or its path from the root (/Folder); the root itself when empty"`
}

type createArgs struct {
	Content string   `json:"content" jsonschema:"the markdown content; its first heading becomes the title and its #hashtags become tags"`
	Folder  string   `json:"folder,omitempty" jsonschema:"the folder to create the note in: its uuid, or its path from the root (/Folder); the root itself when empty"`
	Tags    []string `json:"tags,omitempty" jsonschema:"tags to add to the content, without the #"`
}

type editArgs struct {
	Note    string   `json:"note" jsonschema:"the note to rewrite: its uuid, or its path from the root (/Folder/Note)"`
	Content string   `json:"content" jsonschema:"the new markdown content; it replaces the whole note"`
	Tags    []string `json:"tags,omitempty" jsonschema:"tags to add to the content, without the #"`
}

type createFolderArgs struct {
	Title  string `json:"title" jsonschema:"the new folder's title"`
	Folder string `json:"folder,omitempty" jsonschema:"the folder to create it in: its uuid, or its path from the root (/Folder); the root itself when empty"`
}

type renameFolderArgs struct {
	Folder string `json:"folder" jsonschema:"the folder to rename: its uuid, or its path from the root (/Folder)"`
	Title  string `json:"title" jsonschema:"its new title"`
}

type moveArgs struct {
	Item   string `json:"item" jsonschema:"the note or folder to move: its uuid, or its path from the root (/Folder/Note)"`
	Folder string `json:"folder,omitempty" jsonschema:"the folder to move it into: its uuid, or its path from the root (/Folder); the root itself when empty"`
}

type trashedArgs struct {
	Item string `json:"item" jsonschema:"the uuid of the trashed note or folder, as list_trash gives it"`
}

type tagArgs struct {
	Tag string `json:"tag" jsonschema:"the tag, without the #"`
}

// Serve speaks MCP on stdin and stdout: nothing else may be printed there.
func Serve(cobra *cobra.Command, _ []string) error {
	server := mcp.NewServer(&mcp.Implementation{Name: "notes", Title: "Notes", Version: cobra.Root().Version}, nil)
	// The account comes from the CLI session, in the command's context: the handlers act for it.
	AddTools(server, cobra.Context())
	return server.Run(cobra.Context(), &mcp.StdioTransport{})
}

// AddTools gives the server the tools acting for the account of ctx.
func AddTools(server *mcp.Server, ctx context.Context) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_notes",
		Description: "Find the notes whose title or content contains a text, with the line where it was found.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args searchArgs) (*mcp.CallToolResult, []searchnotes.Match, error) {
		process := searchnotes.Process(ctx, postgrey.GetDb(), args.Text)
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}
		return nil, process.GetResult(), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "recent_notes",
		Description: "List the notes changed most recently, newest first.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, []listrecent.Recent, error) {
		process := listrecent.Process(ctx, postgrey.GetDb())
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}
		return nil, process.GetResult(), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_note",
		Description: "Read a note: its content, its path and its tags.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args noteArgs) (*mcp.CallToolResult, note, error) {
		id, err := resolve(ctx, domain.NOTE, args.Note)
		if err != nil {
			return nil, note{}, err
		}
		read, err := readNote(ctx, id)
		return nil, read, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_folder",
		Description: "List the notes and folders directly inside a folder.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args folderArgs) (*mcp.CallToolResult, []item, error) {
		parentID, err := resolveFolder(ctx, args.Folder)
		if err != nil {
			return nil, nil, err
		}

		process := listitems.Process(ctx, postgrey.GetDb(), parentID)
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}

		return nil, made(process.GetResult()...), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_note",
		Description: "Create a note from markdown content.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args createArgs) (*mcp.CallToolResult, note, error) {
		parentID, err := resolveFolder(ctx, args.Folder)
		if err != nil {
			return nil, note{}, err
		}

		process := createnote.Process(ctx, postgrey.GetDb(), createnote.Input{Content: args.Content, ParentID: parentID, Tags: args.Tags})
		process.Run()
		if process.HasErr() {
			return nil, note{}, process.GetErr()
		}
		// Read back: the process answers the item alone, without the path nor the tags it parsed.
		created, err := readNote(ctx, process.GetResult().ID)
		return nil, created, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "edit_note",
		Description: "Replace the content of a note. Read it first: nothing is kept of the old content. A locked note refuses.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args editArgs) (*mcp.CallToolResult, note, error) {
		id, err := resolve(ctx, domain.NOTE, args.Note)
		if err != nil {
			return nil, note{}, err
		}

		process := editnote.Process(ctx, postgrey.GetDb(), editnote.Input{ID: id, Content: args.Content, Tags: args.Tags})
		process.Run()
		if process.HasErr() {
			return nil, note{}, process.GetErr()
		}
		edited, err := readNote(ctx, id)
		return nil, edited, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_folder",
		Description: "Create a folder.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args createFolderArgs) (*mcp.CallToolResult, item, error) {
		parentID, err := resolveFolder(ctx, args.Folder)
		if err != nil {
			return nil, item{}, err
		}

		process := createfolder.Process(ctx, postgrey.GetDb(), createfolder.Input{Title: args.Title, ParentID: parentID})
		process.Run()
		if process.HasErr() {
			return nil, item{}, process.GetErr()
		}
		return nil, one(process.GetResult()), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rename_folder",
		Description: "Change the title of a folder.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args renameFolderArgs) (*mcp.CallToolResult, item, error) {
		id, err := resolve(ctx, domain.FOLDER, args.Folder)
		if err != nil {
			return nil, item{}, err
		}

		process := editfolder.Process(ctx, postgrey.GetDb(), editfolder.Input{ID: id, Title: args.Title})
		process.Run()
		if process.HasErr() {
			return nil, item{}, process.GetErr()
		}
		return nil, one(process.GetResult()), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "move_item",
		Description: "Move a note or a folder into another folder. A folder takes its content with it.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args moveArgs) (*mcp.CallToolResult, []item, error) {
		id, err := resolve(ctx, "", args.Item)
		if err != nil {
			return nil, nil, err
		}
		parentID, err := resolveFolder(ctx, args.Folder)
		if err != nil {
			return nil, nil, err
		}

		process := moveitems.Process(ctx, postgrey.GetDb(), moveitems.Input{IDs: []string{id}, ParentID: parentID})
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}
		return nil, made(process.GetResult()...), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_note",
		Description: "Move a note to the trash, where restore_item can take it back.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args noteArgs) (*mcp.CallToolResult, []item, error) {
		return deleteItem(ctx, domain.NOTE, args.Note)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_folder",
		Description: "Move a folder and everything inside it to the trash, where restore_item can take it back.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args folderArgs) (*mcp.CallToolResult, []item, error) {
		return deleteItem(ctx, domain.FOLDER, args.Folder)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_trash",
		Description: "List what is in the trash, last removed first, with the uuid restore_item takes.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, []listtrash.Trashed, error) {
		process := listtrash.Process(ctx, postgrey.GetDb())
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}
		return nil, process.GetResult(), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "restore_item",
		Description: "Take a note or a folder out of the trash, with everything removed at the same time as it.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args trashedArgs) (*mcp.CallToolResult, []item, error) {
		// No resolve: a trashed item has no path to look up.
		process := restoreitems.Process(ctx, postgrey.GetDb(), []string{args.Item})
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}
		return nil, made(process.GetResult()...), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tags",
		Description: "List the account's tags, most used first, with how many notes carry each.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, []counttags.TagCount, error) {
		process := counttags.Process(ctx, postgrey.GetDb())
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}
		return nil, process.GetResult(), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_tag",
		Description: "List the notes carrying a tag, with their path. Tags of a note are set by create_note and edit_note.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args tagArgs) (*mcp.CallToolResult, readtag.Result, error) {
		process := readtag.Process(ctx, postgrey.GetDb(), args.Tag)
		process.Run()
		if process.HasErr() {
			return nil, readtag.Result{}, process.GetErr()
		}
		return nil, process.GetResult(), nil
	})
}

// deleteItem trashes one note or folder, named by uuid or by path.
func deleteItem(ctx context.Context, kind domain.Kind, ref string) (*mcp.CallToolResult, []item, error) {
	id, err := resolve(ctx, kind, ref)
	if err != nil {
		return nil, nil, err
	}

	process := deleteitems.Process(ctx, postgrey.GetDb(), kind, []string{id})
	process.Run()
	if process.HasErr() {
		return nil, nil, process.GetErr()
	}
	return nil, made(process.GetResult()...), nil
}

// one turns an item the processes answer into what the agent reads.
func one(found domain.Item) item {
	return item{ID: found.ID, Kind: string(found.Kind), Title: string(found.Title)}
}

func made(found ...domain.Item) []item {
	items := make([]item, 0, len(found))
	for _, found := range found {
		items = append(items, one(found))
	}
	return items
}

// resolve turns a path into the uuid the processes take; a uuid goes through untouched.
func resolve(ctx context.Context, kind domain.Kind, ref string) (string, error) {
	process := resolveitem.Process(ctx, postgrey.GetDb(), kind, []string{ref})
	process.Run()
	if process.HasErr() {
		return "", process.GetErr()
	}
	if len(process.GetResult()) == 0 {
		return "", errNotFound
	}
	return process.GetResult()[0], nil
}

// resolveFolder keeps "" as "": the processes read it as the root.
func resolveFolder(ctx context.Context, ref string) (string, error) {
	if ref == "" {
		return "", nil
	}
	return resolve(ctx, domain.FOLDER, ref)
}

func readNote(ctx context.Context, id string) (note, error) {
	process := readitem.Process(ctx, postgrey.GetDb(), domain.NOTE, id)
	process.Run()
	if process.HasErr() {
		return note{}, process.GetErr()
	}

	result := process.GetResult()
	return note{
		ID:       result.Item.ID,
		Title:    string(result.Item.Title),
		Path:     result.Path,
		Content:  result.Content,
		Tags:     result.Tags,
		Readonly: result.Readonly,
	}, nil
}
