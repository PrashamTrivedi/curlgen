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
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature,omitempty"`
	System      string    `json:"system,omitempty"`
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
	Type string `json:"type"`
	Text string `json:"text"`
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

func NewAnthropicClient() (*Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set")
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
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	url := fmt.Sprintf("%s/messages", c.BaseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var createResp CreateMessageResponse
	err = json.Unmarshal(body, &createResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
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
	return Model{}, fmt.Errorf("model not found: %s", name)
}
