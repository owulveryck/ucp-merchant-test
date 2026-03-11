package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/genai"

	"github.com/owulveryck/ucp-merchant-test/demo/internal/a2aclient"
	"github.com/owulveryck/ucp-merchant-test/demo/internal/client"
)

func main() {
	project := flag.String("project", os.Getenv("GOOGLE_CLOUD_PROJECT"), "GCP project for Vertex AI")
	location := flag.String("location", envOrDefault("GOOGLE_CLOUD_LOCATION", "us-central1"), "Vertex AI location")
	model := flag.String("model", "gemini-2.0-flash", "Gemini model name")
	graphURL := flag.String("graph-url", "http://localhost:9000", "Shopping Graph URL")
	obsURL := flag.String("obs-url", "", "Observability Hub URL")
	instruction := flag.String("instruction", "", "one-shot instruction (interactive if empty)")
	flag.Parse()

	if *project == "" {
		log.Fatal("--project or GOOGLE_CLOUD_PROJECT is required")
	}

	ctx := context.Background()

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

	if *instruction != "" {
		result, err := agent.Run(ctx, *instruction)
		if err != nil {
			log.Fatalf("agent error: %v", err)
		}
		fmt.Println(result)
		return
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
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
