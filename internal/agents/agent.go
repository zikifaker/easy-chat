package agents

import (
	"bytes"
	"context"
	"easy-chat/internal/agents/llms"
	"easy-chat/internal/agents/memory"
	"easy-chat/internal/agents/prompts"
	"easy-chat/internal/agents/rag"
	"easy-chat/internal/agents/toolkit"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"
)

var (
	ErrToolNotFound      = errors.New("tool not found")
	ErrWhilePlanningStep = errors.New("error while planning step")
	ErrWhileCallingTool  = errors.New("error while calling tool")
)

type Agent struct {
	LLM     llms.LLM
	Tools   []toolkit.Tool
	Memory  memory.Memory
	MaxStep int
	RAG     rag.RAG
}

type Step struct {
	Thought     string `json:"thought"`
	Action      string `json:"action"`
	ActionInput string `json:"action_input"`
	Observation string `json:"observation"`
	FinalAnswer string `json:"final_answer"`
}

func NewAgent(llm llms.LLM, tools []toolkit.Tool, options ...Option) (*Agent, error) {
	opts := GetDefaultOptions()
	for _, opt := range options {
		opt(opts)
	}

	return &Agent{
		LLM:     llm,
		Tools:   tools,
		Memory:  opts.Memory,
		MaxStep: opts.MaxStep,
	}, nil
}

func (a *Agent) Execute(ctx context.Context, input string) (string, error) {
	var finalAnswer string
	var immediateSteps []Step
	toolMap := a.buildToolMap()
	for i := 0; i < a.MaxStep; i++ {
		step, err := a.plan(ctx, input, immediateSteps)
		if err != nil {
			log.Printf("%v: %v", ErrWhilePlanningStep, err)
			return "", err
		}

		if step.FinalAnswer != "" {
			finalAnswer = step.FinalAnswer
			break
		}

		if step.Action != "" {
			observation, err := callTool(ctx, toolMap, step)
			if err != nil {
				log.Printf("%v: %s", ErrWhileCallingTool, step.Action)
				step.Observation = err.Error()
			} else {
				log.Println("observation: ", observation)
				step.Observation = observation
			}
		}

		immediateSteps = append(immediateSteps, *step)
	}

	a.Memory.AddMessage(ctx, []memory.Message{
		{Role: memory.MessageRoleUser, Content: input},
		{Role: memory.MessageRoleAI, Content: finalAnswer},
	})

	return finalAnswer, nil
}

func (a *Agent) buildToolMap() map[string]toolkit.Tool {
	toolMap := make(map[string]toolkit.Tool)
	for _, tool := range a.Tools {
		toolMap[strings.ToUpper(tool.Name())] = tool
	}
	return toolMap
}

func (a *Agent) plan(ctx context.Context, input string, immediateSteps []Step) (*Step, error) {
	toolDetail := a.getToolDetail()
	toolNames := a.getToolNames()
	chatHistory := a.getChatHistory(ctx)
	agentScratchpad := buildAgentScratchpad(immediateSteps)

	prompt, err := renderPromptTemplate(prompts.ReActPromptTemplate, map[string]any{
		"max_step":         a.MaxStep,
		"current_step":     len(immediateSteps) + 1,
		"tool_detail":      toolDetail,
		"tool_names":       toolNames,
		"chat_history":     chatHistory,
		"agent_scratchpad": agentScratchpad,
		"question":         input,
	})
	if err != nil {
		return nil, err
	}

	result, err := a.LLM.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, err
	}
	log.Println("agent step output:\n", result)

	step, err := parseOutput(result)
	if err != nil {
		return nil, err
	}
	return step, nil
}

func (a *Agent) getToolDetail() string {
	var result strings.Builder
	for _, tool := range a.Tools {
		result.WriteString(tool.Name() + ": " + tool.Description() + "\n\n")
	}
	return result.String()
}

func (a *Agent) getToolNames() string {
	names := make([]string, len(a.Tools))
	for i, tool := range a.Tools {
		names[i] = tool.Name()
	}
	return strings.Join(names, ",")
}

func (a *Agent) getChatHistory(ctx context.Context) string {
	var result strings.Builder
	messages := a.Memory.GetMessages(ctx)
	for _, message := range messages {
		result.WriteString(message.Role + ": " + message.Content + "\n")
	}
	return result.String()
}

func buildAgentScratchpad(immediateSteps []Step) string {
	var result strings.Builder
	for _, step := range immediateSteps {
		result.WriteString("Thought: " + step.Thought + "\n")
		appendFieldIfNotEmpty(&result, "Action", step.Action)
		appendFieldIfNotEmpty(&result, "Action Input", step.ActionInput)
		appendFieldIfNotEmpty(&result, "Observation", step.Observation)
		result.WriteString("\n")
	}
	return result.String()
}

func appendFieldIfNotEmpty(builder *strings.Builder, fieldName, fieldValue string) {
	if fieldValue != "" {
		builder.WriteString(fieldName + ": " + fieldValue + "\n")
	}
}

func renderPromptTemplate(promptTemplate string, placeholders map[string]any) (string, error) {
	tmpl, err := template.New("react_prompt_template").Parse(promptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, placeholders); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func parseOutput(output string) (*Step, error) {
	step := &Step{}

	fieldPatterns := map[string]string{
		"Thought":     `Thought:\s*(.*)`,
		"Action":      `Action:\s*(.*)`,
		"ActionInput": `Action Input:\s*(.*)`,
		"Observation": `Observation:\s*(.*)`,
		"FinalAnswer": `Final Answer:\s*(.*)`,
	}

	for fieldName, pattern := range fieldPatterns {
		value := extractField(output, pattern)
		switch fieldName {
		case "Thought":
			step.Thought = value
		case "Action":
			step.Action = value
		case "ActionInput":
			step.ActionInput = value
		case "Observation":
			step.Observation = value
		case "FinalAnswer":
			step.FinalAnswer = value
		}
	}

	return step, nil
}

func extractField(output string, pattern string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(output)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func callTool(ctx context.Context, toolMap map[string]toolkit.Tool, step *Step) (string, error) {
	tool, exists := toolMap[strings.ToUpper(step.Action)]
	if !exists {
		return "", fmt.Errorf("%w: %s", ErrToolNotFound, step.Action)
	}

	result, err := tool.Execute(ctx, step.ActionInput)
	if err != nil {
		return "", err
	}

	return result, nil
}
