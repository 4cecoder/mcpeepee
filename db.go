package main

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dbName = "mcpeepee.db"
)

// MCP represents a single MCP server configuration in the database
type MCP struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex"` // MCP name like "desktop-commander"
	Command  string
	ArgsJSON string // Store args as JSON string
	Enabled  bool   // Whether this MCP is active in the final config
}

// AppSettings stores application-level settings like API keys
type AppSettings struct {
	gorm.Model
	KeyName  string `gorm:"uniqueIndex"` // Setting key (e.g., "SmitheryAPIKey")
	KeyValue string // Setting value
}

// OtherConfig represents other top-level keys in the config file stored in the DB
type OtherConfig struct {
	gorm.Model
	Key      string `gorm:"uniqueIndex"` // Top-level key like "ghidra"
	DataJSON string // Store the value as JSON string
}

var (
	db      *gorm.DB
	dbMutex sync.Mutex
)

// setupDatabase initializes the database connection and migrates schemas.
func setupDatabase() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	var err error
	db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Keep GORM quiet
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&MCP{}, &OtherConfig{}, &AppSettings{})
	if err != nil {
		return fmt.Errorf("failed to migrate database schema: %w", err)
	}
	log.Println("Database setup and migration complete.")
	return nil
}

// getSetting retrieves a setting value from the database.
func getSetting(keyName string) (string, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return "", fmt.Errorf("database not initialized")
	}

	var setting AppSettings
	result := db.Where("key_name = ?", keyName).First(&setting)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil // Not found is not an error, just return empty string
		}
		return "", fmt.Errorf("failed to get setting '%s': %w", keyName, result.Error)
	}
	return setting.KeyValue, nil
}

// saveSetting saves or updates a setting value in the database.
func saveSetting(keyName string, keyValue string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	setting := AppSettings{
		KeyName:  keyName,
		KeyValue: keyValue,
	}

	// Use Assign to update if exists (based on KeyName), or create if not.
	if err := db.Where(AppSettings{KeyName: keyName}).Assign(AppSettings{KeyValue: keyValue}).FirstOrCreate(&setting).Error; err != nil {
		return fmt.Errorf("failed to save setting '%s': %w", keyName, err)
	}
	log.Printf("Saved setting: %s", keyName)
	return nil
}

// loadMCPsFromDB retrieves all MCP configurations from the database.
func loadMCPsFromDB() ([]MCP, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var mcps []MCP
	if err := db.Order("name asc").Find(&mcps).Error; err != nil {
		return nil, fmt.Errorf("failed to load MCPs from database: %w", err)
	}
	log.Printf("Loaded %d MCPs from DB", len(mcps))
	return mcps, nil
}

// updateMCPEnabledStatus updates the 'enabled' flag for a specific MCP.
func updateMCPEnabledStatus(mcpID uint, enabled bool) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	result := db.Model(&MCP{}).Where("id = ?", mcpID).Update("enabled", enabled)
	if result.Error != nil {
		return fmt.Errorf("failed to update enabled status for MCP ID %d: %w", mcpID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("MCP with ID %d not found for status update", mcpID)
	}
	log.Printf("Updated enabled status for MCP ID %d to %v", mcpID, enabled)
	return nil
}

// addMCP adds a new MCP configuration to the database.
func addMCP(mcp MCP) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Check for existing name
	var existing MCP
	result := db.Where("name = ?", mcp.Name).First(&existing)
	if result.Error == nil {
		return fmt.Errorf("MCP with name '%s' already exists", mcp.Name)
	} else if result.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("database error checking for existing MCP name '%s': %w", mcp.Name, result.Error)
	}

	// Create the new MCP
	if err := db.Create(&mcp).Error; err != nil {
		return fmt.Errorf("failed to create new MCP '%s': %w", mcp.Name, err)
	}
	log.Printf("Successfully added new MCP: %s", mcp.Name)
	return nil
}

// loadOtherConfigsFromDB retrieves all OtherConfig records from the database.
func loadOtherConfigsFromDB() ([]OtherConfig, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var others []OtherConfig
	if err := db.Order("key asc").Find(&others).Error; err != nil {
		return nil, fmt.Errorf("failed to load other configs from database: %w", err)
	}
	return others, nil
}

// saveOtherConfig adds or updates an OtherConfig record.
func saveOtherConfig(other OtherConfig) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Use Assign to update if exists (based on Key), or create if not.
	if err := db.Where(OtherConfig{Key: other.Key}).Assign(OtherConfig{DataJSON: other.DataJSON}).FirstOrCreate(&other).Error; err != nil {
		return fmt.Errorf("failed to save other config '%s': %w", other.Key, err)
	}
	return nil
}
