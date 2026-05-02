package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/Adhithya-J/safeshell/internal/models"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	config  models.Config
	history []models.Message
	httpClient HTTPClient
}

func NewClient(cfg models.Config) *Client {
	return &Client{
		config: cfg,
		history: []models.Message{
			{
				Role:    "system",
				Content: "You are a senior DevOps engineer and security expert. Your task is to generate bash scripts for the user's natural language requests. You must prioritize safety. If a request is impossible or dangerously malicious (e.g., trying to escape the container), set is_safe to false and provide an explanation. You always return valid JSON matching the provided schema.",
			},
		},
		httpClient: http.DefaultClient,
	}
}

func (c *Client) SetHTTPClient(client HTTPClient) {
	c.httpClient = client
}

func (c *Client) GetBashScript(userInput string) (*models.AgentResponse, error) {
	if c.config.UseMock {
		return &models.AgentResponse{
			Script:      "echo 'Hello from Safeshell Mock! You asked for: " + userInput + "'",
			Explanation: "This is a mock response because the application is running in mock mode.",
			IsSafe:      true,
		}, nil
	}
	c.history = append(c.history, models.Message{Role: "user", Content: userInput})

	// Truncate history to last 20 messages (excluding system prompt at index 0)
	if len(c.history) > 21 {
		c.history = append([]models.Message{c.history[0]}, c.history[len(c.history)-20:]...)
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":    c.config.Model,
		"messages": c.history,
		"response_format": map[string]interface{}{
			"type": "json_schema",
			"json_schema": map[string]interface{}{
				"name":   "agent_response",
				"strict": true,
				"schema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"script":      map[string]string{"type": "string"},
						"explanation": map[string]string{"type": "string"},
						"is_safe":     map[string]string{"type": "boolean"},
					},
					"required":             []string{"script", "explanation", "is_safe"},
					"additionalProperties": false,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.OpenAIBaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.OpenAIAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	var agentResponse models.AgentResponse
	if err := json.Unmarshal([]byte(apiResponse.Choices[0].Message.Content), &agentResponse); err != nil {
		return nil, err
	}

	c.history = append(c.history, models.Message{Role: "assistant", Content: apiResponse.Choices[0].Message.Content})

	return &agentResponse, nil
}
