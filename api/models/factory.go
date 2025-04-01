package models

import (
	"code-generator-ai/api"
	"fmt"
)

// GetModelClient returns a ModelClient based on the model name
func GetModelClient(modelName string, config api.Config) (ModelClient, error) {
	switch modelName {
	case "gemini":
		return NewGeminiClient(config.Gemini), nil
	default:
		return nil, fmt.Errorf("unsupported model: %s", modelName)
	}
}
