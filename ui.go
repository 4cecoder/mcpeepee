package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	// Assuming module path is mcpeepee or similar, adjust if necessary
	"mcpeepee/smithery"
)

// UI Globals (consider wrapping in a UI struct later for better encapsulation)
var (
	mcpList     *fyne.Container          // Container for the checklist
	mcpCheckMap map[string]*widget.Check // Map MCP name to its checkbox

	// Smithery Tab UI elements
	smitheryServerList   *widget.List
	smitheryServersData  []smithery.Server // Use type from smithery package - LIST DATA
	smitherySearchEntry  *widget.Entry
	smitheryStatusLabel  *widget.Label
	smitheryTabContainer *fyne.Container // Main container for the smithery tab content (list or details)
	// smitheryNavStack     *fyne.Container // No longer needed, using modals for details
)

// --- Smithery API Structs ---
// Based on https://smithery.ai/docs/registry/llms.txt
type SmitheryServer struct {
	QualifiedName string `json:"qualifiedName"`
	DisplayName   string `json:"displayName"`
	Description   string `json:"description"`
	Homepage      string `json:"homepage"`
	UseCount      string `json:"useCount"` // Kept as string as per doc
	IsDeployed    bool   `json:"isDeployed"`
	CreatedAt     string `json:"createdAt"`
}

type SmitheryPagination struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalPages  int `json:"totalPages"`
	TotalCount  int `json:"totalCount"`
}

type SmitheryListResponse struct {
	Servers    []SmitheryServer   `json:"servers"`
	Pagination SmitheryPagination `json:"pagination"`
}

// --- Smithery API Client ---
const smitheryRegistryURL = "https://registry.smithery.ai/servers"

func fetchSmitheryServers(apiKey string, query string, page int) (*SmitheryListResponse, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Smithery API key is required")
	}

	req, err := http.NewRequest("GET", smitheryRegistryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	// Add query parameters
	q := req.URL.Query()
	if query != "" {
		q.Add("q", query)
	}
	if page > 1 {
		q.Add("page", fmt.Sprintf("%d", page))
	}
	// q.Add("pageSize", "20") // Example: request more items per page
	req.URL.RawQuery = q.Encode()

	log.Println("Fetching Smithery servers from:", req.URL.String())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Smithery API Error: Status %d, Body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("Smithery API request failed with status %d", resp.StatusCode)
	}

	var listResponse SmitheryListResponse
	if err := json.Unmarshal(body, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Smithery response: %w", err)
	}

	return &listResponse, nil
}

// --- UI Functions ---

// refreshMCPListUI loads MCPs from the DB and updates the UI checklist.
func refreshMCPListUI() error {
	mcps, err := loadMCPsFromDB() // Uses func from db.go
	if err != nil {
		return fmt.Errorf("failed to load MCPs from database for UI refresh: %w", err)
	}

	// Ensure UI elements are initialized (should be done in buildUI ideally)
	if mcpList == nil {
		log.Println("WARN: mcpList container not initialized before refresh.")
		mcpList = container.NewVBox() // Initialize defensively
	}
	if mcpCheckMap == nil {
		log.Println("WARN: mcpCheckMap not initialized before refresh.")
		mcpCheckMap = make(map[string]*widget.Check) // Initialize defensively
	}

	// Clear existing UI items
	mcpList.Objects = nil
	// Clear map - important if MCPs can be deleted/renamed
	mcpCheckMap = make(map[string]*widget.Check)

	for i := range mcps {
		mcp := mcps[i] // Create a local copy for the closure

		check := widget.NewCheck(mcp.Name, func(checked bool) {
			go func(mcpToUpdate *MCP, isChecked bool) { // MCP is defined in db.go
				log.Printf("Checkbox changed for '%s': %v", mcpToUpdate.Name, isChecked)
				if err := updateMCPEnabledStatus(mcpToUpdate.ID, isChecked); err != nil { // Uses func from db.go
					log.Printf("ERROR updating MCP '%s' status: %v", mcpToUpdate.Name, err)
					// TODO: Show error to user, maybe revert checkbox?
				} else {
					log.Printf("Successfully updated enabled status for MCP '%s'", mcpToUpdate.Name)
				}
			}(&mcp, checked) // Pass the address of the loop variable copy
		})
		check.SetChecked(mcp.Enabled)
		mcpCheckMap[mcp.Name] = check
		mcpList.Add(check)
	}

	mcpList.Refresh()
	log.Printf("Refreshed MCP UI list with %d items", len(mcps))
	return nil
}

