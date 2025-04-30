package main

import (
	"log"

	"fyne.io/fyne/v2/app"
	// "fyne.io/fyne/v2/theme" // Keep commented out if not used directly in main.go
)

// MCP represents a single MCP server configuration
// type MCP struct {
// 	gorm.Model
// 	Name     string `gorm:"uniqueIndex"` // MCP name like "desktop-commander"
// 	Command  string
// 	ArgsJSON string // Store args as JSON string
// 	Enabled  bool   // Whether this MCP is active in mcp.json
// }

// AppSettings stores application-level settings like API keys
// type AppSettings struct {
// 	gorm.Model
// 	KeyName  string `gorm:"uniqueIndex"` // Setting key (e.g., "SmitheryAPIKey")
// 	KeyValue string // Setting value
// }

// OtherConfig represents other top-level keys in mcp.json
// type OtherConfig struct {
// 	gorm.Model
// 	Key      string `gorm:"uniqueIndex"` // Top-level key like "ghidra"
// 	DataJSON string // Store the value as JSON string
// }

/* // Moved UI Globals to ui.go
var (
	mcpList     *fyne.Container // Container for the checklist
	mcpCheckMap map[string]*widget.Check // Map MCP name to its checkbox
)
*/

/* // Moved to config.go
func getClaudeLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	// Standard path for macOS logs
	// Assuming the folder is named "Claude", adjust if needed
	logPath := filepath.Join(home, "Library", "Logs", claudeConfigDirName)
	return logPath, nil
}
*/

