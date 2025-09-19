package misc

import (
	"time"

	"github.com/briandowns/spinner"
)

// Spinner represents a spinner instance
type Spinner = *spinner.Spinner

// StartSpinner starts a spinner animation with the given message
// It returns a Spinner that can be stopped using StopSpinner
// Usage:
//
//	s := StartSpinner("Processing")
//	// do work
//	StopSpinner(s)
//	// Print your success message here
func StartSpinner(message string) Spinner {
	// Create a new spinner with dot style and 100ms update frequency
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()
	return s
}

// StopSpinner stops the spinner animation
// This is a blocking call that ensures the spinner is fully stopped
// before returning
func StopSpinner(s Spinner) {
	s.Stop()
}
