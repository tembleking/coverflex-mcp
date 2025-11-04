package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolGetCompensation struct {
	coverflexClient *coverflex.Client
}

func NewToolGetCompensation(client *coverflex.Client) *ToolGetCompensation {
	return &ToolGetCompensation{
		coverflexClient: client,
	}
}

func (t *ToolGetCompensation) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	compensation, err := t.coverflexClient.GetCompensation()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("error getting compensation", err), nil
	}

	return mcp.NewToolResultJSON(compensation)
}

func (t *ToolGetCompensation) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("get_compensation",
		mcp.WithDescription("Retrieve user compensation summary."),
		mcp.WithOutputSchema[*coverflex.CompensationSummary](),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolGetCompensation) CanBeUsed() bool {
	return t.coverflexClient != nil && t.coverflexClient.IsLoggedIn()
}
