package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

func generateCurls(taskContent, filesContent, prompt, model, exampleCallContent, apiGatewayJSON string) {
	promptGen := newPromptGenerator(taskContent, filesContent, apiGatewayJSON, prompt, exampleCallContent)
	client, err := determineClient(model)
	if err != nil {
		fmt.Printf("Error determining client: %v\n", err)
		return
	}

	var curlCommands []CurlCommand

	switch c := client.(type) {
	case *openai.Client:
		promptContent, tools := promptGen.generateOpenAIPrompt()
		curlCommands, err = generateCurlsWithOpenAI(c, promptContent, tools)
	case *Client:
		promptContent, tools := promptGen.generateAnthropicPrompt()
		curlCommands, err = generateCurlsWithAnthropic(c, promptContent, tools)
	default:
		fmt.Println("Unsupported client type")
		return
	}

	if err != nil {
		fmt.Printf("Error generating curl commands: %v\n", err)
		return
	}

	for _, cmd := range curlCommands {
		fmt.Printf("Command: %s\nExplanation: %s\n\n", cmd.Command, cmd.Explanation)
		if executeCurl {
			output, err := runCommand(cmd.Command)
			if err != nil {
				fmt.Printf("Error executing curl command: %v\n", err)
			} else {
				fmt.Printf("Output:\n%s\n\n", output)
			}
		}
	}
}

func generateCurlsWithOpenAI(client *openai.Client, promptContent string, tools []openai.Tool) ([]CurlCommand, error) {
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
			Tools: tools,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error creating chat completion: %v", err)
	}

	var toolResponse ToolResponse
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &toolResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling tool response: %v", err)
	}

	return toolResponse.CurlCommands, nil
}

func generateCurlsWithAnthropic(client *Client, promptContent string, tools []Tool) ([]CurlCommand, error) {
	resp, err := client.CreateMessage(CreateMessageRequest{
		Model:    model,
		Messages: []Message{{Role: "user", Content: promptContent}},
		Tools:    tools,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating message: %v", err)
	}

	var toolResponse ToolResponse
	err = json.Unmarshal([]byte(resp.Content[0].Text), &toolResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling tool response: %v", err)
	}

	return toolResponse.CurlCommands, nil
}
