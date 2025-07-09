package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8" // Import for accurate character counting

	"github.com/Rohan-Shah-312003/tui-gpt/internal/groq"
	"github.com/gdamore/tcell/v2"

	"github.com/rivo/tview"
)

type MainLayout struct {
	app              *App
	conversationView *tview.TextView // This will hold all chat messages as a single, formatted string
	inputField       *tview.InputField
	statusBar        *tview.TextView
	sidebar          *tview.TextView
	// messageContainer *tview.Flex // Removed: No longer needed for individual bubbles
}

func NewMainLayout(app *App) *MainLayout {
	return &MainLayout{
		app: app,
	}
}

func (ml *MainLayout) Create() *tview.Flex {
	header := ml.createHeader()
	ml.sidebar = ml.createSidebar()
	ml.conversationView = ml.createConversationView() // Now returns a TextView
	ml.inputField = ml.createInputField()
	buttonFlex := ml.createButtonFlex()
	ml.statusBar = ml.createStatusBar()

	inputSection := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ml.inputField, 4, 1, true).
		AddItem(buttonFlex, 3, 1, false)

	mainContent := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(ml.conversationView, 0, 4, false). // conversationView is the scrollable wrapper
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

// createConversationView returns a TextView that will display all messages.
func (ml *MainLayout) createConversationView() *tview.TextView {
	conversationView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true). // Keep word wrap TRUE here!
		SetRegions(true).  // Enable regions for potential future highlighting
		SetMaxLines(0)     // Allow unlimited lines

	conversationView.SetBorder(true).
		SetTitle(" 💬 Conversation ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorLightGreen)

	// Add input capture for copy functionality and scrolling
	conversationView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			ml.copySelectedText()
			return nil
		case tcell.KeyPgUp:
			ml.conversationView.ScrollToBeginning()
			return nil
		case tcell.KeyPgDn:
			ml.conversationView.ScrollToEnd()
			return nil
		case tcell.KeyUp:
			row, col := ml.conversationView.GetScrollOffset()
			if row > 0 {
				ml.conversationView.ScrollTo(row-1, col)
			}
			return nil
		case tcell.KeyDown:
			row, col := ml.conversationView.GetScrollOffset()
			ml.conversationView.ScrollTo(row+1, col)
			return nil
		}
		return event
	})

	// Set a changed func to ensure it always scrolls to end when new content is added
	conversationView.SetChangedFunc(func() {
		ml.app.app.QueueUpdateDraw(func() {
			ml.conversationView.ScrollToEnd()
		})
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

	newChatButton := tview.NewButton("📝 New Chat").
		SetSelectedFunc(ml.app.newChat).
		SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorBlueViolet).Foreground(tcell.ColorBlack))

	modelButton := tview.NewButton("🤖 Models").
		SetSelectedFunc(ml.app.modelListModal.Show).
		SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorPurple).Foreground(tcell.ColorBlack))

	clearButton := tview.NewButton("🧹 Clear").
		SetSelectedFunc(ml.app.clearChat).
		SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorOrange).Foreground(tcell.ColorBlack))

	quitButton := tview.NewButton("🚪 Quit").SetSelectedFunc(func() {
		ml.app.saveCurrentChat()
		ml.app.app.Stop()
	}).SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorBlack))
	quitButton.SetRect(0, 0, 10, 1)

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

// stripANSICodes removes Tview color tags from a string.
func stripANSICodes(s string) string {
	re := regexp.MustCompile(`\[[^\]]*\]`)
	return re.ReplaceAllString(s, "")
}

// runeCountInString counts the number of runes in a string, effectively ignoring Tview color tags
func runeCountInString(s string) int {
	return utf8.RuneCountInString(stripANSICodes(s))
}

