package model

// EmployeeInfo contains information about an employee
type EmployeeInfo struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	Title           string `json:"title"`
	Deactivated     bool   `json:"deactivated"`
	DeactivatedDate string `json:"deactivated_date,omitempty"`
}