/* // Moved to config.go
func syncDBWithClaudeConfig() error {
	// This function is now primarily for the *initial* sync or a manual refresh.
	// Real-time enabling/disabling happens via checkboxes updating the DB directly.
	// dbMutex.Lock()
	// defer dbMutex.Unlock()

	log.Println("Syncing database with Claude config...")
	mcpServers, otherConfigsFromFile, err := readClaudeConfig()
	if err != nil {
		// Specific handling for "file not found" moved to readClaudeConfig
		return fmt.Errorf("failed to read Claude config for sync: %w", err)
	}

	// Get all MCPs currently in the DB
	// var currentDBMCPs []MCP
	// if err := db.Find(&currentDBMCPs).Error; err != nil {
	// 	return fmt.Errorf("failed to fetch MCPs from DB for sync: %w", err)
	// }
	// dbMCPMap := make(map[string]*MCP)
	// for i := range currentDBMCPs {
	// 	dbMCPMap[currentDBMCPs[i].Name] = &currentDBMCPs[i]
	// }

	// Process MCPs found in the config file
	for name, serverData := range mcpServers {
		argsBytes, err := json.Marshal(serverData.Args)
		if err != nil {
			log.Printf("WARN: Failed to marshal args for MCP '%s': %v. Skipping update.", name, err)
			continue // Skip this MCP if args are invalid
		}
		argsJSON := string(argsBytes)

		// if existingMCP, found := dbMCPMap[name]; found {
		// 	// Found in DB: Update command/args if changed, mark as enabled
		// 	updates := map[string]interface{}{"enabled": true}
		// 	changed := false
		// 	if existingMCP.Command != serverData.Command {
		// 		updates["command"] = serverData.Command
		// 		changed = true
		// 	}
		// 	if existingMCP.ArgsJSON != argsJSON {
		// 		updates["args_json"] = argsJSON
		// 		changed = true
		// 	}
		// 	if changed || !existingMCP.Enabled { // Update if data changed or if it was disabled
		// 		log.Printf("Updating existing MCP '%s' (found in config file)", name)
		// 		if err := db.Model(existingMCP).Updates(updates).Error; err != nil {
		// 			log.Printf("ERROR updating MCP '%s': %v", name, err)
		// 		}
		// 	} else {
		// 		log.Printf("MCP '%s' already up-to-date and enabled.", name)
		// 	}
		// 	delete(dbMCPMap, name) // Remove from map as it's been processed
		// } else {
		// 	// Not found in DB: Create new MCP record, mark as enabled
		// 	log.Printf("Adding new MCP '%s' from config file", name)
		// 	newMCP := MCP{
		// 		Name:     name,
		// 		Command:  serverData.Command,
		// 		ArgsJSON: argsJSON,
		// 		Enabled:  true,
		// 	}
		// 	if err := db.Create(&newMCP).Error; err != nil {
		// 		log.Printf("ERROR creating new MCP '%s': %v", name, err)
		// 	}
		// }
	}

	// Process MCPs remaining in dbMCPMap (these were in DB but NOT in the config file)
	// for name, mcp := range dbMCPMap {
	// 	if mcp.Enabled { // Only update if it's currently marked as enabled
	// 		log.Printf("Disabling MCP '%s' (not found in config file)", name)
	// 		if err := db.Model(mcp).Update("enabled", false).Error; err != nil {
	// 			log.Printf("ERROR disabling MCP '%s': %v", name, err)
	// 		}
	// 	}
	// }

	// Process other configs found in the file
	// We'll adopt an "upsert" strategy: update if exists, insert if not.
	// We won't delete configs from the DB that aren't in the file currently.
	for key, rawValue := range otherConfigsFromFile {
		dataJSON := string(rawValue)
		// var existingOther OtherConfig
		// result := db.Where("key = ?", key).First(&existingOther)

		// if result.Error == nil { // Found it
		// 	if existingOther.DataJSON != dataJSON {
		// 		log.Printf("Updating other config '%s'", key)
		// 		if err := db.Model(&existingOther).Update("data_json", dataJSON).Error; err != nil {
		// 			log.Printf("ERROR updating other config '%s': %v", key, err)
		// 		}
		// 	} else {
		// 		//log.Printf("Other config '%s' is already up-to-date.", key)
		// 	}
		// } else if result.Error == gorm.ErrRecordNotFound { // Not found
		// 	log.Printf("Adding new other config '%s'", key)
		// 	newOther := OtherConfig{
		// 		Key:      key,
		// 		DataJSON: dataJSON,
		// 	}
		// 	if err := db.Create(&newOther).Error; err != nil {
		// 		log.Printf("ERROR creating new other config '%s': %v", key, err)
		// 	}
		// } else { // Other DB error
		// 	log.Printf("ERROR checking for other config '%s': %v", key, result.Error)
		// }
	}

	log.Println("Database sync finished.")
	return nil
}

/* // Moved to config.go
func writeClaudeConfig() error {
	// dbMutex.Lock()
	// defer dbMutex.Unlock()

	log.Println("Writing current state to Claude config:", claudeConfigPath)

	// 1. Fetch enabled MCPs
	// var enabledMCPs []MCP
	// if err := db.Where("enabled = ?", true).Order("name asc").Find(&enabledMCPs).Error; err != nil {
	// 	return fmt.Errorf("failed to fetch enabled MCPs: %w", err)
	// }

	// 2. Fetch all OtherConfig records
	// var otherConfigs []OtherConfig
	// if err := db.Order("key asc").Find(&otherConfigs).Error; err != nil {
	// 	return fmt.Errorf("failed to fetch other configs: %w", err)
	// }

	// 3. Fetch the Smithery API Key
	// savedAPIKey, err := getSetting("SmitheryAPIKey")
	// if err != nil {
	// 	// Log the error but proceed, maybe the key isn't set yet
	// 	log.Printf("WARN: Could not retrieve SmitheryAPIKey from settings: %v", err)
	// 	savedAPIKey = "" // Ensure it's empty if retrieval failed
	// }

	// 4. Construct the target structure
	outputMap := make(map[string]interface{})
	mcpServersMap := make(map[string]MCPServer)

	// for _, mcp := range enabledMCPs {
	// 	var args []string
	// 	if err := json.Unmarshal([]byte(mcp.ArgsJSON), &args); err != nil {
	// 		log.Printf("WARN: Failed to unmarshal args for enabled MCP '%s': %v. Skipping.", mcp.Name, err)
	// 		continue // Skip this MCP in the output if args are corrupted
	// 	}

	// 	// Check if this MCP uses smithery CLI and needs the API key override
	// 	usesSmitheryCLI := false
	// 	hasKeyArg := false
	// 	keyArgIndex := -1
	// 	for i, arg := range args {
	// 		if strings.Contains(arg, "@smithery/cli") {
	// 			usesSmitheryCLI = true
	// 		}
	// 		if arg == "--key" {
	// 			hasKeyArg = true
	// 			keyArgIndex = i
	// 			// Don't break, continue checking for @smithery/cli
	// 		}
	// 	}

	// 	// If it uses smithery CLI and we have a saved API key, modify the args
	// 	if usesSmitheryCLI && savedAPIKey != "" {
	// 		if hasKeyArg && keyArgIndex+1 < len(args) {
	// 			// Found existing --key, replace its value
	// 			log.Printf("Updating --key for MCP '%s' with saved Smithery API Key.", mcp.Name)
	// 			args[keyArgIndex+1] = savedAPIKey
	// 		} else if hasKeyArg && keyArgIndex+1 >= len(args) {
	// 			// Found --key but it has no value (unlikely but handle defensively)
	// 			log.Printf("WARN: Found --key argument without a value for MCP '%s'. Appending saved Smithery API Key.", mcp.Name)
	// 			args = append(args, savedAPIKey) // Append the key
	// 		} else {
	// 			// Did not find --key, add it and the value
	// 			log.Printf("Adding --key to MCP '%s' with saved Smithery API Key.", mcp.Name)
	// 			args = append(args, "--key", savedAPIKey)
	// 		}
	// 	} else if usesSmitheryCLI && savedAPIKey == "" && hasKeyArg && keyArgIndex+1 < len(args) {
	// 		// If no API key is saved in settings, but the config has one, keep the one from the config
	// 		log.Printf("No Smithery API Key saved in settings. Keeping existing --key for MCP '%s'.", mcp.Name)
	// 	} else if usesSmitheryCLI && savedAPIKey == "" && hasKeyArg && keyArgIndex+1 >= len(args) {
	// 		// Remove --key if it has no value and no key is saved
	// 		log.Printf("Removing --key argument with no value for MCP '%s' as no Smithery API Key is saved.", mcp.Name)
	// 		if keyArgIndex >= 0 {
	// 			args = append(args[:keyArgIndex], args[keyArgIndex+1:]...)
	// 		}
	// 	} else if usesSmitheryCLI && savedAPIKey == "" && !hasKeyArg {
	// 		// No saved key, no existing key arg, do nothing.
	// 	}

	mcpServersMap[name] = MCPServer{
		Command: serverData.Command,
		Args:    serverData.Args,
	}
	// }
	if len(mcpServersMap) > 0 {
		outputMap["mcpServers"] = mcpServersMap
	}

	// for _, other := range otherConfigs {
	// 	var dataValue interface{}
	// 	// Attempt to unmarshal to preserve original JSON types (object, array, string, etc.)
	// 	if err := json.Unmarshal([]byte(other.DataJSON), &dataValue); err != nil {
	// 		log.Printf("WARN: Failed to unmarshal data for other config '%s': %v. Storing as raw string.", other.Key, err)
	// 		// Fallback to storing as a string if unmarshalling fails
	// 		outputMap[other.Key] = other.DataJSON
	// 	} else {
	// 		outputMap[other.Key] = dataValue
	// 	}
	// }

	// 4. Marshal the map to JSON
	outputJSON, err := json.MarshalIndent(outputMap, "", "  ") // Use 2-space indent like example
	if err != nil {
		return fmt.Errorf("failed to marshal output config to JSON: %w", err)
	}

	// 5. Create parent directories if they don't exist
	configDir := filepath.Dir(claudeConfigPath)
	if err := os.MkdirAll(configDir, 0750); err != nil { // Use 0750 permissions
		return fmt.Errorf("failed to create config directory '%s': %w", configDir, err)
	}

	// 6. Write the JSON to claudeConfigPath
	if err := os.WriteFile(claudeConfigPath, outputJSON, 0640); err != nil { // Use 0640 permissions
		return fmt.Errorf("failed to write config file '%s': %w", claudeConfigPath, err)
	}

	log.Println("Successfully wrote config to", claudeConfigPath)
	return nil
}
*/

