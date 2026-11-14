// Package cmdmcp serves the account's notes to an agent over MCP, a third front end on the
// shared processes next to the CLI and the API.
package cmdmcp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"testing/fstest"

	"github.com/cyril-fl/notes/notes-go/app/shared/domain"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/counttags"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfile"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/createnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deleteitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/deletetag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editfolder"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/editnote"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/edittag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/exportnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/importnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listbrokenlinks"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listfiles"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listrecent"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listtrash"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/listversions"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/moveitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readfile"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readneighbors"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readtag"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/readversion"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/resolveitem"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/restoreitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/searchnotes"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var (
	errNotFound = errors.New("no note or folder of this name")
	errBadPath  = errors.New("not the relative path of a .md file")
)

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

type tagsArgs struct {
	Tag string `json:"tag,omitempty" jsonschema:"the tag, without the #; the list of tags when empty"`
}

// tags is the account's tags with their count, or one tag with the notes carrying it.
type tags struct {
	Tags  []counttags.TagCount `json:"tags,omitempty"`
	Tag   *domain.Tag          `json:"tag,omitempty"`
	Notes []readtag.Note       `json:"notes,omitempty"`
}

type renameTagArgs struct {
	Tag string `json:"tag" jsonschema:"the tag to rename, without the #"`
	To  string `json:"to" jsonschema:"its new name, without the #"`
}

type historyArgs struct {
	Note    string `json:"note" jsonschema:"the note: its uuid, or its path from the root (/Folder/Note)"`
	Version string `json:"version,omitempty" jsonschema:"the uuid of a version to read, as the list gives it; the list of versions when empty"`
}

// history is the list of versions, their texts left out, or the one version asked with its text.
type history struct {
	Versions []domain.NoteVersion `json:"versions,omitempty"`
	Version  *domain.NoteVersion  `json:"version,omitempty"`
}

type linksArgs struct {
	Note string `json:"note,omitempty" jsonschema:"the note: its uuid, or its path from the root (/Folder/Note); the broken links of every note when empty"`
}

// links joins a note to the others, the note itself left out (read_note reads it), or lists the broken links of every note.
type links struct {
	Links     []readneighbors.Linked   `json:"links,omitempty"`
	Backlinks []readneighbors.Linked   `json:"backlinks,omitempty"`
	Neighbors []readneighbors.Neighbor `json:"neighbors,omitempty"`
	Broken    []listbrokenlinks.Link   `json:"broken,omitempty"`
}

type fileArgs struct {
	File string `json:"file,omitempty" jsonschema:"the uuid of a file, as a note cites it (/api/files/<uuid>) or the list gives it; the list of files when empty"`
}

type attachArgs struct {
	Name string `json:"name" jsonschema:"the file name, with its extension"`
	Data string `json:"data" jsonschema:"the bytes of the file in base64: an image (PNG, JPEG, GIF, WebP), a PDF, a zip, an audio or a video file"`
}

// attached is the stored file and the markdown a note cites it with.
type attached struct {
	ID       string `json:"id"`
	Markdown string `json:"markdown"`
}

type exportArgs struct {
	Folder string   `json:"folder,omitempty" jsonschema:"the folder to export: its uuid, or its path from the root (/Folder); everything when empty"`
	Items  []string `json:"items,omitempty" jsonschema:"notes or folders to export instead of a folder: their uuid, or their path from the root"`
}

// exported is a note as a markdown file, or a file it cites, whose bytes read_file gives.
type exported struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	File    string `json:"file,omitempty"`
}

type importArgs struct {
	Files  []importFile `json:"files" jsonschema:"the markdown files; their directories become folders"`
	Folder string       `json:"folder,omitempty" jsonschema:"the folder to import into: its uuid, or its path from the root (/Folder); the root itself when empty"`
}

