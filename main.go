package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"code-generator-ai/api"
	"code-generator-ai/api/models"
)

type ModelTemplate struct {
	Name string
}

type model struct {
	templates    []ModelTemplate
	cursor       int
	selected     bool
	selectedItem int
	errorMsg     string
	apiResponse  string
	loading      bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			PaddingLeft(4).
			PaddingRight(4).
			MarginBottom(1)

	infoStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#ABABAB"))

	exitHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7D56F4")).
				Bold(true)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DDDDDD"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#FF0000")).
			PaddingLeft(2).
			PaddingRight(2).
			MarginTop(1).
			MarginBottom(1)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true)
)

// initialModel crea e restituisce lo stato iniziale del tuo modello.
func getTemplatesFromDirectory() []ModelTemplate {
	templates := []ModelTemplate{}
	templateDir := "templates"

	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return templates
	}

	for _, entry := range entries {
		if entry.IsDir() {
			templates = append(templates, ModelTemplate{
				Name: entry.Name(),
			})
		}
	}

	return templates
}

// Update initialModel to use the new function
func initialModel() model {
	return model{
		templates:    getTemplatesFromDirectory(),
		cursor:       0,
		selected:     false,
		selectedItem: -1,
	}
}

// Init è un comando (Cmd) da eseguire all'avvio dell'applicazione.
func (m model) Init() tea.Cmd {
	return nil
}

// Update gestisce i messaggi (Msg) in arrivo, come input da tastiera, timer, ecc.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case models.ResponseMsg:
		m.loading = false
		if msg.Err != nil {
			errorMsg := fmt.Sprintf("API Error: %v", msg.Err)
			log.Error().Err(msg.Err).Msg("API Error")
			m.errorMsg = errorMsg
			m.selected = false
		} else {
			log.Info().Int("responseLength", len(msg.Response)).Msg("API call successful")
			m.apiResponse = msg.Response
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if !m.selected && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if !m.selected && m.cursor < len(m.templates)-1 {
				m.cursor++
			}

		case "enter":
			if !m.selected {
				m.selected = true
				m.selectedItem = m.cursor

				// Check if prompt.txt exists
				if len(m.templates) > 0 {
					templateDir := filepath.Join("templates", m.templates[m.selectedItem].Name)
					promptFile := filepath.Join(templateDir, "prompt.txt")

					log.Info().
						Str("template", m.templates[m.selectedItem].Name).
						Str("promptFile", promptFile).
						Msg("Attempting to use template")

					if _, err := os.Stat(promptFile); os.IsNotExist(err) {
						errorMsg := fmt.Sprintf("Il file prompt.txt non esiste in %s", templateDir)
						log.Error().Str("templateDir", templateDir).Msg("prompt.txt file not found")
						m.errorMsg = errorMsg
						m.selected = false
					} else {
						m.errorMsg = ""
						log.Info().Msg("Found prompt.txt file, attempting to read")

						// Read prompt.txt content
						promptContent, err := os.ReadFile(promptFile)
						if err != nil {
							errorMsg := fmt.Sprintf("Errore nella lettura di prompt.txt: %v", err)
							log.Error().Err(err).Msg("Error reading prompt.txt")
							m.errorMsg = errorMsg
							m.selected = false
						} else {
							log.Info().Int("bytes", len(promptContent)).Msg("Successfully read prompt.txt")

							// Load config
							config, err := api.LoadConfig()
							if err != nil {
								errorMsg := fmt.Sprintf("Error loading config: %v", err)
								log.Error().Msg(errorMsg)
								m.errorMsg = errorMsg
								m.selected = false
								return m, nil
							}

							// Get the appropriate model client
							// For now, we're hardcoding "gemini" as the model name
							client, err := models.GetModelClient("gemini", config)
							if err != nil {
								errorMsg := fmt.Sprintf("Error getting model client: %v", err)
								log.Error().Msg(errorMsg)
								m.errorMsg = errorMsg
								m.selected = false
								return m, nil
							}

							// Call the model API with prompt content
							m.loading = true
							return m, client.GenerateContent(string(promptContent))
						}
					}
				}
			}

		case "esc", "backspace":
			if m.selected {
				m.selected = false
				m.errorMsg = ""
				m.apiResponse = ""
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	// Check if there are no templates
	if len(m.templates) == 0 {
		title := titleStyle.Render("Seleziona un Modello")
		noTemplates := infoStyle.Render("Nessun template rilevato")
		exitHint := exitHintStyle.Render("Premi 'q' o 'Ctrl+C' per uscire.")

		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			noTemplates,
			"",
			exitHint,
		)
	}

	if m.selected {
		title := titleStyle.Render("Modello Selezionato")
		selectedModel := fmt.Sprintf("Hai selezionato: %s", m.templates[m.selectedItem].Name)

		var content string
		if m.loading {
			content = loadingStyle.Render("Caricamento risposta dall'API Gemini...")
		} else if m.apiResponse != "" {
			content = "Risposta API:\n\n" + m.apiResponse
		} else {
			content = infoStyle.Render("In attesa della risposta API...")
		}

		backHint := infoStyle.Render("Premi 'ESC' o 'Backspace' per tornare alla selezione")
		exitHint := exitHintStyle.Render("Premi 'q' o 'Ctrl+C' per uscire.")

		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			selectedModel,
			"",
			content,
			"",
			backHint,
			"",
			exitHint,
		)
	}

	// Mostra la lista di selezione
	title := titleStyle.Render("Seleziona un Modello")

	// Show error message if present
	errorSection := ""
	if m.errorMsg != "" {
		errorSection = errorStyle.Render(m.errorMsg)
	}

	// Costruisci la lista di elementi
	var items []string
	for i, t := range m.templates {
		if i == m.cursor {
			items = append(items, selectedItemStyle.Render("> "+t.Name))
		} else {
			items = append(items, itemStyle.Render("  "+t.Name))
		}
	}

	list := lipgloss.JoinVertical(lipgloss.Left, items...)
	info := infoStyle.Render("Usa le frecce su/giù per navigare e premi Enter per selezionare")
	exitHint := exitHintStyle.Render("Premi 'q' o 'Ctrl+C' per uscire.")

	// Include error message in the output if present
	if errorSection != "" {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			list,
			"",
			errorSection,
			"",
			info,
			"",
			exitHint,
		)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		list,
		"",
		info,
		"",
		exitHint,
	)
}

// main è il punto di ingresso del programma.
func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339 // Use RFC3339 format (YYYY-MM-DDTHH:MM:SSZ)

	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Printf("Failed to create logs directory: %v\n", err)
		os.Exit(1)
	}

	// Get current date in American format (YYYY-MM-DD)
	currentDate := time.Now().Format("2006-01-02")
	logFileName := filepath.Join(logsDir, currentDate+".log")

	// Open log file
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	// Set up logger to write only to file (no console output)
	log.Logger = zerolog.New(logFile).With().Timestamp().Logger()

	log.Info().Msg("Application starting")

	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Error().Err(err).Msg("Application error")
		fmt.Printf("Si è verificato un errore durante l'esecuzione: %v\n", err)
		os.Exit(1)
	}
}