/* // Moved to db.go (loadMCPsFromDB) and ui.go (UI logic)
func loadMCPsFromDB() error {
	// dbMutex.Lock()
	// defer dbMutex.Unlock()

	var mcps []MCP
	if err := db.Order("name asc").Find(&mcps).Error; err != nil {
		return fmt.Errorf("failed to load MCPs from database: %w", err)
	}

	// Clear existing data first (important for refresh)
	mcpData = make(map[string]*MCP)
	mcpCheckMap = make(map[string]*widget.Check)
	if mcpList != nil {
		mcpList.Objects = nil // Clear UI list items
	} else {
		mcpList = container.NewVBox()
	}

	for i := range mcps {
		mcp := mcps[i] // Create a local copy for the closure
		mcpData[mcp.Name] = &mcp

		check := widget.NewCheck(mcp.Name, func(checked bool) {
			go func(mcpToUpdate *MCP, isChecked bool) { // Run DB update in goroutine
				// dbMutex.Lock()
				// defer dbMutex.Unlock()
				log.Printf("Updating MCP '%s' enabled status to %v", mcpToUpdate.Name, isChecked)
				err := db.Model(&mcpToUpdate).Update("enabled", isChecked).Error
				if err != nil {
					log.Printf("ERROR updating MCP '%s' status: %v", mcpToUpdate.Name, err)
					// Optionally: Show error to user, revert checkbox state
				} else {
					mcpToUpdate.Enabled = isChecked // Update in-memory state as well
					log.Printf("Successfully updated MCP '%s'", mcpToUpdate.Name)
				}

			}(&mcp, checked) // Pass the correct mcp instance
		})
		check.SetChecked(mcp.Enabled)
		mcpCheckMap[mcp.Name] = check
		mcpList.Add(check)
	}

	if mcpList != nil {
		mcpList.Refresh()
	}
	log.Printf("Loaded %d MCPs from DB", len(mcps))
	return nil
}
*/

