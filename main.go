package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logger         *slog.Logger
	taskFile       string
	changedFiles   []string
	prompt         string
	model          string
	key            string
	anthropicKey   string
	exampleCall    string
	apiGatewayJSON string
)

func init() {
	rootCmd.Flags().StringVarP(&taskFile, "task", "t", "", "File containing task description or 'pbpaste' for clipboard content")
	rootCmd.Flags().StringSliceVar(&changedFiles, "files", []string{}, "Files with changes or helpful for code generation, use files a.txt,b.js,c.json way")
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Prompt explaining the changes")
	rootCmd.Flags().StringVarP(&model, "model", "m", "default", "AI model to use")
	rootCmd.Flags().StringVarP(&exampleCall, "example", "e", "", "File containing example API call (curl command) or 'pbpaste' for clipboard content")
	rootCmd.Flags().StringVarP(&apiGatewayJSON, "api-gateway", "g", "", "File containing API Gateway JSON schema")

	rootCmd.MarkFlagRequired("task")
	rootCmd.MarkFlagRequired("files")
	rootCmd.MarkFlagRequired("example")
	rootCmd.MarkFlagRequired("api-gateway")

	// Allow interspersed flags and positional arguments
	rootCmd.Flags().SetInterspersed(true)

	viper.BindPFlag("task", rootCmd.Flags().Lookup("task"))
	viper.BindPFlag("files", rootCmd.Flags().Lookup("files"))
	viper.BindPFlag("prompt", rootCmd.Flags().Lookup("prompt"))
	viper.BindPFlag("model", rootCmd.Flags().Lookup("model"))
	viper.BindPFlag("example", rootCmd.Flags().Lookup("example"))
	viper.BindPFlag("api-gateway", rootCmd.Flags().Lookup("api-gateway"))

	configCmd.Flags().StringVarP(&key, "openai-key", "o", "", "OpenAI API Key")
	configCmd.Flags().StringVarP(&anthropicKey, "anthropic-key", "a", "", "Anthropic API Key")

	rootCmd.AddCommand(listModelsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(printConfigCmd)
}

var printConfigCmd = &cobra.Command{
	Use:   "printConfig",
	Short: "Print all configurations",
	Run: func(cmd *cobra.Command, args []string) {
		printConfigurations()
	},
}

func printConfigurations() {
	allSettings := ReadAllConfig()
	fmt.Println("Current configurations:")
	for key, value := range allSettings {
		if key == OpenaiKey || key == AnthropicKey {
			fmt.Printf("%s: %s\n", key, maskAPIKey(value.(string)))
		} else {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

var rootCmd = &cobra.Command{
	Use:   "curlgen",
	Short: "Generate curl commands using AI models",
	Run: func(cmd *cobra.Command, args []string) {
		taskContent := getTaskContent()
		filesContent := getFilesContent()
		exampleCallContent := getExampleCallContent()
		generateCurls(taskContent, filesContent, prompt, model, exampleCallContent, apiGatewayJSON)
	},
}

var listModelsCmd = &cobra.Command{
	Use:   "listModels",
	Short: "List available models from OpenAI and Anthropic",
	Run: func(cmd *cobra.Command, args []string) {
		listModels()
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set OpenAI Key and file paths to store assistants and chats",
	Long:  `Set OpenAI key to use for API calls. File paths are optional and will default to ~/.assistants/assistants.json and ~/.assistants/chats.json (In windows this will be \%USERPROFILE%\.assistants\assistants.json and \%USERPROFILE%\.assistants\chats.json)`,
	Run: func(cmd *cobra.Command, args []string) {
		configToSet := map[string]string{}
		configToSet[OpenaiKey] = key
		configToSet[AnthropicKey] = anthropicKey
		WriteConfig(configToSet)
	},
}

func getTaskContent() string {
	if taskFile == "" {
		fmt.Println("Task file is required")
		os.Exit(1)
	}
	if taskFile == "pbpaste" {
		content, err := getClipboardContent()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return content
	}

	content, err := os.ReadFile(taskFile)
	if err != nil {
		fmt.Printf("Error reading task file: %v\n", err)
		os.Exit(1)
	}
	return string(content)
}

func getExampleCallContent() string {
	if exampleCall == "" {
		fmt.Println("Example API call is required")
		os.Exit(1)
	}
	if exampleCall == "pbpaste" {
		content, err := getClipboardContent()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return content
	}

	content, err := os.ReadFile(exampleCall)
	if err != nil {
		fmt.Printf("Error reading example API call file: %v\n", err)
		os.Exit(1)
	}
	return string(content)
}

func getFilesContent() string {
	var content strings.Builder
	for _, file := range changedFiles {
		fileContent, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", file, err)
			continue
		}
		content.WriteString(fmt.Sprintf("File: %s\n%s\n\n", file, string(fileContent)))
	}
	return content.String()
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

func generateCurls(taskContent, filesContent, prompt, model, exampleCallContent, apiGatewayJSONPath string) {
	client, err := determineClient(model)
	if err != nil {
		fmt.Printf("Error determining client: %v\n", err)
		return
	}

	apiGatewayJSONContent := ""
	if apiGatewayJSONPath != "" {
		content, err := os.ReadFile(apiGatewayJSONPath)
		if err != nil {
			fmt.Printf("Error reading API Gateway JSON file: %v\n", err)
		} else {
			apiGatewayJSONContent = string(content)
		}
	}

	promptGenerator := NewPromptGenerator(taskContent, filesContent, apiGatewayJSONContent, prompt, exampleCallContent)

	switch c := client.(type) {
	case *openai.Client:
		openaiClient := initOpenAIClient()
		fmt.Println("Using OpenAI client")
		generatedPrompt, tools := promptGenerator.GenerateOpenAIPrompt()
		messages := []openai.ChatCompletionMessage{}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: generatedPrompt,
		})

		resp, err := openaiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    model,
				Messages: messages,
				Tools:    tools,
			},
		)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
		}

		responseMessage := resp.Choices[0].Message
		// Arguments is a stringified json, convert it into a map
		functionMessages, err := handleFunctionCall(responseMessage, openaiClient, resp, messages, tools)

		if err != nil {
			fmt.Println("Error handling function call: ", err.Error())

		}
		isFunctionHandled := err == nil && len(functionMessages) > 0
		if isFunctionHandled {
			messages = append(messages, functionMessages...)
		} else {
			messages = append(messages, resp.Choices[0].Message)
			fmt.Println(resp.Choices[0].Message.Content)
		}
	case *Client:
		fmt.Println("Using Anthropic client")
		anthropicClient := initAnthropicClient()
		generatedPrompt, tools := promptGenerator.GenerateAnthropicPrompt()
		// Use with Anthropic client
		// logger.Info("Generated Prompt", "Antrhopic", generatedPrompt)
		logger.Info("Generated Tool", "Antrhopic", tools)

		anthropicResp, err := anthropicClient.CreateMessage(CreateMessageRequest{
			Model:     "claude-3-5-sonnet-20240620",
			Messages:  []Message{{Role: "user", Content: generatedPrompt}},
			MaxTokens: 1024,
			Tools:     tools,
		})
		if err != nil {
			log.Fatalf("Error calling Anthropic API: %v", err)
		}
		logger.Debug("Response", "AnthropicResp", anthropicResp)
		logger.Debug("Response", "Anthropic", anthropicResp.Content)
		// loop through anthropicResp.Content
		for _, content := range anthropicResp.Content {
			if content.Type == "text" {
				fmt.Println(content.Text)
			} else if content.Type == "tool_use" {

				// Input value is an interface, print the type, keys and values of it
				// logger.Debug("Tool Use", "Name", content.Input)

				logger.Debug("Tool Use", "Input", content.Input["curl_commands"])
				curlCommandJson := content.Input["curl_commands"]
				logger.Debug("Tool Use", "CurlCoomandJSON", curlCommandJson)
				curlCommandJson = strings.ReplaceAll(curlCommandJson, "\n", "")
				curlCommandJson = strings.ReplaceAll(curlCommandJson, "  ", " ")
				curlCommandJson = strings.ReplaceAll(curlCommandJson, "\\--", "--")
				// var toolResponse ToolResponse
				// preprocessedJSON := strings.ReplaceAll(content.Input["curl_commands"], "\n", "")
				// preprocessedJSON = strings.ReplaceAll(preprocessedJSON, "\\", "")
				// // preprocessedJSON = strings.ReplaceAll(preprocessedJSON, "\\\"", "\"")
				// logger.Debug("Tool Use", "PreprocessedJSON", preprocessedJSON)
				var toolResponse ToolResponse
				err := json.Unmarshal([]byte(curlCommandJson), &toolResponse.CurlCommands)
				if err != nil {
					fmt.Printf("Error unmarshalling: %v\n", err)
				} else {
					fmt.Printf("Successfully unmarshalled: %+v\n", toolResponse)
					curlCommands := toolResponse.CurlCommands
					for _, cmd := range curlCommands {
						fmt.Printf("=== %s ===\n", cmd.Explanation)
						fmt.Printf("Executing command: %s\n", cmd.Command)
						// result, err := RunCommand([]string{cmd.Command})
						// if err != nil {
						// 	fmt.Printf("Error: %s\n", err.Error())
						// } else {
						// 	fmt.Printf("Result:\n%s\n", result)
						// }
						fmt.Println()
					}
				}

				// logger.Info("Tool Use", "Name", content.Input)
				// for key, valueJson := range content.Input {
				// 	fmt.Printf("%s\n", key)
				// 	fmt.Printf("%s\n", valueJson)
				// }
			}
		}

	default:
		fmt.Printf("Unknown client type: %T\n", c)
		return
	}

	// TODO: Use the generated prompt and tools to make the API call and process the response
}

