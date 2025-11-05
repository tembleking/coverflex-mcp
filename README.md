# Coverflex MCP Server

This repository contains a CLI and MCP server for interacting with Coverflex.

## Overview

The `coverflex-mcp` is a command-line interface (CLI) tool designed to streamline your interactions with Coverflex services. It allows you to authenticate, manage tokens, and perform various operations directly from your terminal.

It also exposes an MCP server that allows AI agents to interact with Coverflex services.

## Available Tools

The MCP server exposes the following tools:

-   **`get_benefits`**: Retrieve user benefits.
-   **`get_cards`**: Retrieve user cards.
-   **`get_company`**: Retrieve company information.
-   **`get_compensation`**: Retrieve user compensation summary.
-   **`get_family`**: Retrieve user family members.
-   **`get_operations`**: Retrieve user operations with optional pagination and filtering.
-   **`is_logged_in`**: Check if the user is currently logged in.
-   **`request_otp`**: Initiates the login process by requesting an OTP to be sent to the user's phone.
-   **`trust_device_via_otp`**: Submits the One-Time Password (OTP) received via SMS to complete the login process and trust the device.

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
