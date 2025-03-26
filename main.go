package main

import (
	"fmt"
	"os" // Necessario per os.Exit in caso di errore

	tea "github.com/charmbracelet/bubbletea"
	// Potresti voler importare anche lipgloss per lo stile in futuro:
	// "github.com/charmbracelet/lipgloss"
)

// Il 'model' rappresenta lo stato della tua applicazione TUI.
// Per ora è semplice, ma qui aggiungerai i dati che la tua UI deve gestire.
type model struct {
	// Esempio: potresti aggiungere campi come cursor int, choices []string, ecc.
}

// initialModel crea e restituisce lo stato iniziale del tuo modello.
func initialModel() model {
	return model{
		// Qui puoi inizializzare i campi del tuo modello, se necessario.
	}
}

// Init è un comando (Cmd) da eseguire all'avvio dell'applicazione.
// Spesso è nil se non devi fare nulla di speciale all'inizio (es. caricare dati).
func (m model) Init() tea.Cmd {
	return nil // Nessun comando iniziale per ora
}

// Update gestisce i messaggi (Msg) in arrivo, come input da tastiera, timer, ecc.
// Restituisce il modello aggiornato e un eventuale comando successivo da eseguire.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// È un messaggio dalla tastiera?
	case tea.KeyMsg:
		// Controlla quale tasto è stato premuto.
		switch msg.String() {
		// Se viene premuto 'ctrl+c' o 'q', esci dal programma.
		case "ctrl+c", "q":
			return m, tea.Quit // tea.Quit è il comando per terminare Bubble Tea

			// Qui potresti aggiungere la gestione per altri tasti:
			// case "up", "k":
			//   // Logica per muovere su
			// case "down", "j":
			//   // Logica per muovere giù
			// case "enter":
			//   // Logica per confermare
		}
	}

	// Se il messaggio non è stato gestito, restituisci il modello corrente e nessun comando.
	return m, nil
}

// View genera la stringa che rappresenta la UI visualizzata nel terminale,
// basandosi sullo stato corrente del modello.
func (m model) View() string {
	// Costruisci la stringa della tua UI.
	s := "Ciao da Bubble Tea!\n\n"
	s += "Premi 'q' o 'Ctrl+C' per uscire.\n"

	// Qui aggiungerai la logica per visualizzare lo stato del tuo modello 'm'.
	// Potresti usare Lip Gloss qui per lo stile.

	// Restituisci la stringa finale.
	return s
}

// main è il punto di ingresso del programma.
func main() {
	// Crea il programma Bubble Tea passando il modello iniziale.
	p := tea.NewProgram(initialModel())

	// Esegui il programma. Run() bloccherà finché non viene restituito tea.Quit.
	// Catturiamo eventuali errori durante l'esecuzione.
	if _, err := p.Run(); err != nil {
		fmt.Printf("Si è verificato un errore durante l'esecuzione: %v\n", err)
		os.Exit(1) // Esci se c'è stato un errore
	}
}
