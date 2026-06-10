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
	"sync"
	"time"

	"google.golang.org/genai"

	"github.com/owulveryck/ucp-merchant-test/demo/internal/a2aclient"
)

// agentHTTPClient is used for non-SSE HTTP calls (search, obs events) with a timeout.
var agentHTTPClient = &http.Client{Timeout: 10 * time.Second}

// detectOptimizationMode infers the user's intent from their instruction.
// Returns "fastest" if they want speed, "cheapest" otherwise (default).
func detectOptimizationMode(instruction string) string {
	lower := strings.ToLower(instruction)

	// Fast/rapid delivery keywords
	fastKeywords := []string{
		"rapide", "vite", "express", "urgent", "fastest", "quickest",
		"quick", "fast", "speed", "rapid", "asap", "soon",
	}

	for _, kw := range fastKeywords {
		if strings.Contains(lower, kw) {
			return "fastest"
		}
	}

	// Default to cheapest
	return "cheapest"
}

func buildSystemPrompt(merchantCount int, optimizeFor string) string {
	shippingSelection := "pick the cheapest shipping option id"
	merchantSelection := "at the cheapest merchant"
	explanation := "lowest total after discounts and shipping"

	if optimizeFor == "fastest" {
		shippingSelection = "pick the FASTEST shipping option (express > standard, prefer shorter delivery times over price)"
		merchantSelection = "at the merchant with the FASTEST delivery (shortest estimated days)"
		explanation = "fastest delivery time"
	}

	return fmt.Sprintf(`You are a shopping assistant that finds products across multiple merchants.

OPTIMIZATION MODE: %s

WORKFLOW:
1. Use search_products to find matching products across all merchants (limit is set to %d automatically)
2. Call get_product_details for ALL %d results (from different merchants) to verify stock — issue ALL calls in a single step
3. Call create_checkout at all qualifying merchants (up to %d) to start checkout sessions — issue ALL calls in a single step
4. Call list_promotions at all merchants — issue ALL calls in a single step
5. If promotions are available, use apply_discount_codes with the discovered codes
6. Use update_checkout to set buyer info (email: john.doe@example.com, first_name: John, last_name: Doe) AND fulfillment_type "shipping"
7. Use get_checkout_summary to read available destinations from checkout fulfillment
8. Use update_checkout with selected_destination_id (pick the first destination from fulfillment.methods[0].destinations[0].id)
9. Use get_checkout_summary to read available shipping options from fulfillment.methods[0].groups[0].options
10. Use update_checkout with selected_option_id (%s)
11. Use get_checkout_summary to compare final results from all merchants (including shipping option titles for delivery estimation)
12. Use complete_checkout %s (handler_id: "mock_payment_handler", token: "success_token")
13. Use cancel_checkout at the other merchants

IMPORTANT: When multiple calls are independent (same tool on different merchants), issue ALL of them in a single step. This enables parallel execution and is critical for performance.

IMPORTANT: Fulfillment is progressive. You MUST do steps 6-10 for EACH merchant checkout before comparing.
Each update_checkout call for fulfillment builds on the previous state. Do not skip steps.

SHIPPING SPEED DETECTION:
- "Express" or "Expedited" in title = fast (1-2 days estimated)
- "Standard" or "Regular" in title = normal (3-5 days estimated)
- When optimizing for fastest delivery, ALWAYS prefer express options even if more expensive.

Always show a clear comparison before purchasing. Format prices as dollars (divide cents by 100).
When applying discount codes, try all hints provided in the search results.
Explain which merchant won and why (%s).`, optimizeFor, merchantCount, merchantCount, merchantCount, shippingSelection, merchantSelection, explanation)
}

// Agent is the Gemini-powered shopping assistant.
type Agent struct {
	genaiClient          *genai.Client
	model                string
	a2aClient            *a2aclient.Client
	graphURL             string
	obsURL               string
	maxIterations        int
	currentMerchantCount int // set per Run() call
}

// NewAgent creates a new client agent.
func NewAgent(genaiClient *genai.Client, model string, a2aClient *a2aclient.Client, graphURL, obsURL string) *Agent {
	return &Agent{
		genaiClient:   genaiClient,
		model:         model,
		a2aClient:     a2aClient,
		graphURL:      graphURL,
		obsURL:        obsURL,
		maxIterations: 50,
	}
}

