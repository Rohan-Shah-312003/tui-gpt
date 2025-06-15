// ui/tui.go
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rohan-Shah-312003/tui-gpt/internal/groq"
	"github.com/Rohan-Shah-312003/tui-gpt/internal/storage"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app               *tview.Application
	pages             *tview.Pages
	chatHistory       []storage.ChatMessage
	conversationView  *tview.TextView
	inputField        *tview.InputField
	statusBar         *tview.TextView
	sidebar           *tview.TextView
	chatList          *tview.List
	storageManager    *storage.Storage
	currentSession    *storage.ChatSession
	isShowingChatList bool
)

func StartApp() {
	app = tview.NewApplication()

	// Initialize storage
	storageManager = storage.NewStorage()
	if err := storageManager.Initialize(); err != nil {
		panic(fmt.Sprintf("Failed to initialize storage: %v", err))
	}

	// Create new session
	startNewChat()

	// Create the main layout
	setupUI()

	// Set up key bindings
	setupKeyBindings()

	// Start the application
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func startNewChat() {
	currentSession = &storage.ChatSession{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []storage.ChatMessage{},
	}
	chatHistory = []storage.ChatMessage{}
}

func setupUI() {
	// Create pages container
	pages = tview.NewPages()

	// Create main chat interface
	mainLayout := createMainLayout()

	// Create help modal
	helpModal := createHelpModal()

	// Create chat list modal
	chatListModal := createChatListModal()

	// Add pages
	pages.AddPage("main", mainLayout, true, true)
	pages.AddPage("help", helpModal, true, false)
	pages.AddPage("chatlist", chatListModal, true, false)
}

func createMainLayout() *tview.Flex {
	// Header
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[::bu]🤖 TUI-GPT Chat Assistant [::-]\n[dim]Press Ctrl+H for help, Ctrl+O for chat history, Ctrl+C to quit")
	header.SetBorder(true).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(" Welcome ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)

	// Sidebar with conversation stats
	sidebar = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
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
		SetPlaceholder("Type your message here... (Press Enter to send)").
		SetFieldTextColor(tcell.ColorBlack)
	inputField.SetBorder(true).
		SetTitle(" Input ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorDarkBlue)

	// Button area
	buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	sendButton := tview.NewButton("📤 Send").
		SetSelectedFunc(sendMessage).
		SetLabelColor(tcell.ColorBlack)
	sendButton.SetBorder(true).
		SetBorderColor(tcell.ColorGreen)

	newChatButton := tview.NewButton("📝 New").
		SetSelectedFunc(func() {
			saveCurrentChat()
			startNewChat()
			updateConversationView()
			updateSidebar()
			updateStatus("[blue]🆕 Started new chat!")
		}).
		SetLabelColor(tcell.ColorBlack)
	newChatButton.SetBorder(true).
		SetBorderColor(tcell.ColorBlue)

	saveChatButton := tview.NewButton("💾 Save").
		SetSelectedFunc(func() {
			saveCurrentChat()
			updateStatus("[green]💾 Chat saved!")
		}).
		SetLabelColor(tcell.ColorBlack)
	saveChatButton.SetBorder(true).
		SetBorderColor(tcell.ColorPurple)

	clearButton := tview.NewButton("🗑️ Clear").
		SetSelectedFunc(clearChat).
		SetLabelColor(tcell.ColorBlack)
	clearButton.SetBorder(true).
		SetBorderColor(tcell.ColorOrange)

	quitButton := tview.NewButton("❌ Quit").
		SetSelectedFunc(func() {
			saveCurrentChat()
			app.Stop()
		}).
		SetLabelColor(tcell.ColorBlack)
	quitButton.SetBorder(true).
		SetBorderColor(tcell.ColorRed)

	buttonFlex.AddItem(sendButton, 0, 1, false).
		AddItem(newChatButton, 0, 1, false).
		AddItem(saveChatButton, 0, 1, false).
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
• Ctrl+C       - Quit application (auto-saves)
• Ctrl+H       - Show/hide this help
• Ctrl+L       - Clear conversation
• Ctrl+N       - Start new chat
• Ctrl+S       - Save current chat
• Ctrl+O       - Open chat history
• Tab          - Navigate between elements
• Shift+Tab    - Navigate backwards
• Ctrl+U       - Clear input field

💾 Chat Storage:
• Chats are automatically saved locally
• Access previous chats with Ctrl+O
• Each chat gets a title from first message
• Delete unwanted chats from history

💡 Tips:
• Type your message and press Enter
• Use "New" button to start fresh chat
• All chats are saved in 'chat_history' folder
• Chat titles are auto-generated from content

🎨 Features:
• Real-time chat with AI
• Persistent message history
• Beautiful colored interface
• Responsive design`

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.HidePage("help")
		}).
		SetTextColor(tcell.ColorBlack)

	modal.SetBorderColor(tcell.ColorYellow)
	modal.SetTitle(" Help & Instructions ")
	modal.SetTitleColor(tcell.ColorBlack)

	return modal
}

func createChatListModal() *tview.Flex {
	// Create chat list
	chatList = tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true).
		SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			loadChatFromList(index)
		})

	chatList.SetBorder(true).
		SetTitle(" Chat History ").
		SetBorderColor(tcell.ColorDarkCyan)

	// Instructions
	instructions := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]📚 Chat History\n\n[white]• Use ↑/↓ to navigate\n• Press Enter to load chat\n• Press 'd' to delete selected\n• Press Escape to close").
		SetTextAlign(tview.AlignLeft)
	instructions.SetBorder(true).
		SetTitle(" Instructions ").
		SetBorderColor(tcell.ColorGreen)

	// Button area for chat list
	chatButtonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	loadButton := tview.NewButton("📂 Load").
		SetSelectedFunc(func() {
			index := chatList.GetCurrentItem()
			if index >= 0 {
				loadChatFromList(index)
			}
		})

	deleteButton := tview.NewButton("🗑️ Delete").
		SetSelectedFunc(func() {
			index := chatList.GetCurrentItem()
			if index >= 0 {
				deleteChatFromList(index)
			}
		})

	closeButton := tview.NewButton("❌ Close").
		SetSelectedFunc(func() {
			pages.HidePage("chatlist")
			isShowingChatList = false
		})

	chatButtonFlex.AddItem(loadButton, 0, 1, false).
		AddItem(deleteButton, 0, 1, false).
		AddItem(closeButton, 0, 1, false)

	// Complete chat list layout
	chatListLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(instructions, 8, 1, false).
		AddItem(chatList, 0, 1, true).
		AddItem(chatButtonFlex, 3, 1, false)

	// Set up key bindings for chat list
	chatList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			pages.HidePage("chatlist")
			isShowingChatList = false
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'd', 'D':
				index := chatList.GetCurrentItem()
				if index >= 0 {
					deleteChatFromList(index)
				}
				return nil
			}
		}
		return event
	})

	return chatListLayout
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
			if !isShowingChatList {
				if pages.HasPage("help") {
					name, _ := pages.GetFrontPage()
					if name == "help" {
						pages.HidePage("help")
					} else {
						pages.ShowPage("help")
					}
				}
			}
			return nil
		case tcell.KeyCtrlL:
			if !isShowingChatList {
				clearChat()
			}
			return nil
		case tcell.KeyCtrlN:
			if !isShowingChatList {
				saveCurrentChat()
				startNewChat()
				updateConversationView()
				updateSidebar()
				updateStatus("[blue]🆕 Started new chat!")
			}
			return nil
		case tcell.KeyCtrlS:
			if !isShowingChatList {
				saveCurrentChat()
				updateStatus("[green]💾 Chat saved!")
			}
			return nil
		case tcell.KeyCtrlO:
			if !isShowingChatList {
				showChatList()
			}
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
	userMsg := storage.ChatMessage{
		Role:      "user",
		Content:   prompt,
		Timestamp: time.Now(),
	}
	chatHistory = append(chatHistory, userMsg)
	currentSession.Messages = chatHistory

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
				errorMsg := storage.ChatMessage{
					Role:      "error",
					Content:   fmt.Sprintf("Error: %v", err),
					Timestamp: time.Now(),
				}
				chatHistory = append(chatHistory, errorMsg)
				updateStatus("[red]❌ Error occurred!")
			} else {
				aiMsg := storage.ChatMessage{
					Role:      "assistant",
					Content:   reply,
					Timestamp: time.Now(),
				}
				chatHistory = append(chatHistory, aiMsg)
				updateStatus("[green]✅ Response received!")
			}

			currentSession.Messages = chatHistory
			updateConversationView()
			updateSidebar()

			// Auto-save after each exchange
			go saveCurrentChat()

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
	chatHistory = []storage.ChatMessage{}
	currentSession.Messages = chatHistory
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

func saveCurrentChat() {
	if len(chatHistory) == 0 {
		return // Don't save empty chats
	}

	currentSession.Messages = chatHistory
	if err := storageManager.SaveChat(currentSession); err != nil {
		app.QueueUpdateDraw(func() {
			updateStatus(fmt.Sprintf("[red]❌ Save failed: %v", err))
		})
	}
}

func showChatList() {
	summaries, err := storageManager.GetChatSummaries()
	if err != nil {
		updateStatus(fmt.Sprintf("[red]❌ Failed to load chats: %v", err))
		return
	}

	chatList.Clear()

	if len(summaries) == 0 {
		chatList.AddItem("No saved chats", "Start a conversation to create your first chat!", 0, nil)
	} else {
		for _, summary := range summaries {
			mainText := summary.Title
			secondaryText := fmt.Sprintf("%d messages • Updated: %s",
				summary.MessageCount,
				summary.UpdatedAt.Format("Jan 2, 15:04"))
			chatList.AddItem(mainText, secondaryText, 0, nil)
		}
	}

	pages.ShowPage("chatlist")
	isShowingChatList = true
	app.SetFocus(chatList)
}

func loadChatFromList(index int) {
	summaries, err := storageManager.GetChatSummaries()
	if err != nil || index >= len(summaries) {
		updateStatus("[red]❌ Failed to load chat")
		return
	}

	// Save current chat before loading new one
	saveCurrentChat()

	// Load selected chat
	session, err := storageManager.LoadChat(summaries[index].ID)
	if err != nil {
		updateStatus(fmt.Sprintf("[red]❌ Failed to load chat: %v", err))
		return
	}

	currentSession = session
	chatHistory = session.Messages

	updateConversationView()
	updateSidebar()
	updateStatus("[green]📂 Chat loaded successfully!")

	pages.HidePage("chatlist")
	isShowingChatList = false
	app.SetFocus(inputField)
}

func deleteChatFromList(index int) {
	summaries, err := storageManager.GetChatSummaries()
	if err != nil || index >= len(summaries) {
		updateStatus("[red]❌ Failed to delete chat")
		return
	}

	if err := storageManager.DeleteChat(summaries[index].ID); err != nil {
		updateStatus(fmt.Sprintf("[red]❌ Failed to delete chat: %v", err))
		return
	}

	updateStatus("[yellow]🗑️ Chat deleted!")

	// Refresh the chat list
	showChatList()
}

func updateConversationView() {
	var content strings.Builder

	if len(chatHistory) == 0 {
		content.WriteString("[dim]🌟 Welcome to TUI-GPT!\n\n")
		content.WriteString("Start a conversation by typing a message below.\n")
		content.WriteString("Ask me anything - I'm here to help! 🤖[white]\n\n")
		content.WriteString("[cyan]💾 Your chats are automatically saved!\n")
		content.WriteString("Press Ctrl+O to access your chat history.[white]\n\n")
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
	content.WriteString("• Ctrl+N for new chat\n")
	content.WriteString("• Ctrl+S to save\n")
	content.WriteString("• Ctrl+H for help\n")

	sidebar.SetText(content.String())
}

func updateStatus(status string) {
	statusBar.SetText(status)
}
