package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolGetCards struct {
	coverflexClient *coverflex.Client
}

func NewToolGetCards(client *coverflex.Client) *ToolGetCards {
	return &ToolGetCards{
		coverflexClient: client,
	}
}

func (t *ToolGetCards) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cards, err := t.coverflexClient.GetCards()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("error getting cards", err), nil
	}

	return mcp.NewToolResultJSON(cards)
}

func (t *ToolGetCards) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("get_cards",
		mcp.WithDescription("Retrieve user cards."),
		mcp.WithOutputSchema[[]coverflex.Card](),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolGetCards) CanBeUsed() bool {
	return t.coverflexClient != nil && t.coverflexClient.IsLoggedIn()
}
