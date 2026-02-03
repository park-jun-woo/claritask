package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"parkjunwoo.com/claritask/internal/db"
)

// Init Phase constants
const (
	InitPhaseDBInit    = 1
	InitPhaseAnalysis  = 2
	InitPhaseApproval  = 3
	InitPhaseSpecsGen  = 4
	InitPhaseFeedback  = 5
	InitPhaseComplete  = 6
)

// InitConfig holds configuration for the init process
type InitConfig struct {
	ProjectID      string
	Name           string
	Description    string
	SkipAnalysis   bool
	SkipSpecs      bool
	NonInteractive bool
	Force          bool
	WorkDir        string // Working directory (default: current directory)
}

// InitState holds the current state of the init process
type InitState struct {
	Phase         int                    `json:"phase"`
	ProjectID     string                 `json:"project_id"`
	StartedAt     time.Time              `json:"started_at"`
	Tech          map[string]interface{} `json:"tech,omitempty"`
	Design        map[string]interface{} `json:"design,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	CurrentSpecs  string                 `json:"current_specs,omitempty"`
	SpecsRevision int                    `json:"specs_revision"`
}

// InitResult holds the result of the init process
type InitResult struct {
	Success   bool   `json:"success"`
	ProjectID string `json:"project_id"`
	DBPath    string `json:"db_path"`
	SpecsPath string `json:"specs_path,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ValidateProjectID validates the project ID format
func ValidateProjectID(projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	// Only allow lowercase letters, numbers, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-z0-9_-]+$`, projectID)
	if !matched {
		return fmt.Errorf("project ID must contain only lowercase letters, numbers, hyphens, and underscores")
	}
	return nil
}

// RunInit runs the entire init process
func RunInit(config InitConfig) (*InitResult, error) {
	// Validate project ID
	if err := ValidateProjectID(config.ProjectID); err != nil {
		return &InitResult{Success: false, Error: err.Error()}, err
	}

	// Set default work directory
	if config.WorkDir == "" {
		var err error
		config.WorkDir, err = os.Getwd()
		if err != nil {
			return &InitResult{Success: false, Error: err.Error()}, err
		}
	}

	// Set default name
	if config.Name == "" {
		config.Name = config.ProjectID
	}

	dbPath := filepath.Join(config.WorkDir, ".claritask", "db.clt")

	// Phase 1: DB Init
	PrintProgress(1, 5, "데이터베이스 초기화 중...")
	database, err := InitPhase1_DBInit(config)
	if err != nil {
		PrintError(err.Error())
		return &InitResult{Success: false, Error: err.Error()}, err
	}
	defer database.Close()

	// Save initial state
	state := &InitState{
		Phase:     InitPhaseDBInit,
		ProjectID: config.ProjectID,
		StartedAt: time.Now(),
	}
	if err := SaveInitState(database, state); err != nil {
		return &InitResult{Success: false, Error: err.Error()}, err
	}

	// If skip all, return early
	if config.SkipAnalysis && config.SkipSpecs {
		PrintInfo("분석 및 스펙 생성을 건너뜁니다.")
		state.Phase = InitPhaseComplete
		SaveInitState(database, state)
		result := &InitResult{
			Success:   true,
			ProjectID: config.ProjectID,
			DBPath:    dbPath,
		}
		PrintFinalResult(config.ProjectID, dbPath, "")
		return result, nil
	}

	// Phase 2: Analysis
	if !config.SkipAnalysis {
		PrintProgress(2, 5, "프로젝트 분석 중...")
		analysisResult, err := InitPhase2_Analysis(database, config.WorkDir, config.Description)
		if err != nil {
			PrintError(err.Error())
			return &InitResult{Success: false, Error: err.Error()}, err
		}

		state.Tech = analysisResult.Tech
		state.Design = analysisResult.Design
		state.Context = analysisResult.Context
		state.Phase = InitPhaseAnalysis
		SaveInitState(database, state)

		// Phase 3: Approval
		PrintProgress(3, 5, "분석 결과 승인 대기 중...")
		if err := InitPhase3_Approval(database, analysisResult, config.NonInteractive); err != nil {
			if err.Error() == "cancelled by user" {
				PrintInfo("사용자가 취소했습니다.")
				return &InitResult{Success: false, Error: "cancelled"}, nil
			}
			PrintError(err.Error())
			return &InitResult{Success: false, Error: err.Error()}, err
		}
		state.Phase = InitPhaseApproval
		SaveInitState(database, state)
	}

	// Phase 4 & 5: Specs generation and feedback
	specsPath := ""
	if !config.SkipSpecs {
		PrintProgress(4, 5, "스펙 문서 생성 중...")
		specs, err := InitPhase4_SpecsGen(database, config)
		if err != nil {
			PrintError(err.Error())
			return &InitResult{Success: false, Error: err.Error()}, err
		}

		state.CurrentSpecs = specs
		state.Phase = InitPhaseSpecsGen
		SaveInitState(database, state)

		// Phase 5: Feedback loop
		PrintProgress(5, 5, "스펙 문서 검토 중...")
		finalSpecs, err := InitPhase5_Feedback(database, specs, config.NonInteractive)
		if err != nil {
			if err.Error() == "cancelled by user" {
				PrintInfo("사용자가 취소했습니다.")
				return &InitResult{Success: false, Error: "cancelled"}, nil
			}
			PrintError(err.Error())
			return &InitResult{Success: false, Error: err.Error()}, err
		}

		// Save specs to file
		specsPath = filepath.Join(config.WorkDir, "specs", config.ProjectID+".md")
		if err := saveSpecsFile(specsPath, finalSpecs); err != nil {
			PrintError(err.Error())
			return &InitResult{Success: false, Error: err.Error()}, err
		}
	}

	// Mark complete
	state.Phase = InitPhaseComplete
	SaveInitState(database, state)

	result := &InitResult{
		Success:   true,
		ProjectID: config.ProjectID,
		DBPath:    dbPath,
		SpecsPath: specsPath,
	}

	PrintFinalResult(config.ProjectID, dbPath, specsPath)
	return result, nil
}

// InitPhase1_DBInit initializes the database
func InitPhase1_DBInit(config InitConfig) (*db.DB, error) {
	claritaskDir := filepath.Join(config.WorkDir, ".claritask")
	dbPath := filepath.Join(claritaskDir, "db.clt")

	// Check if DB already exists
	if _, err := os.Stat(dbPath); err == nil {
		if !config.Force {
			return nil, fmt.Errorf("database already exists at %s (use --force to overwrite)", dbPath)
		}
		// Remove existing DB and WAL files
		os.Remove(dbPath)
		os.Remove(dbPath + "-wal")
		os.Remove(dbPath + "-shm")
	}

	// Create directory
	if err := os.MkdirAll(claritaskDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .claritask directory: %w", err)
	}

	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create project record
	if err := CreateProject(database, config.ProjectID, config.Name, config.Description); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return database, nil
}

// InitPhase2_Analysis scans project and calls LLM for analysis
func InitPhase2_Analysis(database *db.DB, dir, description string) (*ContextAnalysisResult, error) {
	// Scan project files
	scanResult, err := ScanProjectFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	// Build prompt
	prompt := BuildContextAnalysisPrompt(scanResult, description)

	// Call LLM
	llmResponse, err := CallClaude(LLMRequest{
		Prompt:  prompt,
		Timeout: 120 * time.Second,
		Retries: 3,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	if !llmResponse.Success {
		return nil, fmt.Errorf("LLM call failed: %s", llmResponse.Error)
	}

	// Parse result
	result, err := ParseContextAnalysis(llmResponse.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM output: %w", err)
	}

	return result, nil
}

// InitPhase3_Approval handles user approval of analysis results
func InitPhase3_Approval(database *db.DB, result *ContextAnalysisResult, nonInteractive bool) error {
	if nonInteractive {
		// Auto-approve in non-interactive mode
		return saveAnalysisToTech(database, result)
	}

	for {
		// Display results
		PrintAnalysisResult(result.Tech, result.Design, result.Context)

		// Get user choice
		approval, err := PromptApproval(GetStandardApprovalOptions())
		if err != nil {
			return err
		}

		switch approval.Action {
		case "approve":
			return saveAnalysisToTech(database, result)

		case "edit":
			// Let user edit
			key, value, err := PromptEdit(result.Tech)
			if err != nil {
				PrintError(err.Error())
				continue
			}
			result.Tech[key] = value
			continue

		case "reanalyze":
			return fmt.Errorf("reanalyze requested")

		case "cancel":
			return fmt.Errorf("cancelled by user")
		}
	}
}

// saveAnalysisToTech saves analysis results to database
func saveAnalysisToTech(database *db.DB, result *ContextAnalysisResult) error {
	// Save tech
	if err := SetTech(database, result.Tech); err != nil {
		return fmt.Errorf("failed to save tech: %w", err)
	}

	// Save design
	if err := SetDesign(database, result.Design); err != nil {
		return fmt.Errorf("failed to save design: %w", err)
	}

	// Save context
	if err := SetContext(database, result.Context); err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	return nil
}

// InitPhase4_SpecsGen generates specs document
func InitPhase4_SpecsGen(database *db.DB, config InitConfig) (string, error) {
	// Load analysis results from DB
	state, err := LoadInitState(database)
	if err != nil {
		return "", err
	}

	// Build prompt
	prompt := BuildSpecsGenerationPrompt(
		config.ProjectID,
		config.Name,
		config.Description,
		state.Tech,
		state.Design,
		state.Context,
	)

	// Call LLM
	llmResponse, err := CallClaude(LLMRequest{
		Prompt:  prompt,
		Timeout: 180 * time.Second,
		Retries: 3,
	})
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	if !llmResponse.Success {
		return "", fmt.Errorf("LLM call failed: %s", llmResponse.Error)
	}

	// Parse result
	specs, err := ParseSpecsDocument(llmResponse.Output)
	if err != nil {
		return "", fmt.Errorf("failed to parse specs: %w", err)
	}

	return specs, nil
}

// InitPhase5_Feedback handles feedback loop for specs
func InitPhase5_Feedback(database *db.DB, specs string, nonInteractive bool) (string, error) {
	if nonInteractive {
		// Auto-approve in non-interactive mode
		return specs, nil
	}

	currentSpecs := specs
	revision := 0

	for {
		// Display specs
		PrintSpecs(currentSpecs)

		// Get user choice
		options := []PromptOption{
			{Key: "A", Label: "approve", Description: "승인하고 저장"},
			{Key: "F", Label: "feedback", Description: "피드백 제공"},
			{Key: "Q", Label: "cancel", Description: "취소"},
		}

		approval, err := PromptApproval(options)
		if err != nil {
			return "", err
		}

		switch approval.Action {
		case "approve":
			return currentSpecs, nil

		case "feedback":
			// Get feedback
			feedback, err := PromptMultilineInput("피드백을 입력하세요:")
			if err != nil {
				PrintError(err.Error())
				continue
			}

			// Call LLM for revision
			PrintInfo("스펙 문서를 수정 중...")
			prompt := BuildSpecsRevisionPrompt(currentSpecs, feedback)
			llmResponse, err := CallClaude(LLMRequest{
				Prompt:  prompt,
				Timeout: 180 * time.Second,
				Retries: 3,
			})
			if err != nil {
				PrintError(err.Error())
				continue
			}

			if !llmResponse.Success {
				PrintError(llmResponse.Error)
				continue
			}

			revisedSpecs, err := ParseSpecsDocument(llmResponse.Output)
			if err != nil {
				PrintError(err.Error())
				continue
			}

			currentSpecs = revisedSpecs
			revision++

			// Update state
			state, _ := LoadInitState(database)
			state.CurrentSpecs = currentSpecs
			state.SpecsRevision = revision
			SaveInitState(database, state)

		case "cancel":
			return "", fmt.Errorf("cancelled by user")
		}
	}
}

// SaveInitState saves the init state to database
func SaveInitState(database *db.DB, state *InitState) error {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return SetState(database, "init_state", string(stateJSON))
}

// LoadInitState loads the init state from database
func LoadInitState(database *db.DB) (*InitState, error) {
	stateJSON, err := GetState(database, "init_state")
	if err != nil {
		return nil, err
	}

	var state InitState
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// ResumeInit resumes an interrupted init process
func ResumeInit(workDir string) (*InitResult, error) {
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return &InitResult{Success: false, Error: err.Error()}, err
		}
	}

	dbPath := filepath.Join(workDir, ".claritask", "db.clt")

	// Check if DB exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return &InitResult{Success: false, Error: "no init in progress"}, fmt.Errorf("no init in progress")
	}

	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return &InitResult{Success: false, Error: err.Error()}, err
	}
	defer database.Close()

	// Load state
	state, err := LoadInitState(database)
	if err != nil {
		return &InitResult{Success: false, Error: "no init state found"}, fmt.Errorf("no init state found")
	}

	PrintInfo(fmt.Sprintf("Phase %d에서 재개합니다...", state.Phase))

	// Create config from state
	config := InitConfig{
		ProjectID: state.ProjectID,
		WorkDir:   workDir,
	}

	// Resume from the current phase
	// This is simplified - in a real implementation, you'd need to handle each phase
	return RunInit(config)
}

// saveSpecsFile saves specs to a file
func saveSpecsFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create specs directory: %w", err)
	}

	return os.WriteFile(path, []byte(content), 0644)
}
