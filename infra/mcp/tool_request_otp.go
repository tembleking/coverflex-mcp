package mcp

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolRequestOTP struct {
	coverflexClient *coverflex.Client
}

func NewToolRequestOTP(client *coverflex.Client) *ToolRequestOTP {
	return &ToolRequestOTP{
		coverflexClient: client,
	}
}

func (t *ToolRequestOTP) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	user := os.Getenv("COVERFLEX_USERNAME")
	pass := os.Getenv("COVERFLEX_PASSWORD")

	if user == "" || pass == "" {
		return mcp.NewToolResultError("COVERFLEX_USERNAME and COVERFLEX_PASSWORD env vars must be set"), nil
	}

	if t.coverflexClient.IsLoggedIn() {
		return mcp.NewToolResultText("already logged in"), nil
	}

	err := t.coverflexClient.RequestOTP(user, pass)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("error requesting OTP", err), nil
	}

	return mcp.NewToolResultText("OTP requested successfully. Please let the user provide the OTP and configure it using the 'trust_device_via_otp' tool."), nil
}

func (t *ToolRequestOTP) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("request_otp",
		mcp.WithDescription("Initiates the login process by requesting an OTP to be sent to the user's phone."),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolRequestOTP) CanBeUsed() bool {
	return t.coverflexClient != nil && !t.coverflexClient.IsLoggedIn()
}
