package mcp

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tembleking/coverflex-mcp/infra/coverflex"
)

type ToolTrustDeviceViaOTP struct {
	coverflexClient *coverflex.Client
}

func NewToolTrustDeviceViaOTP(client *coverflex.Client) *ToolTrustDeviceViaOTP {
	return &ToolTrustDeviceViaOTP{
		coverflexClient: client,
	}
}

func (t *ToolTrustDeviceViaOTP) handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		return mcp.NewToolResultErrorFromErr("error submitting OTP", err), nil
	}

	return mcp.NewToolResultText("OTP submitted successfully. Device trusted, refresh the MCP servers to see the available tools."), nil
}

func (t *ToolTrustDeviceViaOTP) RegisterInServer(s *server.MCPServer) {
	tool := mcp.NewTool("trust_device_via_otp",
		mcp.WithDescription("Submits the One-Time Password (OTP) received via SMS to complete the login process and trust the device."),
		mcp.WithString("otp", mcp.Description("The One-Time Password (OTP) received via SMS for 2FA.")),
	)

	s.AddTool(tool, t.handle)
}

func (t *ToolTrustDeviceViaOTP) CanBeUsed() bool {
	return t.coverflexClient != nil && !t.coverflexClient.IsLoggedIn()
}
