package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolIsLoggedIn struct {
	coverflexClient *coverflex.Client
}

func NewToolIsLoggedIn(client *coverflex.Client) *ToolIsLoggedIn {
	return &ToolIsLoggedIn{
		coverflexClient: client,
	}
}

func (t *ToolIsLoggedIn) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	loggedIn := t.coverflexClient.IsLoggedIn()
	return mcp.NewToolResultJSON(loggedIn)
}

func (t *ToolIsLoggedIn) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("is_logged_in",
		mcp.WithDescription("Check if the user is currently logged in."),
		mcp.WithOutputSchema[bool](),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolIsLoggedIn) CanBeUsed() bool {
	return t.coverflexClient != nil
}
