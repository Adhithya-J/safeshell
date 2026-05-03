package app

import (
	"context"
	"errors"
	"testing"

	"github.com/Adhithya-J/safeshell/internal/models"
)

type mockScriptGenerator struct {
	getFunc func(userInput string) (*models.AgentResponse, error)
}

func (m *mockScriptGenerator) GetBashScript(userInput string) (*models.AgentResponse, error) {
	return m.getFunc(userInput)
}

type mockRunner struct {
	runFunc func(ctx context.Context, script string) (string, error)
}

func (m *mockRunner) RunScript(ctx context.Context, script string) (string, error) {
	return m.runFunc(ctx, script)
}

func TestGenerateAndValidateScript(t *testing.T) {
	t.Run("EmptyInput", func(t *testing.T) {
		_, err := GenerateAndValidateScript("")
		if err == nil {
			t.Error("Expected error for empty input, got nil")
		}
	})

	t.Run("Success", func(t *testing.T) {
		mockAI := &mockScriptGenerator{
			getFunc: func(userInput string) (*models.AgentResponse, error) {
				return &models.AgentResponse{
					Script:      "echo 'hello'",
					Explanation: "Test",
					IsSafe:      true,
				}, nil
			},
		}
		SetAIClient(mockAI)

		resp, err := GenerateAndValidateScript("say hello")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.Script != "echo 'hello'" {
			t.Errorf("Expected script 'echo hello', got %s", resp.Script)
		}
	})

	t.Run("AISaysUnsafe", func(t *testing.T) {
		mockAI := &mockScriptGenerator{
			getFunc: func(userInput string) (*models.AgentResponse, error) {
				return &models.AgentResponse{
					Script:      "rm -rf /",
					Explanation: "Malicious",
					IsSafe:      false,
				}, nil
			},
		}
		SetAIClient(mockAI)

		_, err := GenerateAndValidateScript("delete everything")
		if err == nil {
			t.Error("Expected error when AI says unsafe, got nil")
		}
	})

	t.Run("ValidatorFails", func(t *testing.T) {
		mockAI := &mockScriptGenerator{
			getFunc: func(userInput string) (*models.AgentResponse, error) {
				return &models.AgentResponse{
					Script:      "rm -rf /",
					Explanation: "I missed it",
					IsSafe:      true,
				}, nil
			},
		}
		SetAIClient(mockAI)

		_, err := GenerateAndValidateScript("delete everything")
		if err == nil {
			t.Error("Expected error when validator fails, got nil")
		}
	})
}

func TestExecuteScript(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockR := &mockRunner{
			runFunc: func(ctx context.Context, script string) (string, error) {
				return "output", nil
			},
		}
		SetRunner(mockR)

		out, err := ExecuteScript("echo hello", 0)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if out != "output" {
			t.Errorf("Expected 'output', got '%s'", out)
		}
	})

	t.Run("Error", func(t *testing.T) {
		mockR := &mockRunner{
			runFunc: func(ctx context.Context, script string) (string, error) {
				return "", errors.New("docker error")
			},
		}
		SetRunner(mockR)

		_, err := ExecuteScript("echo hello", 0)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}
