package app

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Adhithya-J/safeshell/internal/ai"
	"github.com/Adhithya-J/safeshell/internal/container"
	"github.com/Adhithya-J/safeshell/internal/models"
	"github.com/Adhithya-J/safeshell/internal/validator"
)

type ScriptGenerator interface {
	GetBashScript(userInput string) (*models.AgentResponse, error)
}

type ScriptRunner interface {
	RunScript(ctx context.Context, script string) (string, error)
}

var (
	aiClient ScriptGenerator
	runner   ScriptRunner
)

func Initialize(cfg models.Config) error {
	aiClient = ai.NewClient(cfg)
	var err error
	dockerRunner, err := container.NewRunner(cfg.DockerImage)
	if err != nil {
		return fmt.Errorf("Error initializing Docker runner: %v\n", err)
	}
	runner = dockerRunner
	return nil
}

func SetAIClient(client ScriptGenerator) {
	aiClient = client
}

func SetRunner(r ScriptRunner) {
	runner = r
}

func ExecuteScript(script string, timeout int) (string, error) {
	fmt.Println("Executing...")

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
	}

	output, err := runner.RunScript(ctx, script)
	if err != nil {
		fmt.Printf("Execution Error: %v\n", err)
		return "", err
	}
	fmt.Println("Execution Complete.")
	return output, nil
}

func GenerateAndValidateScript(input string) (*models.AgentResponse, error) {
	//  correct the code so this returns each type of error when generation or validation fails
	if input == "" {
		return nil, fmt.Errorf("Empty input")
	}
	fmt.Println("Thinking...")
	resp, err := aiClient.GetBashScript(input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil, fmt.Errorf("Unable to generate script")
	}

	fmt.Printf("\nAI Explanation: %s\n", resp.Explanation)

	// this validation depends entirely on LLM deciding what is safe
	if !resp.IsSafe {
		fmt.Println("AI flagged this request as UNSAFE. Refusing to generate script.")
		return nil, fmt.Errorf("AI flagged unsafe")
	}

	// rule based validation as fallback
	if err := validator.Validate(resp.Script); err != nil {
		fmt.Printf("Validation Error: %v\n", err)
		return nil, fmt.Errorf("Failed validation: %v", err)
	}
	fmt.Println("\n--- PROPOSED SCRIPT ---")
	fmt.Println(resp.Script)
	fmt.Println("-----------------------")

	return resp, nil
}

func HandleInput(input string, scanner *bufio.Scanner, dryRun bool, timeout int) {

	resp, err := GenerateAndValidateScript(input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if dryRun {
		fmt.Println("Skipping execution")
		return
	}

	fmt.Print("\nDo you want to execute this script in Docker? (y/N): ")
	if !scanner.Scan() {
	}
	confirm := strings.ToLower(scanner.Text())

	if confirm == "y" || confirm == "yes" {
		output, err := ExecuteScript(resp.Script, timeout)
		if err != nil {
			fmt.Printf("Execution failed: %v\n", err)
		} else {
			fmt.Println("---OUTPUT---")
			fmt.Println(output)
			fmt.Println("------------")
		}
	} else {
		fmt.Println("Execution cancelled.")
	}

}
