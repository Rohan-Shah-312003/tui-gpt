package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rohan-Shah-312003/tui-gpt/internal/groq"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app              *tview.Application
	pages            *tview.Pages
	chatHistory      []ChatMessage
	conversationView *tview.TextView
	inputField       *tview.InputField
	statusBar        *tview.TextView
	sidebar          *tview.TextView
)

type ChatMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
}

func StartApp() {
	app = tview.NewApplication()

	// Create the main layout
	setupUI()

	// Set up key bindings
	setupKeyBindings()

	// Start the application
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func setupUI() {
	// Create pages container
	pages = tview.NewPages()

	// Create main chat interface
	mainLayout := createMainLayout()

	// Create help modal
	helpModal := createHelpModal()

	// Add pages
	pages.AddPage("main", mainLayout, true, true)
	pages.AddPage("help", helpModal, true, false)
}

func createMainLayout() *tview.Flex {
	// Header
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[::bu]🤖 TUI-GPT Chat Assistant [::-]\n[dim]Press Ctrl+H for help, Ctrl+C to quit")
	header.SetBorder(true).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(" Welcome ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)

	// Sidebar with conversation stats
	sidebar = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false)
	sidebar.SetBorder(true).
		SetTitle(" Stats ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkMagenta)
	updateSidebar()

	// Main conversation view
	conversationView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	conversationView.SetBorder(true).
		SetTitle(" Conversation ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorDarkGreen)

	// Input area
	inputField = tview.NewInputField().
		SetLabel("💬 You: ").
		SetFieldWidth(0).
		SetPlaceholder("Type your message here... (Press Enter to send)")
	inputField.SetBorder(true).
		SetTitle(" Input ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorDarkBlue)

	// Button area
	buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	sendButton := tview.NewButton("📤 Send").
		SetSelectedFunc(sendMessage)
	sendButton.SetBorder(true).
		SetBorderColor(tcell.ColorGreen)

	clearButton := tview.NewButton("🗑️  Clear").
		SetSelectedFunc(clearChat)
	clearButton.SetBorder(true).
		SetBorderColor(tcell.ColorOrange)

	quitButton := tview.NewButton("❌ Quit").
		SetSelectedFunc(func() {
			app.Stop()
		})
	quitButton.SetBorder(true).
		SetBorderColor(tcell.ColorRed)

	buttonFlex.AddItem(sendButton, 0, 1, false).
		AddItem(clearButton, 0, 1, false).
		AddItem(quitButton, 0, 1, false)

	// Status bar
	statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText("[green]Ready 🟢")
	statusBar.SetBorder(true).
		SetTitle(" Status ").
		SetBorderColor(tcell.ColorDarkCyan)

	// Input section (input field + buttons)
	inputSection := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(inputField, 3, 1, true).
		AddItem(buttonFlex, 3, 1, false)

	// Main content area
	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(conversationView, 0, 4, false).
		AddItem(sidebar, 25, 1, false)

	// Complete layout
	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 4, 1, false).
		AddItem(mainContent, 0, 1, false).
		AddItem(inputSection, 6, 1, true).
		AddItem(statusBar, 3, 1, false)

	return mainLayout
}
func createHelpModal() *tview.Modal {
	helpText := `🚀 TUI-GPT Help

📋 Key Bindings:
• Enter        - Send message
• Ctrl+C       - Quit application
• Ctrl+H       - Show/hide this help
• Ctrl+L       - Clear conversation
• Tab          - Navigate between elements
• Shift+Tab    - Navigate backwards
• Ctrl+U       - Clear input field

💡 Tips:
• Type your message and press Enter
• Use clear button to start fresh
• Scroll through conversation history
• Check stats in the sidebar

🎨 Features:
• Real-time chat with AI
• Message history tracking
• Beautiful colored interface
• Responsive design`

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.HidePage("help")
		})

	// Now use modal to call Box methods without chaining
	modal.SetBorderColor(tcell.ColorYellow)
	modal.SetTitle(" Help & Instructions ")

	return modal
}

func setupKeyBindings() {
	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			sendMessage()
			return nil
		case tcell.KeyCtrlU:
			inputField.SetText("")
			return nil
		}
		return event
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlH:
			if pages.HasPage("help") {
				name, _ := pages.GetFrontPage()
				if name == "help" {
					pages.HidePage("help")
				} else {
					pages.ShowPage("help")
				}
			}
			return nil
		case tcell.KeyCtrlL:
			clearChat()
			return nil
		}
		return event
	})
}