// formatCodeBlocks formats code blocks, inline code, bold, and italic text.
// It uses only foreground and specific background for code sections, no full-line backgrounds.
func (ml *MainLayout) formatCodeBlocks(text string) string {
	codeBlockRegex := regexp.MustCompile("```([a-zA-Z]*)\\n([\\s\\S]*?)\\n```")

	formatted := codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := codeBlockRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match // fallback
		}

		language := parts[1]
		code := parts[2]

		langLabel := ""
		if language != "" {
			langLabel = fmt.Sprintf(" %s ", strings.ToUpper(language))
		} else {
			langLabel = " CODE "
		}

		var result strings.Builder
		// Simple header for code block
		result.WriteString(fmt.Sprintf("\n[black:lightgray]%s[white:black]\n", langLabel))
		// Code lines with dark gray background and white foreground
		for _, line := range strings.Split(code, "\n") {
			result.WriteString(fmt.Sprintf("[white:darkgray]%s[white:black]\n", line))
		}
		// Simple footer for code block
		result.WriteString(fmt.Sprintf("[black:lightgray]%s[white:black]\n", strings.Repeat(" ", runeCountInString(langLabel))))

		return result.String()
	})

	inlineCodeRegex := regexp.MustCompile("`([^`]+)`")
	formatted = inlineCodeRegex.ReplaceAllString(formatted, "[black:lightgray] $1 [white:black]")

	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	formatted = boldRegex.ReplaceAllString(formatted, "[::b]$1[::-]")

	italicRegex := regexp.MustCompile(`\*([^*]+)\*`)
	formatted = italicRegex.ReplaceAllString(formatted, "[::i]$1[::-]")

	return formatted
}

// formatChatMessage prepares a message for display as plain text with role, content, and timestamp.
// No "bubbles" (solid background blocks) are created.
func (ml *MainLayout) formatChatMessage(role, content, timestamp string) string {
	var rolePrefix, contentFgColor, timestampFgColor string

	switch role {
	case "user":
		rolePrefix = "You:"
		contentFgColor = "white"  // User text in white
		timestampFgColor = "gray" // Timestamp in gray
	case "assistant":
		rolePrefix = "AI:"
		contentFgColor = "white"      // Changed from "lightcyan" to "white" to remove highlighting
		timestampFgColor = "darkgray" // Timestamp in dark gray
	case "error":
		rolePrefix = "Error:"
		contentFgColor = "red"
		timestampFgColor = "darkred"
	default:
		rolePrefix = "System:"
		contentFgColor = "yellow"
		timestampFgColor = "orange"
	}

	var messageBuilder strings.Builder

	// Add a newline for separation before each message
	messageBuilder.WriteString("\n")

	// Role prefix (e.g., "You:", "AI:")
	messageBuilder.WriteString(fmt.Sprintf("[::b]%s[::-]", rolePrefix))

	// Timestamp - appended to the role line or below if it doesn't fit
	// For simplicity, let's put it on a new line below the content for clarity
	// and to ensure it doesn't interfere with word wrapping of the main content.
	// timestampFormatted := fmt.Sprintf("[%s]%s[::-]", timestampFgColor, timestamp)

	// Main content with its color
	formattedContent := ml.formatCodeBlocks(content) // Apply code block formatting here

	// Add content
	messageBuilder.WriteString(fmt.Sprintf(" [%s]%s[::-]\n", contentFgColor, formattedContent))

	// Add timestamp on a new line, right-aligned within the *available space*
	// This will not have a solid background, just text color.
	_, _, viewWidth, _ := ml.conversationView.GetRect()
	effectiveWidth := viewWidth - 4 // Account for conversationView borders

	timestampText := fmt.Sprintf("[%s]%s[::-]", timestampFgColor, timestamp)
	timestampPadded := strings.Repeat(" ", effectiveWidth-runeCountInString(timestampText)) + timestampText
	messageBuilder.WriteString(timestampPadded)

	messageBuilder.WriteString("\n") // Newline after message for next one or spacing

	return messageBuilder.String()
}

func (ml *MainLayout) updateConversationView() {
	chatHistory := ml.app.GetChatHistory()

	// Clear existing content
	ml.conversationView.Clear()

	// Build the conversation text
	var conversation strings.Builder

	if len(chatHistory) == 0 {
		// Welcome message - simplified to plain text, no full backgrounds
		welcomeMsg := `
[::b]🌟 Welcome to TUI-GPT! 🌟[::-]

✨ Ready to chat! Start typing...

[yellow]Press Ctrl+H for help & shortcuts[::-]
`
		conversation.WriteString(welcomeMsg)
	} else {
		// Add chat messages as formatted text lines
		for _, msg := range chatHistory {
			timestamp := msg.Timestamp.Format("15:04")
			formattedMessage := ml.formatChatMessage(msg.Role, msg.Content, timestamp)
			conversation.WriteString(formattedMessage)
		}
	}

	// Set the text and scroll to the bottom
	ml.conversationView.SetText(conversation.String())
	ml.conversationView.ScrollToEnd()
}

