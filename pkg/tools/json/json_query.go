package json

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/asaintsever/ama-employees-ai-agent/pkg/model"
	"github.com/thedevsaddam/gojsonq/v2"
)

// JSONQuery provides functionality for querying and manipulating JSON data
type JSONQuery struct{}

// NewJSONQuery creates a new instance of JSONQuery
func NewJSONQuery() *JSONQuery {
	return &JSONQuery{}
}

// ProcessQuery handles different types of queries on employee data using gojsonq
func (q *JSONQuery) ProcessQuery(jsonData []byte, query string) (string, error) {
	fmt.Printf("🔍 Processing query: %s\n", query)

	// Create a new gojsonq instance with the JSON data
	jq := gojsonq.New().FromString(string(jsonData))

	// Count total employees before any filtering
	totalCount := jq.Count()
	fmt.Printf("📊 Initial dataset: %d employees\n", totalCount)

	// Reset the query to start fresh
	jq.Reset()

	// Convert query to lowercase for case-insensitive matching
	query = strings.ToLower(query)

	// Apply filters based on query
	if strings.Contains(query, "deactivat") || strings.Contains(query, "terminat") {
		jq.Where("deactivated", "=", true)
		fmt.Println("🔎 Filtered to deactivated employees")
	} else if strings.Contains(query, "active") && !strings.Contains(query, "deactivat") {
		jq.Where("deactivated", "=", false)
		fmt.Println("🔎 Filtered to active employees")
	}

	// Check if we need to find a specific employee
	if q.isSpecificEmployeeSearch(query) {
		fmt.Println("🔍 Searching for specific employee...")
		return q.findSpecificEmployee(jq, query)
	}

	// Get the filtered data
	result := jq.Get()

	// Convert result to []model.EmployeeInfo
	var employees []model.EmployeeInfo
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), err
	}

	err = json.Unmarshal(resultBytes, &employees)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), err
	}

	fmt.Printf("🔎 Found %d employees after filtering\n", len(employees))

	// Sort by deactivation date if needed
	if strings.Contains(query, "last") || strings.Contains(query, "recent") ||
		strings.Contains(query, "sort by date") || strings.Contains(query, "sort by deactivation") {
		// Sort employees by deactivation date
		sort.Slice(employees, func(i, j int) bool {
			dateI := employees[i].DeactivatedDate
			dateJ := employees[j].DeactivatedDate

			// Handle empty dates
			if dateI == "" && dateJ == "" {
				return false
			}
			if dateI == "" {
				return false
			}
			if dateJ == "" {
				return true
			}

			// Parse dates
			timeI, errI := time.Parse("2006-01-02", dateI)
			timeJ, errJ := time.Parse("2006-01-02", dateJ)

			if errI != nil && errJ != nil {
				return false
			}
			if errI != nil {
				return false
			}
			if errJ != nil {
				return true
			}

			// Sort descending (most recent first)
			return timeI.After(timeJ)
		})
		fmt.Println("📅 Sorted employees by deactivation date (most recent first)")
	}

	// Limit results if needed
	originalCount := len(employees)

	// Look for patterns like "last 5", "top 10", "50 employees", etc.
	words := strings.Fields(query)
	var limitApplied bool

	// First look for explicit numeric limits
	for i, word := range words {
		// Check for "last X", "top X", "latest X" patterns
		if (word == "last" || word == "top" || word == "latest") && i+1 < len(words) {
			// Try to parse the next word as a number
			if num, err := strconv.Atoi(words[i+1]); err == nil && num > 0 {
				if num < len(employees) {
					employees = employees[:num]
					limitApplied = true
					break
				}
			}
		}

		// Check for "X employees" pattern
		if i+1 < len(words) && (words[i+1] == "employees" || words[i+1] == "employee") {
			if num, err := strconv.Atoi(word); err == nil && num > 0 {
				if num < len(employees) {
					employees = employees[:num]
					limitApplied = true
					break
				}
			}
		}
	}

	if limitApplied && len(employees) < originalCount {
		fmt.Printf("📏 Limited results to %d employees\n", len(employees))
	}

	// Format the results
	fmt.Printf("📝 Formatting results for %d employees\n", len(employees))
	if strings.Contains(query, "table") || strings.Contains(query, "markdown") {
		fmt.Println("📋 Using markdown table format")
		return q.FormatAsMarkdownTable(employees)
	}

	// Default formatting
	fmt.Println("📋 Using default list format")
	return q.FormatResults(employees)
}

