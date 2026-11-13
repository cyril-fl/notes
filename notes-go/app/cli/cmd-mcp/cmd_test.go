package cmdmcp

import (
	"context"
	"slices"
	"testing"

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
	for _, want := range []string{"search_notes", "recent_notes", "read_note", "list_folder", "create_note", "edit_note", "create_folder", "rename_folder", "move_item", "delete_note", "delete_folder", "list_trash", "restore_item", "list_tags", "read_tag"} {
		if !slices.Contains(names, want) {
			t.Errorf("tools = %v, missing %q", names, want)
		}
	}
}
