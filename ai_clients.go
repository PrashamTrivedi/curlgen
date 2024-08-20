package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/sashabaranov/go-openai"
)

func initOpenAIClient() *openai.Client {
	openAIKey := readConfig(OPENAI_KEY)
	if openAIKey == "" {
		slog.Error("OpenAI key not found. Please set it using the config command.", "error", "missing API key")
		os.Exit(1)
	}
	slog.Debug("Initializing OpenAI client")
	return openai.NewClient(openAIKey)
}

func initAnthropicClient() *Client {
	anthropicKey := readConfig(ANTHROPIC_KEY)
	if anthropicKey == "" {
		slog.Error("Anthropic key not found. Please set it using the config command.", "error", "missing API key")
		os.Exit(1)
	}
	slog.Debug("Initializing Anthropic client")
	anthropicClient, err := NewAnthropicClient(anthropicKey)
	if err != nil {
		slog.Error("Error creating Anthropic client. Please check your API key and network connection.", "error", err)
		os.Exit(1)
	}
	return anthropicClient
}

var modelToClientMap map[string]string

func updateModelToClientMap() error {
	modelToClientMap = make(map[string]string)

	slog.Debug("Updating model to client map")

	// Fetch OpenAI models
	openaiClient := initOpenAIClient()
	openaiModels, err := openaiClient.ListModels(context.Background())
	if err != nil {
		slog.Error("Error fetching OpenAI models", "error", err)
		return fmt.Errorf("error fetching OpenAI models: %v", err)
	}
	for _, model := range openaiModels.Models {
		modelToClientMap[model.ID] = "openai"
		slog.Debug("Added OpenAI model", "model", model.ID)
	}

	// Fetch Anthropic models
	anthropicClient := initAnthropicClient()
	anthropicModels := anthropicClient.ListModels()
	for _, model := range anthropicModels {
		modelToClientMap[model.Name] = "anthropic"
		slog.Debug("Added Anthropic model", "model", model.Name)
	}

	return nil
}

func determineClient(model string) (interface{}, error) {
	if modelToClientMap == nil {
		if err := updateModelToClientMap(); err != nil {
			return nil, err
		}
	}

	clientType, ok := modelToClientMap[model]
	if !ok {
		return nil, fmt.Errorf("unsupported model: %s. Please check the model name and try again", model)
	}

	switch clientType {
	case "openai":
		return initOpenAIClient(), nil
	case "anthropic":
		return initAnthropicClient(), nil
	default:
		return nil, fmt.Errorf("unknown client type for model: %s. This is an internal error, please report it", model)
	}
}

func listModels() {
	if err := updateModelToClientMap(); err != nil {
		fmt.Printf("Error updating model list: %v\n", err)
		return
	}

	fmt.Println("Available Models:")
	for model, clientType := range modelToClientMap {
		fmt.Printf("- %s (%s)\n", model, clientType)
	}
}
