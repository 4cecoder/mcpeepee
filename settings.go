package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

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
			// Trigger refresh of the Smithery tab after saving key
			// This needs to be handled carefully now that the function is in a separate file.
			// Option 1: Pass refresh function as argument.
			// Option 2: Use a notification/event system.
			// Option 3: Refresh periodically or on tab switch (simpler but less immediate).
			// For now, let's call the refresh function directly, assuming it's accessible globally (it is via ui.go's global vars)
			refreshSmitheryTab() // Call refresh defined in ui.go
		}
	}

	dia := dialog.NewForm(
		"Settings", "Save", "Cancel",
		formItems, callback, parentWindow,
	)
	dia.Resize(fyne.NewSize(450, 150))
	dia.Show()
}
