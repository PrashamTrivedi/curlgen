package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	OPENAI_KEY    = "openai_key"
	ANTHROPIC_KEY = "anthropic_api_key"
)

func readConfig(configKey string) string {
	configPath := createConfigPath()
	viper.SetConfigFile(configPath)
	viper.ReadInConfig()
	value := viper.GetString(configKey)
	maskedValue := value
	if configKey == OPENAI_KEY || configKey == ANTHROPIC_KEY {
		maskedValue = maskAPIKey(value)
	}
	slog.Debug("Read config", "key", configKey, "value", maskedValue)
	return value
}

func readAllConfig() map[string]interface{} {
	configPath := createConfigPath()
	viper.SetConfigFile(configPath)
	viper.ReadInConfig()
	return viper.AllSettings()
}

func resetConfig(key string) {
	configPath := createConfigPath()
	viper.SetConfigFile(configPath)
	if key == "" {
		viper.WriteConfig()
		return
	}
	viper.ReadInConfig()
	viper.Set(key, "")
	viper.WriteConfig()
}

func createConfigPath() string {
	if secretPath, exists := os.LookupEnv("SECRET_CONFIG_PATH"); exists {
		configFilePath := secretPath
		fmt.Println("Warning: You are using a secret config path meant for testing. Proceed with caution.")
		return configFilePath
	}
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".curlgen", "config.json")
	// Create ~/.curlgen/config.json if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(home, ".curlgen"), os.ModePerm)
		os.Create(configPath)
	}
	return configPath
}

func writeConfig(keyValueMap map[string]string) {
	configPath := createConfigPath()
	viper.SetConfigFile(configPath)
	for key, value := range keyValueMap {
		if key == OPENAI_KEY || key == ANTHROPIC_KEY {
			fmt.Printf("Warning: You are about to store the %s in %s. This is sensitive information.\n", key, configPath)
			fmt.Print("Are you sure you want to proceed? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Operation cancelled.")
				return
			}
		}
		viper.Set(key, value)
	}
	if err := viper.WriteConfig(); err != nil {
		panic(fmt.Errorf("failed to write config: %w", err))
	}
}

func removeAPIKey(key string) error {
	configPath := createConfigPath()
	viper.SetConfigFile(configPath)
	viper.ReadInConfig()
	viper.Set(key, "")
	return viper.WriteConfig()
}

func getTaskContent() string {
	if taskFile == "" {
		return ""
	}
	if taskFile == "pbpaste" {
		content, err := getClipboardContent()
		if err != nil {
			fmt.Printf("Error reading from clipboard: %v\n", err)
			return ""
		}
		return content
	}
	content, err := os.ReadFile(taskFile)
	if err != nil {
		fmt.Printf("Error reading task file '%s': %v\n", taskFile, err)
		return ""
	}
	return string(content)
}

func getExampleCallContent() string {
	if exampleCall == "" {
		return ""
	}
	if exampleCall == "pbpaste" {
		content, err := getClipboardContent()
		if err != nil {
			fmt.Println("Error reading from clipboard:", err)
			return ""
		}
		return content
	}
	content, err := os.ReadFile(exampleCall)
	if err != nil {
		fmt.Printf("Error reading example API call file '%s': %v\n", exampleCall, err)
		return ""
	}
	return string(content)
}

func getFilesContent() string {
	var content strings.Builder
	for _, file := range changedFiles {
		fileContent, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file '%s': %v\n", file, err)
			continue
		}
		content.WriteString(fmt.Sprintf("File: %s\n%s\n\n", file, string(fileContent)))
	}
	return content.String()
}

func printConfigurations() {
	allSettings := readAllConfig()
	fmt.Println("Current configurations:")
	for key, value := range allSettings {
		if key == OPENAI_KEY || key == ANTHROPIC_KEY {
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

var allowedCommands = map[string]bool{
	"curl": true,
}

func runCommand(command string) (string, error) {
	commandToRun := strings.TrimSpace(command)
	commandToRun = strings.TrimPrefix(commandToRun, "```")
	commandToRun = strings.TrimPrefix(commandToRun, "`")
	commandToRun = strings.TrimPrefix(commandToRun, "```sh")
	commandToRun = strings.TrimPrefix(commandToRun, "```shell")
	commandToRun = strings.TrimSuffix(commandToRun, "```")
	commandToRun = strings.TrimSuffix(commandToRun, "`")
	slog.Debug("Command", "CommandToRun", commandToRun)
	parts := strings.Fields(commandToRun)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	if !allowedCommands[parts[0]] {
		return "", fmt.Errorf("command not allowed: %s", parts[0])
	}

	for i, part := range parts {
		parts[i] = filepath.Clean(part)
	}

	fmt.Println("Warning: Executing AI-generated command. Please review for safety.")
	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing command: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

func formatOutput(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var formattedOutput strings.Builder

	for _, line := range lines {
		formattedOutput.WriteString("| " + line + "\n")
	}

	return formattedOutput.String()
}
