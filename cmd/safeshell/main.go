package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Adhithya-J/safeshell/internal/app"
	"github.com/Adhithya-J/safeshell/internal/models"
	"github.com/joho/godotenv"
)

const Version string = "0.1.0"

func main() {

	// LLM config
	modelPtr := flag.String("model", os.Getenv("OPENAI_MODEL"), "Model to use for generating the script")

	// mocking behaviour
	mockPtr := flag.Bool("mock", false, "Control mocking behaviour")

	// Docker config
	dockerimgPtr := flag.String("docker-img", "alpine:latest", "The docker image to execute the scripts in")

	// Shell config
	versionPtr := flag.Bool("version", false, "Displays program version")
	dryRunPtr := flag.Bool("dry-run", false, "Prints command without executing it")
	timeoutPtr := flag.Int("timeout", 10, "Timeout in seconds")
	promptPtr := flag.String("prompt", "", "Allows user to pass prompt to run the program in non-interactive mode")
	// outputPtr := flag.String("output", "", "Allows user to store the generated script to the specified file")
	envFilePtr := flag.String("env-file", "", "Allows user to specify the path to the .env file")
	// readOnlyPtr := flag.Bool("read-only", false, "Allow only read only commands")
	// verbosePtr := flag.Bool("verbose", false, "Control verbosity of the output")

	// This would be an improvement to the application later on to allow running other shell commands
	// shellPtr := flag.String("shell", "bash", "Shell to genrate the command in")

	flag.Parse()

	if *versionPtr {
		fmt.Printf("%s", "safeshell v"+Version)
		return
	}

	if *envFilePtr != "" {
		err := godotenv.Load(*envFilePtr)
		if err != nil {
			fmt.Printf("Error loading environment file at : %s", *envFilePtr)
		} else {
			godotenv.Load()
		}
	}

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
	if *mockPtr {
		fmt.Println("Running in MOCK mode.")
		cfg.UseMock = true
	} else if cfg.OpenAIAPIKey == "" {
		fmt.Errorf("Error: OPENAI_API_KEY is not set.")
		os.Exit(1)
	}

	if err := app.Initialize(cfg); err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Safeshell AI Agent Initialized.")

	if *promptPtr != "" {
		app.HandleInput(*promptPtr, scanner, *dryRunPtr, *timeoutPtr)
		return
	}

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

		app.HandleInput(input, scanner, *dryRunPtr, *timeoutPtr)

	}
}
