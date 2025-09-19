package json

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/callbacks"
)

// JSONQueryTool implements the langchaingo Tool interface for querying JSON data
type JSONQueryTool struct {
	CallbacksHandler callbacks.Handler
	jsonQuery        *JSONQuery
}

// NewJSONQueryTool creates a new instance of JSONQueryTool
func NewJSONQueryTool() *JSONQueryTool {
	return &JSONQueryTool{
		jsonQuery: NewJSONQuery(),
	}
}

// Name returns the name of the tool
func (t *JSONQueryTool) Name() string {
	return "QueryJSON"
}

// Description returns a description of the tool for the AI to understand its purpose
func (t *JSONQueryTool) Description() string {
	return `Queries and manipulates JSON EmployeeInfo data to extract specific information.

This tool accepts a file path to a JSON file containing an array of EmployeeInfo objects, along with a query operation.

This tool can perform the following operations:
- Filter data based on field values (active/deactivated status)
- Sort data by deactivation date
- Limit results to a specific number
- Find specific employees by name
- Format results as a markdown table or text list

The input should be a JSON object with the following structure:
{
  "file_path": "<Path to the JSON file containing employee data>",
  "query": "<query string describing the operation to perform>"
}

Example queries:
- "Find the last 5 deactivated employees"
- "When John Doe was deactivated?"
- "List all deactivated engineering managers"
- "How many employees are active?"

The tool will return the query results as a string, formatted appropriately for the query type.`
}

// Call executes the tool with the given input
func (t *JSONQueryTool) Call(ctx context.Context, input string) (string, error) {
	// Start the tool execution
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	// Variables to store the result and error
	var output string
	var err error

	// Defer the end callback to ensure it's always called
	defer func() {
		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolEnd(ctx, output)
		}
	}()

	// Parse the input JSON
	var queryInput struct {
		FilePath string `json:"file_path"`
		Query    string `json:"query"`
	}

	err = json.Unmarshal([]byte(input), &queryInput)
	if err != nil {
		output = fmt.Sprintf("Error: %v", err)
		return "", fmt.Errorf("failed to parse input: %v", err)
	}

	// Verify file path is provided
	if queryInput.FilePath == "" {
		output = "Error: No file path provided"
		return "", fmt.Errorf("no file path provided")
	}

	// Clean up file path and ensure it exists
	filePath := filepath.Clean(queryInput.FilePath)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		output = fmt.Sprintf("Error: Could not access file at %s: %v", filePath, err)
		return "", fmt.Errorf("could not access file at %s: %v", filePath, err)
	}

	if fileInfo.IsDir() {
		output = fmt.Sprintf("Error: %s is a directory, not a file", filePath)
		return "", fmt.Errorf("%s is a directory, not a file", filePath)
	}

	// Read the file contents
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		output = fmt.Sprintf("Error: Failed to read file %s: %v", filePath, err)
		return "", fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	fmt.Printf("📄 Reading employee data from file: %s\n", filePath)

	// Process the query using the gojsonq implementation
	output, err = t.jsonQuery.ProcessQuery(fileContents, queryInput.Query)
	if err != nil {
		output = fmt.Sprintf("Error: %v", err)
		return "", err
	}

	return output, nil
}
