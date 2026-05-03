# SafeShell

SafeShell is a safety-constrained LLM agent that converts natural language requests into validated, sandboxed shell scripts. It prioritizes execution safety by running all generated scripts within an isolated Docker container, preventing any harm to your host system.

## 🚀 Features

- **Natural Language to Bash**: Describe what you want to do in plain English.
- **AI-Powered Safety Validation**: The AI agent evaluates the safety of every request before generating a script.
- **Sandboxed Execution**: All scripts run inside an ephemeral Docker container (default: `alpine:latest`).
- **Interactive Workflow**: Review the AI's explanation and the proposed script before confirming execution.
- **Mock Mode**: Test the interface without an OpenAI API key.
- **History Aware**: Maintains context across multiple turns for more complex tasks.

## 🏗️ Architecture

SafeShell is built with a modular Go architecture:

- **AI Engine (`internal/ai`)**: Handles communication with OpenAI and manages conversation history.
- **App Logic (`internal/app`)**: Orchestrates the workflow between AI generation, validation, and execution.
- **Container Runner (`internal/container`)**: Orchestrates Docker lifecycle (pull, create, start, logs, cleanup).
- **CLI Interface (`cmd/safeshell`)**: Provides a user-friendly terminal loop for interaction.

## 📋 Prerequisites

- **Go**: 1.26.2 or higher.
- **Docker**: Must be installed and running on your system.
- **OpenAI API Key**: Required for AI-powered script generation (optional for mock mode).

## 🛠️ Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/Adhithya-J/safeshell.git
   cd safeshell
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

## ⚙️ Configuration

SafeShell uses both command-line flags and environment variables for configuration:

| Flag | Variable | Description | Default |
| ---- | -------- | ----------- | ------- |
| `-model` | `OPENAI_MODEL` | AI model to use | `gpt-4o` |
| `-mock` | - | Run in mock mode | `false` |
| `-docker-img`| - | Docker image for sandbox | `alpine:latest` |
| `-timeout` | - | Execution timeout in seconds | `10` |
| `-dry-run` | - | Show script without executing | `false` |
| `-prompt` | - | Run a single prompt (non-interactive)| `""` |
| - | `OPENAI_API_KEY` | Your OpenAI API Key | (Required for AI) |
| - | `OPENAI_BASE_URL`| API base URL | `https://api.openai.com/v1` |

## 💻 Usage

### Running the Agent

You can start the agent directly using `go run`:

```bash
export OPENAI_API_KEY="your-key-here"
go run cmd/safeshell/main.go
```

Or build and run the binary:

```bash
go build -o safeshell ./cmd/safeshell
./safeshell -mock
```

### Non-Interactive Mode

Run a single command without entering the interactive loop:

```bash
./safeshell -prompt "list files in current directory" -dry-run
```

## 🛡️ Safety & Sandboxing

- **Isolation**: Every script execution happens in a fresh Docker container. Files created or modified do not persist on your host.
- **Multi-Layer Validation**: 
  - **AI Guardrails**: The system prompt instructs the AI to refuse dangerously malicious requests.
  - **Rule-Based Validation**: A secondary validator checks for high-risk commands (e.g., `rm -rf /`) as a fallback.
- **User Confirmation**: No script is executed without your explicit 'y' confirmation.

## 🛠️ Development

### Project Structure

- `cmd/`: Application entry points.
- `configs/`: Default configuration files.
- `internal/`: Private library code.
  - `ai/`: OpenAI client implementation.
  - `app/`: Core application logic and workflow orchestration.
  - `container/`: Docker runner logic using the Docker SDK.
  - `models/`: Shared data structures and configuration.
  - `validator/`: Basic shell script validation logic.

### Building

```bash
go build -o safeshell ./cmd/safeshell/main.go
```

### Running Tests

SafeShell includes a suite of unit tests with high coverage using interfaces for dependency injection.

Run all tests:

```bash
go test ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
