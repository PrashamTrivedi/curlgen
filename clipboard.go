package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func getClipboardContent() (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "linux":
		cmd = exec.Command("xclip", "-o")
	case "windows":
		cmd = exec.Command("powershell.exe", "Get-Clipboard")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error accessing clipboard: %v", err)
	}

	return string(output), nil
}
