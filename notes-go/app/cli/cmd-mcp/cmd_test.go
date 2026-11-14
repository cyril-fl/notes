package cmdmcp

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/exportnotes"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestAddTools catches a schema the SDK cannot infer from the arguments: the tools would be
// announced broken, and only an agent calling them would find out.
func TestAddTools(t *testing.T) {
	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "notes"}, nil)
	AddTools(server, ctx)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	session, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	client, err := mcp.NewClient(&mcp.Implementation{Name: "test"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	result, err := client.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
		if tool.InputSchema == nil || tool.Description == "" {
			t.Errorf("tool %q: schema %v, description %q", tool.Name, tool.InputSchema, tool.Description)
		}
	}
	for _, want := range []string{"search_notes", "recent_notes", "read_note", "list_folder", "create_note", "edit_note", "create_folder", "rename_folder", "move_item", "delete_note", "delete_folder", "list_trash", "restore_item", "tags", "rename_tag", "delete_tag", "note_history", "note_links", "read_file", "attach_file", "export_notes", "import_notes"} {
		if !slices.Contains(names, want) {
			t.Errorf("tools = %v, missing %q", names, want)
		}
	}
}

// TestImportDir refuses what would climb out of the import or is no note.
func TestImportDir(t *testing.T) {
	dir, err := importDir([]importFile{{Path: "Folder/Note.md", Content: "# Note"}})
	if err != nil || string(dir["Folder/Note.md"].Data) != "# Note" {
		t.Fatalf("importDir = %v, %v", dir, err)
	}
	for _, bad := range []string{"../Note.md", "/Note.md", "Folder/../Note.md", "image.png", ""} {
		if _, err := importDir([]importFile{{Path: bad}}); !errors.Is(err, errBadPath) {
			t.Errorf("importDir(%q) = %v, want errBadPath", bad, err)
		}
	}
}

// TestExportedFiles drops the directories and names a file by its id.
func TestExportedFiles(t *testing.T) {
	id := "0b6f8e2a-1c3d-4e5f-8a9b-0c1d2e3f4a5b"
	got := exportedFiles([]exportnotes.File{
		{Path: "Folder", Dir: true},
		{Path: "Folder/Note.md", Content: "text"},
		{Path: exportnotes.FilesDir + "/" + id + ".png", Key: "user/" + id},
	})
	want := []exported{{Path: "Folder/Note.md", Content: "text"}, {Path: exportnotes.FilesDir + "/" + id + ".png", File: id}}
	if !slices.Equal(got, want) {
		t.Errorf("exportedFiles = %v, want %v", got, want)
	}
}
