package main

import (
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

type CurlCommand struct {
	Command     string
	Explanation string
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

Generate a set of curl commands that thoroughly test the API endpoint described above, using the placeholders for API URL and API Key.
Respond with a list of curl commands and their explanations in the following format:

Here are the updated curl commands and explanations:
1. Command: <curl_command_1>
Explanation: <explanation_1>
2. Command: <curl_command_2>
Explanation: <explanation_2>
3. Command: <curl_command_3>
Explanation: <explanation_3>

Continue this format for all generated curl commands. Ensure that each command starts with a number followed by 'Command:' on a new line, and 'Explanation:' is on a separate line. Do not include any JSON in your response.`

	return strings.TrimSpace(prompt)
}
