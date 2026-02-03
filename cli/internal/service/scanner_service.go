package service

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// ScannedFile represents a scanned file with its metadata
type ScannedFile struct {
	Path    string `json:"path"`
	Type    string `json:"type"`    // "package_json", "go_mod", "readme", etc.
	Content string `json:"content"` // File content (partial)
}

// ScanResult represents the result of scanning a project directory
type ScanResult struct {
	Files       []ScannedFile `json:"files"`
	Directories []string      `json:"directories"` // src/, internal/, lib/, etc.
}

// fileTypeMap maps file names to their types
var fileTypeMap = map[string]string{
	"package.json":      "package_json",
	"go.mod":            "go_mod",
	"requirements.txt":  "requirements_txt",
	"pyproject.toml":    "pyproject_toml",
	"Cargo.toml":        "cargo_toml",
	"pom.xml":           "pom_xml",
	"build.gradle":      "build_gradle",
	"README.md":         "readme",
	"readme.md":         "readme",
	"docker-compose.yml": "docker_compose",
	"docker-compose.yaml": "docker_compose",
	"Makefile":          "makefile",
	".env.example":      "env_example",
}

// knownDirectories lists directories to check for existence
var knownDirectories = []string{
	"src",
	"internal",
	"lib",
	"pkg",
	"cmd",
	"api",
	"app",
	"test",
	"tests",
	"docs",
}

// ScanProjectFiles scans a project directory and collects metadata files
func ScanProjectFiles(dir string) (*ScanResult, error) {
	result := &ScanResult{
		Files:       []ScannedFile{},
		Directories: []string{},
	}

	// Scan for known files
	for filename, fileType := range fileTypeMap {
		filePath := filepath.Join(dir, filename)
		if _, err := os.Stat(filePath); err == nil {
			content, err := readFilePartial(filePath, 500)
			if err != nil {
				continue // Skip files that can't be read
			}
			result.Files = append(result.Files, ScannedFile{
				Path:    filename,
				Type:    fileType,
				Content: content,
			})
		}
	}

	// Scan for known directories
	for _, dirname := range knownDirectories {
		dirPath := filepath.Join(dir, dirname)
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			result.Directories = append(result.Directories, dirname)
		}
	}

	return result, nil
}

// readFilePartial reads up to maxLines lines from a file
func readFilePartial(filePath string, maxLines int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()
		// Check if the line is valid UTF-8 (not binary)
		if !utf8.ValidString(line) {
			return "", fmt.Errorf("binary file detected")
		}
		lines = append(lines, line)
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

// FormatScanResultForLLM formats scan result for LLM prompt
func FormatScanResultForLLM(result *ScanResult) string {
	var sb strings.Builder

	sb.WriteString("## Project Structure\n\n")

	// Directories
	if len(result.Directories) > 0 {
		sb.WriteString("### Directories\n")
		for _, dir := range result.Directories {
			sb.WriteString(fmt.Sprintf("- %s/\n", dir))
		}
		sb.WriteString("\n")
	}

	// Files
	if len(result.Files) > 0 {
		sb.WriteString("### Configuration Files\n\n")
		for _, file := range result.Files {
			sb.WriteString(fmt.Sprintf("#### %s (%s)\n", file.Path, file.Type))
			sb.WriteString("```\n")
			// Limit content to first 100 lines for LLM prompt
			lines := strings.Split(file.Content, "\n")
			if len(lines) > 100 {
				lines = lines[:100]
				sb.WriteString(strings.Join(lines, "\n"))
				sb.WriteString("\n... (truncated)\n")
			} else {
				sb.WriteString(file.Content)
				sb.WriteString("\n")
			}
			sb.WriteString("```\n\n")
		}
	}

	return sb.String()
}
