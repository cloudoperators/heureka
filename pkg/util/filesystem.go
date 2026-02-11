// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func SetEnvVars(f string) error {
	//nolint: gosec
	file, err := os.Open(f)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("error during file closing: %s", err)
		}
	}()

	// Create a map to store the environment variables
	envVars := make(map[string]string)
	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// skip in case line length is to small
		if len(line) < 3 {
			continue
		}
		// Split the line into key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Store the key-value pair in the map
			envVars[key] = value
		}
	}
	// Check for any errors while reading the file
	if err := scanner.Err(); err != nil {
		return err
	}
	// Set the environment variables
	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	return err
}

func GetProjectRoot() (string, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}
	// Find the project root directory by traversing up the directory tree
	projectRoot := findProjectRoot(cwd)
	return projectRoot, nil
}

func findProjectRoot(cwd string) string {
	for {
		// Check if the current directory contains a specific file or directory
		// that indicates it as the project root (e.g., a go.mod file)
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}
		// Move up one directory level
		parent := filepath.Dir(cwd)
		// If we have reached the root directory ("/"), break the loop
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return ""
}
