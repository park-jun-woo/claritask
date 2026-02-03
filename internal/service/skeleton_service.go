package service

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"parkjunwoo.com/claritask/internal/db"
	"parkjunwoo.com/claritask/internal/model"
)

// CreateSkeleton creates a skeleton record
func CreateSkeleton(database *db.DB, featureID int64, filePath, layer string) (int64, error) {
	checksum, _ := CalculateFileChecksum(filePath)
	now := db.TimeNow()

	result, err := database.Exec(
		`INSERT INTO skeletons (feature_id, file_path, layer, checksum, created_at) VALUES (?, ?, ?, ?, ?)`,
		featureID, filePath, layer, checksum, now,
	)
	if err != nil {
		return 0, fmt.Errorf("create skeleton: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}
	return id, nil
}

// GetSkeleton retrieves a skeleton by ID
func GetSkeleton(database *db.DB, id int64) (*model.Skeleton, error) {
	row := database.QueryRow(
		`SELECT id, feature_id, file_path, layer, checksum, created_at FROM skeletons WHERE id = ?`, id,
	)

	var s model.Skeleton
	var createdAt string
	err := row.Scan(&s.ID, &s.FeatureID, &s.FilePath, &s.Layer, &s.Checksum, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get skeleton: %w", err)
	}
	s.CreatedAt, _ = db.ParseTime(createdAt)
	return &s, nil
}

// ListSkeletonsByFeature lists all skeletons for a feature
func ListSkeletonsByFeature(database *db.DB, featureID int64) ([]model.Skeleton, error) {
	rows, err := database.Query(
		`SELECT id, feature_id, file_path, layer, checksum, created_at FROM skeletons WHERE feature_id = ? ORDER BY id`,
		featureID,
	)
	if err != nil {
		return nil, fmt.Errorf("list skeletons: %w", err)
	}
	defer rows.Close()

	var skeletons []model.Skeleton
	for rows.Next() {
		var s model.Skeleton
		var createdAt string
		if err := rows.Scan(&s.ID, &s.FeatureID, &s.FilePath, &s.Layer, &s.Checksum, &createdAt); err != nil {
			return nil, fmt.Errorf("scan skeleton: %w", err)
		}
		s.CreatedAt, _ = db.ParseTime(createdAt)
		skeletons = append(skeletons, s)
	}
	return skeletons, nil
}

// DeleteSkeleton deletes a skeleton
func DeleteSkeleton(database *db.DB, id int64) error {
	_, err := database.Exec(`DELETE FROM skeletons WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete skeleton: %w", err)
	}
	return nil
}

// DeleteSkeletonsByFeature deletes all skeletons for a feature
func DeleteSkeletonsByFeature(database *db.DB, featureID int64) error {
	_, err := database.Exec(`DELETE FROM skeletons WHERE feature_id = ?`, featureID)
	if err != nil {
		return fmt.Errorf("delete skeletons by feature: %w", err)
	}
	return nil
}

// CalculateFileChecksum calculates SHA256 checksum of a file
func CalculateFileChecksum(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}

// UpdateSkeletonChecksum updates skeleton checksum
func UpdateSkeletonChecksum(database *db.DB, id int64, checksum string) error {
	_, err := database.Exec(`UPDATE skeletons SET checksum = ? WHERE id = ?`, checksum, id)
	if err != nil {
		return fmt.Errorf("update skeleton checksum: %w", err)
	}
	return nil
}

// HasSkeletonChanged checks if skeleton file has changed
func HasSkeletonChanged(database *db.DB, id int64) (bool, error) {
	skeleton, err := GetSkeleton(database, id)
	if err != nil {
		return false, err
	}

	currentChecksum, err := CalculateFileChecksum(skeleton.FilePath)
	if err != nil {
		return true, nil // File doesn't exist, consider it changed
	}

	return currentChecksum != skeleton.Checksum, nil
}

// ReadSkeletonContent reads skeleton file content
func ReadSkeletonContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read skeleton file: %w", err)
	}
	return string(content), nil
}

// GetSkeletonAtLine reads code around a specific line
func GetSkeletonAtLine(filePath string, line, contextLines int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var lines []string

	startLine := line - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := line + contextLines

	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && lineNum <= endLine {
			lines = append(lines, fmt.Sprintf("%4d: %s", lineNum, scanner.Text()))
		}
		if lineNum > endLine {
			break
		}
	}

	return strings.Join(lines, "\n"), nil
}

// TODOLocation represents a TODO location in code
type TODOLocation struct {
	Line     int    `json:"line"`
	Function string `json:"function"`
	Content  string `json:"content"`
}

// ExtractTODOLocations extracts TODO locations from a file
func ExtractTODOLocations(filePath string) ([]TODOLocation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var locations []TODOLocation
	scanner := bufio.NewScanner(file)
	lineNum := 0
	todoPattern := regexp.MustCompile(`(?i)#\s*TODO|//\s*TODO|/\*\s*TODO|\*\s*TODO`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if todoPattern.MatchString(line) {
			funcName, _ := GetFunctionAtLine(filePath, lineNum)
			locations = append(locations, TODOLocation{
				Line:     lineNum,
				Function: funcName,
				Content:  strings.TrimSpace(line),
			})
		}
	}

	return locations, nil
}

// GetFunctionAtLine extracts function name at a specific line
func GetFunctionAtLine(filePath string, targetLine int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	// Patterns for different languages
	pythonDefPattern := regexp.MustCompile(`^\s*(async\s+)?def\s+(\w+)\s*\(`)
	goFuncPattern := regexp.MustCompile(`^\s*func\s+(?:\(\w+\s+\*?\w+\)\s+)?(\w+)\s*\(`)
	jsFuncPattern := regexp.MustCompile(`^\s*(async\s+)?function\s+(\w+)|^\s*(const|let|var)\s+(\w+)\s*=\s*(async\s+)?(\(|function)`)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	lastFuncName := ""

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check Python def
		if matches := pythonDefPattern.FindStringSubmatch(line); len(matches) > 2 {
			lastFuncName = matches[2]
		}
		// Check Go func
		if matches := goFuncPattern.FindStringSubmatch(line); len(matches) > 1 {
			lastFuncName = matches[1]
		}
		// Check JavaScript/TypeScript function
		if matches := jsFuncPattern.FindStringSubmatch(line); len(matches) > 0 {
			for i := len(matches) - 1; i >= 0; i-- {
				if matches[i] != "" && matches[i] != "const" && matches[i] != "let" && matches[i] != "var" && matches[i] != "async" && matches[i] != "(" && matches[i] != "function" {
					lastFuncName = matches[i]
					break
				}
			}
		}

		if lineNum >= targetLine {
			break
		}
	}

	return lastFuncName, nil
}

