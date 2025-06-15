# TUI-GPT - Command Line Chat Assistant

A modern, terminal-based chat interface built with Go that provides a beautiful and interactive experience for chatting with AI.

## Features

- 🎨 Beautiful colored interface with a clean design
- 🤖 Real-time chat with AI (using Groq API)
- 📚 Persistent chat history with automatic saving
- 🔄 Auto-generated chat titles based on content
- 📝 Quick chat management (new, save, load, delete)
- 📱 Responsive design that works in any terminal
- 📝 Comprehensive help system

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

4. Copy the example environment file and add your Groq API key:
```bash
cp .env.example .env
# Edit .env and add your Groq API key
```

## Usage

Run the application:
```bash
go run cmd/main.go
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
├── cmd/              # Main application entry point
├── internal/         # Internal packages
│   ├── groq/        # Groq API integration
│   └── storage/     # Chat storage system
├── ui/              # User interface components
└── chat_history/    # Stored chat sessions
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
