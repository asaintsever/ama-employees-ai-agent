package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/asaintsever/ama-employees-ai-agent/pkg/agent"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Define styles for the terminal UI
var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4") // Purple
	secondaryColor = lipgloss.Color("#FF9D00") // Orange/gold
	accentColor    = lipgloss.Color("#FF5252") // Red for warnings/errors
	successColor   = lipgloss.Color("#00CC8F") // Green for success
)

// Text styles
var titleStyle = lipgloss.NewStyle().
	Foreground(primaryColor).
	Bold(true).
	MarginBottom(1)

var subtitleStyle = lipgloss.NewStyle().
	Foreground(secondaryColor).
	Bold(true)

var highlightStyle = lipgloss.NewStyle().
	Foreground(primaryColor).
	Bold(true)

var successStyle = lipgloss.NewStyle().
	Foreground(successColor)

var errorStyle = lipgloss.NewStyle().
	Foreground(accentColor).
	Bold(true)

var warningStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFCC00"))

var promptStyle = lipgloss.NewStyle().
	Foreground(secondaryColor).
	Bold(true)

var resultHeaderStyle = lipgloss.NewStyle().
	Foreground(successColor).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(primaryColor).
	Padding(0, 1).
	MarginLeft(0).
	Width(20).
	Align(lipgloss.Left).
	Bold(true)

// Box styles
var boxStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(primaryColor).
	Padding(1, 2).
	MarginTop(1).
	MarginBottom(1)

