package mcp

import (
	"context"
	"io"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/server"
)

type Handler struct {
	server *server.MCPServer
}

type mcpTool interface {
	RegisterInServer(server *server.MCPServer)
	CanBeUsed() bool
}

func NewHandler() *Handler {
	s := server.NewMCPServer(
		"Coverflex MCP Server",
		"1.0.0",
		server.WithInstructions(`You are a helpful assistant for managing Coverflex benefits. You have access to a set of tools to retrieve information about the user's benefits, cards, company details, and more.

If no tools are available, it means the user is not logged in. To help the user, follow these steps:
1. First, check if the 'COVERFLEX_USERNAME' and 'COVERFLEX_PASSWORD' environment variables are set.
2. If the environment variables are set but the user is not logged in, the device may not be trusted. In this case, suggest using the 'trust_device' tool with the OTP.
3. If the environment variables are not set, guide the user to authenticate manually by running the 'login' command with the '--user' and '--pass' flags.`),
		server.WithToolCapabilities(true),
	)
	return &Handler{
		server: s,
	}
}

func NewHandlerWithTools(tools ...mcpTool) *Handler {
	h := NewHandler()
	h.RegisterTools(tools...)
	return h
}

func (h *Handler) RegisterTools(tools ...mcpTool) {
	for _, tool := range tools {
		if tool.CanBeUsed() {
			tool.RegisterInServer(h.server)
		}
	}
}

func (h *Handler) ServeStdio(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	return server.NewStdioServer(h.server).Listen(ctx, stdin, stdout)
}

func (h *Handler) ServeInProcessClient() (*client.Client, error) {
	return client.NewInProcessClient(h.server)
}
