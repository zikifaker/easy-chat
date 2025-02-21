package agents

import (
	"bytes"
	"context"
	"easy-chat/agents/llms"
	"easy-chat/agents/prompts"
	"easy-chat/agents/toolkit"
	"easy-chat/consts"
	"easy-chat/dao"
	"easy-chat/request"
	"errors"
	"fmt"
	"github.com/dlclark/regexp2"
	"log"
	"strings"
	"text/template"
	"time"
)

var (
	ErrToolNotFound        = errors.New("tool not found")
	ErrWhilePlanningStep   = errors.New("error while planning step")
	ErrWhileCallingTool    = errors.New("error while calling tool")
	ErrWhileCompilingRegex = errors.New("error while compiling regex")
	ErrWhileMatchingRegex  = errors.New("error while matching regex")
)

type Agent struct {
	LLM     llms.LLM
	Tools   []toolkit.Tool
	MaxStep int
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
		MaxStep: opts.MaxStep,
	}, nil
}

func (a *Agent) Execute(ctx context.Context, request *request.ChatRequest) (string, error) {
	var finalAnswer string
	var immediateSteps []Step
	toolMap := a.buildToolMap()

	for i := 0; i < a.MaxStep; i++ {
		step, err := a.plan(ctx, request, immediateSteps)
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
				log.Printf("%v %s: %v", ErrWhileCallingTool, step.Action, err)
				step.Observation = err.Error()
			} else {
				log.Println("observation:", observation)
				step.Observation = observation
			}
		}

		immediateSteps = append(immediateSteps, *step)
	}

	return finalAnswer, nil
}

func (a *Agent) buildToolMap() map[string]toolkit.Tool {
	toolMap := make(map[string]toolkit.Tool)
	for _, tool := range a.Tools {
		toolMap[strings.ToUpper(tool.Name())] = tool
	}
	return toolMap
}

func (a *Agent) plan(ctx context.Context, request *request.ChatRequest, immediateSteps []Step) (*Step, error) {
	currentTime := buildCurrentTime()
	toolDetail := a.getToolDetail()
	toolNames := a.getToolNames()
	chatHistory := a.getChatHistory(request)
	agentScratchpad := buildAgentScratchpad(immediateSteps)

	prompt, err := renderPromptTemplate(prompts.ReActPromptTemplate, map[string]interface{}{
		"max_step":         a.MaxStep,
		"current_step":     len(immediateSteps) + 1,
		"current_time":     currentTime,
		"tool_detail":      toolDetail,
		"tool_names":       toolNames,
		"chat_history":     chatHistory,
		"agent_scratchpad": agentScratchpad,
		"question":         request.Query,
	})
	if err != nil {
		return nil, err
	}

	streamFunc, exists := ctx.Value(consts.KeyStreamFunc).(llms.StreamFunc)
	if !exists {
		return nil, fmt.Errorf("%w: %s", consts.ErrInvalidContextKey, consts.KeyStreamFunc)
	}

	result, err := a.LLM.GenerateContent(ctx, prompt, llms.WithStreamFunc(streamFunc))
	if err != nil {
		return nil, err
	}

	// force to output a new line for a new step
	if err := streamFunc(ctx, []byte("\n\n")); err != nil {
		return nil, err
	}

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

func (a *Agent) getChatHistory(request *request.ChatRequest) string {
	var result strings.Builder

	chatHistories, err := dao.GetChatHistoryBySessionID(request.SessionID)
	if err != nil {
		return ""
	}

	for _, chatHistory := range chatHistories {
		result.WriteString(chatHistory.MessageType + ": " + chatHistory.Content + "\n")
	}

	return result.String()
}

func buildCurrentTime() string {
	currentTime := time.Now()
	return currentTime.Format(time.RFC3339)
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

func renderPromptTemplate(promptTemplate string, placeholders map[string]interface{}) (string, error) {
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
		"Thought":     `(?s)(?<=\*\*Thought\*\*:)\s*(.*?)(?=\*\*Action\*\*:\s|\Z)`,
		"Action":      `(?s)(?<=\*\*Action\*\*:)\s*(.*?)(?=\*\*Action Input\*\*:\s|\Z)`,
		"ActionInput": `(?s)(?<=\*\*Action Input\*\*:)\s*(.*?)(?=\*\*Observation\*\*:\s|\Z)`,
		"Observation": `(?s)(?<=\*\*Observation\*\*:)\s*(.*?)(?=\*\*Final Answer\*\*:\s|\Z)`,
		"FinalAnswer": `(?s)(?<=\*\*Final Answer\*\*:)\s*(.*)`,
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
	re, err := regexp2.Compile(pattern, regexp2.None)
	if err != nil {
		log.Printf("%v: %v", ErrWhileCompilingRegex, err)
		return ""
	}

	match, err := re.FindStringMatch(output)
	if err != nil {
		log.Printf("%v: %v", ErrWhileMatchingRegex, err)
		return ""
	}

	if match == nil {
		return ""
	}

	if len(match.Groups()) > 0 {
		return strings.TrimSpace(match.Groups()[0].String())
	}
	return ""
}

func callTool(ctx context.Context, toolMap map[string]toolkit.Tool, step *Step) (string, error) {
	step.Action = strings.TrimSpace(step.Action)
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