// showAddMCPDialog displays a dialog for adding a new MCP configuration.
func showAddMCPDialog(parentWindow fyne.Window) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Unique MCP Name (e.g., my-custom-mcp)")

	commandEntry := widget.NewEntry()
	commandEntry.SetPlaceHolder("Executable command (e.g., python3, npx)")

	argsEntry := widget.NewMultiLineEntry()
	argsEntry.SetPlaceHolder(`Arguments as JSON array (e.g., ["script.py", "--port", "8080"])`)
	argsEntry.SetMinRowsVisible(3)

	formItems := []*widget.FormItem{
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Command", commandEntry),
		widget.NewFormItem("Args (JSON Array)", argsEntry),
	}

	callback := func(confirmed bool) {
		if !confirmed {
			return
		}
		log.Println("Save clicked in Add MCP Dialog")
		name := strings.TrimSpace(nameEntry.Text)
		command := strings.TrimSpace(commandEntry.Text)
		argsJSONString := strings.TrimSpace(argsEntry.Text)

		// --- Validation ---
		if name == "" || command == "" {
			dialog.ShowError(fmt.Errorf("MCP Name and Command cannot be empty"), parentWindow)
			return
		}

		// Validate Args JSON
		var args []string
		if argsJSONString == "" {
			argsJSONString = "[]"
		} else {
			if err := json.Unmarshal([]byte(argsJSONString), &args); err != nil {
				dialog.ShowError(fmt.Errorf("invalid JSON array format for Args: %w", err), parentWindow)
				return
			}
			argsBytes, _ := json.Marshal(args) // Re-marshal for consistent formatting
			argsJSONString = string(argsBytes)
		}

		// --- Save to DB ---
		newMCP := MCP{ // MCP is defined in db.go
			Name:     name,
			Command:  command,
			ArgsJSON: argsJSONString,
			Enabled:  false, // Add as disabled by default
		}
		if err := addMCP(newMCP); err != nil { // Uses func from db.go
			log.Printf("ERROR creating new MCP '%s': %v", name, err)
			if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "already exists") {
				dialog.ShowError(fmt.Errorf("MCP with name '%s' already exists", name), parentWindow)
			} else {
				dialog.ShowError(fmt.Errorf("failed to save new MCP to database: %w", err), parentWindow)
			}
			return
		}

		log.Printf("Successfully added new MCP: %s", name)
		dialog.ShowInformation("Success", fmt.Sprintf("MCP '%s' added successfully.", name), parentWindow)

		// Refresh the main list
		go func() {
			if err := refreshMCPListUI(); err != nil {
				log.Printf("ERROR refreshing MCP list after add: %v", err)
			}
		}()
	}

	dia := dialog.NewForm(
		"Add New MCP Configuration", "Save", "Cancel",
		formItems, callback, parentWindow,
	)
	dia.Resize(fyne.NewSize(400, 300))
	dia.Show()
}

