# MC PeePee - Claude MCP Manager

## Overview

MC PeePee is a desktop application built with Go and Fyne designed to manage the Model Compute Providers (MCPs) for Anthropic's Claude Desktop application. It provides a graphical user interface to easily add, configure, enable/disable, and apply MCP settings, simplifying the process compared to manually editing the `claude_mcp_config.json` file. Additionally, it integrates with the Smithery AI registry to allow browsing and adding pre-configured Smithery servers as MCPs.

## Features

*   **MCP Management:**
    *   List currently configured MCPs from `claude_mcp_config.json` or the internal database.
    *   Enable or disable MCPs with checkboxes.
    *   Add new custom MCP configurations (command and arguments).
    *   Apply the current enabled/disabled state back to the `claude_mcp_config.json` file used by Claude Desktop.
*   **Smithery Integration:**
    *   Browse and search the Smithery AI server registry (requires API Key).
    *   View details of Smithery servers, including configuration schema and provided tools.
    *   Add selected Smithery servers directly as MCPs with pre-filled `npx` commands.
    *   Automatically runs the `npx @smithery/cli run <server>` command in the background when a Smithery MCP is added.
*   **Convenience Features:**
    *   Directly open the `claude_mcp_config.json` file in Cursor (if installed).
    *   Open the Claude Desktop log folder.
    *   Restart Claude Desktop (implementation details pending).
    *   Configure settings, such as the Smithery API Key.

## Prerequisites

*   **Go:** Version 1.17 or later.
*   **C Compiler:** Required by Fyne (e.g., GCC on Linux/macOS, MinGW on Windows).
*   **System Development Tools:** (e.g., `build-essential` on Debian/Ubuntu, Xcode Command Line Tools on macOS).
*   **Claude Desktop:** The application needs Claude Desktop installed to manage its configuration.
*   **(Optional) Cursor:** Required for the "Open Config" button to function.
*   **(Optional) Smithery API Key:** Required to use the Smithery registry browser tab. Obtainable from [smithery.ai](https://smithery.ai/).

## Installation & Building

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd mcpeepee
    ```
2.  **Fetch dependencies:**
    ```bash
    go mod tidy
    ```
3.  **Build the application:**
    ```bash
    go build -o mcpeepee
    ```
    (On Windows, use `go build -o mcpeepee.exe`)

## Usage

1.  **Run the executable:**
    ```bash
    ./mcpeepee
    ```
    (Or `.\mcpeepee.exe` on Windows)

2.  **MCP Manager Tab:**
    *   View your existing MCPs.
    *   Check/uncheck boxes to enable/disable MCPs.
    *   Click **"Add New MCP"** to define a custom provider. Fill in the name, command, and JSON arguments.
    *   Click **"Apply to Claude"** to save the enabled/disabled status to `claude_mcp_config.json`. Claude Desktop typically needs a restart to pick up changes.
    *   Use the other buttons (**"Open Config"**, **"Open Logs"**, **"Restart Claude"**, **"Settings"**) for convenience.

3.  **Smithery Tab:**
    *   Go to **"Settings"** (in the MCP Manager tab) and enter your Smithery API Key.
    *   Return to the **"Smithery"** tab. It should now fetch and display servers.
    *   Use the search bar to find specific servers.
    *   Click **"Details"** on a server to view more information in a pop-up window.
    *   Click **"Add to Config"** to add the selected Smithery server as an enabled MCP. The application will attempt to run the necessary `npx` setup command in the background.

## Configuration

*   **Claude MCP Config:** The primary target file is `claude_mcp_config.json`, usually located in Claude Desktop's application support directory (e.g., `~/Library/Application Support/Claude/` on macOS). The application attempts to locate this automatically.
*   **Internal Database:** MC PeePee uses a local SQLite database (`mcpeepee.db` in the same directory as the executable) to store MCP configurations and settings persistently. It syncs with `claude_mcp_config.json` on startup.
*   **Settings:** Application settings (like the Smithery API Key) are stored in the internal database.

## Troubleshooting

*   **Fyne Errors on Startup:** Ensure you have a C compiler and the necessary development libraries installed for your operating system. See the [Fyne documentation](https://developer.fyne.io/started/) for details.
*   **Cannot Find `claude_mcp_config.json`:** The application tries to automatically detect the path. If it fails, check the logs for errors. Manual path configuration might be needed in future versions.
*   **Smithery Tab Blank:** Ensure you have entered a valid Smithery API Key in Settings and have an internet connection. Check the application logs (printed to the terminal where you ran the app) for API errors.
*   **Smithery Setup Fails:** If adding a Smithery MCP shows an error dialog after confirmation, check the terminal logs for detailed output from the `npx` command. Ensure `npx` (Node.js) is installed and working on your system.

## License

(Specify your license here, e.g., MIT License) 