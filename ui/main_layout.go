package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

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
	messageContainer *tview.Flex
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
		AddItem(ml.inputField, 4, 1, true).
		AddItem(buttonFlex, 3, 1, false)

	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(ml.conversationView, 0, 4, false).
		AddItem(ml.sidebar, 30, 1, false)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 4, 1, false).
		AddItem(mainContent, 0, 1, false).
		AddItem(inputSection, 7, 1, true).
		AddItem(ml.statusBar, 3, 1, false)

	ml.updateConversationView()
	ml.updateSidebar()

	return mainLayout
}

func (ml *MainLayout) createHeader() *tview.TextView {
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[::bu]🚀 TUI-GPT - Enhanced Chat Experience [::-]\n[dim]Press Ctrl+H for help • Ctrl+O for history • Ctrl+- for models • Ctrl+C to copy • Ctrl+V to paste")
	header.SetBorder(true).
		SetBorderPadding(0, 0, 1, 1).
		SetTitle(" ✨ Welcome to Enhanced TUI-GPT ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorLightBlue).
		SetBorderStyle(tcell.StyleDefault.Foreground(tcell.ColorLightBlue))
	return header
}

func (ml *MainLayout) createSidebar() *tview.TextView {
	sidebar := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	sidebar.SetBorder(true).
		SetTitle(" 📊 Chat Analytics ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorPurple)
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
		SetTitle(" 💬 Conversation ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorLightGreen)

	// Add input capture for copy functionality
	conversationView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			ml.copySelectedText()
			return nil
		}
		return event
	})

	return conversationView
}

func (ml *MainLayout) createInputField() *tview.InputField {
	inputField := tview.NewInputField().
		SetLabel("💭 Message: ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorNavy).
		SetPlaceholder("Type your message... (Enter to send, Ctrl+V to paste)").
		SetFieldTextColor(tcell.ColorWhite)

	inputField.SetBorder(true).
		SetTitle(" ✍️ Your Input ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorBlue)

	return inputField
}

func (ml *MainLayout) createButtonFlex() *tview.Flex {
	buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	sendButton := tview.NewButton("🚀 Send").
		SetSelectedFunc(ml.app.sendMessage).
		SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack))
	// sendButton.SetBorder(true).SetBorderColor(tcell.ColorGreen)

	newChatButton := tview.NewButton("📝 New Chat").
		SetSelectedFunc(ml.app.newChat).
		SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorBlueViolet).Foreground(tcell.ColorBlack))
	// newChatButton.SetBorder(true).SetBorderColor(tcell.ColorBlueViolet)

	modelButton := tview.NewButton("🤖 Models").
		SetSelectedFunc(ml.app.modelListModal.Show).
		SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorPurple).Foreground(tcell.ColorBlack))
	// modelButton.SetBorder(true).SetBorderColor(tcell.ColorPurple)

	clearButton := tview.NewButton("🧹 Clear").
		SetSelectedFunc(ml.app.clearChat).
		SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorOrange).Foreground(tcell.ColorBlack))
	// clearButton.SetBorder(true).SetBorderColor(tcell.ColorOrange)

	quitButton := tview.NewButton("🚪 Quit").SetSelectedFunc(func() {
		ml.app.saveCurrentChat()
		ml.app.app.Stop()
	}).SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorBlack))
	quitButton.SetRect(0, 0, 10, 1)
	// quitButton.SetBorder(true).SetBorderColor(tcell.ColorRed)

	buttonFlex.
		AddItem(tview.NewBox(), 1, 0, false). // Left padding
		AddItem(sendButton, 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false). // Spacing between buttons
		AddItem(newChatButton, 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false). // Spacing between buttons
		AddItem(modelButton, 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false). // Spacing between buttons
		AddItem(clearButton, 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false). // Spacing between buttons
		AddItem(quitButton, 0, 1, false).
		AddItem(tview.NewBox(), 1, 0, false) // Right padding

	return buttonFlex
}

func (ml *MainLayout) createStatusBar() *tview.TextView {
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[green]🟢 Ready - Enhanced UI Mode")
	statusBar.SetBorder(true).
		SetTitle(" 📡 Status ").
		SetBorderColor(tcell.ColorDarkCyan)
	return statusBar
}