// findSpecificEmployee searches for a specific employee by name using gojsonq
func (q *JSONQuery) findSpecificEmployee(jq *gojsonq.JSONQ, query string) (string, error) {
	// Extract potential names from the query
	words := strings.Fields(query)

	// Try different combinations of adjacent words as potential names
	for i := 0; i < len(words)-1; i++ {
		potentialFirstName := words[i]
		potentialLastName := words[i+1]

		// Skip short words and common words that are unlikely to be names
		if len(potentialFirstName) < 3 || len(potentialLastName) < 3 {
			continue
		}

		// Reset query for new search
		jq.Reset()

		// Search for first name and last name
		result := jq.OrWhere("first_name", "contains", potentialFirstName).
			OrWhere("last_name", "contains", potentialLastName).Get()

		// Convert result to []model.EmployeeInfo
		var employees []model.EmployeeInfo
		resultBytes, err := json.Marshal(result)
		if err != nil {
			continue
		}

		err = json.Unmarshal(resultBytes, &employees)
		if err != nil || len(employees) == 0 {
			continue
		}

		// Found at least one matching employee
		fmt.Println("✅ Employee found!")

		// Format the first matching employee
		var resultBuilder strings.Builder
		emp := employees[0]

		resultBuilder.WriteString(fmt.Sprintf("Employee: %s %s\n", emp.FirstName, emp.LastName))

		if emp.Title != "" {
			resultBuilder.WriteString(fmt.Sprintf("Title: %s\n", emp.Title))
		}

		if emp.Email != "" {
			resultBuilder.WriteString(fmt.Sprintf("Email: %s\n", emp.Email))
		}

		if emp.Deactivated {
			resultBuilder.WriteString("Status: Deactivated\n")
			if emp.DeactivatedDate != "" {
				resultBuilder.WriteString(fmt.Sprintf("Deactivation Date: %s\n", emp.DeactivatedDate))
			}
		} else {
			resultBuilder.WriteString("Status: Active\n")
		}

		return resultBuilder.String(), nil
	}

	fmt.Println("❌ Employee not found")
	return "Employee not found in the dataset.", nil
}

// FormatAsMarkdownTable formats the employee data as a markdown table
func (q *JSONQuery) FormatAsMarkdownTable(employees []model.EmployeeInfo) (string, error) {
	if len(employees) == 0 {
		return "No employees found matching the criteria.", nil
	}

	var result strings.Builder

	// Write table header
	result.WriteString("| Name | Title | Email | Status | Deactivation Date |\n")
	result.WriteString("|------|-------|-------|--------|------------------|\n")

	// Write table rows
	for _, emp := range employees {
		name := emp.FirstName + " " + emp.LastName

		status := "Active"
		deactivationDate := ""

		if emp.Deactivated {
			status = "Deactivated"
			deactivationDate = emp.DeactivatedDate
		}

		result.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			name, emp.Title, emp.Email, status, deactivationDate))
	}

	return result.String(), nil
}

// isSpecificEmployeeSearch determines if the query is looking for a specific person
func (q *JSONQuery) isSpecificEmployeeSearch(query string) bool {
	// Common patterns for specific employee searches
	specificPatterns := []string{
		"when was", "when did", "what date", "who is", "information about", "details for", "details about",
		"find employee", "search for", "look for", "locate", "get info on",
	}

	// Check if query contains any specific employee search patterns
	for _, pattern := range specificPatterns {
		if strings.Contains(query, pattern) {
			return true
		}
	}

	// Check for "find" followed by what appears to be a name (not "find last X" pattern)
	if strings.Contains(query, "find") {
		// Check if it's a "find last X" or "find top X" pattern
		if strings.Contains(query, "find last") || strings.Contains(query, "find top") ||
			strings.Contains(query, "find the last") || strings.Contains(query, "find the top") {
			return false
		}

		// If it contains "find" but not in the patterns above, it's likely a specific search
		return true
	}

	return false
}

// FormatResults formats the employee data as a simple text list
func (q *JSONQuery) FormatResults(employees []model.EmployeeInfo) (string, error) {
	if len(employees) == 0 {
		return "No employees found matching the criteria.", nil
	}

	var result strings.Builder

	result.WriteString(fmt.Sprintf("Found %d employees:\n\n", len(employees)))

	for i, emp := range employees {
		result.WriteString(fmt.Sprintf("%d. %s %s", i+1, emp.FirstName, emp.LastName))

		if emp.Title != "" {
			result.WriteString(fmt.Sprintf(" - %s", emp.Title))
		}

		if emp.Deactivated {
			if emp.DeactivatedDate != "" {
				result.WriteString(fmt.Sprintf(" (Deactivated on %s)", emp.DeactivatedDate))
			} else {
				result.WriteString(" (Deactivated)")
			}
		}

		result.WriteString("\n")
	}

	return result.String(), nil
}
