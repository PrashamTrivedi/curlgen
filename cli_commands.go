package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
		if err := updateModelToClientMap(); err != nil {
			fmt.Printf("Error updating model list: %v. Please check your API keys and network connection.\n", err)
			return
		}
		listModels()
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set or remove API keys",
	Long:  `Set or remove OpenAI and Anthropic API keys. Use with caution as these are sensitive information.`,
	Run: func(cmd *cobra.Command, args []string) {
		configToSet := map[string]string{}
		if openAIKey != "" {
			configToSet[OPENAI_KEY] = openAIKey
		}
		if anthropicKey != "" {
			configToSet[ANTHROPIC_KEY] = anthropicKey
		}
		if len(configToSet) > 0 {
			writeConfig(configToSet)
		}
		if removeOpenAIKey {
			if err := removeAPIKey(OPENAI_KEY); err != nil {
				fmt.Printf("Error removing OpenAI key: %v. Please check if the key exists in the configuration.\n", err)
			} else {
				fmt.Println("OpenAI key removed successfully.")
			}
		}
		if removeAnthropicKey {
			if err := removeAPIKey(ANTHROPIC_KEY); err != nil {
				fmt.Printf("Error removing Anthropic key: %v. Please check if the key exists in the configuration.\n", err)
			} else {
				fmt.Println("Anthropic key removed successfully.")
			}
		}
	},
}

var printConfigCmd = &cobra.Command{
	Use:   "printConfig",
	Short: "Print all configurations",
	Run: func(cmd *cobra.Command, args []string) {
		printConfigurations()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&taskFile, "task", "t", "", "File containing task description or 'pbpaste' for clipboard content")
	rootCmd.Flags().StringSliceVarP(&changedFiles, "files", "f", []string{}, "Files with changes or helpful for code generation, use files a.txt,b.js,c.json way")
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Prompt explaining the changes")
	rootCmd.Flags().StringVarP(&model, "model", "m", "default", "AI model to use")
	rootCmd.Flags().StringVarP(&exampleCall, "example", "e", "", "File containing example API call (curl command) or 'pbpaste' for clipboard content")
	rootCmd.Flags().StringVarP(&apiGatewayJSON, "api-gateway", "g", "", "File containing API Gateway JSON schema")
	rootCmd.Flags().BoolVarP(&executeCurl, "executecurl", "x", false, "Execute curl commands (default is false, only print commands)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "API Key to replace {{API_KEY}} in the prompt response")
	rootCmd.Flags().StringVar(&apiURL, "api-url", "", "API URL to replace {{API_URL}} in the prompt response")

	rootCmd.MarkFlagRequired("files")
	rootCmd.MarkFlagRequired("api-gateway")

	rootCmd.Flags().SetInterspersed(true)

	configCmd.Flags().StringVarP(&openAIKey, "openai-key", "o", "", "OpenAI API Key")
	configCmd.Flags().StringVarP(&anthropicKey, "anthropic-key", "a", "", "Anthropic API Key")
	configCmd.Flags().BoolVar(&removeOpenAIKey, "remove-openai", false, "Remove OpenAI API Key")
	configCmd.Flags().BoolVar(&removeAnthropicKey, "remove-anthropic", false, "Remove Anthropic API Key")

	rootCmd.AddCommand(listModelsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(printConfigCmd)
}

var (
	removeOpenAIKey    bool
	removeAnthropicKey bool
)