// refreshSmitheryTab fetches servers and updates the Smithery tab UI
func refreshSmitheryTab() {
	// Remove the initial UI readiness check from here
	// if smitheryStatusLabel == nil || smitheryServerList == nil || smitheryTabContainer == nil {
	// 	log.Println("WARN: Smithery tab UI elements not ready for refresh.")
	// 	return // UI not fully built yet
	// }

	// Always try to get the key first
	apiKey, err := getSetting("SmitheryAPIKey")
	log.Printf("refreshSmitheryTab: Retrieved API Key: [%s] (Error: %v)", apiKey, err)
	if err != nil {
		log.Printf("ERROR getting Smithery API key for refresh: %v", err)
		// Only update label if it exists
		if smitheryStatusLabel != nil {
			smitheryStatusLabel.SetText("Error: Could not get API key.")
		}
		return
	}

	if apiKey == "" {
		log.Println("refreshSmitheryTab: API Key is empty.")
		// Only update label if it exists
		if smitheryStatusLabel != nil {
			smitheryStatusLabel.SetText("Smithery API Key not set. Enter in Settings.")
		}
		// Clear data and refresh list if it exists
		if smitheryServerList != nil {
			smitheryServersData = nil
			smitheryServerList.Refresh()
		}
		return
	}

	// If we have an API key, proceed to fetch, even if UI isn't fully ready yet.
	// Update status label immediately if possible.
	if smitheryStatusLabel != nil {
		smitheryStatusLabel.SetText("Fetching servers...")
	}
	// Clear data and refresh list immediately if possible.
	if smitheryServerList != nil {
		smitheryServersData = nil // Clear old data
		smitheryServerList.Refresh()
	}

	// Perform fetch in background
	go func() {
		query := ""
		if smitherySearchEntry != nil {
			query = smitherySearchEntry.Text
		}
		// Use function from smithery package
		resp, fetchErr := smithery.FetchServers(apiKey, query, 1)

		// NOW check if UI elements are ready before updating them
		if smitheryStatusLabel == nil || smitheryServerList == nil {
			log.Println("WARN: Smithery UI elements not ready when fetch completed. UI will update on next view/refresh.")
			// Store the fetched data anyway, so it's ready for the next refresh
			if fetchErr == nil && resp != nil {
				smitheryServersData = resp.Servers
			} else {
				smitheryServersData = nil
			}
			return // Skip UI updates
		}

		// --- UI Update Logic (runs only if elements were ready) ---
		if fetchErr != nil {
			log.Printf("ERROR fetching Smithery servers: %v", fetchErr)
			smitheryStatusLabel.SetText(fmt.Sprintf("Error fetching servers:\n%v", fetchErr))
			smitheryServersData = nil
		} else if resp == nil || len(resp.Servers) == 0 {
			smitheryStatusLabel.SetText("No servers found.")
			smitheryServersData = nil
		} else {
			smitheryStatusLabel.SetText(fmt.Sprintf("Found %d servers.", resp.Pagination.TotalCount))
			smitheryServersData = resp.Servers // Assign data from smithery package type
		}
		// Refresh must happen after data update
		smitheryServerList.Refresh()

	}()
}

// buildSmitheryTabUI creates the content for the Smithery tab
func buildSmitheryTabUI() fyne.CanvasObject {
	smitheryStatusLabel = widget.NewLabel("Enter API Key in Settings to search Smithery.")
	smitherySearchEntry = widget.NewEntry()
	smitherySearchEntry.SetPlaceHolder("Search Smithery Registry...")
	smitherySearchEntry.OnSubmitted = func(s string) {
		log.Println("Smithery search submitted:", s)
		refreshSmitheryTab()
	}

	searchButton := widget.NewButton("Search", func() {
		log.Println("Smithery search button clicked")
		refreshSmitheryTab()
	})

	// Initialize the data slice
	smitheryServersData = []smithery.Server{}

	// Create the list widget
	smitheryServerList = widget.NewList(
		func() int {
			return len(smitheryServersData)
		},
		func() fyne.CanvasObject { // CreateItem template
			nameLabel := widget.NewLabel("Server Name")
			descLabel := widget.NewLabel("Description")
			detailsButton := widget.NewButton("Details", nil)
			detailsButton.Importance = widget.LowImportance
			addButton := widget.NewButton("Add to Config", nil)
			addButton.Importance = widget.LowImportance

			textInfo := container.NewVBox(nameLabel, descLabel)

			// Use HBox with Spacer for better control
			// Make sure the text area expands and buttons stay right
			buttons := container.NewHBox(detailsButton, addButton)
			return container.NewHBox(textInfo, layout.NewSpacer(), buttons)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) { // UpdateItem
			if i >= len(smitheryServersData) {
				log.Printf("[List UpdateItem %d] Index out of bounds (data length %d)", i, len(smitheryServersData))
				o.Hide()
				return
			}
			o.Show()

			server := smitheryServersData[i]
			hbox := o.(*fyne.Container)
			textInfo := hbox.Objects[0].(*fyne.Container) // VBox is first item
			buttons := hbox.Objects[2].(*fyne.Container)  // HBox containing buttons is third item
			detailsButton := buttons.Objects[0].(*widget.Button)
			addButton := buttons.Objects[1].(*widget.Button)
			nameLabel := textInfo.Objects[0].(*widget.Label)
			descLabel := textInfo.Objects[1].(*widget.Label)

			nameLabel.SetText(server.DisplayName)

			descText := server.Description
			maxDescLen := 180
			if len(descText) > maxDescLen {
				descText = descText[:maxDescLen] + "..."
			}
			descLabel.SetText(descText)

			// Update button actions
			detailsButton.OnTapped = func() {
				log.Printf("Details button clicked for: %s (%s)", server.DisplayName, server.QualifiedName)
				// Get API key (handle error)
				apiKey, err := getSetting("SmitheryAPIKey")
				if err != nil || apiKey == "" {
					log.Printf("ERROR: Cannot fetch details, Smithery API key not available: %v", err)
					dialog.ShowError(fmt.Errorf("Smithery API Key is not set. Please add it in Settings."), fyne.CurrentApp().Driver().AllWindows()[0])
					return
				}

				// Show loading indicator (optional)
				loading := dialog.NewProgressInfinite("Fetching Details", "Fetching details for "+server.QualifiedName+"...", fyne.CurrentApp().Driver().AllWindows()[0])
				loading.Show()

				go func(qName string) {
					detail, fetchErr := smithery.FetchServerDetail(apiKey, qName)
					loading.Hide()

					if fetchErr != nil {
						log.Printf("ERROR fetching server detail for %s: %v", qName, fetchErr)
						dialog.ShowError(fmt.Errorf("Failed to fetch server details: %w", fetchErr), fyne.CurrentApp().Driver().AllWindows()[0])
						return
					}

					if detail != nil {
						// Pass both the original server (for description) and the detail
						detailViewContent := buildServerDetailView(server, *detail) // Get content
						// Create and show a modal dialog
						detailDialog := dialog.NewCustom(
							server.DisplayName, // Title
							"Close",            // Dismiss button text
							detailViewContent,  // Content
							fyne.CurrentApp().Driver().AllWindows()[0], // Parent window
						)
						detailDialog.Resize(fyne.NewSize(500, 400)) // Adjust size as needed
						detailDialog.Show()
					} else {
						log.Printf("WARN: Fetched server detail was nil for %s", qName)
						// Handle case where detail is nil even without error (shouldn't happen ideally)
					}
				}(server.QualifiedName) // Pass qualified name to goroutine

			}
			addButton.OnTapped = func() {
				log.Printf("Add button clicked for: %s (%s)", server.DisplayName, server.QualifiedName)
				showAddSmitheryMCPDialog(fyne.CurrentApp().Driver().AllWindows()[0], server)
			}
		},
	)

	topBar := container.NewBorder(nil, nil, nil, searchButton, smitherySearchEntry)
	// This container holds the list view components
	listViewContainer := container.NewBorder(topBar, smitheryStatusLabel, nil, nil, smitheryServerList)

	// Initial fetch
	go refreshSmitheryTab() // Might need adjustment if detail view is active

	// Return the stack, which will initially show the list view - RETURN LIST VIEW CONTAINER DIRECTLY
	return listViewContainer
}

