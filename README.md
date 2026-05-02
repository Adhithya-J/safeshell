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
- **Container Runner (`internal/container`)**: Orchestrates Docker lifecycle (pull, create, start, logs, cleanup).
- **CLI Interface (`cmd/safeshell`)**: Provides a user-friendly terminal loop for interaction.

## 📋 Prerequisites

- **Go**: 1.22 or higher.
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

SafeShell uses environment variables for sensitive configuration:

| Variable          | Description         | Default                     |
| ----------------- | ------------------- | --------------------------- |
| `OPENAI_API_KEY`  | Your OpenAI API Key | (Required for AI)           |
| `OPENAI_BASE_URL` | API base URL        | `https://api.openai.com/v1` |
| `OPENAI_MODEL`    | AI model to use     | `gpt-4o`                    |

## 💻 Usage

### Running the Agent

You can start the agent directly using `go run`:

```bash
export OPENAI_API_KEY="your-key-here"
go run cmd/safeshell/main.go
```

If no `OPENAI_API_KEY` is provided, SafeShell will automatically start in **Mock Mode**.

### Example Interaction

```text
Safeshell AI Agent Initialized.
Type your task in natural language (or 'exit' to quit):

> create a file named hello.txt with 'Hello World' inside and then list the directory

Thinking...

AI Explanation: I will create a file named hello.txt containing the text 'Hello World' using the echo command, and then use the ls -l command to list the contents of the current directory to verify its creation.

--- PROPOSED SCRIPT ---
echo 'Hello World' > hello.txt
ls -l
-----------------------

Do you want to execute this script in Docker? (y/N): y
Executing...
total 4
-rw-r--r--    1 root     root            12 May  2 12:00 hello.txt
Execution Complete.
```

## 🛡️ Safety & Sandboxing

- **Isolation**: Every script execution happens in a fresh Docker container. Files created or modified do not persist on your host.
- **AI Guardrails**: The system prompt instructs the AI to refuse dangerously malicious requests (e.g., attempts to escape the container or network attacks).
- **User Confirmation**: No script is executed without your explicit 'y' confirmation.

## 🛠️ Development

### Project Structure

- `cmd/`: Application entry points.
- `configs/`: Default configuration files.
- `internal/`: Private library code.
  - `ai/`: OpenAI client implementation with mockable HTTP client.
  - `container/`: Docker runner logic using the Docker API interface.
  - `models/`: Shared data structures.
  - `validator/`: Basic shell script validation logic.

### Building

```bash
go build -o safeshell ./cmd/safeshell
```

### Running Tests

SafeShell includes a suite of unit tests that use manual mocks to verify the AI client and Docker runner without requiring external API keys or a running Docker daemon.

Run all tests:

```bash
go test ./internal/...
```

To see verbose output:

```bash
go test -v ./internal/...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
