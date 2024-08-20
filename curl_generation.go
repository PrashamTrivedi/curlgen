package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func generateCurls(taskContent, filesContent, prompt, model, exampleCallContent, apiGatewayJSON string) {
	slog.Debug("Generating curl commands", "model", model)
	promptGen := newPromptGenerator(taskContent, filesContent, apiGatewayJSON, prompt, exampleCallContent)
	client, err := determineClient(model)
	if err != nil {
		slog.Error("Error determining client", "error", err)
		return
	}

	var curlCommands []CurlCommand

	switch c := client.(type) {
	case *openai.Client:
		promptContent := promptGen.generatePrompt()
		slog.Debug("Generating curls with OpenAI")
		curlCommands, err = generateCurlsWithOpenAI(c, promptContent)
	case *Client:
		promptContent := promptGen.generatePrompt()
		slog.Debug("Generating curls with Anthropic")
		curlCommands, err = generateCurlsWithAnthropic(c, promptContent)
	default:
		slog.Error("Unsupported client type")
		return
	}

	if err != nil {
		slog.Error("Error generating curl commands", "error", err)
		return
	}

	for i, cmd := range curlCommands {
		slog.Debug("Generated curl command", "index", i, "command", cmd.Command)

		// Replace placeholders with actual values if provided
		if apiKey != "" {
			cmd.Command = strings.ReplaceAll(cmd.Command, "{{API_KEY}}", apiKey)
		}
		if apiURL != "" {
			cmd.Command = strings.ReplaceAll(cmd.Command, "{{API_URL}}", apiURL)
		}

		fmt.Printf("Command: %s\nExplanation: %s\n\n", cmd.Command, cmd.Explanation)
		if executeCurl {
			slog.Debug("Executing curl command", "index", i)
			output, err := runCommand(cmd.Command)
			if err != nil {
				slog.Error("Error executing curl command", "error", err)
			} else {
				slog.Debug("Curl command executed successfully", "index", i)
				fmt.Printf("Output:\n%s\n\n", output)
			}
		}
	}
}

func generateCurlsWithOpenAI(client *openai.Client, promptContent string) ([]CurlCommand, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: promptContent,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error creating chat completion with OpenAI: %w", err)
	}

	slog.Debug("OpenAI response", "response", resp)
	var toolResponse ToolResponse
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &toolResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling OpenAI JSON response: %w", err)
	}

	return toolResponse.CurlCommands, nil
}

func generateCurlsWithAnthropic(client *Client, promptContent string) ([]CurlCommand, error) {
	resp, err := client.CreateMessage(CreateMessageRequest{
		Model:    model,
		Messages: []Message{{Role: "user", Content: promptContent}},
		System:   "Respond only with a JSON object containing an array of curl commands.",
	})

	if err != nil {
		return nil, fmt.Errorf("error creating message with Anthropic: %w", err)
	}
	slog.Debug("Anthropic response", "response", resp)

	var toolResponse ToolResponse
	err = json.Unmarshal([]byte(resp.Content[0].Text), &toolResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling Anthropic JSON response: %w", err)
	}

	return toolResponse.CurlCommands, nil
}
