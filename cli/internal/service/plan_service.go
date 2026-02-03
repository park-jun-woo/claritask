package service

import (
	"fmt"

	"parkjunwoo.com/claritask/internal/db"
)

// PlannedFeature represents a feature suggested by LLM planning
type PlannedFeature struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Priority     int      `json:"priority"` // 1: core, 2: important, 3: nice-to-have
	Dependencies []string `json:"dependencies"`
}

// FeaturePlan represents the result of feature planning
type FeaturePlan struct {
	Features   []PlannedFeature `json:"features"`
	TotalCount int              `json:"total_count"`
	Reasoning  string           `json:"reasoning"`
	Prompt     string           `json:"prompt"`
}

// FeaturePlanContext provides context for LLM feature planning
type FeaturePlanContext struct {
	ProjectName string                 `json:"project_name"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Tech        map[string]interface{} `json:"tech,omitempty"`
	Design      map[string]interface{} `json:"design,omitempty"`
	Existing    []ExistingFeature      `json:"existing_features,omitempty"`
}

// ExistingFeature represents an already defined feature
type ExistingFeature struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Spec string `json:"spec,omitempty"`
}

// PreparePlanFeatures prepares context for LLM to plan features
func PreparePlanFeatures(database *db.DB) (*FeaturePlan, error) {
	plan := &FeaturePlan{
		Features: []PlannedFeature{},
	}

	// Get project
	project, err := GetProject(database)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}

	// Build context
	ctx := FeaturePlanContext{
		ProjectName: project.ID,
	}

	// Get context
	context, _ := GetContext(database)
	if context != nil {
		ctx.Context = context
	}

	// Get tech
	tech, _ := GetTech(database)
	if tech != nil {
		ctx.Tech = tech
	}

	// Get design
	design, _ := GetDesign(database)
	if design != nil {
		ctx.Design = design
	}

	// Get existing features
	features, _ := ListFeatures(database, project.ID)
	for _, f := range features {
		ctx.Existing = append(ctx.Existing, ExistingFeature{
			ID:   f.ID,
			Name: f.Name,
			Spec: f.Spec,
		})
	}

	// Build prompt
	plan.Prompt = buildFeaturePlanPrompt(ctx)

	return plan, nil
}

// buildFeaturePlanPrompt creates the LLM prompt for feature planning
func buildFeaturePlanPrompt(ctx FeaturePlanContext) string {
	prompt := fmt.Sprintf("You are analyzing a software project named '%s' to identify and plan features.\n\n", ctx.ProjectName)

	if ctx.Context != nil {
		prompt += "## Project Context\n"
		for key, value := range ctx.Context {
			prompt += fmt.Sprintf("- %s: %v\n", key, value)
		}
		prompt += "\n"
	}

	if ctx.Tech != nil {
		prompt += "## Technology Stack\n"
		for key, value := range ctx.Tech {
			prompt += fmt.Sprintf("- %s: %v\n", key, value)
		}
		prompt += "\n"
	}

	if ctx.Design != nil {
		prompt += "## Design Decisions\n"
		for key, value := range ctx.Design {
			prompt += fmt.Sprintf("- %s: %v\n", key, value)
		}
		prompt += "\n"
	}

	if len(ctx.Existing) > 0 {
		prompt += "## Existing Features (do not duplicate)\n"
		for _, f := range ctx.Existing {
			prompt += fmt.Sprintf("- %s", f.Name)
			if f.Spec != "" {
				prompt += fmt.Sprintf(": %s", f.Spec)
			}
			prompt += "\n"
		}
		prompt += "\n"
	}

	prompt += `## Instructions

Based on the project context, identify features that need to be implemented.
For each feature, provide:
1. name: A short, descriptive identifier (use snake_case or kebab-case)
2. description: What this feature does
3. priority: 1 (core - required for MVP), 2 (important - needed for good UX), 3 (nice-to-have)
4. dependencies: List of feature names this feature depends on

Prioritization guidelines:
- Priority 1: Features critical for the application to function at all
- Priority 2: Features that significantly enhance usability or value
- Priority 3: Additional features for polish or edge cases

Return a JSON object with this structure:
{
  "features": [
    {
      "name": "feature_name",
      "description": "What this feature does",
      "priority": 1,
      "dependencies": ["other_feature"]
    }
  ],
  "reasoning": "Brief explanation of your analysis approach"
}`

	return prompt
}