// SkeletonGeneratorResult represents skeleton generator result
type SkeletonGeneratorResult struct {
	GeneratedFiles []GeneratedFile `json:"generated_files"`
	Errors         []string        `json:"errors,omitempty"`
}

// GeneratedFile represents a generated skeleton file
type GeneratedFile struct {
	Path     string `json:"path"`
	Layer    string `json:"layer"`
	Checksum string `json:"checksum"`
}

// RunSkeletonGenerator runs Python skeleton generator
func RunSkeletonGenerator(fdlPath, outputDir, backend, frontend string, force bool) (*SkeletonGeneratorResult, error) {
	args := []string{
		"scripts/skeleton_generator.py",
		"--fdl", fdlPath,
		"--output-dir", outputDir,
		"--backend", backend,
		"--frontend", frontend,
	}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command("python3", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run skeleton generator: %w\nOutput: %s", err, string(output))
	}

	var result SkeletonGeneratorResult
	if err := json.Unmarshal(output, &result); err != nil {
		// If not JSON, assume success with file list in output
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.GeneratedFiles = append(result.GeneratedFiles, GeneratedFile{
					Path: strings.TrimSpace(line),
				})
			}
		}
	}

	return &result, nil
}

// RunSkeletonGeneratorDryRun returns list of files that would be generated
func RunSkeletonGeneratorDryRun(fdlPath, outputDir, backend, frontend string) (*SkeletonGeneratorResult, error) {
	args := []string{
		"scripts/skeleton_generator.py",
		"--fdl", fdlPath,
		"--output-dir", outputDir,
		"--backend", backend,
		"--frontend", frontend,
		"--dry-run",
	}

	cmd := exec.Command("python3", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run skeleton generator dry-run: %w\nOutput: %s", err, string(output))
	}

	var result SkeletonGeneratorResult
	if err := json.Unmarshal(output, &result); err != nil {
		// If not JSON, assume success with file list in output
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.GeneratedFiles = append(result.GeneratedFiles, GeneratedFile{
					Path: strings.TrimSpace(line),
				})
			}
		}
	}

	return &result, nil
}

// RunSkeletonGeneratorForFeature runs skeleton generator for a feature
func RunSkeletonGeneratorForFeature(database *db.DB, featureID int64, outputDir string, force bool) (*SkeletonGeneratorResult, error) {
	// Get feature
	feature, err := GetFeature(database, featureID)
	if err != nil {
		return nil, fmt.Errorf("get feature: %w", err)
	}

	if feature.FDL == "" {
		return nil, fmt.Errorf("no FDL defined for feature %d", featureID)
	}

	// Get tech stack
	tech, _ := GetTech(database)
	backend := "python"
	frontend := "none"
	if tech != nil {
		if b, ok := tech["backend"].(string); ok {
			backend = strings.ToLower(b)
		}
		if f, ok := tech["frontend"].(string); ok {
			frontend = strings.ToLower(f)
		}
	}

	// Write FDL to temp file
	tmpFile, err := os.CreateTemp("", "fdl-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(feature.FDL); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write FDL: %w", err)
	}
	tmpFile.Close()

	// Run generator
	result, err := RunSkeletonGenerator(tmpFile.Name(), outputDir, backend, frontend, force)
	if err != nil {
		return nil, err
	}

	// Record skeletons in database
	for _, f := range result.GeneratedFiles {
		_, err := CreateSkeleton(database, featureID, f.Path, f.Layer)
		if err != nil {
			// Log warning but continue
			result.Errors = append(result.Errors, fmt.Sprintf("failed to record skeleton: %v", err))
		}
	}

	// Mark feature as skeleton generated
	feature.SkeletonGenerated = true
	UpdateFeature(database, feature)

	return result, nil
}

// GetSkeletonInfo builds SkeletonInfo for task pop response
func GetSkeletonInfo(database *db.DB, skeletonID int64, targetLine int) (*model.SkeletonInfo, error) {
	skeleton, err := GetSkeleton(database, skeletonID)
	if err != nil {
		return nil, err
	}

	content := ""
	if targetLine > 0 {
		content, _ = GetSkeletonAtLine(skeleton.FilePath, targetLine, 10)
	} else {
		content, _ = ReadSkeletonContent(skeleton.FilePath)
		// Truncate if too long
		if len(content) > 2000 {
			content = content[:2000] + "\n... (truncated)"
		}
	}

	line := targetLine
	if line == 0 {
		line = 1
	}

	return &model.SkeletonInfo{
		File:    skeleton.FilePath,
		Line:    line,
		Content: content,
	}, nil
}
