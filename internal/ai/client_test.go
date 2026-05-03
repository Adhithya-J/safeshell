package ai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/Adhithya-J/safeshell/internal/models"
)

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestGetBashScript_Success(t *testing.T) {
	cfg := models.Config{
		OpenAIAPIKey:  "test-key",
		OpenAIBaseURL: "https://api.openai.com/v1",
		Model:         "gpt-4o",
	}
	client := NewClient(cfg)

	expectedResponse := models.AgentResponse{
		Script:      "echo 'hello'",
		Explanation: "Just a greeting",
		IsSafe:      true,
	}

	responseBody, _ := json.Marshal(map[string]interface{}{
		"choices": []map[string]interface{}{
			{
				"message": map[string]interface{}{
					"content": func() string {
						b, _ := json.Marshal(expectedResponse)
						return string(b)
					}(),
				},
			},
		},
	})

	client.SetHTTPClient(&mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
			}, nil
		},
	})

	resp, err := client.GetBashScript("say hello")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Script != expectedResponse.Script {
		t.Errorf("Expected script %s, got %s", expectedResponse.Script, resp.Script)
	}
	if resp.IsSafe != expectedResponse.IsSafe {
		t.Errorf("Expected IsSafe %v, got %v", expectedResponse.IsSafe, resp.IsSafe)
	}
}

func TestGetBashScript_HistoryTruncation(t *testing.T) {
	cfg := models.Config{
		OpenAIAPIKey: "test-key",
	}
	client := NewClient(cfg)

	// Fill history with more than 20 messages
	for i := 0; i < 25; i++ {
		client.history = append(client.history, models.Message{Role: "user", Content: "test"})
		client.history = append(client.history, models.Message{Role: "assistant", Content: "{}"})
	}

	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			var body map[string]interface{}
			json.NewDecoder(req.Body).Decode(&body)
			messages := body["messages"].([]interface{})
			
			// System prompt + 20 messages = 21
			if len(messages) != 21 {
				t.Errorf("Expected 21 messages in history, got %d", len(messages))
			}
			
			// Verify system prompt is still there
			if messages[0].(map[string]interface{})["role"] != "system" {
				t.Errorf("Expected first message to be system, got %v", messages[0])
			}

			respContent, _ := json.Marshal(models.AgentResponse{IsSafe: true})
			respBody, _ := json.Marshal(map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]interface{}{
							"content": string(respContent),
						},
					},
				},
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(respBody)),
			}, nil
		},
	}
	client.SetHTTPClient(mockClient)

	_, err := client.GetBashScript("trigger truncation")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestGetBashScript_UseMock(t *testing.T) {
	cfg := models.Config{
		UseMock: true,
	}
	client := NewClient(cfg)

	resp, err := client.GetBashScript("test command")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Script == "" {
		t.Error("Expected a script in mock mode, got empty string")
	}
	if !resp.IsSafe {
		t.Error("Expected mock response to be safe")
	}
}
