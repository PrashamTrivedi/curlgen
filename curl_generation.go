package main

import (
	"context"
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

	promptContent := promptGen.generatePrompt()
	var curlCommands []CurlCommand

	switch c := client.(type) {
	case *openai.Client:
		slog.Debug("Generating curls with OpenAI")
		curlCommands, err = generateCurlsWithOpenAI(c, promptContent)
	case *Client:
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

	slog.Debug("Generated curl commands", "commands", len(curlCommands))

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
				fmt.Printf("Output:\n%s\n%s\n%s\n\n", 
					strings.Repeat("-", 80),
					formatOutput(output),
					strings.Repeat("-", 80))
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
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error creating chat completion with OpenAI: %w", err)
	}

	slog.Debug("OpenAI response", "response", resp)
	return parseCurlCommands(resp.Choices[0].Message.Content)
}

func generateCurlsWithAnthropic(client *Client, promptContent string) ([]CurlCommand, error) {
	model, err := client.GetModelByName(model)
	if err != nil {
		return nil, fmt.Errorf("error getting model by name: %w", err)
	}
	resp, err := client.CreateMessage(CreateMessageRequest{
		Model:     model.ApiName,
		Messages:  []Message{{Role: "user", Content: promptContent}},
		MaxTokens: 4096,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating message with Anthropic: %w", err)
	}
	slog.Debug("Anthropic response", "response", resp)

	var content string
	for _, message := range resp.Content {
		if message.Type == "text" {
			content += message.Text
		}
	}

	return parseCurlCommands(content)
}

func parseCurlCommands(content string) ([]CurlCommand, error) {
	var curlCommands []CurlCommand
	lines := strings.Split(content, "\n")

	slog.Debug("Parsing curl commands", "lines", len(lines))

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		slog.Debug("Line", "line", line)
		if strings.Contains(line, ". Command:") {
			parts := strings.SplitN(line, ". Command:", 2)
			if len(parts) != 2 {
				continue
			}
			command := strings.TrimSpace(parts[1])
			slog.Debug("Parsed command", "command", command)

			// Find the explanation
			explanation := ""
			for j := i + 1; j < len(lines); j++ {
				if strings.HasPrefix(lines[j], "Explanation:") {
					explanation = strings.TrimPrefix(strings.TrimSpace(lines[j]), "Explanation:")
					break
				}
			}

			curlCommands = append(curlCommands, CurlCommand{
				Command:     command,
				Explanation: strings.TrimSpace(explanation),
			})
		}
	}

	return curlCommands, nil
}
