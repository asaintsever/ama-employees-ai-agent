package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
)

// SlackAMAEmployeesTool implements the langchaingo Tool interface
type SlackAMAEmployeesTool struct {
	CallbacksHandler callbacks.Handler
	slackTool        *SlackTool
}

// NewSlackAMAEmployeesTool creates a new instance of SlackAMAEmployeesTool
func NewSlackAMAEmployeesTool(token string) *SlackAMAEmployeesTool {
	return &SlackAMAEmployeesTool{
		slackTool: NewSlackTool(token),
	}
}

// Name returns the name of the tool
func (t *SlackAMAEmployeesTool) Name() string {
	return "SearchAMAEmployees"
}

// Description returns a description of the tool for the AI to understand its purpose
func (t *SlackAMAEmployeesTool) Description() string {
	return `Searches for employees information in Slack.

The input to this tool should specify which type of employees you want to retrieve:
- For all employees, use "all" or leave input empty
- For active employees only, include the word "active" in your input
- For deactivated/terminated/deleted employees only, include the word "deactivated" in your input

The tool returns a file path to a JSON file containing the employee data.

The JSON file contains an array of employee objects with the following structure:

[
    {
        "first_name": "John",
        "last_name": "Doe",
		"email": "john.doe@example.com",
		"deactivated": true,
        "deactivated_date": "2021-01-01",
        "title": "Software Engineer"
    },
	{
        "first_name": "Jane",
        "last_name": "Doe",
		"email": "jane.doe@example.com",
		"deactivated": false,
        "title": "Marketing Manager"
    }
]
`
}

// Call executes the tool with the given input
func (t *SlackAMAEmployeesTool) Call(ctx context.Context, input string) (string, error) {
	// Start the tool execution
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	// Variables to store the result and error
	var output string = ""
	var err error = nil

	// Defer the end callback to ensure it's always called
	defer func() {
		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolEnd(ctx, output)
		}
	}()

	// Determine filter type from input
	filter := FilterAll

	// Convert input to lowercase for case-insensitive comparison
	inputLower := strings.ToLower(input)

	// Check if input contains specific filter keywords
	if strings.Contains(inputLower, "active") && !strings.Contains(inputLower, "deactivated") {
		filter = FilterActive
	} else if strings.Contains(inputLower, "deactivated") {
		filter = FilterDeactivated
	}

	// Search for employees information with the determined filter
	employees, err := t.slackTool.SearchAMAEmployees(filter)
	if err != nil {
		output = fmt.Sprintf("Error: %v", err)
		return output, fmt.Errorf("error searching for employees information: %v", err)
	}

	// Convert the employees to JSON for writing to file
	employeesJSON, err := json.Marshal(employees)
	if err != nil {
		output = fmt.Sprintf("Error: %v", err)
		return output, fmt.Errorf("error marshalling employees data: %v", err)
	}

	// Create data directory if it doesn't exist
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		output = fmt.Sprintf("Error creating data directory: %v", err)
		return output, fmt.Errorf("error creating data directory: %v", err)
	}

	// Create a timestamped filename to avoid overwrites
	timestamp := time.Now().Format("20060102-150405")
	filterType := "all"
	switch filter {
	case FilterActive:
		filterType = "active"
	case FilterDeactivated:
		filterType = "deactivated"
	}

	fileName := fmt.Sprintf("employees-%s-%s.json", filterType, timestamp)
	filePath := filepath.Join(dataDir, fileName)

	// Write the JSON data to the file
	if err := os.WriteFile(filePath, employeesJSON, 0644); err != nil {
		output = fmt.Sprintf("Error writing employees data to file: %v", err)
		return output, fmt.Errorf("error writing employees data to file: %v", err)
	}

	// Get absolute path for better clarity
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath // Fall back to relative path if absolute fails
	}

	employeeCount := len(employees)
	output = fmt.Sprintf("Saved %d employees to file: %s", employeeCount, absPath)
	fmt.Printf("💾 Saved %d employees to file: %s\n", employeeCount, absPath)

	return absPath, nil
}
