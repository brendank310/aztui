package azcli

import (
	"bytes"

	"fmt"
	"os/exec"
	"strings"
)

var gAzResourceCommandMap map[string]string

// If we don't have an implementation for a given command, fall back to
// shelling out to azcli.
func GetResourceCommands(subcommand string) (map[string]string, error) {
	resourceCommands := make(map[string]string)
	gAzResourceCommandMap := make(map[string]string)
	cmd := exec.Command("az", subcommand, "--help")

	// Create buffers to capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Run the command
	err := cmd.Run()
	if err != nil {
		return resourceCommands, fmt.Errorf("Command execution failed with error: %v\n", err)
	}

	// Read and print stdout and stderr
	stdoutStr := stdoutBuf.String()
	subCommandPairs := strings.Split(stdoutStr, "Commands:")[1]
	for _, commandPair := range strings.Split(subCommandPairs, "\n") {
		if commandPair == "" {
			continue
		}

		pair := strings.Split(commandPair, ":")
		if len(pair) != 2 {
			continue
		}

		resourceCommands[strings.TrimSpace(pair[0])] = pair[1]
		gAzResourceCommandMap[strings.TrimSpace(pair[0])] = pair[1]
	}

	return resourceCommands, nil
}

// Function to check if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func RunAzCommand(args []string, handleErrorFunc func([]string, error) error) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("empty args list")
	}

	// Check for no
	if args[0] == "" {
		return "", fmt.Errorf("empty subcommand")
	}

	// Slice to store the keys
	var keys []string

	// Iterate over the map to get the keys
	for key := range gAzResourceCommandMap {
		keys = append(keys, key)
	}

	if contains(keys, args[0]) {
		return "", fmt.Errorf("invalid: %v not in %v", args[0], keys)
	}
	azcli := exec.Command("az", args...)

	// Create buffers to capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	azcli.Stdout = &stdoutBuf
	azcli.Stderr = &stderrBuf

	// Run the command
	err := azcli.Run()

	if err != nil {
		if handleErrorFunc != nil {
			handleErrorFunc(args, fmt.Errorf("%v", stderrBuf.String()))
		}
	}

	return stdoutBuf.String(), nil
}

func RunAzCommandPromptMissingArgs(args []string, promptUser func(string) (string,error)) (string, error) {

	return "", nil
}
