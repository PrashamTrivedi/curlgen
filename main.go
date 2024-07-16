package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	taskFile     string
	changedFiles []string
	prompt       string
	model        string
	key          string
	anthropicKey string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&taskFile, "task", "t", "", "File containing task description or 'pbpaste' for clipboard content")
	rootCmd.PersistentFlags().StringSliceVarP(&changedFiles, "files", "f", []string{}, "Files with changes or helpful for code generation")
	rootCmd.PersistentFlags().StringVarP(&prompt, "prompt", "p", "", "Prompt explaining the changes")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "default", "AI model to use")
	rootCmd.MarkFlagRequired("task")
	rootCmd.MarkFlagRequired("files")

	viper.BindPFlag("task", rootCmd.PersistentFlags().Lookup("task"))
	viper.BindPFlag("files", rootCmd.PersistentFlags().Lookup("files"))
	viper.BindPFlag("prompt", rootCmd.PersistentFlags().Lookup("prompt"))
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))

	configCmd.Flags().StringVarP(&key, "openai-key", "o", "", "OpenAI API Key")
	configCmd.Flags().StringVarP(&anthropicKey, "anthropic-key", "a", "", "Anthropic API Key")

	rootCmd.AddCommand(listModelsCmd)
	rootCmd.AddCommand(configCmd)

}

var rootCmd = &cobra.Command{
	Use:   "curlgen",
	Short: "Generate curl commands using AI models",
	Run: func(cmd *cobra.Command, args []string) {
		taskContent := getTaskContent()
		filesContent := getFilesContent()
		generateCurls(taskContent, filesContent, prompt, model)
	},
}

var listModelsCmd = &cobra.Command{
	Use:   "list-models",
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
		out, err := exec.Command("pbpaste").Output()
		if err != nil {
			fmt.Println("Error accessing clipboard:", err)
			os.Exit(1)
		}
		return string(out)
	}

	content, err := os.ReadFile(taskFile)
	if err != nil {
		fmt.Printf("Error reading task file: %v\n", err)
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

func generateCurls(taskContent, filesContent, prompt, model string) {
	// TODO: Implement API calls to OpenAI and Anthropic
	fmt.Printf("Generating curl commands with:\nTask: %s\nFiles: %s\nPrompt: %s\nModel: %s\n",
		taskContent, filesContent, prompt, model)
}

func listModels() {

	openaiKey := ReadConfig(OpenaiKey)
	if openaiKey == "" {
		fmt.Println("OpenAI key not found. Please set it using the config command.")
		os.Exit(1)
	}
	client := openai.NewClient(openaiKey)
	anthropicClient, error := NewAnthropicClient()
	if error != nil {
		fmt.Println("Error creating anthropic client:", error)
		os.Exit(1)
	}
	fmt.Println("Fetching Models")
	openaiModels, err := client.ListModels(context.Background())
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
		fmt.Printf("%d, %s\n", i, model.ID)
		// fmt.Printf("%+v\n", model)
	}
	fmt.Println("Anthropic: ")
	for i, model := range anthropicModels {
		fmt.Printf("%d, %s (%s)\n", i, model.Name, model.ApiName)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
