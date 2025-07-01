package ui

import (
	"fmt"
	"strings"

	"github.com/Rohan-Shah-312003/tui-gpt/internal/groq"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MainLayout struct {
	app              *App
	conversationView *tview.TextView
	inputField       *tview.InputField
	statusBar        *tview.TextView
	sidebar          *tview.TextView
}

func NewMainLayout(app *App) *MainLayout {
	return &MainLayout{
		app: app,
	}
}

func (ml *MainLayout) Create() *tview.Flex {
	header := ml.createHeader()
	ml.sidebar = ml.createSidebar()
	ml.conversationView = ml.createConversationView()
	ml.inputField = ml.createInputField()
	buttonFlex := ml.createButtonFlex()
	ml.statusBar = ml.createStatusBar()

	inputSection := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ml.inputField, 3, 1, true).
		AddItem(buttonFlex, 3, 1, false)

	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(ml.conversationView, 0, 4, false).
		AddItem(ml.sidebar, 25, 1, false)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 4, 1, false).
		AddItem(mainContent, 0, 1, false).
		AddItem(inputSection, 6, 1, true).
		AddItem(ml.statusBar, 3, 1, false)

	ml.updateConversationView()
	ml.updateSidebar()

	return mainLayout
}

func (ml *MainLayout) createHeader() *tview.TextView {
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[::bu]🤖 TUI-GPT Chat Assistant [::-]\n[dim]Press Ctrl+H for help, Ctrl+O for chat history, Ctrl+- for models, Ctrl+C to quit")
	header.SetBorder(true).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(" Welcome ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)
	return header
}

func (ml *MainLayout) createSidebar() *tview.TextView {
	sidebar := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	sidebar.SetBorder(true).
		SetTitle(" Stats ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkMagenta)
	return sidebar
}

func (ml *MainLayout) createConversationView() *tview.TextView {
	conversationView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetChangedFunc(func() { ml.app.app.Draw() })
	conversationView.SetBorder(true).
		SetTitle(" Conversation ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorDarkGreen)
	return conversationView
}

func (ml *MainLayout) createInputField() *tview.InputField {
	inputField := tview.NewInputField().
		SetLabel("💬 You: ").
		SetFieldWidth(0).SetFieldBackgroundColor(tcell.ColorWheat).
		SetPlaceholder("Type your message here... (Press Enter to send)").
		SetFieldTextColor(tcell.ColorBlack)
	inputField.SetBorder(true).
		SetTitle(" Input ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorDarkBlue)
	return inputField
}

func (ml *MainLayout) createButtonFlex() *tview.Flex {
	buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	sendButton := tview.NewButton("📤 Send").
		SetSelectedFunc(ml.app.sendMessage).
		SetLabelColor(tcell.ColorBlack)
	sendButton.SetBorder(true).SetBorderColor(tcell.ColorGreen)

	newChatButton := tview.NewButton("📝 New").
		SetSelectedFunc(ml.app.newChat).
		SetLabelColor(tcell.ColorBlack)
	newChatButton.SetBorder(true).SetBorderColor(tcell.ColorBlue)

	modelButton := tview.NewButton("🤖 Model").
		SetSelectedFunc(ml.app.modelListModal.Show).
		SetLabelColor(tcell.ColorBlack)
	modelButton.SetBorder(true).SetBorderColor(tcell.ColorPurple)

	clearButton := tview.NewButton("🗑️ Clear").
		SetSelectedFunc(ml.app.clearChat).
		SetLabelColor(tcell.ColorBlack)
	clearButton.SetBorder(true).SetBorderColor(tcell.ColorOrange)

	quitButton := tview.NewButton("❌ Quit").SetSelectedFunc(func() {
		ml.app.saveCurrentChat()
		ml.app.app.Stop()
	}).SetLabelColor(tcell.ColorBlack)
	quitButton.SetBorder(true).SetBorderColor(tcell.ColorRed)

	buttonFlex.AddItem(sendButton, 0, 1, false).
		AddItem(newChatButton, 0, 1, false).
		AddItem(modelButton, 0, 1, false).
		AddItem(clearButton, 0, 1, false).
		AddItem(quitButton, 0, 1, false)

	return buttonFlex
}

func (ml *MainLayout) createStatusBar() *tview.TextView {
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[green]Ready 🟢")
	statusBar.SetBorder(true).
		SetTitle(" Status ").
		SetBorderColor(tcell.ColorDarkCyan)
	return statusBar
}

func (ml *MainLayout) updateConversationView() {
	var content strings.Builder
	chatHistory := ml.app.GetChatHistory()

	if len(chatHistory) == 0 {
		content.WriteString("[dim]🌟 Welcome to TUI-GPT!\n\n")
		content.WriteString("Start a conversation by typing a message below.\n")
		content.WriteString("Ask me anything - I'm here to help! 🤖[white]\n\n")
		content.WriteString("[cyan]💾 Your chats are automatically saved!\n")
		content.WriteString("Press Ctrl+O to access your chat history.[white]\n\n")
		content.WriteString("[magenta]🤖 Press Ctrl+- to switch AI models![white]\n\n")
	}

	for i, msg := range chatHistory {
		timestamp := msg.Timestamp.Format("15:04:05")
		switch msg.Role {
		case "user":
			content.WriteString(fmt.Sprintf("[blue]📤 You [dim](%s)[white]\n", timestamp))
			content.WriteString(fmt.Sprintf("[white]%s[white]\n\n", msg.Content))
		case "assistant":
			content.WriteString(fmt.Sprintf("[green]🤖 AI Assistant [dim](%s)[white]\n", timestamp))
			content.WriteString(fmt.Sprintf("[white]%s[white]\n\n", msg.Content))
		case "error":
			content.WriteString(fmt.Sprintf("[red]❌ Error [dim](%s)[white]\n", timestamp))
			content.WriteString(fmt.Sprintf("[red]%s[white]\n\n", msg.Content))
		}
		if i < len(chatHistory)-1 {
			content.WriteString("[dim]" + strings.Repeat("─", 50) + "[white]\n\n")
		}
	}
	ml.conversationView.SetText(content.String())
	ml.conversationView.ScrollToEnd()
}

func (ml *MainLayout) updateSidebar() {
	var content strings.Builder
	chatHistory := ml.app.GetChatHistory()
	currentSession := ml.app.GetCurrentSession()

	content.WriteString("[yellow]📊 Statistics[white]\n\n")
	content.WriteString(fmt.Sprintf("💬 Messages: %d\n", len(chatHistory)))

	userCount, aiCount, errorCount := 0, 0, 0
	for _, msg := range chatHistory {
		switch msg.Role {
		case "user":
			userCount++
		case "assistant":
			aiCount++
		case "error":
			errorCount++
		}
	}

	content.WriteString(fmt.Sprintf("📤 Your messages: %d\n", userCount))
	content.WriteString(fmt.Sprintf("🤖 AI responses: %d\n", aiCount))
	if errorCount > 0 {
		content.WriteString(fmt.Sprintf("❌ Errors: %d\n", errorCount))
	}

	currentModel := groq.GetCurrentModel()
	models := groq.GetAvailableModels()
	if modelName, exists := models[currentModel]; exists {
		content.WriteString(fmt.Sprintf("\n[cyan]🤖 Model[white]\n%s\n", modelName))
	}

	content.WriteString("\n[cyan]🕒 Session Info[white]\n\n")
	if currentSession != nil {
		content.WriteString(fmt.Sprintf("📝 Created: %s\n", currentSession.CreatedAt.Format("Jan 2, 15:04")))
		if currentSession.Title != "" {
			content.WriteString(fmt.Sprintf("🏷️  Title: %s\n", currentSession.Title))
		}
	}

	if len(chatHistory) > 0 {
		lastMsg := chatHistory[len(chatHistory)-1]
		content.WriteString(fmt.Sprintf("⏰ Last: %s\n", lastMsg.Timestamp.Format("15:04:05")))
	}

	content.WriteString("\n[magenta]🎯 Quick Tips[white]\n\n")
	content.WriteString("• Enter to send\n")
	content.WriteString("• Ctrl+O for history\n")
	content.WriteString("• Ctrl+- for models\n")
	content.WriteString("• Ctrl+N for new chat\n")
	content.WriteString("• Ctrl+H for help\n")

	ml.sidebar.SetText(content.String())
}

func (ml *MainLayout) updateStatus(status string) {
	ml.statusBar.SetText(status)
}
