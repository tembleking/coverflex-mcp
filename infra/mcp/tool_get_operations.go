package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolGetOperations struct {
	coverflexClient *coverflex.Client
}

func NewToolGetOperations(client *coverflex.Client) *ToolGetOperations {
	return &ToolGetOperations{
		coverflexClient: client,
	}
}

func (t *ToolGetOperations) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	page := request.GetInt("page", 0)
	perPage := request.GetInt("per_page", 0)
	filterType := request.GetString("filter_type", "")

	var opts []coverflex.GetOperationsOption
	if page > 0 {
		opts = append(opts, coverflex.WithOperationsPage(int(page)))
	}
	if perPage > 0 {
		opts = append(opts, coverflex.WithOperationsPerPage(int(perPage)))
	}
	if filterType != "" {
		opts = append(opts, coverflex.WithOperationsFilterType(filterType))
	}

	operations, err := t.coverflexClient.GetOperations(opts...)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("error getting operations", err), nil
	}

	return mcp.NewToolResultJSON(map[string][]coverflex.Operation{"result": operations})
}

func (t *ToolGetOperations) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("get_operations",
		mcp.WithDescription("Retrieve user operations with optional pagination and filtering. To retrieve all operations, the LLM needs to paginate."),
		mcp.WithNumber("page", mcp.Description("The page number for pagination."), mcp.DefaultNumber(1)),
		mcp.WithNumber("per_page", mcp.Description("The number of items per page."), mcp.DefaultNumber(20)),
		mcp.WithString("filter_type", mcp.Description("The type of operation to filter by.")),
		mcp.WithOutputSchema[map[string][]coverflex.Operation](),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolGetOperations) CanBeUsed() bool {
	return t.coverflexClient != nil && t.coverflexClient.IsLoggedIn()
}
