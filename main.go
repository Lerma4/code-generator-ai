package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Add config structures
type Config struct {
	Database DatabaseConfig `json:"database"`
	Gemini   GeminiConfig   `json:"gemini"`
}

type DatabaseConfig struct {
	Driver          string `json:"driver"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	DBName          string `json:"dbname"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
}

type GeminiConfig struct {
	APIKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
}

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
// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response structure from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Function to load configuration
func loadConfig() (Config, error) {
	var config Config
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(configFile, &config)
	return config, err
}

// Function to call Gemini API using the official client
func callGeminiAPI(prompt string) tea.Cmd {
	return func() tea.Msg {
		// Load config
		config, err := loadConfig()
		if err != nil {
			errorMsg := fmt.Sprintf("Error loading config: %v", err)
			log.Error().Msg(errorMsg)
			return apiResponseMsg{err: fmt.Errorf(errorMsg)}
		}

		// Log API call attempt
		log.Info().
			Str("model", config.Gemini.ModelName).
			Int("promptLength", len(prompt)).
			Msg("Attempting API call to Gemini")

		// Create a new client
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(config.Gemini.APIKey))
		if err != nil {
			errorMsg := fmt.Sprintf("Error creating Gemini client: %v", err)
			log.Error().Err(err).Msg("Failed to create Gemini client")
			return apiResponseMsg{err: fmt.Errorf(errorMsg)}
		}
		defer client.Close()

		// Get the model
		model := client.GenerativeModel(config.Gemini.ModelName)
		log.Info().Msg("Created Gemini client and model successfully")

		// Generate content
		log.Info().Msg("Sending request to Gemini API...")
		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			log.Error().Err(err).Msg("Error generating content")
			return apiResponseMsg{err: fmt.Errorf("Error generating content: %v", err)}
		}

		// Log response details
		log.Info().Int("candidates", len(resp.Candidates)).Msg("Response received")

		// Extract text from response
		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			// The genai library uses a different structure, so we need to extract the text differently
			part := resp.Candidates[0].Content.Parts[0]
			responseText, ok := part.(genai.Text)
			if !ok {
				log.Error().Msg("Failed to convert response part to text")
				return apiResponseMsg{err: fmt.Errorf("Failed to convert response part to text")}
			}

			log.Info().Int("responseLength", len(string(responseText))).Msg("Successfully received response")
			return apiResponseMsg{response: string(responseText)}
		}

		log.Error().
			Int("candidates", len(resp.Candidates)).
			Msg("No response from API: empty candidates or parts")
		
		if len(resp.Candidates) > 0 {
			log.Error().Int("firstCandidateParts", len(resp.Candidates[0].Content.Parts)).Msg("Response structure details")
		}
		
		return apiResponseMsg{err: fmt.Errorf("No response from API: empty candidates or parts")}
	}
}

// Message type for API response
type apiResponseMsg struct {
	response string
	err      error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case apiResponseMsg:
		m.loading = false
		if msg.err != nil {
			errorMsg := fmt.Sprintf("API Error: %v", msg.err)
			log.Error().Err(msg.err).Msg("API Error")
			m.errorMsg = errorMsg
			m.selected = false
		} else {
			log.Info().Int("responseLength", len(msg.response)).Msg("API call successful")
			m.apiResponse = msg.response
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
							// Call Gemini API with prompt content
							m.loading = true
							return m, callGeminiAPI(string(promptContent))
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
