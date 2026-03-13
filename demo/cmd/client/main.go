package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/owulveryck/ucp-merchant-test/demo/internal/a2aclient"
	"github.com/owulveryck/ucp-merchant-test/demo/internal/client"
)

func main() {
	project := flag.String("project", os.Getenv("GOOGLE_CLOUD_PROJECT"), "GCP project for Vertex AI")
	location := flag.String("location", envOrDefault("GOOGLE_CLOUD_LOCATION", "us-central1"), "Vertex AI location")
	model := flag.String("model", "gemini-2.5-flash", "Gemini model name")
	graphURL := flag.String("graph-url", "http://localhost:9000", "Shopping Graph URL")
	obsURL := flag.String("obs-url", "", "Observability Hub URL")
	instruction := flag.String("instruction", "", "one-shot instruction (interactive if empty)")
	mcpPort := flag.Int("mcp-port", 0, "MCP server port (0 = disabled)")
	flag.Parse()

	if *project == "" {
		log.Fatal("--project or GOOGLE_CLOUD_PROJECT is required")
	}

	ctx := context.Background()

	// Start SSE command listener EARLY (before slow genai init)
	// so the obs-hub sees the agent as "online" immediately.
	commandCh := make(chan string, 8)
	if *obsURL != "" {
		go listenCommandsSSE(ctx, *obsURL, commandCh)
	}

	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  *project,
		Location: *location,
	})
	if err != nil {
		log.Fatalf("create genai client: %v", err)
	}

	a2aClient := a2aclient.NewClient("john.doe@example.com", "US", *obsURL)
	agent := client.NewAgent(genaiClient, *model, a2aClient, *graphURL, *obsURL)

	// Process commands received (possibly queued) from the SSE listener.
	if *obsURL != "" {
		go func() {
			for instr := range commandCh {
				result, err := agent.Run(ctx, instr)
				if err != nil {
					log.Printf("command run error: %v", err)
					continue
				}
				log.Printf("command result: %s", result)
			}
		}()
	}

	if *mcpPort > 0 {
		mcpSrv := client.NewMCPServer(agent)
		mux := http.NewServeMux()
		mux.Handle("/mcp", mcpSrv)
		go func() {
			addr := fmt.Sprintf(":%d", *mcpPort)
			log.Printf("MCP endpoint: http://localhost%s/mcp", addr)
			log.Fatal(http.ListenAndServe(addr, mux))
		}()
	}

	if *instruction != "" {
		result, err := agent.Run(ctx, *instruction)
		if err != nil {
			log.Fatalf("agent error: %v", err)
		}
		fmt.Println(result)
		return
	}

	// If obsURL is set and stdin is not a terminal, run as daemon
	if *obsURL != "" {
		fi, err := os.Stdin.Stat()
		if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
			log.Println("Daemon mode: listening for commands via SSE")
			select {} // block forever, ListenCommands goroutine handles commands
		}
	}

	// Interactive REPL
	fmt.Println("Shopping Assistant (type 'quit' to exit)")
	fmt.Println("=========================================")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			break
		}

		result, err := agent.Run(ctx, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Println(result)
	}

	// If stdin closed (EOF) but obs-url is set, stay alive in daemon mode
	// so the SSE command listener keeps running.
	if *obsURL != "" {
		log.Println("Stdin closed, continuing in daemon mode")
		select {}
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// listenCommandsSSE connects to the obs-hub /commands SSE endpoint and
// forwards received instructions to ch. It reconnects on error and blocks
// until ctx is cancelled.
func listenCommandsSSE(ctx context.Context, obsURL string, ch chan<- string) {
	for {
		if err := listenCommandsSSEOnce(ctx, obsURL, ch); err != nil {
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

func listenCommandsSSEOnce(ctx context.Context, obsURL string, ch chan<- string) error {
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
		ch <- cmd.Instruction
	}
	return scanner.Err()
}
