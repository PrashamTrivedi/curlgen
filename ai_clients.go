package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

func initOpenAIClient() *openai.Client {
	openAIKey := readConfig(OPENAI_KEY)
	if openAIKey == "" {
		fmt.Println("OpenAI key not found. Please set it using the config command.")
		os.Exit(1)
	}
	return openai.NewClient(openAIKey)
}

func initAnthropicClient() *Client {
	anthropicKey := readConfig(ANTHROPIC_KEY)
	if anthropicKey == "" {
		fmt.Println("Anthropic key not found. Please set it using the config command.")
		os.Exit(1)
	}
	anthropicClient, err := NewAnthropicClient(anthropicKey)
	if err != nil {
		fmt.Println("Error creating Anthropic client:", err)
		os.Exit(1)
	}
	return anthropicClient
}

func determineClient(model string) (interface{}, error) {
	openaiClient := initOpenAIClient()
	anthropicClient := initAnthropicClient()

	openaiModels, err := openaiClient.ListModels(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error listing OpenAI models: %v", err)
	}

	for _, openaiModel := range openaiModels.Models {
		if openaiModel.ID == model {
			return openaiClient, nil
		}
	}

	anthropicModels := anthropicClient.ListModels()
	for _, anthropicModel := range anthropicModels {
		if anthropicModel.Name == model || anthropicModel.ApiName == model {
			return anthropicClient, nil
		}
	}

	return nil, fmt.Errorf("model %s not found in either OpenAI or Anthropic", model)
}

func listModels() {
	openaiClient := initOpenAIClient()
	anthropicClient := initAnthropicClient()

	fmt.Println("OpenAI Models:")
	openaiModels, err := openaiClient.ListModels(context.Background())
	if err != nil {
		logger.Error("Error listing OpenAI models", "error", err)
		fmt.Printf("Error listing OpenAI models: %v\n", err)
	} else {
		for _, model := range openaiModels.Models {
			fmt.Printf("- %s\n", model.ID)
		}
	}

	fmt.Println("\nAnthropic Models:")
	anthropicModels := anthropicClient.ListModels()
	for _, model := range anthropicModels {
		fmt.Printf("- %s (API Name: %s)\n", model.Name, model.ApiName)
	}
}
