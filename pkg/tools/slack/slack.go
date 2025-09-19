package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"

	"github.com/asaintsever/ama-employees-ai-agent/pkg/misc"
	"github.com/asaintsever/ama-employees-ai-agent/pkg/model"
)

const (
	maxUsersPerPage       = 500 // Recommended by Slack for optimal performance
	maxPaginationAttempts = 10  // Prevent infinite loops but allow up to 4000 users (20 * 200)
)

// SlackTool handles interactions with Slack API
type SlackTool struct {
	client *slack.Client
	token  string
}

// NewSlackTool creates a new instance of the Slack tool
func NewSlackTool(token string) *SlackTool {
	return &SlackTool{
		client: slack.New(token),
		token:  token,
	}
}

// FilterType defines the type of employee filter
type FilterType string

const (
	// FilterAll returns all employees
	FilterAll FilterType = "all"
	// FilterActive returns only active employees
	FilterActive FilterType = "active"
	// FilterDeactivated returns only deactivated employees
	FilterDeactivated FilterType = "deactivated"
)

// SearchAMAEmployees searches for employees on Slack
// filter parameter can be "all", "active", or "deactivated"
func (s *SlackTool) SearchAMAEmployees(filter FilterType) ([]model.EmployeeInfo, error) {
	spinner := misc.StartSpinner("🔌 Connecting to Slack workspace...")

	// Test the authentication
	authTest, err := s.client.AuthTest()

	misc.StopSpinner(spinner)

	if err != nil {
		return nil, fmt.Errorf("slack authentication failed: %v", err)
	}

	// Print success message after spinner is cleared
	fmt.Printf("✅ Successfully authenticated to Slack as %s in team %s\n", authTest.User, authTest.Team)

	var employees []model.EmployeeInfo
	fetchSpinner := misc.StartSpinner("🔍 Fetching employees data...")
	employees, err = s.searchAMAEmployeesUsingStandardAPI(filter)
	misc.StopSpinner(fetchSpinner)

	// Handle the result
	if err != nil {
		return nil, fmt.Errorf("error searching for employees: %v", err)
	}

	fmt.Printf("👤 Found %d employees\n", len(employees))
	return employees, nil
}

// searchAMAEmployeesUsingStandardAPI uses the standard Slack API to search for employees
// Uses GetUsersPaginated for efficient pagination
func (s *SlackTool) searchAMAEmployeesUsingStandardAPI(filter FilterType) ([]model.EmployeeInfo, error) {
	employees := []model.EmployeeInfo{}
	paginationCount := 0 // Start at 0 since the first page is just initialization
	totalUsers := 0
	ctx := context.Background()

	standardApiSpinner := misc.StartSpinner("📥 Fetching users with pagination...")

	// Get paginated users - this just initializes the pagination structure
	pagination := s.client.GetUsersPaginated(slack.GetUsersOptionLimit(maxUsersPerPage))

	// Process pages with actual fetching
	for paginationCount < maxPaginationAttempts {
		var err error
		pagination, err = pagination.Next(ctx)

		// Check if this is the end of pagination or if there's a failure
		if pagination.Done(err) {
			break
		}

		paginationCount++

		if pagination.Failure(err) != nil {
			fmt.Printf("❌ Error fetching next page: %v\n", pagination.Failure(err))
			break
		}

		fetchedCount := len(pagination.Users)
		totalUsers += fetchedCount

		// Process users from this page
		for _, user := range pagination.Users {
			if !user.IsBot {
				processUser(&employees, user, filter)
			}
		}
	}

	if paginationCount >= maxPaginationAttempts {
		fmt.Printf("⚠️ Reached maximum pagination attempts (%d), stopping\n", maxPaginationAttempts)
	}

	misc.StopSpinner(standardApiSpinner)
	fmt.Printf("✅ Completed fetching users via standard API (total: %d users)\n", totalUsers)
	return employees, nil
}

// processUser extracts information from a user and adds it to the employees slice
func processUser(employees *[]model.EmployeeInfo, user slack.User, filter FilterType) {
	// Parse the name parts
	nameParts := strings.Split(user.RealName, " ")
	firstName := user.Profile.FirstName
	lastName := user.Profile.LastName

	// If the profile doesn't have the parts populated, try to extract from real name
	if firstName == "" && len(nameParts) > 0 {
		firstName = nameParts[0]
	}

	if lastName == "" && len(nameParts) > 1 {
		lastName = nameParts[len(nameParts)-1]
	}

	deactivatedDate := ""

	if user.Deleted {
		// Generate a deactivated date from the user's last update time
		deactivatedDate = estimateDeactivatedDateFromJSON(user.Updated)
	}

	employee := model.EmployeeInfo{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           user.Profile.Email,
		Title:           user.Profile.Title,
		Deactivated:     user.Deleted,
		DeactivatedDate: deactivatedDate,
	}

	switch filter {
	case FilterAll:
		*employees = append(*employees, employee)
	case FilterDeactivated:
		if user.Deleted {
			*employees = append(*employees, employee)
		}
	case FilterActive:
		if !user.Deleted {
			*employees = append(*employees, employee)
		}
	}
}

// sortEmployeesByDeactivatedDateDesc sorts the given employees slice by deactivated date in descending order
// func sortEmployeesByDeactivatedDateDesc(employees []EmployeeInfo) {
// 	sort.Slice(employees, func(i, j int) bool {
// 		// Parse dates
// 		dateI, errI := time.Parse("2006-01-02", employees[i].DeactivatedDate)
// 		dateJ, errJ := time.Parse("2006-01-02", employees[j].DeactivatedDate)

// 		// Handle parsing errors (invalid dates come last)
// 		if errI != nil && errJ != nil {
// 			return false // Both invalid, order doesn't matter
// 		}
// 		if errI != nil {
// 			return false // i is invalid, j comes first
// 		}
// 		if errJ != nil {
// 			return true // j is invalid, i comes first
// 		}

// 		// Descending order (newer dates come first)
// 		return dateI.After(dateJ)
// 	})
// }

// estimateDeactivatedDateFromJSON generates a deactivated date based on Slack's JSONTime
// In a real implementation with admin access, we would get the actual deactivation date
func estimateDeactivatedDateFromJSON(jsonTime slack.JSONTime) string {
	// Use the Time() method to convert JSONTime to time.Time
	t := jsonTime.Time()

	// Format as YYYY-MM-DD
	return t.Format("2006-01-02")
}