// buildServerDetailView creates the content for the server detail modal.
// Takes smithery.Server (for list info like description) and smithery.ServerDetail (for detailed info) as input.
func buildServerDetailView(server smithery.Server, serverDetail smithery.ServerDetail) fyne.CanvasObject {
	// Use serverDetail fields
	nameLabel := widget.NewLabelWithStyle(serverDetail.DisplayName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	qualifiedNameLabel := widget.NewLabel(fmt.Sprintf("Qualified Name: %s", serverDetail.QualifiedName))

	// Use description from the original server list item
	descLabel := widget.NewLabel(server.Description)
	descLabel.Wrapping = fyne.TextWrapWord // Enable word wrapping
	descScroll := container.NewVScroll(descLabel)
	descScroll.SetMinSize(fyne.NewSize(0, 50)) // Give description some space

	// Homepage is also not in ServerDetail
	// var homepageWidget fyne.CanvasObject = widget.NewLabel("Homepage: N/A")

	// Display new fields from ServerDetail
	remoteLabel := widget.NewLabel(fmt.Sprintf("Runs Remotely: %t", serverDetail.Remote))

	var iconWidget fyne.CanvasObject = widget.NewLabel("Icon: N/A")
	if serverDetail.IconURL != nil && *serverDetail.IconURL != "" {
		// TODO: Implement image loading/display if desired
		iconWidget = widget.NewLabel(fmt.Sprintf("Icon URL: %s", *serverDetail.IconURL))
	}

	var deployWidget fyne.CanvasObject = widget.NewLabel("Deployment URL: N/A")
	if serverDetail.DeploymentURL != nil && *serverDetail.DeploymentURL != "" {
		deployLink, err := url.Parse(*serverDetail.DeploymentURL)
		if err == nil {
			deployWidget = widget.NewHyperlink(*serverDetail.DeploymentURL, deployLink)
		}
	}

	configSchemaLabel := widget.NewLabel("Config Schema:")
	configSchemaEntry := widget.NewMultiLineEntry()
	configSchemaEntry.SetText(string(serverDetail.ConfigSchema)) // Display raw JSON
	configSchemaEntry.Disable()
	configSchemaEntry.Wrapping = fyne.TextWrapWord
	configSchemaScroll := container.NewVScroll(configSchemaEntry)
	configSchemaScroll.SetMinSize(fyne.NewSize(0, 80))

	connectionsLabel := widget.NewLabel("Connections:")
	connectionsBox := container.NewVBox()
	if len(serverDetail.Connections) > 0 {
		for _, conn := range serverDetail.Connections {
			connInfo := fmt.Sprintf("- Type: %s", conn.Type)
			if conn.URL != nil {
				connInfo += fmt.Sprintf(", URL: %s", *conn.URL)
			}
			// Could add config schema preview here too
			connectionsBox.Add(widget.NewLabel(connInfo))
		}
	} else {
		connectionsBox.Add(widget.NewLabel("- None"))
	}

	securityLabel := widget.NewLabel("Security Scan Passed:")
	var securityWidget fyne.CanvasObject = widget.NewLabel("N/A")
	if serverDetail.Security != nil {
		securityWidget = widget.NewLabel(fmt.Sprintf("%t", serverDetail.Security.ScanPassed))
	}

	toolsLabel := widget.NewLabel("Provided Tools:")
	toolsBox := container.NewVBox()
	if len(serverDetail.Tools) > 0 {
		for _, tool := range serverDetail.Tools {
			// Tool Name (Bold)
			toolNameLabel := widget.NewLabelWithStyle(tool.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			toolsBox.Add(toolNameLabel)

			// Tool Description (if exists, wrapped)
			if tool.Description != nil && *tool.Description != "" {
				toolDescLabel := widget.NewLabel(*tool.Description)
				toolDescLabel.Wrapping = fyne.TextWrapWord
				toolsBox.Add(toolDescLabel)
			}

			// Separator between tools
			toolsBox.Add(widget.NewSeparator())
		}
	} else {
		toolsBox.Add(widget.NewLabel("- None listed"))
	}

	// UseCount, IsDeployed, CreatedAt are not in ServerDetail according to docs
	// useCountLabel := widget.NewLabel(fmt.Sprintf("Use Count: %d", server.UseCount))
	// deployedLabel := widget.NewLabel(fmt.Sprintf("Is Deployed: %t", server.IsDeployed))
	// createdLabel := widget.NewLabel(fmt.Sprintf("Created At: %s", server.CreatedAt))

	details := container.NewVBox(
		nameLabel,
		qualifiedNameLabel,
		widget.NewSeparator(),
		descScroll, // Shows placeholder
		// homepageWidget,
		remoteLabel,
		iconWidget,
		deployWidget,
		widget.NewSeparator(),
		securityLabel,
		securityWidget,
		widget.NewSeparator(),
		connectionsLabel,
		connectionsBox,
		widget.NewSeparator(),
		toolsLabel,
		toolsBox,
		widget.NewSeparator(),
		configSchemaLabel,
		configSchemaScroll,
		// useCountLabel,
		// deployedLabel,
		// createdLabel,
	)

	// Use Border layout: Back button at bottom, details take up the rest
	// Wrap details in a scroll container for potentially long content
	detailScroll := container.NewVScroll(details)
	// Return just the scrollable content, the dialog provides the border/buttons
	return detailScroll
}

// showAddSmitheryMCPDialog shows a confirmation dialog to add a Smithery MCP
// Takes smithery.Server as input (this remains correct, uses list data)
func showAddSmitheryMCPDialog(parentWindow fyne.Window, server smithery.Server) {
	mcpName := server.QualifiedName
	command := "npx"
	args := []string{
		"-y",
		"@smithery/cli@latest",
		"run",
		server.QualifiedName, // The server to run
		// We will add --key later when writing the final config based on settings
	}

	argsJSONBytes, _ := json.Marshal(args)
	argsJSONString := string(argsJSONBytes)

	message := fmt.Sprintf("Add MCP '%s'?\nCommand: %s\nArgs: %s\n(API key will be added automatically if set in Settings)",
		mcpName, command, argsJSONString)

	dialog.ShowConfirm("Add Smithery MCP", message, func(confirm bool) {
		if !confirm {
			return
		}

		newMCP := MCP{
			Name:     mcpName,
			Command:  command,
			ArgsJSON: argsJSONString,
			Enabled:  true, // Add as enabled by default when added from Smithery tab
		}

		if err := addMCP(newMCP); err != nil {
			log.Printf("ERROR adding Smithery MCP '%s': %v", mcpName, err)
			if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "already exists") {
				dialog.ShowError(fmt.Errorf("MCP with name '%s' already exists in your config.", mcpName), parentWindow)
			} else {
				dialog.ShowError(fmt.Errorf("failed to save new MCP to database: %w", err), parentWindow)
			}
			return
		}

		log.Printf("Successfully added Smithery MCP: %s", mcpName)
		dialog.ShowInformation("Success", fmt.Sprintf("MCP '%s' added. Apply changes to update Claude.", mcpName), parentWindow)

		// Refresh the MCP manager list immediately
		go refreshMCPListUI()

	}, parentWindow)
}

// buildUI creates the main application window and its content.
func buildUI(a fyne.App) fyne.Window {
	w := a.NewWindow("MC PeePee - Claude MCP Manager")
	w.Resize(fyne.NewSize(600, 500))

	// --- Build MCP Manager Tab ---
	mcpList = container.NewVBox()
	mcpCheckMap = make(map[string]*widget.Check)

	// Initial sync DB <-> Config File
	if err := syncDBWithClaudeConfig(); err != nil { // Uses func from config.go
		log.Printf("Initial sync failed: %v", err)
		dialog.ShowInformation("Sync Warning", fmt.Sprintf("Initial sync failed: %s\nThis might happen if the config file is malformed.\nPlease check the logs or the file itself.", err.Error()), w)
	}
	// Load initial state into MCP Manager UI
	if err := refreshMCPListUI(); err != nil {
		dialog.ShowError(fmt.Errorf("CRITICAL: Failed to load MCPs for UI: %w", err), w)
		log.Printf("CRITICAL: Failed to load MCPs for UI: %v", err)
		// Handle error, maybe disable the tab?
	}

	// MCP Manager Buttons...
	applyButton := widget.NewButton("Apply to Claude", func() {
		log.Println("Apply button clicked")
		if err := writeClaudeConfig(); err != nil { // Uses func from config.go
			log.Printf("ERROR applying changes to Claude config: %v", err)
			dialog.ShowError(fmt.Errorf("Failed to write config: %w", err), w)
		} else {
			log.Println("Successfully applied changes to Claude config.")
			dialog.ShowInformation("Success", "Configuration applied to Claude.", w)
		}
	})
	addNewButton := widget.NewButton("Add New MCP", func() { showAddMCPDialog(w) })
	openInCursorButton := widget.NewButton("Open Config", func() {
		go func() {
			cmd := exec.Command("cursor", claudeConfigPath) // Uses var from config.go
			if err := cmd.Start(); err != nil {
				log.Printf("ERROR trying to open config in cursor: %v", err)
				dialog.ShowError(fmt.Errorf("Failed to open config in Cursor: %v", err), w)
			}
		}()
	})
	openLogFolderButton := widget.NewButton("Open Logs", func() {
		logPath, err := getClaudeLogPath() // Uses func from config.go
		if err != nil {
			dialog.ShowError(fmt.Errorf("Could not determine log path: %w", err), w)
			return
		}
		go func() {
			cmd := exec.Command("open", logPath)
			if err := cmd.Start(); err != nil {
				log.Printf("ERROR opening log folder: %v", err)
				dialog.ShowError(fmt.Errorf("Failed to open log folder: %w", err), w)
			}
		}()
	})
	restartClaudeButton := widget.NewButton("Restart Claude", func() {
		dialog.ShowConfirm("Confirm Restart", "Restart Claude Desktop?", func(confirm bool) {
			if !confirm {
				return
			}
			go func() { /* ... restart logic ... */ }()
		}, w)
	})
	settingsButton := widget.NewButton("Settings", func() {
		log.Println("Settings button clicked")
		showSettingsDialog(w)
	})

	mcpScrollContainer := container.NewVScroll(mcpList)
	topButtons := container.NewGridWithColumns(5, addNewButton, openInCursorButton, openLogFolderButton, settingsButton, applyButton)
	mcpBottomButtons := container.NewVBox(topButtons, restartClaudeButton)
	mcpManagerContent := container.NewBorder(nil, mcpBottomButtons, nil, nil, mcpScrollContainer)

	// --- Build Smithery Tab ---
	smitheryContent := buildSmitheryTabUI()

	// --- Create Tabs ---
	mcpManagerTab := container.NewTabItem("MCP Manager", mcpManagerContent)
	smitheryTab := container.NewTabItem("Smithery", smitheryContent)
	appTabs := container.NewAppTabs(mcpManagerTab, smitheryTab)

	w.SetContent(appTabs)
	return w
}
