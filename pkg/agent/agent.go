package agent

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
	"github.com/tmc/langchaingo/tools"

	"github.com/asaintsever/ama-employees-ai-agent/pkg/tools/json"
	"github.com/asaintsever/ama-employees-ai-agent/pkg/tools/slack"
)

// Agent represents the AMA Employees Agent
type Agent struct {
	bedrockClient *bedrockruntime.Client
	llm           llms.Model
	agentExecutor *agents.Executor
	slackTool     *slack.SlackAMAEmployeesTool
	jsonQueryTool *json.JSONQueryTool
}

// NewAgent creates a new instance of the AMA Employees Agent
func NewAgent(slackToken string, debug bool) (*Agent, error) {
	// Configure AWS SDK to use SSO login
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	// Create a Bedrock client for Claude
	bedrockClient := bedrockruntime.NewFromConfig(cfg)

	// Initialize tools
	slackTool := slack.NewSlackAMAEmployeesTool(slackToken)
	jsonQueryTool := json.NewJSONQueryTool()

	// Create a bedrock LLM for the agent
	llm, err := bedrock.New(
		bedrock.WithClient(bedrockClient),
		bedrock.WithModel("anthropic.claude-3-5-sonnet-20241022-v2:0"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Bedrock LLM: %v", err)
	}

	// Create tools array
	tools := []tools.Tool{
		slackTool,
		jsonQueryTool,
	}

	// Initialize the agent executor with custom prompt
	// IMPORTANT: we MUST prepend the response with "Final Answer: " to avoid parsing errors (see https://github.com/tmc/langchaingo/blob/v0.1.13/agents/mrkl.go#L135)
	agentPrompt := `Today is {{.today}}.
You are the AMA Employees Agent, designed to provide information about employees.
Focus only on providing the requested information about employees as asked.
Adopt a neutral tone and be super concise, do not share thoughts or reasoning.

Do not summarize the results, just provide the results as is in markdown format.
Always prepend the response with "Final Answer: ".

You have access to the following tools:
	
{{.tool_descriptions}}`

	// Create a Zero-Shot ReAct agent
	// Prepare agent options
	agentOpts := []agents.Option{agents.WithPromptPrefix(agentPrompt)}

	// Add debug logging if debug mode is enabled
	if debug {
		fmt.Println("🔍 Debug mode enabled - detailed agent operations will be logged")
		var logHandler callbacks.Handler = callbacks.LogHandler{}

		agentOpts = append(agentOpts, agents.WithCallbacksHandler(logHandler))
		slackTool.CallbacksHandler = logHandler
		jsonQueryTool.CallbacksHandler = logHandler
	}

	// Create the agent with options
	zeroShotAgent := agents.NewOneShotAgent(
		llm,
		tools,
		agentOpts...,
	)

	// Create the executor with the agent
	agentExecutor := agents.NewExecutor(
		zeroShotAgent,
		agents.WithMaxIterations(5),
	)
	// No error handling needed here as NewOneShotAgent and NewExecutor don't return errors

	return &Agent{
		bedrockClient: bedrockClient,
		llm:           llm,
		agentExecutor: agentExecutor,
		slackTool:     slackTool,
		jsonQueryTool: jsonQueryTool,
	}, nil
}

// ProcessPrompt processes user prompts and returns responses
func (a *Agent) ProcessPrompt(prompt string) (string, error) {
	ctx := context.Background()

	// Run the agent executor
	result, err := a.agentExecutor.Call(
		ctx,
		map[string]any{"input": prompt},
	)

	// Check for parsing errors in the LangChain executor
	if err != nil {
		return "", fmt.Errorf("error running agent executor: %v", err)
	}

	// Extract the output from the result
	outputInterface, ok := result["output"]
	if !ok {
		return "", fmt.Errorf("missing output key in agent response")
	}

	output, ok := outputInterface.(string)
	if !ok {
		return "", fmt.Errorf("output is not a string")
	}

	return output, nil
}