func sendMessage() {
	prompt := strings.TrimSpace(inputField.GetText())
	if prompt == "" {
		updateStatus("[red]⚠️  Empty message!")
		return
	}

	// Add user message to history
	userMsg := ChatMessage{
		Role:      "user",
		Content:   prompt,
		Timestamp: time.Now(),
	}
	chatHistory = append(chatHistory, userMsg)

	// Clear input and update display
	inputField.SetText("")
	updateConversationView()
	updateStatus("[yellow]🤔 AI is thinking...")
	updateSidebar()

	// Send to API in goroutine
	go func() {
		reply, err := groq.SendPrompt(prompt)

		app.QueueUpdateDraw(func() {
			if err != nil {
				errorMsg := ChatMessage{
					Role:      "error",
					Content:   fmt.Sprintf("Error: %v", err),
					Timestamp: time.Now(),
				}
				chatHistory = append(chatHistory, errorMsg)
				updateStatus("[red]❌ Error occurred!")
			} else {
				aiMsg := ChatMessage{
					Role:      "assistant",
					Content:   reply,
					Timestamp: time.Now(),
				}
				chatHistory = append(chatHistory, aiMsg)
				updateStatus("[green]✅ Response received!")
			}

			updateConversationView()
			updateSidebar()

			// Reset status after 3 seconds
			go func() {
				time.Sleep(3 * time.Second)
				app.QueueUpdateDraw(func() {
					updateStatus("[green]Ready 🟢")
				})
			}()
		})
	}()
}

func clearChat() {
	chatHistory = []ChatMessage{}
	updateConversationView()
	updateSidebar()
	updateStatus("[blue]🧹 Chat cleared!")

	go func() {
		time.Sleep(2 * time.Second)
		app.QueueUpdateDraw(func() {
			updateStatus("[green]Ready 🟢")
		})
	}()
}

func updateConversationView() {
	var content strings.Builder

	if len(chatHistory) == 0 {
		content.WriteString("[dim]🌟 Welcome to TUI-GPT!\n\n")
		content.WriteString("Start a conversation by typing a message below.\n")
		content.WriteString("Ask me anything - I'm here to help! 🤖[white]\n\n")
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

		// Add separator between messages (except for the last one)
		if i < len(chatHistory)-1 {
			content.WriteString("[dim]" + strings.Repeat("─", 50) + "[white]\n\n")
		}
	}

	conversationView.SetText(content.String())
	conversationView.ScrollToEnd()
}

func updateSidebar() {
	var content strings.Builder

	content.WriteString("[yellow]📊 Statistics[white]\n\n")
	content.WriteString(fmt.Sprintf("💬 Messages: %d\n", len(chatHistory)))

	userCount := 0
	aiCount := 0
	errorCount := 0

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

	content.WriteString("\n[cyan]🕒 Session Info[white]\n\n")
	content.WriteString(fmt.Sprintf("⏰ Started: %s\n", time.Now().Format("15:04")))

	if len(chatHistory) > 0 {
		lastMsg := chatHistory[len(chatHistory)-1]
		content.WriteString(fmt.Sprintf("📝 Last: %s\n", lastMsg.Timestamp.Format("15:04:05")))
	}

	content.WriteString("\n[magenta]🎯 Quick Tips[white]\n\n")
	content.WriteString("• Press Enter to send\n")
	content.WriteString("• Ctrl+H for help\n")
	content.WriteString("• Ctrl+L to clear\n")
	content.WriteString("• Ctrl+C to quit\n")

	sidebar.SetText(content.String())
}

func updateStatus(status string) {
	statusBar.SetText(status)
}