// Run executes the agent with the given instruction.
func (a *Agent) Run(ctx context.Context, instruction string, merchantCount int) (string, error) {
	if merchantCount < 1 {
		merchantCount = 3
	}
	a.currentMerchantCount = merchantCount

	// Detect optimization intent from instruction
	optimizeFor := detectOptimizationMode(instruction)
	a.emitEvent("agent_start", fmt.Sprintf("Instruction: %s (merchants: %d, mode: %s)", instruction, merchantCount, optimizeFor))

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(buildSystemPrompt(merchantCount, optimizeFor), genai.RoleUser),
		Tools:             defineTools(),
	}

	contents := []*genai.Content{
		genai.NewContentFromText(instruction, genai.RoleUser),
	}

	for i := range a.maxIterations {
		a.emitEvent("agent_step", fmt.Sprintf("Step %d/%d", i+1, a.maxIterations))

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

		// Execute function calls concurrently and build response parts
		responseParts := make([]*genai.Part, len(functionCalls))
		var wg sync.WaitGroup
		for i, fc := range functionCalls {
			wg.Add(1)
			go func(idx int, fc *genai.FunctionCall) {
				defer wg.Done()
				log.Printf("Tool call: %s(%v)", fc.Name, fc.Args)
				start := time.Now()
				result, err := a.executeTool(fc.Name, fc.Args)
				durationMs := time.Since(start).Milliseconds()
				if err != nil {
					log.Printf("Tool error: %s: %v", fc.Name, err)
					a.emitEvent("tool_error", fmt.Sprintf("Tool %s failed: %v", fc.Name, err), map[string]any{
						"action":      fc.Name,
						"params":      fc.Args,
						"error":       err.Error(),
						"duration_ms": durationMs,
					})
					responseParts[idx] = &genai.Part{
						FunctionResponse: &genai.FunctionResponse{
							ID:       fc.ID,
							Name:     fc.Name,
							Response: map[string]any{"error": err.Error()},
						},
					}
				} else {
					var parsed any
					if json.Unmarshal([]byte(result), &parsed) == nil {
						responseParts[idx] = &genai.Part{
							FunctionResponse: &genai.FunctionResponse{
								ID:       fc.ID,
								Name:     fc.Name,
								Response: map[string]any{"result": parsed},
							},
						}
						a.emitEvent("tool_result", fmt.Sprintf("Tool %s returned (%d bytes)", fc.Name, len(result)), map[string]any{
							"action":      fc.Name,
							"params":      fc.Args,
							"response":    parsed,
							"duration_ms": durationMs,
						})
					} else {
						responseParts[idx] = &genai.Part{
							FunctionResponse: &genai.FunctionResponse{
								ID:       fc.ID,
								Name:     fc.Name,
								Response: map[string]any{"result": result},
							},
						}
						a.emitEvent("tool_result", fmt.Sprintf("Tool %s returned (%d bytes)", fc.Name, len(result)), map[string]any{
							"action":      fc.Name,
							"params":      fc.Args,
							"response":    result,
							"duration_ms": durationMs,
						})
					}
				}
			}(i, fc)
		}
		wg.Wait()

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
			Instruction   string `json:"instruction"`
			MerchantCount int    `json:"merchant_count"`
		}
		if json.Unmarshal([]byte(payload), &cmd) != nil || cmd.Instruction == "" {
			continue
		}
		count := cmd.MerchantCount
		if count < 1 {
			count = 3
		}
		log.Printf("Received command: %s (merchants: %d)", cmd.Instruction, count)
		go func(instr string, n int) {
			result, err := a.Run(ctx, instr, n)
			if err != nil {
				log.Printf("command run error: %v", err)
				return
			}
			log.Printf("command result: %s", result)
		}(cmd.Instruction, count)
	}
	return scanner.Err()
}

func (a *Agent) searchGraph(query string, limit int) (map[string]any, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"query": query,
		"limit": limit,
	})

	resp, err := agentHTTPClient.Post(a.graphURL+"/search", "application/json", bytes.NewReader(reqBody))
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

func (a *Agent) emitEvent(eventType, summary string, eventData ...any) {
	if a.obsURL == "" {
		return
	}

	event := map[string]any{
		"source":  "client-agent",
		"type":    eventType,
		"summary": summary,
	}
	if len(eventData) > 0 && eventData[0] != nil {
		event["data"] = eventData[0]
	}

	data, _ := json.Marshal(event)
	resp, err := agentHTTPClient.Post(a.obsURL+"/event", "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("obs event error: %v", err)
		return
	}
	resp.Body.Close()
}
