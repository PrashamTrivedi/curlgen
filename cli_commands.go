package main

import (
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
		listModels()
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set OpenAI Key and file paths to store assistants and chats",
	Long:  `Set OpenAI key to use for API calls. File paths are optional and will default to ~/.assistants/assistants.json and ~/.assistants/chats.json (In windows this will be %USERPROFILE%\.assistants\assistants.json and %USERPROFILE%\.assistants\chats.json)`,
	Run: func(cmd *cobra.Command, args []string) {
		configToSet := map[string]string{}
		configToSet[OPENAI_KEY] = openAIKey
		configToSet[ANTHROPIC_KEY] = anthropicKey
		writeConfig(configToSet)
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
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "API Key to replace {{API_KEY}} in the prompt response")
	rootCmd.Flags().StringVar(&apiURL, "api-url", "", "API URL to replace {{API_URL}} in the prompt response")

	rootCmd.MarkFlagRequired("files")
	rootCmd.MarkFlagRequired("api-gateway")

	rootCmd.Flags().SetInterspersed(true)

	configCmd.Flags().StringVarP(&openAIKey, "openai-key", "o", "", "OpenAI API Key")
	configCmd.Flags().StringVarP(&anthropicKey, "anthropic-key", "a", "", "Anthropic API Key")

	rootCmd.AddCommand(listModelsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(printConfigCmd)
}
