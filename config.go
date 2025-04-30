package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	claudeConfigDirName  = "Claude"
	claudeConfigFileName = "claude_desktop_config.json"
)

var (
	claudeConfigPath string
)

// MCPServer represents the structure within the "mcpServers" map in the JSON config file
type MCPServer struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// SetClaudeConfigPath determines and stores the path to the configuration file.
func SetClaudeConfigPath() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	// Standard path for macOS Application Support
	claudeConfigPath = filepath.Join(home, "Library", "Application Support", claudeConfigDirName, claudeConfigFileName)
	log.Println("Determined Claude config path:", claudeConfigPath)
	return nil
}

// getClaudeLogPath returns the path to the Claude log directory.
func getClaudeLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	// Standard path for macOS logs
	logPath := filepath.Join(home, "Library", "Logs", claudeConfigDirName)
	return logPath, nil
}

// readClaudeConfig reads and parses the claude_desktop_config.json file
func readClaudeConfig() (map[string]MCPServer, map[string]json.RawMessage, error) {
	if claudeConfigPath == "" {
		return nil, nil, fmt.Errorf("Claude config path not set")
	}
	log.Println("Reading Claude config from:", claudeConfigPath)
	data, err := os.ReadFile(claudeConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Claude config file does not exist.")
			return make(map[string]MCPServer), make(map[string]json.RawMessage), nil // Return empty maps, not an error
		}
		return nil, nil, fmt.Errorf("failed to read claude config file: %w", err)
	}

	// Unmarshal into a generic map first to separate known keys from others
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal claude config into raw map: %w", err)
	}

	mcpServers := make(map[string]MCPServer)
	otherConfigs := make(map[string]json.RawMessage)

	for key, rawValue := range rawMap {
		if key == "mcpServers" {
			if err := json.Unmarshal(rawValue, &mcpServers); err != nil {
				log.Printf("WARN: failed to unmarshal mcpServers section: %v", err)
				// Continue processing other keys even if mcpServers fails
			}
		} else {
			otherConfigs[key] = rawValue
		}
	}

	log.Printf("Read %d MCP servers and %d other configs from file.", len(mcpServers), len(otherConfigs))
	return mcpServers, otherConfigs, nil
}

// syncDBWithClaudeConfig updates the database based on the content of claude_desktop_config.json
func syncDBWithClaudeConfig() error {
	log.Println("Syncing database with Claude config...")
	mcpServersFromFile, otherConfigsFromFile, err := readClaudeConfig()
	if err != nil {
		return fmt.Errorf("failed to read Claude config for sync: %w", err)
	}

	// Get all MCPs currently in the DB
	currentDBMCPs, err := loadMCPsFromDB()
	if err != nil {
		return fmt.Errorf("failed to fetch MCPs from DB for sync: %w", err)
	}
	dbMCPMap := make(map[string]*MCP) // Map name to pointer for easy update
	for i := range currentDBMCPs {
		dbMCPMap[currentDBMCPs[i].Name] = &currentDBMCPs[i]
	}

	// Process MCPs found in the config file
	for name, serverData := range mcpServersFromFile {
		argsBytes, err := json.Marshal(serverData.Args)
		if err != nil {
			log.Printf("WARN: Failed to marshal args for MCP '%s' from file: %v. Skipping update.", name, err)
			continue
		}
		argsJSON := string(argsBytes)

		if existingMCP, found := dbMCPMap[name]; found {
			// Found in DB: Update command/args if changed, mark as enabled
			needsDBUpdate := false
			if existingMCP.Command != serverData.Command {
				existingMCP.Command = serverData.Command
				needsDBUpdate = true
			}
			if existingMCP.ArgsJSON != argsJSON {
				existingMCP.ArgsJSON = argsJSON
				needsDBUpdate = true
			}
			if !existingMCP.Enabled { // If it was disabled, mark as enabled because it's in the file
				needsDBUpdate = true
			}

			if needsDBUpdate {
				log.Printf("Updating existing MCP '%s' (found in config file, needs update or enabling)", name)
				mcpToUpdate := *existingMCP // Copy struct
				mcpToUpdate.Enabled = true  // Ensure it's marked enabled
				// Use addMCP logic which handles updates implicitly via GORM's Save or a dedicated Update function
				// For simplicity here, let's use a direct update based on ID.
				if err := db.Model(&MCP{}).Where("id = ?", existingMCP.ID).Updates(map[string]interface{}{
					"command":   mcpToUpdate.Command,
					"args_json": mcpToUpdate.ArgsJSON,
					"enabled":   true,
				}).Error; err != nil {
					log.Printf("ERROR updating MCP '%s' in DB: %v", name, err)
				}
			} else {
				log.Printf("MCP '%s' already up-to-date and enabled.", name)
			}
			delete(dbMCPMap, name) // Remove from map as it's been processed
		} else {
			// Not found in DB: Create new MCP record, mark as enabled
			log.Printf("Adding new MCP '%s' from config file to DB", name)
			newMCP := MCP{
				Name:     name,
				Command:  serverData.Command,
				ArgsJSON: argsJSON,
				Enabled:  true, // It's in the file, so it's considered enabled initially
			}
			if err := addMCP(newMCP); err != nil {
				// addMCP logs the error internally
				log.Printf("Error adding new MCP '%s' during sync: %v", name, err)
			}
		}
	}

	// Process MCPs remaining in dbMCPMap (these were in DB but NOT in the config file)
	for name, mcp := range dbMCPMap {
		if mcp.Enabled { // Only update if it's currently marked as enabled in DB
			log.Printf("Disabling MCP '%s' (not found in config file)", name)
			if err := updateMCPEnabledStatus(mcp.ID, false); err != nil {
				log.Printf("ERROR disabling MCP '%s': %v", name, err)
			}
		}
	}

	// Process other configs found in the file
	// We'll adopt an "upsert" strategy: update if exists, insert if not.
	currentOtherConfigs, err := loadOtherConfigsFromDB()
	if err != nil {
		return fmt.Errorf("failed to load other configs from DB for sync: %w", err)
	}
	otherConfigMap := make(map[string]OtherConfig)
	for _, oc := range currentOtherConfigs {
		otherConfigMap[oc.Key] = oc
	}

	for key, rawValue := range otherConfigsFromFile {
		dataJSON := string(rawValue)
		if existingOther, found := otherConfigMap[key]; found {
			if existingOther.DataJSON != dataJSON {
				log.Printf("Updating other config '%s' in DB", key)
				existingOther.DataJSON = dataJSON
				if err := saveOtherConfig(existingOther); err != nil {
					log.Printf("ERROR updating other config '%s': %v", key, err)
				}
			}
		} else {
			log.Printf("Adding new other config '%s' to DB", key)
			newOther := OtherConfig{
				Key:      key,
				DataJSON: dataJSON,
			}
			if err := saveOtherConfig(newOther); err != nil {
				log.Printf("ERROR creating new other config '%s': %v", key, err)
			}
		}
	}

	log.Println("Database sync with Claude config finished.")
	return nil
}

