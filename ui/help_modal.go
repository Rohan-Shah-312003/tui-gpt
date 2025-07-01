package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type HelpModal struct {
	app *App
}

func NewHelpModal(app *App) *HelpModal {
	return &HelpModal{app: app}
}

func (hm *HelpModal) Create() *tview.Modal {
	helpText := `🚀 TUI-GPT Help

📋 Key Bindings:
• Enter        - Send message
• Ctrl+C       - Quit application (auto-saves)
• Ctrl+H       - Show/hide this help
• Ctrl+L       - Clear conversation
• Ctrl+N       - Start new chat
• Ctrl+S       - Save current chat
• Ctrl+O       - Open chat history
• Ctrl+-       - Switch AI models
• Tab          - Navigate between elements
• Shift+Tab    - Navigate backwards
• Ctrl+U       - Clear input field

🤖 AI Models:
• Switch between different Groq models
• Default: Llama 3 70B for best performance
• Use Ctrl+- to change models anytime

💾 Chat Storage:
• Chats are automatically saved locally
• Access previous chats with Ctrl+O
• Each chat gets a title from first message

💡 Tips:
• Type your message and press Enter
• Use "New" button to start fresh chat
• All chats are saved in 'chat_history' folder`

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			hm.app.pages.HidePage("help")
		}).
		SetTextColor(tcell.ColorBlack)

	modal.SetBorderColor(tcell.ColorYellow)
	modal.SetTitle(" Help & Instructions ")
	modal.SetTitleColor(tcell.ColorBlack)

	return modal
}
