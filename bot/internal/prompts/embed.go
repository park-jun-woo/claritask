package prompts

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed common/*.md
var commonFS embed.FS

// Get returns the content of a prompt file from common directory
func Get(name string) (string, error) {
	filename := name
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	data, err := commonFS.ReadFile(path.Join("common", filename))
	if err != nil {
		return "", fmt.Errorf("read prompt %s: %w", name, err)
	}

	return string(data), nil
}

// List returns all prompt names in common directory
func List() ([]string, error) {
	var names []string
	err := fs.WalkDir(commonFS, "common", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".md") {
			name := strings.TrimPrefix(p, "common/")
			name = strings.TrimSuffix(name, ".md")
			names = append(names, name)
		}
		return nil
	})

	return names, err
}

// GetAll returns all prompts concatenated
func GetAll() (string, error) {
	names, err := List()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, name := range names {
		content, err := Get(name)
		if err != nil {
			return "", err
		}
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	return sb.String(), nil
}
