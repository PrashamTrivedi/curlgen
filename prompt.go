package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type promptGenerator struct {
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

func newPromptGenerator(taskDefinition, updatedCode, apiGatewayJSON, additionalInfo, sampleAPICall string) *promptGenerator {
	return &promptGenerator{
		TaskDefinition: taskDefinition,
		UpdatedCode:    updatedCode,
		APIGatewayJSON: apiGatewayJSON,
		AdditionalInfo: additionalInfo,
		SampleAPICall:  sampleAPICall,
	}
}

func (pg *promptGenerator) generatePrompt() string {
	prompt := fmt.Sprintf(`You are an expert API tester. Your task is to generate curl commands for testing an API endpoint based on the following information:

Generate appropriate curl commands for testing the API endpoint based on the following information:

Task Definition:
%s
`, pg.TaskDefinition)

	if pg.UpdatedCode != "" {
		prompt += fmt.Sprintf(`
Updated Code:
%s
`, pg.UpdatedCode)
	}

	if pg.APIGatewayJSON != "" {
		var apiModel APIGatewayRequestModel
		err := json.Unmarshal([]byte(pg.APIGatewayJSON), &apiModel)
		if err != nil {
			fmt.Printf("Warning: Failed to parse API Gateway JSON: %v. Please check the JSON format.\n", err)
			fmt.Println("API Gateway JSON:", pg.APIGatewayJSON)
		}
		prompt += fmt.Sprintf(`
API Gateway Request Model:
%s
`, pg.APIGatewayJSON)
	}

	if pg.AdditionalInfo != "" {
		prompt += fmt.Sprintf(`
Additional Information:
%s
`, pg.AdditionalInfo)
	}

	if pg.SampleAPICall != "" {
		prompt += fmt.Sprintf(`
Sample API Call:
%s
`, pg.SampleAPICall)
	}

	prompt += `
Consider the following when generating curl commands:

1. Use the API Gateway Request Model to structure the request body (if provided).
2. Incorporate any authentication or headers required by the API.
3. Include examples for different HTTP methods (GET, POST, PUT, DELETE) if applicable.
4. Provide variations of the curl commands to test different scenarios or edge cases.
5. Include any necessary query parameters or path variables.
6. Use the placeholder {{API_URL}} for the API URL.
7. Use the placeholder {{API_KEY}} for the API Key in the Authorization header.

Please generate a set of curl commands that thoroughly test the API endpoint described above, using the placeholders for API URL and API Key.`

	return strings.TrimSpace(prompt)
}

func (pg *promptGenerator) generateAnthropicPrompt() (string, []Tool) {
	basePrompt := pg.generatePrompt()

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

	return strings.TrimSpace(basePrompt), tools
}

func (pg *promptGenerator) generateOpenAIPrompt() (string, []openai.Tool) {
	basePrompt := pg.generatePrompt()

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

	return strings.TrimSpace(basePrompt), tools
}