/* // Moved to ui.go
// refreshMCPListUI loads MCPs from the DB and updates the UI checklist.
func refreshMCPListUI() error {
	mcps, err := loadMCPsFromDB() // Uses func from db.go
	if err != nil {
		return fmt.Errorf("failed to load MCPs from database for UI refresh: %w", err)
	}

	// Ensure UI elements are initialized
	if mcpList == nil {
		mcpList = container.NewVBox()
	}
	if mcpCheckMap == nil {
		mcpCheckMap = make(map[string]*widget.Check)
	}

	// Clear existing UI items
	mcpList.Objects = nil
	// Clear map - important if MCPs can be deleted
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
					// In-memory update is less critical now as we reload from DB on refresh
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
*/

/* // Moved to ui.go
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
*/

/* // Moved to ui.go
// showSettingsDialog displays a dialog for managing application settings.
func showSettingsDialog(parentWindow fyne.Window) {
	apiKeyEntry := widget.NewEntry()
	apiKeyEntry.SetPlaceHolder("Enter your Smithery API Key")
	apiKeyEntry.Password = true

	currentKey, err := getSetting("SmitheryAPIKey") // Uses func from db.go
	if err != nil {
		log.Printf("ERROR loading SmitheryAPIKey for settings dialog: %v", err)
	} else {
		apiKeyEntry.SetText(currentKey)
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("Smithery API Key", apiKeyEntry),
	}

	callback := func(confirmed bool) {
		if !confirmed {
			return
		}
		newKey := apiKeyEntry.Text
		err := saveSetting("SmitheryAPIKey", newKey) // Uses func from db.go
		if err != nil {
			log.Printf("ERROR saving SmitheryAPIKey: %v", err)
			dialog.ShowError(fmt.Errorf("Failed to save API Key: %w", err), parentWindow)
		} else {
			dialog.ShowInformation("Success", "Smithery API Key saved.", parentWindow)
		}
	}

	dia := dialog.NewForm(
		"Settings", "Save", "Cancel",
		formItems, callback, parentWindow,
	)
	dia.Resize(fyne.NewSize(450, 150))
	dia.Show()
}
*/