func main() {
	// Define command-line flags
	promptFlag := flag.String("prompt", "", "Prompt to process (non-interactive mode)")
	quietFlag := flag.Bool("quiet", false, "Minimal output, only show response (for scripting)")
	debugFlag := flag.Bool("debug", false, "Enable debug output to see agent's decision-making process")

	// Parse command-line flags
	flag.Parse()

	// Get Slack token from environment
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		errorMsg := errorStyle.Render("❌ ERROR: SLACK_TOKEN environment variable not set") + "\n" +
			"🔑 Please set it with your Slack OAuth token"
		errorBox := boxStyle.BorderForeground(accentColor).Render(errorMsg)
		fmt.Fprintln(os.Stderr, errorBox)
		os.Exit(1)
	}

	// Check for AWS credentials (except in quiet mode)
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" && !*quietFlag {
		warningMsg := warningStyle.Render("⚠️ Warning: No AWS credentials found") + "\n" +
			"🔄 Please run 'aws sso login' followed by 'aws configure export-credentials --format=env' before starting this agent\n" +
			"🔐 AWS credentials are required for Bedrock API access to Claude"
		warningBox := boxStyle.BorderForeground(lipgloss.Color("#FFCC00")).Render(warningMsg)
		fmt.Fprintln(os.Stderr, warningBox)
	}

	// Initialize agent
	if !*quietFlag {
		fmt.Println(highlightStyle.Render("🚀 Initializing AMA Employees AI Agent..."))
		// Small delay for visual effect
		time.Sleep(300 * time.Millisecond)
	}

	agent, err := agent.NewAgent(slackToken, *debugFlag)

	if err != nil {
		errorMsg := errorStyle.Render("❌ Error initializing agent:") + "\n" + err.Error()
		errorBox := boxStyle.BorderForeground(accentColor).Render(errorMsg)
		fmt.Fprintln(os.Stderr, errorBox)
		os.Exit(1)
	}

	// Non-interactive mode: process a single prompt and exit
	if *promptFlag != "" {
		if !*quietFlag {
			fmt.Println(highlightStyle.Render("⏳ Processing your query..."))
		}

		// Process the prompt
		response, err := agent.ProcessPrompt(*promptFlag)

		// No need for spinner cleanup

		if err != nil {
			errorMsg := errorStyle.Render("❌ Error processing prompt:") + "\n" + err.Error()
			errorBox := boxStyle.BorderForeground(accentColor).Render(errorMsg)
			fmt.Fprintln(os.Stderr, errorBox)
			os.Exit(1)
		}

		// Render markdown response in the terminal
		renderedResponse, err := renderMarkdown(response)
		if err != nil {
			fmt.Fprintf(os.Stderr, warningStyle.Render("⚠️ Error rendering markdown: %v\n"), err)
			// Fall back to plain text if rendering fails
			fmt.Println("📄 " + response)
		} else {
			// Show results in a nice box
			resultHeader := resultHeaderStyle.Render("📊 Results")
			fmt.Println(resultHeader)
			// Add a small margin to the rendered response for better alignment
			formattedResponse := lipgloss.NewStyle().
				MarginLeft(1).
				MarginTop(1).
				Render(renderedResponse)
			fmt.Print(formattedResponse)
			fmt.Println() // Add a newline at the end
		}
		os.Exit(0)
	}

	// Interactive mode
	if !*quietFlag {
		title := titleStyle.Render("👤 AMA Employees Agent")
		subtitle := subtitleStyle.Render("🔍 This Agent provides identities of employees")
		instructions := "💡 " + highlightStyle.Render("Type 'exit' to quit")

		welcomeContent := title + "\n\n" +
			subtitle + "\n" +
			instructions + "\n\n" +
			successStyle.Render("✅ Agent initialized successfully!")
		welcomeBox := boxStyle.BorderForeground(primaryColor).Render(welcomeContent)

		fmt.Println(welcomeBox)

		// Example queries in a separate box
		examplesBox := boxStyle.BorderForeground(secondaryColor).Render(
			subtitleStyle.Render("📝 Example queries:") + "\n\n" +
				"❓ " + highlightStyle.Render("Who are the latest 30 deactivated employees?") + "\n" +
				"❓ " + highlightStyle.Render("When <employee name> has been deactivated?"),
		)

		fmt.Println(examplesBox)
	}

	// Start CLI loop for interactive mode
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !*quietFlag {
			prompt := promptStyle.Render("🔎 > ")
			fmt.Print(prompt)
		}

		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		if strings.ToLower(input) == "exit" {
			if !*quietFlag {
				exitMsg := boxStyle.
					BorderForeground(successColor).
					Padding(0, 1).
					Render("👋 " + highlightStyle.Render("Exiting..."))
				fmt.Println(exitMsg)
			}
			break
		}

		// Process the prompt with or without visual feedback
		var response string
		var err error

		if !*quietFlag {
			// Process with timing
			fmt.Println(highlightStyle.Render("⏳ Processing your query..."))

			// Process the prompt
			startTime := time.Now()
			response, err = agent.ProcessPrompt(input)
			elapsedTime := time.Since(startTime)

			if err != nil {
				errorMsg := errorStyle.Render("❌ Error:") + "\n" + err.Error()
				errorBox := boxStyle.BorderForeground(accentColor).Render(errorMsg)
				fmt.Fprintln(os.Stderr, errorBox)
				continue
			}

			// Show success message with timing
			fmt.Printf("%s (completed in %s)\n",
				successStyle.Render("✨ Results found!"),
				highlightStyle.Render(elapsedTime.Round(time.Millisecond).String()))
		} else {
			// Quiet mode - just process without spinner
			response, err = agent.ProcessPrompt(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
				continue
			}
		}

		// Render markdown response in the terminal
		renderedResponse, err := renderMarkdown(response)
		if err != nil {
			fmt.Fprintf(os.Stderr, warningStyle.Render("⚠️ Error rendering markdown: %v\n"), err)
			// Fall back to plain text if rendering fails
			plainTextMsg := "📄 " + response
			fmt.Println(boxStyle.BorderForeground(secondaryColor).Render(plainTextMsg))
		} else {
			// Show results in a nice box
			resultHeader := resultHeaderStyle.Render("📊 Results")
			fmt.Println(resultHeader)
			// Add a small margin to the rendered response for better alignment
			formattedResponse := lipgloss.NewStyle().
				MarginLeft(1).
				MarginTop(1).
				Render(renderedResponse)
			fmt.Print(formattedResponse)
		}
		if !*quietFlag {
			fmt.Println()
		}
	}

	if scanner.Err() != nil {
		errorBox := boxStyle.BorderForeground(accentColor).Render(
			errorStyle.Render("❌ Error reading input:") + "\n" +
				scanner.Err().Error(),
		)
		fmt.Fprintln(os.Stderr, errorBox)
	}

	if !*quietFlag {
		// Create a fancy goodbye message
		goodbyeMsg := "👋 " + titleStyle.Render("Thank you for using the AMA Employees AI Agent!") + "\n\n" +
			subtitleStyle.Render("Have a great day! 👤✨")
		goodbyeBox := boxStyle.
			BorderForeground(successColor).
			Padding(1, 2).
			Render(goodbyeMsg)
		fmt.Println(goodbyeBox)
	}
}

// renderMarkdown renders markdown text as formatted terminal output
func renderMarkdown(markdown string) (string, error) {
	// Create a new renderer with dark theme and emoji support
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
		glamour.WithEmoji(),
	)
	if err != nil {
		return "", err
	}

	// Render the markdown
	return r.Render(markdown)
}
