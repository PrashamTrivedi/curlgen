package main

import (
	"fmt"
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
	return viper.GetString(configKey)
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
		viper.Set(key, value)
	}
	if err := viper.WriteConfig(); err != nil {
		panic(err)
	}
}

func getTaskContent() string {
	if taskFile == "" {
		return ""
	}
	if taskFile == "pbpaste" {
		content, err := getClipboardContent()
		if err != nil {
			fmt.Println("Error reading from clipboard:", err)
			return ""
		}
		return content
	}
	content, err := os.ReadFile(taskFile)
	if err != nil {
		fmt.Printf("Error reading task file: %v\n", err)
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
		fmt.Printf("Error reading example API call file: %v\n", err)
		return ""
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

func printConfigurations() {
	allSettings := readAllConfig()
	fmt.Println("Current configurations:")
	for key, value := range allSettings {
		if key == openAIKey || key == anthropicKey {
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

func runCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing command: %v, output: %s", err, string(output))
	}
	return string(output), nil
}