func (ml *MainLayout) formatCodeBlocks(text string) string {
	codeBlockRegex := regexp.MustCompile("```([a-zA-Z]*)\\n([\\s\\S]*?)\\n```")

	formatted := codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := codeBlockRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match // fallback
		}

		code := parts[2]

		// Header Starting Line
		header := fmt.Sprintf("[yellow]┌────────┐[white]")

		// Format lines with line numbers
		lines := strings.Split(code, "\n")
		var formattedLines []string
		maxWidth := 0
		for i, line := range lines {
			lineContent := line
			if strings.TrimSpace(line) == "" {
				lineContent = " "
			}
			lineStr := fmt.Sprintf("[yellow]│[gray]%2d[white] %s", i+1, lineContent)
			formattedLines = append(formattedLines, lineStr)
			if len(lineContent) > maxWidth {
				maxWidth = len(lineContent)
			}
		}

		// Footer width based on max line content length
		footer := "[yellow]└" + strings.Repeat("─", len(header)-10) + "┘[white]"

		return fmt.Sprintf("\n%s\n%s\n%s\n",
			header,
			strings.Join(formattedLines, "\n"),
			footer,
		)
	})

	// Format inline code: `code`
	inlineCodeRegex := regexp.MustCompile("`([^`]+)`")
	formatted = inlineCodeRegex.ReplaceAllString(formatted, "[lightblue]`[cyan]$1[lightblue]`[white]")

	return formatted
}

func (ml *MainLayout) createChatBubble(role, content, timestamp string) *tview.TextView {
	// Create a text view for the bubble
	bubble := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetWordWrap(true)

	// Set colors and icons based on role
	var color tcell.Color
	var icon, label string

	switch role {
	case "user":
		color = tcell.ColorBlue
		icon, label = "👤", "You"
	case "assistant":
		color = tcell.ColorGreen
		icon, label = "🤖", "AI Assistant"
	case "error":
		color = tcell.ColorRed
		icon, label = "❌", "Error"
	default:
		color = tcell.ColorRed
		icon, label = "❓", "Unknown"
	}

	// Format the header with timestamp
	header := fmt.Sprintf("%s %s • %s", icon, label, timestamp)

	// Set bubble styling
	bubble.SetBorder(true).
		SetTitle(header).
		SetTitleColor(color).
		SetBorderColor(color).
		SetBackgroundColor(tcell.ColorBlack)

	// Format and set the content
	formattedContent := ml.formatCodeBlocks(content)
	bubble.SetText(formattedContent)

	// Add some padding
	bubble.SetText(fmt.Sprintf("\n%s\n", formattedContent)).
		SetBackgroundColor(tcell.ColorBlack).
		SetBorder(true).
		SetBorderColor(tcell.ColorDarkGray)

	return bubble
}

func (ml *MainLayout) updateConversationView() {
	chatHistory := ml.app.GetChatHistory()

	// Clear existing content
	ml.conversationView.Clear()

	if len(chatHistory) == 0 {
		// Create a welcome message
		welcome := tview.NewTextView().
			SetDynamicColors(true).
			SetText(
				"\n[cyan]══════════════════════════════════════════════\n" +
					"              🌟 Welcome! 🌟                  \n" +
					"══════════════════════════════════════════════\n" +
					"[white] Welcome to TUI-GPT!            [cyan]\n" +
					"[white] Features:                               [cyan]\n" +
					"[white] • Auto-save conversations locally               [cyan]\n" +
					"[white] • Multiple AI models                    [cyan]\n" +
					"══════════════════════════════════════════════\n" +
					"[white] Start typing below to begin! 👇         [cyan]\n" +
					"══════════════════════════════════════════════[white]")

		welcome.SetBorder(true).
			SetBorderColor(tcell.ColorDarkCyan).
			SetTitle(" Welcome to TUI-GPT ").
			SetTitleColor(tcell.ColorDarkCyan)

		// Add welcome message to conversation view
		fmt.Fprintf(ml.conversationView, "%s\n\n", welcome.GetText(true))
		return
	}

	// Add chat messages
	for _, msg := range chatHistory {
		timestamp := msg.Timestamp.Format("15:04:05")
		bubble := ml.createChatBubble(msg.Role, msg.Content, timestamp)
		fmt.Fprintf(ml.conversationView, "%s\n\n", bubble.GetText(true))
	}

	// Auto-scroll to bottom
	ml.conversationView.ScrollToEnd()
}

