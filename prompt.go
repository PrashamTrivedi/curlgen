package main

import (
	"encoding/json"
	"fmt"
	"strings"
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

First, generate a set of curl commands that thoroughly test the API endpoint described above, using the placeholders for API URL and API Key.
Then respond with a JSON object containing an array of curl commands. Each command should be an object with 'command' and 'explanation' fields.`

	return strings.TrimSpace(prompt)
}
