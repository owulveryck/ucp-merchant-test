package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/owulveryck/ucp-merchant-test/demo/internal/a2aclient"
)

const systemPrompt = `You are a shopping assistant that finds the best deals across multiple merchants.

WORKFLOW:
1. Use search_products to find matching products across all merchants
2. Use get_product_details for the top 3 results (from different merchants) to verify stock
3. Use create_checkout at all qualifying merchants (up to 3) to start checkout sessions
4. Use apply_discount_codes with any discount hints from search results (try all available codes)
5. Use update_checkout to set buyer info (email: john.doe@example.com, first_name: John, last_name: Doe) AND fulfillment_type "shipping"
6. Use get_checkout_summary to read available destinations from checkout fulfillment
7. Use update_checkout with selected_destination_id (pick the first destination from fulfillment.methods[0].destinations[0].id)
8. Use get_checkout_summary to read available shipping options from fulfillment.methods[0].groups[0].options
9. Use update_checkout with selected_option_id (pick the cheapest shipping option id)
10. Use get_checkout_summary to compare final totals from all merchants
11. Use complete_checkout at the cheapest merchant (handler_id: "mock_payment_handler", token: "success_token")
12. Use cancel_checkout at the other merchants

IMPORTANT: Fulfillment is progressive. You MUST do steps 5-9 for EACH merchant checkout before comparing prices.
Each update_checkout call for fulfillment builds on the previous state. Do not skip steps.

Always show a clear price comparison before purchasing. Format prices as dollars (divide cents by 100).
When applying discount codes, try all hints provided in the search results.
Explain which merchant won and why (lowest total after discounts and shipping).`

// Agent is the Gemini-powered shopping assistant.
type Agent struct {
	genaiClient *genai.Client
	model       string
	a2aClient   *a2aclient.Client
	graphURL    string
	obsURL      string
}

// NewAgent creates a new client agent.
func NewAgent(genaiClient *genai.Client, model string, a2aClient *a2aclient.Client, graphURL, obsURL string) *Agent {
	return &Agent{
		genaiClient: genaiClient,
		model:       model,
		a2aClient:   a2aClient,
		graphURL:    graphURL,
		obsURL:      obsURL,
	}
}

// Run executes the agent with the given instruction.
func (a *Agent) Run(ctx context.Context, instruction string) (string, error) {
	a.emitEvent("agent_start", fmt.Sprintf("Instruction: %s", instruction))

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		Tools:             defineTools(),
	}

	contents := []*genai.Content{
		genai.NewContentFromText(instruction, genai.RoleUser),
	}

	for i := 0; i < 20; i++ {
		a.emitEvent("agent_step", fmt.Sprintf("Step %d", i+1))

		resp, err := a.genaiClient.Models.GenerateContent(ctx, a.model, contents, config)
		if err != nil {
			return "", fmt.Errorf("generate content: %w", err)
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			a.emitEvent("agent_error", "Empty response from model")
			return "", fmt.Errorf("empty response from model")
		}

		content := resp.Candidates[0].Content
		contents = append(contents, content)

		// Check for function calls
		var functionCalls []*genai.FunctionCall
		var textParts []string

		for _, part := range content.Parts {
			if part.FunctionCall != nil {
				functionCalls = append(functionCalls, part.FunctionCall)
			}
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			}
		}

		if len(textParts) > 0 {
			a.emitEvent("agent_thinking", strings.Join(textParts, "\n"))
		}

		if len(functionCalls) == 0 {
			result := ""
			for _, t := range textParts {
				result += t
			}
			a.emitEvent("agent_done", result)
			return result, nil
		}

		// Execute function calls and build response parts
		responseParts := make([]*genai.Part, 0, len(functionCalls))
		for _, fc := range functionCalls {
			log.Printf("Tool call: %s(%v)", fc.Name, fc.Args)
			result, err := a.executeTool(fc.Name, fc.Args)
			if err != nil {
				log.Printf("Tool error: %s: %v", fc.Name, err)
				a.emitEvent("tool_error", fmt.Sprintf("Tool %s failed: %v", fc.Name, err))
				responseParts = append(responseParts, &genai.Part{
					FunctionResponse: &genai.FunctionResponse{
						ID:       fc.ID,
						Name:     fc.Name,
						Response: map[string]any{"error": err.Error()},
					},
				})
			} else {
				var parsed any
				if json.Unmarshal([]byte(result), &parsed) == nil {
					responseParts = append(responseParts, &genai.Part{
						FunctionResponse: &genai.FunctionResponse{
							ID:       fc.ID,
							Name:     fc.Name,
							Response: map[string]any{"result": parsed},
						},
					})
				} else {
					responseParts = append(responseParts, &genai.Part{
						FunctionResponse: &genai.FunctionResponse{
							ID:       fc.ID,
							Name:     fc.Name,
							Response: map[string]any{"result": result},
						},
					})
				}
				a.emitEvent("tool_result", fmt.Sprintf("Tool %s returned (%d bytes)", fc.Name, len(result)))
			}
		}

		contents = append(contents, &genai.Content{
			Role:  "user",
			Parts: responseParts,
		})
	}

	a.emitEvent("agent_error", "Exceeded maximum iterations")
	return "", fmt.Errorf("exceeded maximum iterations")
}

// ListenCommands connects to the obs hub SSE command stream and runs
// instructions as they arrive. It blocks until ctx is cancelled.
func (a *Agent) ListenCommands(ctx context.Context, obsURL string) {
	for {
		if err := a.listenCommandsOnce(ctx, obsURL); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("commands SSE error: %v, reconnecting...", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

func (a *Agent) listenCommandsOnce(ctx context.Context, obsURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, obsURL+"/commands", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		var cmd struct {
			Instruction string `json:"instruction"`
		}
		if json.Unmarshal([]byte(payload), &cmd) != nil || cmd.Instruction == "" {
			continue
		}
		log.Printf("Received command: %s", cmd.Instruction)
		go func(instr string) {
			result, err := a.Run(ctx, instr)
			if err != nil {
				log.Printf("command run error: %v", err)
				return
			}
			log.Printf("command result: %s", result)
		}(cmd.Instruction)
	}
	return scanner.Err()
}

func (a *Agent) searchGraph(query string, limit int) (map[string]any, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"query": query,
		"limit": limit,
	})

	resp, err := http.Post(a.graphURL+"/search", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("search graph: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search: %w", err)
	}
	return result, nil
}

func (a *Agent) emitEvent(eventType, summary string) {
	if a.obsURL == "" {
		return
	}

	event := map[string]any{
		"source":  "client-agent",
		"type":    eventType,
		"summary": summary,
	}

	data, _ := json.Marshal(event)
	resp, err := http.Post(a.obsURL+"/event", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("obs event error: %v", err)
		return
	}
	resp.Body.Close()
}
