# TUI-GPT - Command Line Chat Assistant

A modern, terminal-based chat interface built with Go that provides a beautiful and interactive experience for chatting with AI.

## Features

- 🎨 Beautiful colored interface with a clean design
- 🤖 Real-time chat with AI
- 📚 Persistent chat history with automatic saving
- 🔄 Auto-generated chat titles based on content
- 📝 Quick chat management (new, save, load, delete)
- 📱 Responsive design that works in any terminal
- 📝 Comprehensive help system
- 📋 Model selection interface

## Installation

1. Ensure you have Go installed (version 1.20 or higher)

2. Clone the repository:
```bash
git clone https://github.com/Rohan-Shah-312003/tui-gpt.git
cd tui-gpt
```

3. Install dependencies:
```bash
go mod download
```

4. Copy the example environment file and configure your settings:
```bash
cp .env.example .env
# Edit .env with your configuration
```

## Usage

Run the application:
```bash
go run main.go
```

### Keyboard Shortcuts

- `Enter` - Send message
- `Ctrl+C` - Quit application (auto-saves)
- `Ctrl+H` - Show/hide help
- `Ctrl+L` - Clear conversation
- `Ctrl+N` - Start new chat
- `Ctrl+S` - Save current chat
- `Ctrl+O` - Open chat history
- `Tab` - Navigate between elements
- `Shift+Tab` - Navigate backwards
- `Ctrl+U` - Clear input field

### Chat History Management

- All chats are automatically saved in the `chat_history` folder
- Chat titles are auto-generated from the first message
- Access previous chats using `Ctrl+O`
- Delete unwanted chats from the history

## Project Structure

```
tui-gpt/
├── .env              # Environment configuration
├── go.mod           # Go module definition
├── go.sum           # Go module dependencies
├── internal/        # Internal packages
│   └── storage/     # Chat storage system
├── main.go          # Application entry point
├── ui/              # User interface components
└── chat_history/    # Stored chat sessions
```