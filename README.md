# MC PeePee - Claude MCP Manager

[![Go Build Status](https://img.shields.io/github/actions/workflow/status/4cecoder/mcpeepee/go.yml?branch=main)](https://github.com/4cecoder/mcpeepee/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/4cecoder/mcpeepee)](https://goreportcard.com/report/github.com/4cecoder/mcpeepee)

**Manage your Claude Desktop Model Compute Providers (MCPs) with ease!**

MC PeePee provides a user-friendly interface built with Go and [Fyne](https://fyne.io/) to manage MCP configurations for Anthropic's Claude Desktop. Forget manually editing JSON files – add, configure, enable/disable MCPs, and integrate with the [Smithery AI](https://smithery.ai/) registry effortlessly.

---

## ✨ Features

| Feature                 | Description                                                                                                                                                           |
| :---------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **MCP Management**      | List, enable/disable existing MCPs. Add new custom MCP configurations (command & arguments). Apply changes directly to Claude's configuration file.                |
| **Smithery Integration**| Browse and search the Smithery AI server registry (requires API Key). View server details (config schema, tools). Add Smithery servers directly as MCPs.              |
| **Automatic Setup**     | When adding a Smithery MCP, automatically runs `npx -y @smithery/cli@latest install <server_name> --client claude [--key <api_key>]` in the background for setup. |
| **Convenience Tools**   | Open Claude config file in Cursor (if installed). Open Claude log folder. Configure settings (e.g., Smithery API Key). *(Restart Claude feature planned)*           |

---

## 🚀 Getting Started

### Prerequisites

Make sure you have the following installed:

| Requirement               | Minimum Version / Notes                                                                 |
| :------------------------ | :-------------------------------------------------------------------------------------- |
| **Go Compiler**           | `1.17` or later                                                                         |
| **C Compiler & Dev Tools**| Required by Fyne (e.g., `gcc`/`build-essential` on Linux, Xcode Command Line Tools on macOS) |
| **Claude Desktop**        | Must be installed for MC PeePee to manage its configuration.                             |
| **Node.js & npm/npx**   | Required for the automatic Smithery CLI execution (`install` command).                 |
| **(Optional) Cursor**     | For the "Open Config" button.                                                          |
| **(Optional) Smithery Key**| For browsing the Smithery registry. Get one at [smithery.ai](https://smithery.ai/).     |

### Installation & Running

We recommend using the provided `Makefile` for ease of use:

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/4cecoder/mcpeepee.git
    cd mcpeepee
    ```

2.  **Prepare Dependencies:**
    ```bash
    make tidy
    ```
    *(This runs `go mod tidy`)*

3.  **Build the Application:**
    ```bash
    make build
    ```
    *(This runs `go build -o mcpeepee`)*

4.  **Run MC PeePee:**
    ```bash
    make run
    ```
    *(This executes `./mcpeepee`)*

Alternatively, you can use standard Go commands (`go mod tidy`, `go build`, `./mcpeepee`).

---

## 💻 Usage Guide

1.  **Launch the application** using `make run` or `./mcpeepee`.

2.  **MCP Manager Tab:**
    *   Your current MCPs are listed here.
    *   ✅ Check/uncheck boxes to enable/disable.
    *   ➕ Click **"Add New MCP"** for custom providers (name, command, JSON args).
    *   💾 Click **"Apply to Claude"** to save the current state to `claude_mcp_config.json`. *Remember to restart Claude Desktop to see the changes.*
    *   ⚙️ Use **"Settings"** to add your Smithery API Key.
    *   Other buttons provide quick access to logs and config files.

3.  **Smithery Tab:**
    *   Requires a **Smithery API Key** set in Settings.
    *   Browse or search for servers.
    *   📄 Click **"Details"** for more info.
    *   ➕ Click **"Add to Config"** to add the server as an MCP.
        *   This saves the MCP config (enabled by default).
        *   It *also* runs the `npx ... install <server> --client claude ...` command in the background to perform initial setup. You'll see a confirmation or error dialog.

---

## 🔧 Configuration Explained

*   **Claude Config:** The app reads/writes `claude_mcp_config.json` (usually in `~/Library/Application Support/Claude/` on macOS, auto-detected).
*   **Internal DB:** Uses `mcpeepee.db` (SQLite) in the app's directory for storing config and settings. Syncs with Claude's config on startup.
*   **Settings:** Stored in `mcpeepee.db`.

---

## 🤔 Troubleshooting

| Issue                            | Suggestion                                                                                                                                                              |
| :------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Fyne Errors on Startup**       | Ensure C compiler & dev tools are installed. See [Fyne Getting Started](https://developer.fyne.io/started/).                                                           |
| **Cannot Find Claude Config**    | Check logs for detection errors. Auto-detection might fail in non-standard setups.                                                                                      |
| **Smithery Tab Blank**           | Verify API Key in Settings & internet connection. Check terminal logs for API errors.                                                                                    |
| **Smithery Setup Fails**         | Adding a Smithery MCP runs `npx ... install ... --client claude`. Check terminal logs for detailed `npx` output. Ensure Node.js/npx is installed and working correctly. |
| **Build Fails**                  | Run `make clean` then `make tidy` and `make build` again. Check Go & C compiler setup.                                                                                  |

---

## 🛠️ Using the Makefile

The `Makefile` simplifies common development tasks:

| Command         | Action                                                  |
| :-------------- | :------------------------------------------------------ |
| `make tidy`     | Downloads/cleans up Go module dependencies (`go mod tidy`). |
| `make build`    | Compiles the Go application (`go build -o mcpeepee`).   |
| `make run`      | Executes the compiled application (`./mcpeepee`).        |
| `make clean`    | Removes the compiled `mcpeepee` binary.                 |
| `make all`      | Runs `tidy` followed by `build`.                        |

---

## 📜 License

*(Specify your license here, e.g., MIT License)* 