func (ml *MainLayout) updateSidebar() {
	var content strings.Builder
	chatHistory := ml.app.GetChatHistory()
	currentSession := ml.app.GetCurrentSession()

	content.WriteString("[yellow]╔═══ 📊 ANALYTICS ═══╗[white]\n")
	content.WriteString(fmt.Sprintf("[yellow]║[white] Total Messages: %-4d[yellow]║[white]\n", len(chatHistory)))

	userCount, aiCount, errorCount := 0, 0, 0
	totalChars := 0
	for _, msg := range chatHistory {
		totalChars += len(msg.Content)
		switch msg.Role {
		case "user":
			userCount++
		case "assistant":
			aiCount++
		case "error":
			errorCount++
		}
	}

	content.WriteString(fmt.Sprintf("[yellow]║[white] 👤 Your msgs: %-6d[yellow]║[white]\n", userCount))
	content.WriteString(fmt.Sprintf("[yellow]║[white] 🤖 AI replies: %-5d[yellow]║[white]\n", aiCount))
	if errorCount > 0 {
		content.WriteString(fmt.Sprintf("[yellow]║[white] ❌ Errors: %-8d[yellow]║[white]\n", errorCount))
	}
	content.WriteString(fmt.Sprintf("[yellow]║[white] 📝 Characters: %-4d[yellow]║[white]\n", totalChars))
	content.WriteString("[yellow]╚═══════════════════╝[white]\n\n")

	// Current Model Info
	currentModel := groq.GetCurrentModel()
	models := groq.GetAvailableModels()
	if modelName, exists := models[currentModel]; exists {
		content.WriteString("[cyan]╔═══ 🤖 MODEL ═══╗[white]\n")
		modelDisplayName := strings.Replace(modelName, "Meta ", "", 1)
		if len(modelDisplayName) > 15 {
			modelDisplayName = modelDisplayName[:15] + "..."
		}
		content.WriteString(fmt.Sprintf("[cyan]║[white] %-15s[cyan]║[white]\n", modelDisplayName))
		content.WriteString("[cyan]╚═════════════════╝[white]\n\n")
	}

	// Session Info
	content.WriteString("[magenta]╔═══ 🕒 SESSION ═══╗[white]\n")
	if currentSession != nil {
		content.WriteString(fmt.Sprintf("[magenta]║[white] Started: %s[magenta]║[white]\n", currentSession.CreatedAt.Format("15:04")))
		if currentSession.Title != "" && len(currentSession.Title) > 0 {
			title := currentSession.Title
			if len(title) > 15 {
				title = title[:12] + "..."
			}
			content.WriteString(fmt.Sprintf("[magenta]║[white] Title: %-9s[magenta]║[white]\n", title))
		}
	}
	if len(chatHistory) > 0 {
		lastMsg := chatHistory[len(chatHistory)-1]
		content.WriteString(fmt.Sprintf("[magenta]║[white] Last: %s[magenta]║[white]\n", lastMsg.Timestamp.Format("15:04:05")))
	}
	content.WriteString("[magenta]╚═══════════════════╝[white]\n\n")

	// Quick Tips
	content.WriteString("[white]╔══ 💡 SHORTCUTS ══╗[white]\n")
	content.WriteString("[white]║[yellow] Enter[white] - Send msg     ║[white]\n")
	content.WriteString("[white]║[yellow] Ctrl+C[white] - Copy text   ║[white]\n")
	content.WriteString("[white]║[yellow] Ctrl+V[white] - Paste text  ║[white]\n")
	content.WriteString("[white]║[yellow] Ctrl+O[white] - Chat history ║[white]\n")
	content.WriteString("[white]║[yellow] Ctrl+-[white] - Change model ║[white]\n")
	content.WriteString("[white]║[yellow] Ctrl+N[white] - New chat     ║[white]\n")
	content.WriteString("[white]║[yellow] Ctrl+H[white] - Help menu    ║[white]\n")
	content.WriteString("[white]╚═══════════════════╝[white]\n")

	ml.sidebar.SetText(content.String())
}

func (ml *MainLayout) updateStatus(status string) {
	ml.statusBar.SetText(status)
}

// Copy functionality
func (ml *MainLayout) copySelectedText() {
	// For now, copy the last AI response or create a simple copy mechanism
	chatHistory := ml.app.GetChatHistory()
	if len(chatHistory) > 0 {
		lastMsg := chatHistory[len(chatHistory)-1]
		ml.app.clipboard = lastMsg.Content
		ml.updateStatus("[green]📋 Text copied to clipboard!")

		// Clear status after 2 seconds
		go func() {
			time.Sleep(2 * time.Second)
			ml.app.app.QueueUpdateDraw(func() {
				ml.updateStatus("[green]🟢 Ready")
			})
		}()
	}
}
