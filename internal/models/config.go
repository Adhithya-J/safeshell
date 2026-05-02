package models

// Message represents a single turn in the AI conversation history.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AgentResponse defines the structured output expected from the AI.
type AgentResponse struct {
	Script      string `json:"script"`
	Explanation string `json:"explanation"`
	IsSafe      bool   `json:"is_safe"`
}

// Config holds the application configuration.
type Config struct {
	OpenAIAPIKey  string
	OpenAIBaseURL string
	Model         string
	DockerImage   string
	UseMock       bool
}
