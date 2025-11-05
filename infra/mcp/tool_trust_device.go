package mcp

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolTrustDevice struct {
	coverflexClient *coverflex.Client
}

func NewToolTrustDevice(client *coverflex.Client) *ToolTrustDevice {
	return &ToolTrustDevice{
		coverflexClient: client,
	}
}

func (t *ToolTrustDevice) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	otp := request.GetString("otp", "")
	if otp == "" {
		return mcp.NewToolResultError("otp is required"), nil
	}

	user := os.Getenv("COVERFLEX_USERNAME")
	pass := os.Getenv("COVERFLEX_PASSWORD")

	if user == "" || pass == "" {
		return mcp.NewToolResultError("COVERFLEX_USERNAME and COVERFLEX_PASSWORD env vars must be set"), nil
	}

	if t.coverflexClient.IsLoggedIn() {
		return mcp.NewToolResultText("already logged in"), nil
	}

	err := t.coverflexClient.Login(user, pass, otp)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("error logging in", err), nil
	}

	return mcp.NewToolResultText("login successful"), nil
}

func (t *ToolTrustDevice) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("trust_device",
		mcp.WithDescription("Trust the device by providing the OTP sent to your phone. It uses COVERFLEX_USERNAME and COVERFLEX_PASSWORD environment variables to login."),
		mcp.WithString("otp", mcp.Description("The One-Time Password (OTP) received via SMS for 2FA.")),
		mcp.WithOutputSchema[string](),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolTrustDevice) CanBeUsed() bool {
	return t.coverflexClient != nil
}