// writeClaudeConfig writes the enabled MCPs and other configs from DB to claude_desktop_config.json
func writeClaudeConfig() error {
	if claudeConfigPath == "" {
		return fmt.Errorf("Claude config path not set")
	}
	log.Println("Writing current state to Claude config:", claudeConfigPath)

	// 1. Fetch enabled MCPs from DB
	enabledMCPs, err := loadMCPsFromDB() // Load all first
	if err != nil {
		return fmt.Errorf("failed to fetch MCPs from DB: %w", err)
	}

	// Filter only enabled ones for writing
	mcpServersMap := make(map[string]MCPServer)
	for _, mcp := range enabledMCPs {
		if !mcp.Enabled {
			continue // Skip disabled MCPs
		}
		var args []string
		if err := json.Unmarshal([]byte(mcp.ArgsJSON), &args); err != nil {
			log.Printf("WARN: Failed to unmarshal args for enabled MCP '%s': %v. Skipping.", mcp.Name, err)
			continue
		}
		mcpServersMap[mcp.Name] = MCPServer{
			Command: mcp.Command,
			Args:    args,
		}
	}

	// 2. Fetch all OtherConfig records from DB
	otherConfigs, err := loadOtherConfigsFromDB()
	if err != nil {
		return fmt.Errorf("failed to fetch other configs from DB: %w", err)
	}

	// 3. Construct the final output map
	outputMap := make(map[string]interface{})
	if len(mcpServersMap) > 0 {
		outputMap["mcpServers"] = mcpServersMap
	}

	for _, other := range otherConfigs {
		var dataValue interface{}
		if err := json.Unmarshal([]byte(other.DataJSON), &dataValue); err != nil {
			log.Printf("WARN: Failed to unmarshal data for other config '%s': %v. Storing as raw string.", other.Key, err)
			outputMap[other.Key] = other.DataJSON // Fallback
		} else {
			outputMap[other.Key] = dataValue
		}
	}

	// 4. Apply Smithery API Key override if present
	savedAPIKey, err := getSetting("SmitheryAPIKey")
	if err != nil {
		log.Printf("WARN: Could not retrieve SmitheryAPIKey from settings during write: %v", err)
	} else if savedAPIKey != "" {
		if mcpMapForWrite, ok := outputMap["mcpServers"].(map[string]MCPServer); ok {
			for name, serverData := range mcpMapForWrite {
				usesSmitheryCLI := false
				hasKeyArg := false
				keyArgIndex := -1
				currentArgs := serverData.Args // Work with a copy?
				for i, arg := range currentArgs {
					if strings.Contains(arg, "@smithery/cli") {
						usesSmitheryCLI = true
					}
					if arg == "--key" {
						hasKeyArg = true
						keyArgIndex = i
					}
				}

				if usesSmitheryCLI {
					modifiedArgs := currentArgs // Start with original args
					if hasKeyArg && keyArgIndex+1 < len(modifiedArgs) {
						log.Printf("Updating --key for MCP '%s' with saved Smithery API Key for write.", name)
						modifiedArgs[keyArgIndex+1] = savedAPIKey
					} else if hasKeyArg {
						log.Printf("Found --key without value for MCP '%s', appending saved key for write.", name)
						modifiedArgs = append(modifiedArgs, savedAPIKey)
					} else {
						log.Printf("Adding --key to MCP '%s' with saved Smithery API Key for write.", name)
						modifiedArgs = append(modifiedArgs, "--key", savedAPIKey)
					}
					// Update the map with modified args
					serverData.Args = modifiedArgs
					mcpMapForWrite[name] = serverData
				}
			}
			outputMap["mcpServers"] = mcpMapForWrite // Put the potentially modified map back
		}
	}

	// 5. Marshal the map to JSON
	outputJSON, err := json.MarshalIndent(outputMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output config to JSON: %w", err)
	}

	// 6. Create parent directories if they don't exist
	configDir := filepath.Dir(claudeConfigPath)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory '%s': %w", configDir, err)
	}

	// 7. Write the JSON to claudeConfigPath
	if err := os.WriteFile(claudeConfigPath, outputJSON, 0640); err != nil {
		return fmt.Errorf("failed to write config file '%s': %w", claudeConfigPath, err)
	}

	log.Println("Successfully wrote config to", claudeConfigPath)
	return nil
}