type importFile struct {
	Path    string `json:"path" jsonschema:"the file path, relative (Folder/Note.md)"`
	Content string `json:"content" jsonschema:"the markdown content"`
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

		process := listitems.Process(ctx, postgrey.GetDb(), parentID, "")
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
		Name:        "tags",
		Description: "List the notes carrying a tag, with their path. Without a tag, the account's tags, most used first. Tags of a note are set by create_note and edit_note.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args tagsArgs) (*mcp.CallToolResult, tags, error) {
		if args.Tag == "" {
			process := counttags.Process(ctx, postgrey.GetDb())
			process.Run()
			return nil, tags{Tags: process.GetResult()}, process.GetErr()
		}

		process := readtag.Process(ctx, postgrey.GetDb(), args.Tag)
		process.Run()
		if process.HasErr() {
			return nil, tags{}, process.GetErr()
		}
		result := process.GetResult()
		return nil, tags{Tag: &result.Tag, Notes: result.Notes}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "rename_tag",
		Description: "Rename a tag on every note carrying it, in their text too.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args renameTagArgs) (*mcp.CallToolResult, domain.Tag, error) {
		process := edittag.Process(ctx, postgrey.GetDb(), args.Tag, args.To)
		process.Run()
		if process.HasErr() {
			return nil, domain.Tag{}, process.GetErr()
		}
		return nil, process.GetResult(), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_tag",
		Description: "Remove a tag from every note carrying it, in their text too. A note holding nothing but the tag goes to the trash.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args tagArgs) (*mcp.CallToolResult, []domain.Tag, error) {
		process := deletetag.Process(ctx, postgrey.GetDb(), []string{args.Tag})
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}
		return nil, process.GetResult(), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "note_history",
		Description: "List the earlier texts of a note, newest first, or read one: what edit_note replaced.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args historyArgs) (*mcp.CallToolResult, history, error) {
		id, err := resolve(ctx, domain.NOTE, args.Note)
		if err != nil {
			return nil, history{}, err
		}

		if args.Version == "" {
			process := listversions.Process(ctx, postgrey.GetDb(), id)
			process.Run()
			return nil, history{Versions: process.GetResult()}, process.GetErr()
		}
		process := readversion.Process(ctx, postgrey.GetDb(), id, args.Version)
		process.Run()
		if process.HasErr() {
			return nil, history{}, process.GetErr()
		}
		version := process.GetResult()
		return nil, history{Version: &version}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "note_links",
		Description: "List the notes a note links to, those linking to it, and those sharing its tags. Without a note, the broken links of every note.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args linksArgs) (*mcp.CallToolResult, links, error) {
		if args.Note == "" {
			process := listbrokenlinks.Process(ctx, postgrey.GetDb())
			process.Run()
			return nil, links{Broken: process.GetResult()}, process.GetErr()
		}

		id, err := resolve(ctx, domain.NOTE, args.Note)
		if err != nil {
			return nil, links{}, err
		}
		process := readneighbors.Process(ctx, postgrey.GetDb(), id)
		process.Run()
		if process.HasErr() {
			return nil, links{}, process.GetErr()
		}
		result := process.GetResult()
		return nil, links{Links: result.Links, Backlinks: result.Backlinks, Neighbors: result.Neighbors}, nil
	})

	// Out is any: a read answers the file as content (an image the agent sees), not as JSON.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_file",
		Description: "Read an image or attachment a note cites. Without a file, list the account's files, largest first.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args fileArgs) (*mcp.CallToolResult, any, error) {
		if args.File == "" {
			process := listfiles.Process(ctx, postgrey.GetDb())
			process.Run()
			return nil, process.GetResult(), process.GetErr()
		}

		// Silo is opened here, not at start: without S3, every other tool still works.
		store, err := silo.New(ctx)
		if err != nil {
			return nil, nil, err
		}
		process := readfile.Process(ctx, postgrey.GetDb(), store, args.File)
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}
		body := process.GetBody()
		defer body.Close()
		data, err := io.ReadAll(body)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{fileContent(process.GetResult(), data)}}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "attach_file",
		Description: "Store an image or attachment and answer the markdown citing it, to put in a note with create_note or edit_note.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args attachArgs) (*mcp.CallToolResult, attached, error) {
		data, err := base64.StdEncoding.DecodeString(args.Data)
		if err != nil {
			return nil, attached{}, fmt.Errorf("data is not base64: %w", err)
		}
		store, err := silo.New(ctx)
		if err != nil {
			return nil, attached{}, err
		}

		process := createfile.Process(ctx, postgrey.GetDb(), store, createfile.Input{Name: args.Name, Data: data})
		process.Run()
		if process.HasErr() {
			return nil, attached{}, process.GetErr()
		}
		file := process.GetResult()
		return nil, attached{ID: file.ID, Markdown: citation(file, args.Name)}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_notes",
		Description: "Export notes as markdown files, [[links]] written back, their folders as directories.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args exportArgs) (*mcp.CallToolResult, []exported, error) {
		folderID, err := resolveFolder(ctx, args.Folder)
		if err != nil {
			return nil, nil, err
		}
		ids := make([]string, 0, len(args.Items))
		for _, ref := range args.Items {
			id, err := resolve(ctx, "", ref)
			if err != nil {
				return nil, nil, err
			}
			ids = append(ids, id)
		}

		process := exportnotes.Process(ctx, postgrey.GetDb(), folderID, ids...)
		process.Run()
		if process.HasErr() {
			return nil, nil, process.GetErr()
		}
		return nil, exportedFiles(process.GetResult()), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "import_notes",
		Description: "Import markdown files as notes, their directories as folders, all or nothing: a folder name already taken refuses.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args importArgs) (*mcp.CallToolResult, importnotes.Result, error) {
		parentID, err := resolveFolder(ctx, args.Folder)
		if err != nil {
			return nil, importnotes.Result{}, err
		}
		dir, err := importDir(args.Files)
		if err != nil {
			return nil, importnotes.Result{}, err
		}

		// Silo even for text: a note linking to another whose bytes look like a PDF stores it as a file.
		store, err := silo.New(ctx)
		if err != nil {
			return nil, importnotes.Result{}, err
		}
		process := importnotes.Process(ctx, postgrey.GetDb(), store, importnotes.Input{Dir: dir, ParentID: parentID})
		process.Run()
		return nil, process.GetResult(), process.GetErr()
	})
}

