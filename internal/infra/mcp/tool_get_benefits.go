package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/internal/infra/coverflex"
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

	return mcp.NewToolResultJSON(map[string][]coverflex.Benefit{"result": benefits})
}

func (t *ToolGetBenefits) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("get_benefits",
		mcp.WithDescription("Retrieve Coverflex user benefits."),
		mcp.WithOutputSchema[map[string][]coverflex.Benefit](),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolGetBenefits) CanBeUsed() bool {
	return t.coverflexClient != nil && t.coverflexClient.IsLoggedIn()
}