func (ml *MainLayout) updateSidebar() {
	var content strings.Builder
	chatHistory := ml.app.GetChatHistory()
	currentSession := ml.app.GetCurrentSession()

	content.WriteString("[yellow]📊 ANALYTICS[white]\n")
	content.WriteString(fmt.Sprintf("[yellow][white] Total Messages: %-4d[yellow][white]\n", len(chatHistory)))

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

	content.WriteString(fmt.Sprintf("[yellow][white] 👤 Your msgs: %-6d[yellow][white]\n", userCount))
	content.WriteString(fmt.Sprintf("[yellow][white] 🤖 AI replies: %-5d[yellow][white]\n", aiCount))
	if errorCount > 0 {
		content.WriteString(fmt.Sprintf("[yellow][white] ❌ Errors: %-8d[yellow][white]\n", errorCount))
	}
	content.WriteString(fmt.Sprintf("[yellow][white] 📝 Characters: %-4d[yellow][white]\n", totalChars))

	// Current Model Info
	currentModel := groq.GetCurrentModel()
	models := groq.GetAvailableModels()
	if modelName, exists := models[currentModel]; exists {
		content.WriteString("[cyan] 🤖 MODEL [white]\n")
		modelDisplayName := strings.Replace(modelName, "Meta ", "", 1)
		if len(modelDisplayName) > 15 {
			modelDisplayName = modelDisplayName[:15] + "..."
		}
		content.WriteString(fmt.Sprintf("[cyan][white] %-15s[cyan][white]\n", modelDisplayName))
		content.WriteString(fmt.Sprintf("\n\n"))
	}

	// Session Info
	content.WriteString("[magenta]🕒 SESSION [white]\n")
	if currentSession != nil {
		content.WriteString(fmt.Sprintf("[magenta][white] Started: %s[magenta][white]\n", currentSession.CreatedAt.Format("15:04")))
		if currentSession.Title != "" && len(currentSession.Title) > 0 {
			title := currentSession.Title
			if len(title) > 15 {
				title = title[:12] + "..."
			}
			content.WriteString(fmt.Sprintf("[magenta][white] Title: %-9s[magenta][white]\n", title))
		}
	}
	if len(chatHistory) > 0 {
		lastMsg := chatHistory[len(chatHistory)-1]
		content.WriteString(fmt.Sprintf("[magenta][white] Last: %s[magenta][white]\n", lastMsg.Timestamp.Format("15:04:05")))
	}
	content.WriteString("\n\n")

	// Quick Tips
	content.WriteString("[white]💡 SHORTCUTS[white]\n")
	content.WriteString("[white][yellow] Enter[white] - Send msg     [white]\n")
	content.WriteString("[white][yellow] Ctrl+C[white] - Copy text   [white]\n")
	content.WriteString("[white][yellow] Ctrl+V[white] - Paste text  [white]\n")
	content.WriteString("[white][yellow] Ctrl+O[white] - Chat history [white]\n")
	content.WriteString("[white][yellow] Ctrl+-[white] - Change model [white]\n")
	content.WriteString("[white][yellow] Ctrl+N[white] - New chat     [white]\n")
	content.WriteString("[white][yellow] Ctrl+H[white] - Help menu    [white]\n")
	content.WriteString("[white][yellow] PgUp/Dn[white] - Scroll chat [white]\n")
	content.WriteString("[white][yellow] ↑/↓[white] - Line scroll    [white]\n")
	content.WriteString("\n\n")

	ml.sidebar.SetText(content.String())
}

func (ml *MainLayout) updateStatus(status string) {
	ml.statusBar.SetText(status)
}

// Copy functionality
func (ml *MainLayout) copySelectedText() {
	// Get the entire conversation text for copying
	conversationText := ml.conversationView.GetText(false)
	if conversationText != "" {
		// Try to get the last AI response or use the entire conversation
		chatHistory := ml.app.GetChatHistory()
		if len(chatHistory) > 0 {
			lastMsg := chatHistory[len(chatHistory)-1]
			// Strip Tview color codes for clipboard content
			ml.app.clipboard = stripANSICodes(lastMsg.Content)
		} else {
			ml.app.clipboard = stripANSICodes(conversationText)
		}

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

// Additional helper methods for TextView functionality
func (ml *MainLayout) ScrollToTop() {
	ml.conversationView.ScrollToBeginning()
}

func (ml *MainLayout) ScrollToBottom() {
	ml.conversationView.ScrollToEnd()
}

func (ml *MainLayout) GetConversationView() *tview.TextView {
	return ml.conversationView
}
