package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// --- Cyberpunk Theme ---

// Define our color palette
var (
	// Sarif Industries inspired: Black, Gold/Orange, White/Gray
	colorBackground  = color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff} // Very dark gray
	colorPrimary     = color.NRGBA{R: 0xff, G: 0xb7, B: 0x00, A: 0xff} // Gold/Orange
	colorInputBg     = color.NRGBA{R: 0x2a, G: 0x2a, B: 0x2a, A: 0xff} // Slightly lighter dark gray
	colorText        = color.NRGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff} // Light gray
	colorPlaceholder = color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xff} // Medium gray
	colorShadow      = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x40} // Transparent black for shadow
	colorHover       = color.NRGBA{R: 0xff, G: 0xb7, B: 0x00, A: 0x20} // Transparent gold for hover overlay
	colorDisabled    = color.NRGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xff} // Darker gray for disabled elements
)

// cyberpunkTheme implements fyne.Theme
type cyberpunkTheme struct{}

// Ensure cyberpunkTheme implements fyne.Theme
var _ fyne.Theme = (*cyberpunkTheme)(nil)

func (t *cyberpunkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return colorBackground
	case theme.ColorNameButton, theme.ColorNameDisabledButton:
		return color.Transparent // Button itself is transparent, color comes from text/hover
	case theme.ColorNameDisabled:
		return colorDisabled
	case theme.ColorNameError:
		return theme.DefaultTheme().Color(theme.ColorNameError, variant) // Use default red for errors
	case theme.ColorNameFocus:
		return colorPrimary // Use primary gold for focus
	case theme.ColorNameForeground:
		return colorText
	case theme.ColorNameHover:
		return colorHover // Use transparent gold overlay for hover
	case theme.ColorNameInputBackground:
		return colorInputBg
	case theme.ColorNamePlaceHolder:
		return colorPlaceholder
	case theme.ColorNamePressed:
		return colorPrimary // Use primary gold for pressed state too
	case theme.ColorNamePrimary:
		return colorPrimary
	case theme.ColorNameScrollBar:
		return colorInputBg // Use input background color for scrollbar track
	case theme.ColorNameShadow:
		return colorShadow
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns the default theme font for now
func (t *cyberpunkTheme) Font(style fyne.TextStyle) fyne.Resource {
	// TODO: Implement custom font later if desired
	return theme.DefaultTheme().Font(style)
}

// Size returns the default theme size for now
func (t *cyberpunkTheme) Size(name fyne.ThemeSizeName) float32 {
	// Adjust padding slightly for a potentially tighter look
	if name == theme.SizeNamePadding {
		return 2
	}
	return theme.DefaultTheme().Size(name)
}

// Icon returns the default theme icon for now
func (t *cyberpunkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// --- End Cyberpunk Theme ---
