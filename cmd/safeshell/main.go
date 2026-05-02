package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Adhithya-J/safeshell/internal/ai"
	"github.com/Adhithya-J/safeshell/internal/container"
	"github.com/Adhithya-J/safeshell/internal/models"
	"github.com/Adhithya-J/safeshell/internal/validator"
)

func main() {

	// LLM config
	modelPtr := flag.String("model", os.Getenv("OPENAI_MODEL"), "Model to use for generating the script")

	// mocking behaviour
	mockPtr := flag.Bool("mock", false, "Control mocking behaviour")

	// Docker config
	dockerimgPtr := flag.String("docker-img", "alpine:latest", "The docker image to execute the scripts in")

	// Shell config
	dryRunPtr := flag.Bool("dry-run", false, "Prints command without executing it")
	timeoutPtr := flag.Int("timeout", 10, "Timeout in seconds")
	// readOnlyPtr := flag.Bool("read-only", false, "Allow only read only commands")
	// verbosePtr := flag.Bool("verbose", false, "Control verbosity of the output")

	// This would be an improvement to the application later on to allow running other shell commands
	// shellPtr := flag.String("shell", "bash", "Shell to genrate the command in")

	flag.Parse()

	cfg := models.Config{
		OpenAIAPIKey:  os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL: os.Getenv("OPENAI_BASE_URL"),
		Model:         *modelPtr,     // os.Getenv("OPENAI_MODEL"),
		DockerImage:   *dockerimgPtr, // "alpine:latest",
	}

	if cfg.OpenAIBaseURL == "" {
		cfg.OpenAIBaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o"
	}

	if *mockPtr || cfg.OpenAIAPIKey == "" {
		if cfg.OpenAIAPIKey == "" {
			fmt.Println("Warning: OPENAI_API_KEY is not set.")
		}

		fmt.Println("Running in MOCK mode.")
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

		// this validation depends entirely on LLM deciding what is safe
		if !resp.IsSafe {
			fmt.Println("AI flagged this request as UNSAFE. Refusing to generate script.")
			continue
		}

		// rule based validation as fallback
		if err := validator.Validate(resp.Script); err != nil {
			fmt.Printf("Validation Error: %v\n", err)
			continue
		}

		fmt.Println("\n--- PROPOSED SCRIPT ---")
		fmt.Println(resp.Script)
		fmt.Println("-----------------------")

		if *dryRunPtr {
			fmt.Println("Skipping execution")
			continue
		}

		fmt.Print("\nDo you want to execute this script in Docker? (y/N): ")
		if !scanner.Scan() {
			break
		}
		confirm := strings.ToLower(scanner.Text())

		if confirm == "y" || confirm == "yes" {
			fmt.Println("Executing...")

			ctx := context.Background()
			if *timeoutPtr > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), time.Duration(*timeoutPtr*int(time.Second)))
				defer cancel()
			}

			err := runner.RunScript(ctx, resp.Script)
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
