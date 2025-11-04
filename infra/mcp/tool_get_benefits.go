package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolGetBenefits struct {
	coverflexClient *coverflex.Client
}

func NewToolGetBenefits(client *coverflex.Client) *ToolGetBenefits {
	return &ToolGetBenefits{
		coverflexClient: client,
	}
}

func (t *ToolGetBenefits) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	benefits, err := t.coverflexClient.GetBenefits()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("error getting benefits", err), nil
	}

	return mcp.NewToolResultJSON(map[string]any{"benefits": benefits})
}

func (t *ToolGetBenefits) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("get_benefits",
		mcp.WithDescription("Retrieve user benefits."),
		mcp.WithOutputSchema[map[string]any](),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolGetBenefits) CanBeUsed() bool {
	return t.coverflexClient != nil
}