// fileContent shows an image or plays an audio to the agent; anything else comes as a resource of its type.
func fileContent(file domain.File, data []byte) mcp.Content {
	switch {
	case strings.HasPrefix(file.ContentType, "image/"):
		return &mcp.ImageContent{Data: data, MIMEType: file.ContentType}
	case strings.HasPrefix(file.ContentType, "audio/"):
		return &mcp.AudioContent{Data: data, MIMEType: file.ContentType}
	}
	return &mcp.EmbeddedResource{Resource: &mcp.ResourceContents{URI: domain.FileURL(file.ID), MIMEType: file.ContentType, Blob: data}}
}

// citation is the markdown of a file in a note, under the name given (the stored one is sealed): an image shown, anything else linked.
func citation(file domain.File, name string) string {
	link := "[" + name + "](" + domain.FileURL(file.ID) + ")"
	if strings.HasPrefix(file.ContentType, "image/") {
		return "!" + link
	}
	return link
}

// exportedFiles leaves the directories out, implied by the paths of what they hold, and names a file by the id read_file takes.
func exportedFiles(files []exportnotes.File) []exported {
	out := []exported{}
	for _, file := range files {
		switch {
		case file.Dir:
		case file.Key != "":
			// exportnotes names it <id><ext> in its FilesDir.
			out = append(out, exported{Path: file.Path, File: strings.TrimSuffix(path.Base(file.Path), path.Ext(file.Path))})
		default:
			out = append(out, exported{Path: file.Path, Content: file.Content})
		}
	}
	return out
}

// importDir makes a directory of the notes, refusing a path that would climb out of it: an agent sends text, not the
// bytes of an image.
func importDir(files []importFile) (fstest.MapFS, error) {
	dir := fstest.MapFS{}
	for _, file := range files {
		if !fs.ValidPath(file.Path) || !strings.EqualFold(path.Ext(file.Path), ".md") {
			return nil, fmt.Errorf("%w: %q", errBadPath, file.Path)
		}
		dir[file.Path] = &fstest.MapFile{Data: []byte(file.Content)}
	}
	return dir, nil
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
