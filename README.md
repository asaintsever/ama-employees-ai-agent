# AMA (*Ask Me Anything*) Employees AI Agent

This Agent provides information about employees. It uses the LangChain Go framework to implement an AI agent that can autonomously decide how to get information about employees using assigned tools.

> [!NOTE]
>
> This Agent is a simple experiment with room for improvement (in particular, how to properly query the JSON data). As a user of Python's LangChain, I wanted to try out the Go version of the library and see how similar or different the implementation is. The "AMA employees" use case is just a jokish pretext to assess the capabilities of the library.

## Architecture

This project demonstrates an agentic AI approach using LangChain Go:

- **Agent**: Implemented using LangChain's ReAct framework (Reasoning + Acting)
- **Tools**: Custom tools that the agent can use to fetch data
- **LLM**: Anthropic Claude via AWS Bedrock

### LangChain Integration

The agent uses LangChain Go's agent framework to:

1. Process user queries about employees
2. Automatically determine when to use the tools (SlackAMAEmployeesTool and JSONQueryTool)
3. Format responses appropriately based on the user's request

## Tools

### Slack AMA Employees Tool

A custom LangChain tool that implements the [tools.Tool](https://github.com/tmc/langchaingo/blob/v0.1.13/tools/tool.go) interface to connect to your Slack workspace and fetch users information.

### JSON Query Tool

A tool that allows the agent to perform complex queries on JSON data. It relies on the [gojsonq](https://github.com/thedevsaddam/gojsonq/v2) library but is far from being perfect at interpreting the user's query.

> [!NOTE]
>
> A better approach would be to store the JSON dataset in a database and have the LLM generate the SQL query from the user's query.

## Technical details

### Project Structure

The project follows standard Go project layout conventions:

```text
.
├── cmd/
│   └── agent/          # Main application entry point
│       └── main.go
├── pkg/
│   ├── agent/          # Agent implementation
│   │   ├── agent.go
│   │   └── agent_test.go
│   ├── misc/           # Utilities
│   │   └── utils.go
│   ├── model/          # Shared data models
│   │   └── employee.go # Employee data structure
│   └── tools/
│       ├── json/       # JSON query tools implementation
│       │   ├── json_query.go
│       │   └── json_query_tool.go
│       └── slack/      # Slack tools implementation
│           ├── slack.go
│           └── slack_tool.go
├── Makefile           # Build and test commands
└── README.md
```

The Agent is implemented in Go and uses:

- **LangChain Go**: For agent framework, tool interfaces, and LLM integration
- **AWS SDK for Go**: With Bedrock Runtime client for Claude access
- **slack-go**: For Slack API integration

## Prerequisites

- Go 1.24.0 or higher
- AWS credentials configured (via aws sso login)
- Slack API token with access to workspace stats

## Setup

1. Login to AWS SSO:

   ```bash
   aws sso login
   ```

2. Set your Slack API token as environment variables:

   ```bash
   # Required: Slack API token
   export SLACK_TOKEN=xoxp-your-slack-token
   ```

3. Build the agent:

   ```bash
   make build
   ```

## Usage

Run the agent from the command line:

```bash
# Interactive mode
./target/ama-employees-ai-agent

# Non-interactive mode (process a single prompt and exit)
./target/ama-employees-ai-agent -prompt "Who are the latest 30 deactivated employees?"

# Minimal output mode (for scripting)
./target/ama-employees-ai-agent -quiet

# Debug mode (shows agent's decision-making process)
./target/ama-employees-ai-agent -debug
```

### Command-line Arguments

- `-prompt "your prompt here"`: Process a single prompt and exit (non-interactive mode)
- `-quiet`: Minimal output, only show responses (useful for scripting)
- `-debug`: Enable detailed debug output showing the agent's decision-making process

The Agent accepts prompts such as:

- "Who are the latest 30 deactivated employees?"
- "When was `<employee name>` deactivated?"
- "How many employees are active?"

## Testing

Run tests with:

```bash
# Standard tests
make test
```
