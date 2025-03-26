package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModelTemplate struct {
	Name        string
	Description string
}

// Il 'model' rappresenta lo stato della tua applicazione TUI.
type model struct {
	templates    []ModelTemplate
	cursor       int
	selected     bool
	selectedItem int
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
)

// initialModel crea e restituisce lo stato iniziale del tuo modello.
func initialModel() model {
	return model{
		templates: []ModelTemplate{
			{Name: "model1", Description: "Modello base per applicazioni web"},
			{Name: "model2", Description: "Modello per API REST"},
			{Name: "model3", Description: "Modello per applicazioni CLI"},
			{Name: "model4", Description: "Modello per microservizi"},
		},
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
			}
            
		case "esc", "backspace":
			if m.selected {
				m.selected = false
			}
		}
	}

	return m, nil
}

// View genera la stringa che rappresenta la UI visualizzata nel terminale,
// basandosi sullo stato corrente del modello.
func (m model) View() string {
	if m.selected {
		title := titleStyle.Render("Modello Selezionato")
		selectedModel := fmt.Sprintf("Hai selezionato: %s", m.templates[m.selectedItem].Description)
		info := infoStyle.Render("Ora puoi procedere con la generazione del codice")
		backHint := infoStyle.Render("Premi 'ESC' o 'Backspace' per tornare alla selezione")
		exitHint := exitHintStyle.Render("Premi 'q' o 'Ctrl+C' per uscire.")

		return lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			selectedModel,
			"",
			info,
			backHint,
			"",
			exitHint,
		)
	}

	// Mostra la lista di selezione
	title := titleStyle.Render("Seleziona un Modello")

	// Costruisci la lista di elementi
	var items []string
	for i, t := range m.templates {
		item := fmt.Sprintf("%s - %s", t.Name, t.Description)

		if i == m.cursor {
			items = append(items, selectedItemStyle.Render("> "+item))
		} else {
			items = append(items, itemStyle.Render("  "+item))
		}
	}

	list := lipgloss.JoinVertical(lipgloss.Left, items...)
	info := infoStyle.Render("Usa le frecce su/giù per navigare e premi Enter per selezionare")
	exitHint := exitHintStyle.Render("Premi 'q' o 'Ctrl+C' per uscire.")

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
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Si è verificato un errore durante l'esecuzione: %v\n", err)
		os.Exit(1)
	}
}