/* // Moved to ui.go
// buildUI creates the main application window and its content.
func buildUI(a fyne.App) fyne.Window {
	w := a.NewWindow("MC PeePee - Claude MCP Manager")
	w.Resize(fyne.NewSize(600, 500))

	// --- MCP Manager Tab Content ---
	mcpList = container.NewVBox() // Initialize container

	// Initial load and sync (Errors handled below)
	err := syncDBWithClaudeConfig()
	if err != nil {
		log.Printf("Initial sync failed: %v", err)
		// Show non-blocking info dialog, user might need to fix config manually
		dialog.ShowInformation("Sync Warning", fmt.Sprintf("Initial sync failed: %s\nThis might happen if the config file is malformed.\nPlease check the logs or the file itself.", err.Error()), w)
	}
	err = refreshMCPListUI()
	if err != nil {
		// Show fatal error to user and maybe exit?
		dialog.ShowError(fmt.Errorf("CRITICAL: Failed to load MCPs for UI: %w", err), w)
		log.Fatalf("CRITICAL: Failed to load MCPs for UI: %v", err)
		// Alternatively, return w here or disable the MCP tab content
	}

	applyButton := widget.NewButton("Apply to Claude", func() {
		log.Println("Apply button clicked")
		err := writeClaudeConfig()
		if err != nil {
			log.Printf("ERROR applying changes to Claude config: %v", err)
			dialog.ShowError(fmt.Errorf("Failed to write config: %w", err), w)
		} else {
			log.Println("Successfully applied changes to Claude config.")
			dialog.ShowInformation("Success", "Configuration applied to Claude.", w)
		}
	})

	addNewButton := widget.NewButton("Add New MCP", func() {
		log.Println("Add New MCP button clicked")
		showAddMCPDialog(w) // Pass the main window
	})

	openInCursorButton := widget.NewButton("Open Config", func() {
		log.Println("Open Config button clicked. Path:", claudeConfigPath)
		go func() { // Run in goroutine to not block UI thread
			cmd := exec.Command("cursor", claudeConfigPath)
			// Use Start() instead of Run() or Output() as we just want to launch the app
			err := cmd.Start()
			if err != nil {
				log.Printf("ERROR trying to open config in cursor: %v", err)
				// We need to show the dialog on the main thread
				errorMsg := fmt.Sprintf("Failed to open config in Cursor: %v\nIs 'cursor' command in your PATH?", err)
				dialog.ShowError(fmt.Errorf(errorMsg), w)
			}
		}()
	})

	settingsButton := widget.NewButton("Settings", func() {
		log.Println("Settings button clicked")
		showSettingsDialog(w) // This will now also call updateSmitheryTabState on save
	})

	openLogFolderButton := widget.NewButton("Open Logs", func() {
		log.Println("Open Logs button clicked.")
		logPath, err := getClaudeLogPath()
		if err != nil {
			log.Printf("ERROR determining log path: %v", err)
			dialog.ShowError(fmt.Errorf("Could not determine log path: %w", err), w)
			return
		}

		go func() { // Run in goroutine
			log.Println("Attempting to open log folder:", logPath)
			cmd := exec.Command("open", logPath)
			if err := cmd.Start(); err != nil {
				log.Printf("ERROR trying to open log folder: %v", err)
				dialog.ShowError(fmt.Errorf("Failed to open log folder '%s': %w", logPath, err), w)
			}
		}()
	})

	restartClaudeButton := widget.NewButton("Restart Claude", func() {
		log.Println("Restart Claude button clicked")
		dialog.ShowConfirm("Confirm Restart", "Are you sure you want to restart Claude Desktop?", func(confirm bool) {
			if !confirm {
				return
			}
			go func() { // Run in goroutine
				log.Println("Attempting to kill Claude process...")
				killCmd := exec.Command("pkill", "-f", "Claude") // Use -f to match full process name potentially
				killErr := killCmd.Run()                         // Run waits for completion
				if killErr != nil {
					// pkill returns non-zero if no process is found, which isn't necessarily an error for us
					if exitErr, ok := killErr.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
						log.Println("Claude process not found or already terminated.")
					} else {
						log.Printf("Error attempting to kill Claude: %v", killErr)
						dialog.ShowError(fmt.Errorf("Failed to terminate Claude: %w", killErr), w)
						return // Don't attempt restart if kill failed unexpectedly
					}
				} else {
					log.Println("Claude process terminated successfully (or was not running).")
				}

				// Wait a moment for the process to fully exit
				time.Sleep(1 * time.Second)

				log.Println("Attempting to relaunch Claude...")
				// Standard path for applications on macOS
				claudeAppPath := "/Applications/Claude.app"
				restartCmd := exec.Command("open", claudeAppPath)
				if err := restartCmd.Start(); err != nil { // Use Start() to launch and not wait
					log.Printf("Error attempting to relaunch Claude: %v", err)
					dialog.ShowError(fmt.Errorf("Failed to relaunch Claude from %s: %w", claudeAppPath, err), w)
				} else {
					log.Println("Claude relaunch command executed.")
					dialog.ShowInformation("Restarted", "Claude Desktop restart initiated.", w)
				}
			}()
		}, w)
	})

	mcpScrollContainer := container.NewVScroll(mcpList)

	// Button layout remains similar, but now part of the MCP tab
	topButtons := container.NewGridWithColumns(4, addNewButton, openInCursorButton, openLogFolderButton, settingsButton)
	mcpBottomButtons := container.NewGridWithRows(2,
		topButtons,
		container.NewGridWithColumns(2, restartClaudeButton, applyButton),
	)

	mcpManagerContent := container.NewBorder(nil, mcpBottomButtons, nil, nil, mcpScrollContainer)

	// --- Smithery Tab Content ---
	// Placeholder content for now
	smitheryContent := container.NewCenter(
		widget.NewLabel("Smithery features will appear here."),
	)

	// --- Create Tabs ---
	mcpManagerTab := container.NewTabItem("MCP Manager", mcpManagerContent)
	// Create the Smithery tab item and store it globally
	smitheryTab := container.NewTabItem("Smithery", smitheryContent)

	appTabs := container.NewAppTabs(
		mcpManagerTab,
		smitheryTab, // Add the smithery tab
	)

	// Set the main content of the window to the tabs
	w.SetContent(appTabs)
	return w
}
*/

func main() {
	var err error
	err = SetClaudeConfigPath() // Use function from config.go
	if err != nil {
		log.Fatalf("Fatal: Could not determine Claude config path: %v", err)
	}

	err = setupDatabase() // Use function from db.go
	if err != nil {
		log.Fatalf("Fatal: Failed to setup database: %v", err)
	}

	myApp := app.New()

	// Apply the custom theme (Theme defined in theme.go)
	myApp.Settings().SetTheme(&cyberpunkTheme{})

	myWindow := buildUI(myApp)

	myWindow.ShowAndRun()

	log.Println("MC PeePee exiting.")
}
