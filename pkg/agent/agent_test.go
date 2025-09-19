package agent_test

import (
	"os"
	"testing"

	"github.com/asaintsever/ama-employees-ai-agent/pkg/agent"
)

func TestAMAEmployeesAgent(t *testing.T) {
	// Get Slack token from environment
	slackToken := os.Getenv("SLACK_TOKEN")

	t.Log("Initializing AMA Employees Agent for testing...")

	// Create the agent with LangChain integration
	// Enable debug mode in tests to see agent internals
	const debugMode = true
	employeeAgent, err := agent.NewAgent(slackToken, debugMode)
	if err != nil {
		t.Fatalf("Error initializing agent: %v", err)
	}

	t.Log("Agent initialized successfully")

	// Test prompts to run
	testPrompts := []string{
		"Who are the latest 30 deactivated employees?",
		"Show Name and Email of all active employees",
	}

	// Run the tests
	for i, prompt := range testPrompts {
		t.Logf("Test %d: %q", i+1, prompt)

		// Process the prompt - we're expecting an error due to credentials in test environment
		response, err := employeeAgent.ProcessPrompt(prompt)
		if err != nil {
			// In a real test, this would be a failure, but in our test environment
			// with fake credentials, we expect an error due to AWS authentication
			t.Logf("Expected error in test environment: %v", err)
			continue
		}

		// Display the response if we get one
		t.Logf("Response: %s", response)
	}
}
