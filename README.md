# Coverflex MCP Server

This repository contains a CLI and MCP server for interacting with Coverflex.

## Overview

This project is an MCP server that allows AI agents to interact with Coverflex services.

It includes a command-line interface (CLI) primarily used for authenticating the user, which then enables the tools available in the MCP server.

## Available Tools

The MCP server exposes different tools depending on the user's authentication status.

### When Logged Out

-   **`is_logged_in`**: Check if the user is currently logged in.
-   **`request_otp`**: Initiates the login process by requesting an OTP to be sent to the user's phone.
-   **`trust_device_via_otp`**: Submits the One-Time Password (OTP) received via SMS to complete the login process and trust the device.

### When Logged In

-   **`is_logged_in`**: Check if the user is currently logged in.
-   **`get_benefits`**: Retrieve user benefits.
-   **`get_cards`**: Retrieve user cards.
-   **`get_company`**: Retrieve company information.
-   **`get_compensation`**: Retrieve user compensation summary.
-   **`get_family`**: Retrieve user family members.
-   **`get_operations`**: Retrieve user operations with optional pagination and filtering.

## Getting Started

### Prerequisites

-   Go 1.21 or higher
-   Nix (optional, for development environment)

### Usage

There are two main ways to run the application.

#### Using `go run` (Recommended)

To run the application directly from the remote repository without a local clone:
```sh
# To start the MCP server
go run github.com/tembleking/coverflex-mcp/cmd/server@latest

# To log in (note the -- separator)
go run github.com/tembleking/coverflex-mcp/cmd/server@latest -- login --user <your-email> --pass <your-password>
```

#### Building from source

If you prefer to build the binary yourself:
1.  Clone the repository:
    ```sh
    git clone https://github.com/tembleking/coverflex-mcp.git
    cd coverflex-mcp
    ```

2.  Build the binary:
    ```sh
    go build -o coverflex-mcp ./cmd/server
    ```
This will create a `coverflex-mcp` executable in the current directory.

### Authentication

To use the MCP server, you first need to log in to your Coverflex account. If you built from source, you'll run `./coverflex-mcp`. If you are using `go run`, you'll use the command from the section above.

Example using the built binary:
```sh
./coverflex-mcp login --user <your-email> --pass <your-password>
```

If Two-Factor Authentication (2FA) is enabled, you will receive an OTP on your phone. Re-run the command with the `--otp` flag.

Once authenticated, the tool will save your tokens for future use.

### MCP Server

To start the MCP server, run the appropriate command for your chosen method (`go run` or the built binary).

Example using the built binary:
```sh
./coverflex-mcp
```

The server will start and listen for requests from MCP clients.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