func initOpenAIClient() *openai.Client {
	openaiKey := ReadConfig(OpenaiKey)
	if openaiKey == "" {
		fmt.Println("OpenAI key not found. Please set it using the config command.")
		os.Exit(1)
	}
	return openai.NewClient(openaiKey)
}

func initAnthropicClient() *Client {
	anthropicKey := ReadConfig(AnthropicKey)
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

func listModels() {
	openaiClient := initOpenAIClient()
	anthropicClient := initAnthropicClient()

	fmt.Println("Fetching Models")
	openaiModels, err := openaiClient.ListModels(context.Background())
	anthropicModels := anthropicClient.ListModels()
	// TODO: Implement API calls to list models from OpenAI and Anthropic
	fmt.Println("Available models:")
	fmt.Println("OpenAI: ")
	// Sort models such a way that models containing gpt are listed first
	sort.Slice(openaiModels.Models, func(i, j int) bool {

		return openaiModels.Models[j].CreatedAt < openaiModels.Models[i].CreatedAt
	})
	if err != nil {
		fmt.Println("Error listing models:", err.Error())
	}
	for i, model := range openaiModels.Models {
		fmt.Printf("\t%d, %s\n", i, model.ID)
		// fmt.Printf("%+v\n", model)
	}
	fmt.Println("Anthropic: ")
	for i, model := range anthropicModels {
		fmt.Printf("\t%d, %s (%s)\n", i, model.Name, model.ApiName)
	}
}

func main() {
	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: DefaultLogLevel,
	}))
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handleFunctionCall(responseMessage openai.ChatCompletionMessage, openaiClient *openai.Client,
	resp openai.ChatCompletionResponse, messages []openai.ChatCompletionMessage, tools []openai.Tool) ([]openai.ChatCompletionMessage, error) {

	functionMessages := make([]openai.ChatCompletionMessage, 0)
	if responseMessage.ToolCalls != nil && responseMessage.ToolCalls[0].Function.Name != "" {
		functionMessage := openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: responseMessage.ToolCalls[0].ID,
			Name:       responseMessage.ToolCalls[0].Function.Name,
			Content:    "",
		}
		logger.Info("Function call: ", "Name", resp.Choices[0].Message.ToolCalls[0].Function.Name)
		logger.Info("Function call: ", "Args", resp.Choices[0].Message.ToolCalls[0].Function.Arguments)
		functionMessage.Name = resp.Choices[0].Message.ToolCalls[0].Function.Name

		if resp.Choices[0].Message.ToolCalls[0].Function.Name == "execute_curl" {
			var toolResponse ToolResponse
			err := json.Unmarshal([]byte(resp.Choices[0].Message.ToolCalls[0].Function.Arguments), &toolResponse)
			if err != nil {
				logger.Error("Error parsing curl commands", "error", err)
				functionMessage.Content = fmt.Sprintf("Error parsing curl commands: %s", err.Error())
			} else {
				curlCommands := toolResponse.CurlCommands
				var output strings.Builder
				for _, cmd := range curlCommands {
					commandToExecute := CleanCommand(cmd.Command)
					output.WriteString(fmt.Sprintf("=== %s ===\n", cmd.Explanation))
					output.WriteString(fmt.Sprintf("Executing command: %s\n", commandToExecute))
					// result, err := RunCommand([]string{commandToExecute})
					// if err != nil {
					// 	output.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
					// } else {
					// 	output.WriteString(fmt.Sprintf("Result:\n%s\n", result))
					// }
					output.WriteString("\n")
				}
				functionMessage.Content = output.String()
			}
		}

		// // functionMessages = append(functionMessages, functionMessage)
		// // messages := append(messages, responseMessage, functionMessage)
		// // chatRequest := openai.ChatCompletionRequest{
		// // 	Model:    model,
		// // 	Messages: messages,
		// // 	Tools:    tools,
		// // }
		// // afterFuncResponse, err := openaiClient.CreateChatCompletion(
		// // 	context.Background(),
		// // 	chatRequest,
		// // )
		// // if err != nil {
		// // 	logger.Error("ChatCompletion", "error", err)
		// // 	return nil, err
		// // }
		// functionMessages = append(functionMessages, afterFuncResponse.Choices[0].Message)
		// functionResponseMessage := afterFuncResponse.Choices[0].Message
		// fmt.Println(functionResponseMessage.Content)
		// if functionResponseMessage.FunctionCall != nil && functionResponseMessage.FunctionCall.Name != "" {
		// 	// Create a copy of c
		// 	updatedMessagess := messages
		// 	newFunctionMessages, err := handleFunctionCall(functionResponseMessage, openaiClient, resp, updatedMessagess, tools)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	functionMessages = append(functionMessages, newFunctionMessages...)
		// }
	}
	return functionMessages, nil
}

// CleanCommand sanitizes the input command by removing specific characters and sequences.
func CleanCommand(input string) string {
	// Remove newline characters
	result := strings.ReplaceAll(input, "\n", "")
	// Remove backslashes
	result = strings.ReplaceAll(result, "\\", "")
	// Replace escaped double quotes with double quotes
	result = strings.ReplaceAll(result, "\\\"", "\"")
	return result
}

func RunCommand(commands []string) (string, error) {
	var combinedOutput strings.Builder

	logger.Info("Running commands", "commands", commands)
	for _, command := range commands {
		args := strings.Fields(command)
		if len(args) == 0 {
			continue
		}
		logger.Info("Running command", "command", command)

		cmd := exec.Command(args[0], args[1:]...)
		var out strings.Builder
		cmd.Stdout = &out
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			logger.Error("Command failed", "command", command, "error", err)
			return "", fmt.Errorf("error executing command '%s': %v\n%s", command, err, out.String())
		}

		logger.Info("Command executed", "command", command, "output", out.String())
		combinedOutput.WriteString(out.String())
	}

	return combinedOutput.String(), nil
}
