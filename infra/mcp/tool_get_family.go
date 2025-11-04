package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolGetFamily struct {
	coverflexClient *coverflex.Client
}

func NewToolGetFamily(client *coverflex.Client) *ToolGetFamily {
	return &ToolGetFamily{
		coverflexClient: client,
	}
}

func (t *ToolGetFamily) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	family, err := t.coverflexClient.GetFamily()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("error getting family members", err), nil
	}

	return mcp.NewToolResultJSON(family)
}

func (t *ToolGetFamily) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("get_family",
		mcp.WithDescription("Retrieve user family members."),
		mcp.WithOutputSchema[[]coverflex.FamilyMember](),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolGetFamily) CanBeUsed() bool {
	return t.coverflexClient != nil && t.coverflexClient.IsLoggedIn()
}
