# Coverflex MCP Server

This repository contains a CLI and MCP server for interacting with Coverflex.

## Overview

The `coverflex-mcp` is a command-line interface (CLI) tool designed to streamline your interactions with Coverflex services. It allows you to authenticate, manage tokens, and perform various operations directly from your terminal.

It also exposes an MCP server that allows AI agents to interact with Coverflex services.

## Packages

### `cmd/server`

This package contains the entry point of the application.

-   **`root.go`**: Defines the root command `coverflex-mcp`, sets up the MCP server with all the available tools, and initializes the application.
-   **`login.go`**: Implements the `login` command, which handles authentication, token management, OTP verification, and token refresh logic.
-   **`main.go`**: The main function that executes the root command.

### `infra/mcp`

This package is responsible for the MCP server and its tools.

-   **`mcp_handler.go`**: Sets up the MCP server and registers the tools. It dynamically registers tools based on the user's login status.
-   **`tool_*.go`**: Each file implements a specific MCP tool for interacting with the Coverflex API. These tools cover functionalities like retrieving benefits, cards, company information, compensation, family members, and operations. There are also tools for authentication, such as checking the login status, requesting an OTP, and trusting a device via OTP.

### `infra/fs`

This package handles the persistence of authentication tokens.

-   **`token_repository.go`**: Implements a token repository that stores and retrieves authentication and refresh tokens from the filesystem. The tokens are stored in the system's temporary directory.

## Getting Started

### Prerequisites

-   Go 1.21 or higher
-   Nix (optional, for development environment)

### Installation

1.  Clone the repository:
    ```sh
    git clone https://github.com/tembleking/coverflex-mcp.git
    cd coverflex-mcp
    ```

2.  Build the binary:
    ```sh
    go build -o coverflex-mcp ./cmd/server
    ```

### Usage

#### Authentication

To use the MCP server, you first need to log in to your Coverflex account.

```sh
./coverflex-mcp login --user <your-email> --pass <your-password>
```

If Two-Factor Authentication (2FA) is enabled, you will receive an OTP on your phone. Re-run the command with the `--otp` flag:

```sh
./coverflex-mcp login --user <your-email> --pass <your-password> --otp <your-otp>
```

Once authenticated, the tool will save your tokens for future use.

#### MCP Server

To start the MCP server, run the following command:

```sh
./coverflex-mcp
```

The server will start and listen for requests from MCP clients.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
