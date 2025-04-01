package models

import (
	"code-generator-ai/api"
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/generative-ai-go/genai"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
)

// GeminiClient implements the ModelClient interface for Gemini API
type GeminiClient struct {
	APIKey    string
	ModelName string
}

// NewGeminiClient creates a new Gemini client with the provided configuration
func NewGeminiClient(config api.GeminiConfig) *GeminiClient {
	return &GeminiClient{
		APIKey:    config.APIKey,
		ModelName: config.ModelName,
	}
}

// GenerateContent implements the ModelClient interface
func (g *GeminiClient) GenerateContent(prompt string) tea.Cmd {
	return func() tea.Msg {
		// Log API call attempt
		log.Info().
			Str("model", g.ModelName).
			Int("promptLength", len(prompt)).
			Msg("Attempting API call to Gemini")

		// Create a new client
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithAPIKey(g.APIKey))
		if err != nil {
			errorMsg := fmt.Sprintf("Error creating Gemini client: %v", err)
			log.Error().Err(err).Msg("Failed to create Gemini client")
			return ResponseMsg{Err: fmt.Errorf(errorMsg)}
		}
		defer client.Close()

		// Get the model
		model := client.GenerativeModel(g.ModelName)
		log.Info().Msg("Created Gemini client and model successfully")

		// Generate content
		log.Info().Msg("Sending request to Gemini API...")
		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			log.Error().Err(err).Msg("Error generating content")
			return ResponseMsg{Err: fmt.Errorf("Error generating content: %v", err)}
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
				return ResponseMsg{Err: fmt.Errorf("Failed to convert response part to text")}
			}

			log.Info().Int("responseLength", len(string(responseText))).Msg("Successfully received response")
			return ResponseMsg{Response: string(responseText)}
		}

		log.Error().
			Int("candidates", len(resp.Candidates)).
			Msg("No response from API: empty candidates or parts")

		if len(resp.Candidates) > 0 {
			log.Error().Int("firstCandidateParts", len(resp.Candidates[0].Content.Parts)).Msg("Response structure details")
		}

		return ResponseMsg{Err: fmt.Errorf("No response from API: empty candidates or parts")}
	}
}
