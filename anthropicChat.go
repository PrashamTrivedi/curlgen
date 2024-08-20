package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Model struct {
	ApiName string `json:"api_name"`
	Name    string `json:"name"`
}

type Client struct {
	APIKey  string
	BaseURL string
	HTTP    *http.Client
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type CreateMessageRequest struct {
	Model       string      `json:"model"`
	Messages    []Message   `json:"messages"`
	MaxTokens   int         `json:"max_tokens"`
	Temperature float64     `json:"temperature,omitempty"`
	System      string      `json:"system,omitempty"`
	Tools       interface{} `json:"tools,omitempty"`
}

type CreateMessageResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Content      []Content `json:"content"`
	Model        string    `json:"model"`
	StopReason   string    `json:"stop_reason"`
	StopSequence string    `json:"stop_sequence"`
	Usage        Usage     `json:"usage"`
}

type Content struct {
	Type  string            `json:"type"`
	Text  string            `json:"text"`
	Id    string            `json:"id"`
	Name  string            `json:"name"`
	Input map[string]string `json:"input"`
}

type AnhropicToolResponse struct {
	CurlCommands []CurlCommand `json:"curl_commands"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

const (
	DefaultBaseURL = "https://api.anthropic.com/v1"
)

var models = []Model{
	{
		ApiName: "claude-3-5-sonnet-20240620",
		Name:    "claude-3.5-sonnet",
	},
	{
		ApiName: "claude-3-opus-20240229",
		Name:    "claude-3-opus",
	},
	{
		ApiName: "claude-3-sonnet-20240229",
		Name:    "claude-3-sonnet",
	},
	{
		ApiName: "claude-3-haiku-20240307",
		Name:    "claude-3-haiku",
	},
}

func NewAnthropicClient(anthropicKey string) (*Client, error) {
	apiKey := anthropicKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is not provided and environment variable is not set. Please set the API key using the config command or as an environment variable")
		}
	}

	return &Client{
		APIKey:  apiKey,
		BaseURL: DefaultBaseURL,
		HTTP:    &http.Client{},
	}, nil
}

func (c *Client) CreateMessage(req CreateMessageRequest) (*CreateMessageResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling Anthropic API request: %w", err)
	}

	url := fmt.Sprintf("%s/messages", c.BaseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating Anthropic API request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending Anthropic API request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading Anthropic API response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var createResp CreateMessageResponse
	err = json.Unmarshal(body, &createResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling Anthropic API response: %w", err)
	}

	return &createResp, nil
}

func (c *Client) ListModels() []Model {

	return models

}

func (c *Client) GetModelByName(name string) (Model, error) {
	for _, model := range models {
		if model.Name == name {
			return model, nil
		}
	}
	return Model{}, fmt.Errorf("Anthropic model not found: %s. Please check the model name and try again", name)
}
