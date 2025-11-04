package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolGetCompany struct {
	coverflexClient *coverflex.Client
}

func NewToolGetCompany(client *coverflex.Client) *ToolGetCompany {
	return &ToolGetCompany{
		coverflexClient: client,
	}
}

func (t *ToolGetCompany) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	company, err := t.coverflexClient.GetCompany()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("error getting company", err), nil
	}

	return mcp.NewToolResultJSON(company)
}

func (t *ToolGetCompany) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("get_company",
		mcp.WithDescription("Retrieve company information."),
		mcp.WithOutputSchema[map[string]any](),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolGetCompany) CanBeUsed() bool {
	return t.coverflexClient != nil
}
