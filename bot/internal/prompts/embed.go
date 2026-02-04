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

//go:embed dev.platform/*.md
var devPlatformFS embed.FS

//go:embed dev.cli/*.md
var devCliFS embed.FS

//go:embed write.webnovel/*.md
var writeWebnovelFS embed.FS

// Category represents a prompt category
type Category string

const (
	Common        Category = "common"
	DevPlatform   Category = "dev.platform"
	DevCLI        Category = "dev.cli"
	WriteWebnovel Category = "write.webnovel"
)

// Get returns the content of a prompt file
func Get(category Category, name string) (string, error) {
	var fsys embed.FS
	switch category {
	case Common:
		fsys = commonFS
	case DevPlatform:
		fsys = devPlatformFS
	case DevCLI:
		fsys = devCliFS
	case WriteWebnovel:
		fsys = writeWebnovelFS
	default:
		return "", fmt.Errorf("unknown category: %s", category)
	}

	filename := name
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	data, err := fsys.ReadFile(path.Join(string(category), filename))
	if err != nil {
		return "", fmt.Errorf("read prompt %s/%s: %w", category, name, err)
	}

	return string(data), nil
}

// List returns all prompt names in a category
func List(category Category) ([]string, error) {
	var fsys embed.FS
	switch category {
	case Common:
		fsys = commonFS
	case DevPlatform:
		fsys = devPlatformFS
	case DevCLI:
		fsys = devCliFS
	case WriteWebnovel:
		fsys = writeWebnovelFS
	default:
		return nil, fmt.Errorf("unknown category: %s", category)
	}

	var names []string
	err := fs.WalkDir(fsys, string(category), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".md") {
			name := strings.TrimPrefix(p, string(category)+"/")
			name = strings.TrimSuffix(name, ".md")
			names = append(names, name)
		}
		return nil
	})

	return names, err
}

// GetAll returns all prompts in a category concatenated
func GetAll(category Category) (string, error) {
	names, err := List(category)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, name := range names {
		content, err := Get(category, name)
		if err != nil {
			return "", err
		}
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	return sb.String(), nil
}
