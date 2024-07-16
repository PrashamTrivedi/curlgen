package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type PromptGenerator struct {
	TaskDefinition string
	UpdatedCode    string
	APIGatewayJSON string
	AdditionalInfo string
	SampleAPICall  string
}

type APIGatewayRequestModel struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type CurlTool struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	InputSchema CurlToolInputSchema `json:"input_schema"`
}

type CurlToolInputSchema struct {
	Type       string                   `json:"type"`
	Properties CurlToolInputSchemaProps `json:"properties"`
	Required   []string                 `json:"required"`
}

type CurlToolInputSchemaProps struct {
	CurlCommands CurlToolInputSchemaProp `json:"curl_commands"`
}

type CurlToolInputSchemaProp struct {
	Type        string                   `json:"type"`
	Description string                   `json:"description"`
	Items       CurlToolInputSchemaItems `json:"items"`
}

type CurlToolInputSchemaItems struct {
	Type       string                       `json:"type"`
	Properties CurlToolInputSchemaItemProps `json:"properties"`
	Required   []string                     `json:"required"`
}

type CurlToolInputSchemaItemProps struct {
	Command     CurlToolInputSchemaItemProp `json:"command"`
	Explanation CurlToolInputSchemaItemProp `json:"explanation"`
}

type CurlToolInputSchemaItemProp struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
type CurlCommand struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
}

type ToolResponse struct {
	CurlCommands []CurlCommand `json:"curl_commands"`
}

func NewPromptGenerator(taskDefinition, updatedCode, apiGatewayJSON, additionalInfo, sampleAPICall string) *PromptGenerator {
	return &PromptGenerator{
		TaskDefinition: taskDefinition,
		UpdatedCode:    updatedCode,
		APIGatewayJSON: apiGatewayJSON,
		AdditionalInfo: additionalInfo,
		SampleAPICall:  sampleAPICall,
	}
}

func (pg *PromptGenerator) GeneratePrompt() string {
	var apiModel APIGatewayRequestModel
	err := json.Unmarshal([]byte(pg.APIGatewayJSON), &apiModel)
	if err != nil {
		fmt.Printf("Warning: Failed to parse API Gateway JSON: %v\n", err)
		fmt.Println("API Gateway JSON:", pg.APIGatewayJSON)
	}

	prompt := fmt.Sprintf(`
Task Definition:
%s

Updated Code:
%s

API Gateway Request Model:
%s

Additional Information:
%s

Sample API Call:
%s

Based on the provided information, generate appropriate curl commands for testing the API endpoint. Consider the following:

1. Use the API Gateway Request Model to structure the request body.
2. Incorporate any authentication or headers required by the API.
3. Include examples for different HTTP methods (GET, POST, PUT, DELETE) if applicable.
4. Provide variations of the curl commands to test different scenarios or edge cases.
5. Include any necessary query parameters or path variables.
6. If the API uses JWT or other token-based authentication, include a placeholder for the token.

Please generate a set of curl commands that thoroughly test the API endpoint described above.
`, pg.TaskDefinition, pg.UpdatedCode, pg.APIGatewayJSON, pg.AdditionalInfo, pg.SampleAPICall)

	return strings.TrimSpace(prompt)
}

func (pg *PromptGenerator) GenerateAnthropicPrompt() (string, []Tool) {
	basePrompt := pg.GeneratePrompt()

	anthropicPrompt := fmt.Sprintf(`
Human: %s

Assistant: Certainly! I'll analyze the provided information and generate appropriate curl commands for testing the API endpoint. I'll consider the API Gateway Request Model, authentication requirements, different HTTP methods, and various scenarios to create a comprehensive set of test commands.

Here are the curl commands for testing the API endpoint:

`, basePrompt)

	curlTool := CurlTool{

		Name:        "execute_curl",
		Description: "Execute a curl command and return the result",
		InputSchema: CurlToolInputSchema{
			Type: "object",
			Properties: CurlToolInputSchemaProps{
				CurlCommands: CurlToolInputSchemaProp{
					Type:        "array",
					Description: "An array of curl commands to execute",
					Items: CurlToolInputSchemaItems{
						Type: "object",
						Properties: CurlToolInputSchemaItemProps{
							Command: CurlToolInputSchemaItemProp{
								Type:        "string",
								Description: "The curl command to execute",
							},
							Explanation: CurlToolInputSchemaItemProp{
								Type:        "string",
								Description: "A short explanation of what the curl command is doing",
							},
						},
						Required: []string{"command", "explanation"},
					},
				},
			},
			Required: []string{"curl_commands"},
		},
	}

	toolJSON, _ := json.Marshal(curlTool.InputSchema)

	tools := []Tool{
		{
			Name:        curlTool.Name,
			Description: curlTool.Description,
			InputSchema: toolJSON,
		},
	}

	return strings.TrimSpace(anthropicPrompt), tools
}

func (pg *PromptGenerator) GenerateOpenAIPrompt() (string, []openai.Tool) {
	basePrompt := pg.GeneratePrompt()

	openAIPrompt := fmt.Sprintf(`
You are an expert API tester. Your task is to generate curl commands for testing an API endpoint based on the following information:

%s

Please provide a set of curl commands that thoroughly test this API endpoint, considering different scenarios, edge cases, and HTTP methods where applicable.
`, basePrompt)

	params := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"curl_commands": {
				Type:        jsonschema.Array,
				Description: "An array of curl commands to execute",
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"command": {
							Type:        jsonschema.String,
							Description: "The curl command to execute",
						},
						"explanation": {
							Type:        jsonschema.String,
							Description: "A short explanation of what the curl command is doing",
						},
					},
					Required: []string{"command", "explanation"},
				},
			},
		},
		Required: []string{"curl_commands"},
	}
	getCurl := openai.FunctionDefinition{
		Name:        "execute_curl",
		Description: "Execute a curl command and return the result",
		Parameters:  params,
	}

	tools := []openai.Tool{
		{
			Type:     openai.ToolTypeFunction,
			Function: &getCurl,
		},
	}

	return strings.TrimSpace(openAIPrompt), tools
}
