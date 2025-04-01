package models

import (
	"github.com/charmbracelet/bubbletea"
)

// ModelClient defines the interface for AI model clients
type ModelClient interface {
	// GenerateContent sends a prompt to the AI model and returns a tea.Cmd
	// that will produce a response message when executed
	GenerateContent(prompt string) tea.Cmd
}

// ResponseMsg represents a response from an AI model
type ResponseMsg struct {
	Response string
	Err      error
}