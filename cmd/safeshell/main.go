package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Adhithya-J/safeshell/internal/ai"
	"github.com/Adhithya-J/safeshell/internal/container"
	"github.com/Adhithya-J/safeshell/internal/models"
)

func main() {
	cfg := models.Config{
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL: os.Getenv("OPENAI_BASE_URL"),
		Model:         os.Getenv("OPENAI_MODEL"),
		DockerImage:   "alpine:latest",
	}

	if cfg.OpenAIBaseURL == "" {
		cfg.OpenAIBaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o"
	}

	if cfg.OpenAIAPIKey == "" {
		fmt.Println("Warning: OPENAI_API_KEY is not set. Running in MOCK mode.")
		cfg.UseMock = true
	}

	aiClient := ai.NewClient(cfg)
	runner, err := container.NewRunner(cfg.DockerImage)
	if err != nil {
		fmt.Printf("Error initializing Docker runner: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Safeshell AI Agent Initialized.")
	fmt.Println("Type your task in natural language (or 'exit' to quit):")

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		if strings.ToLower(input) == "exit" {
			break
		}

		if input == "" {
			continue
		}

		fmt.Println("Thinking...")
		resp, err := aiClient.GetBashScript(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\nAI Explanation: %s\n", resp.Explanation)

		if !resp.IsSafe {
			fmt.Println("AI flagged this request as UNSAFE. Refusing to generate script.")
			continue
		}

		fmt.Println("\n--- PROPOSED SCRIPT ---")
		fmt.Println(resp.Script)
		fmt.Println("-----------------------")

		fmt.Print("\nDo you want to execute this script in Docker? (y/N): ")
		if !scanner.Scan() {
			break
		}
		confirm := strings.ToLower(scanner.Text())

		if confirm == "y" || confirm == "yes" {
			fmt.Println("Executing...")
			err := runner.RunScript(context.Background(), resp.Script)
			if err != nil {
				fmt.Printf("Execution Error: %v\n", err)
			} else {
				fmt.Println("Execution Complete.")
			}
		} else {
			fmt.Println("Execution cancelled.")
		}
	}
}
