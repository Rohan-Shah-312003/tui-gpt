package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Rohan-Shah-312003/tui-gpt/internal/groq"
	"github.com/gdamore/tcell/v2"
	"github.com/mitchellh/go-wordwrap"
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
		SetWordWrap(true).
		SetRegions(true)

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

	// Set focus change function to handle scrolling
	conversationView.SetChangedFunc(func() {
		ml.app.app.Draw()
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

func (ml *MainLayout) formatCodeBlocks(text string) string {
	codeBlockRegex := regexp.MustCompile("```([a-zA-Z]*)\\n([\\s\\S]*?)\\n```")

	formatted := codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
		parts := codeBlockRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match // fallback
		}

		language := parts[1]
		code := parts[2]

		// Create language label
		langLabel := ""
		if language != "" {
			langLabel = fmt.Sprintf("  %s  ", strings.ToUpper(language))
		} else {
			langLabel = "  CODE  "
		}

		// Calculate code block width
		lines := strings.Split(code, "\n")
		maxWidth := len(langLabel) + 4 // minimum width for header
		for _, line := range lines {
			if len(line) > maxWidth {
				maxWidth = len(line)
			}
		}
		maxWidth += 4 // padding

		// Create rounded code block
		var result strings.Builder
		result.WriteString(fmt.Sprintf("\n[black:lightgray]%s[white:black]\n",
			strings.Repeat(" ", maxWidth)))
		result.WriteString(fmt.Sprintf("[black:lightgray]%s%s%s[white:black]\n",
			strings.Repeat(" ", (maxWidth-len(langLabel))/2),
			langLabel,
			strings.Repeat(" ", maxWidth-(maxWidth-len(langLabel))/2-len(langLabel))))
		result.WriteString(fmt.Sprintf("[black:lightgray]%s[white:black]\n",
			strings.Repeat(" ", maxWidth)))

		// Add code lines with background
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				result.WriteString(fmt.Sprintf("[black:darkgray]%s[white:black]\n",
					strings.Repeat(" ", maxWidth)))
			} else {
				padding := maxWidth - len(line) - 4
				if padding < 0 {
					padding = 0
				}
				result.WriteString(fmt.Sprintf("[white:darkgray]  %s%s  [white:black]\n",
					line, strings.Repeat(" ", padding)))
			}
		}

		result.WriteString(fmt.Sprintf("[black:lightgray]%s[white:black]\n",
			strings.Repeat(" ", maxWidth)))

		return result.String()
	})

	// Format inline code with better styling
	inlineCodeRegex := regexp.MustCompile("`([^`]+)`")
	formatted = inlineCodeRegex.ReplaceAllString(formatted, "[black:lightgray] $1 [white:black]")

	// Format **bold** text
	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	formatted = boldRegex.ReplaceAllString(formatted, "[::b]$1[::-]")

	// Format *italic* text
	italicRegex := regexp.MustCompile(`\*([^*]+)\*`)
	formatted = italicRegex.ReplaceAllString(formatted, "[::i]$1[::-]")

	return formatted
}

func (ml *MainLayout) createModernChatBubble(role, content, timestamp string) string {
	// Get the available width for the text
	_, _, width, _ := ml.conversationView.GetRect()
	maxBubbleWidth := width - 20 // Leave space for margins and alignment
	if maxBubbleWidth < 40 {
		maxBubbleWidth = 40
	}

	// Format and process the content
	formattedContent := ml.formatCodeBlocks(content)
	cleanContent := strings.TrimSpace(formattedContent)

	// Determine bubble properties based on role
	var bubbleColor, alignmentSpaces string
	var isRightAligned bool

	switch role {
	case "user":
		bubbleColor = "[white:blue]"
		// textColor = "[white:blue]"
		isRightAligned = true
	case "assistant":
		bubbleColor = "[black:lightgray]"
		// textColor = "[black:lightgray]"
		isRightAligned = false
	case "error":
		bubbleColor = "[white:red]"
		// textColor = "[white:red]"
		isRightAligned = false
	default:
		bubbleColor = "[black:yellow]"
		// textColor = "[black:yellow]"
		isRightAligned = false
	}

	// Wrap content to fit in bubble
	contentWidth := maxBubbleWidth - 6 // Account for padding and bubble edges

	// Split into paragraphs and wrap each
	paragraphs := strings.Split(cleanContent, "\n\n")
	var wrappedParagraphs []string

	for _, p := range paragraphs {
		if strings.Contains(p, "[black:lightgray]") || strings.Contains(p, "[white:darkgray]") {
			// This is a code block, don't wrap it
			wrappedParagraphs = append(wrappedParagraphs, p)
		} else {
			// Regular text paragraph
			singleLine := strings.ReplaceAll(p, "\n", " ")
			wrapped := wordwrap.WrapString(singleLine, uint(contentWidth))
			wrappedParagraphs = append(wrappedParagraphs, wrapped)
		}
	}

	processedContent := strings.Join(wrappedParagraphs, "\n\n")
	contentLines := strings.Split(processedContent, "\n")

	// Calculate actual bubble width based on content
	actualBubbleWidth := 0
	for _, line := range contentLines {
		// Remove color codes for width calculation
		cleanLine := regexp.MustCompile(`\[[^\]]*\]`).ReplaceAllString(line, "")
		lineWidth := len(cleanLine) + 6 // Add padding
		if lineWidth > actualBubbleWidth {
			actualBubbleWidth = lineWidth
		}
	}

	if actualBubbleWidth > maxBubbleWidth {
		actualBubbleWidth = maxBubbleWidth
	}
	if actualBubbleWidth < 20 {
		actualBubbleWidth = 20
	}

	// Calculate alignment spacing
	if isRightAligned {
		alignmentSpaces = strings.Repeat(" ", width-actualBubbleWidth-10)
	} else {
		alignmentSpaces = "  " // Small left margin for assistant messages
	}

	// Build the modern chat bubble
	var bubble strings.Builder

	// Add some vertical spacing
	bubble.WriteString("\n")

	// Create rounded top
	bubble.WriteString(fmt.Sprintf("%s%s%s%s[white:black]\n",
		alignmentSpaces,
		bubbleColor,
		strings.Repeat(" ", actualBubbleWidth),
		""))

	// Add content lines
	for _, line := range contentLines {
		// Remove color codes for padding calculation
		cleanLine := regexp.MustCompile(`\[[^\]]*\]`).ReplaceAllString(line, "")

		// Handle different line types
		if strings.Contains(line, "[black:lightgray]") || strings.Contains(line, "[white:darkgray]") {
			// Code block line - preserve as is
			bubble.WriteString(fmt.Sprintf("%s%s[white:black]\n", alignmentSpaces, line))
		} else {
			// Regular text line
			padding := actualBubbleWidth - len(cleanLine) - 6
			if padding < 0 {
				padding = 0
			}

			bubble.WriteString(fmt.Sprintf("%s%s   %s%s   %s[white:black]\n",
				alignmentSpaces,
				bubbleColor,
				line,
				strings.Repeat(" ", padding),
				""))
		}
	}

	// Add timestamp line
	timestampText := fmt.Sprintf("  %s  ", timestamp)
	timestampPadding := actualBubbleWidth - len(timestampText)
	if timestampPadding < 0 {
		timestampPadding = 0
	}

	bubble.WriteString(fmt.Sprintf("%s%s%s%s%s[white:black]\n",
		alignmentSpaces,
		"[dim]"+bubbleColor,
		strings.Repeat(" ", timestampPadding/2),
		timestampText,
		strings.Repeat(" ", timestampPadding-timestampPadding/2)))

	// Create rounded bottom
	bubble.WriteString(fmt.Sprintf("%s%s%s%s[white:black]\n",
		alignmentSpaces,
		bubbleColor,
		strings.Repeat(" ", actualBubbleWidth),
		""))

	// Add spacing after bubble
	bubble.WriteString("\n")

	return bubble.String()
}

func (ml *MainLayout) updateConversationView() {
	chatHistory := ml.app.GetChatHistory()

	// Clear existing content
	ml.conversationView.Clear()

	if len(chatHistory) == 0 {
		// Create an attractive welcome message
		var welcome strings.Builder

		// Welcome message with modern styling
		welcome.WriteString("\n\n")
		welcome.WriteString("                    [white:blue]                                      [white:black]\n")
		welcome.WriteString("                    [white:blue]     🌟 Welcome to TUI-GPT! 🌟     [white:black]\n")
		welcome.WriteString("                    [white:blue]                                      [white:black]\n")
		welcome.WriteString("\n")
		welcome.WriteString("                    [black:lightgray]                                      [white:black]\n")
		welcome.WriteString("                    [black:lightgray]   ✨ Ready to chat! Start typing...   [white:black]\n")
		welcome.WriteString("                    [black:lightgray]   Press Ctrl+H for help & shortcuts   [white:black]\n")
		welcome.WriteString("                    [black:lightgray]                                      [white:black]\n")
		welcome.WriteString("\n\n")

		ml.conversationView.SetText(welcome.String())
		return
	}

	// Build the conversation text with modern bubbles
	var conversation strings.Builder

	// Add a small top margin
	conversation.WriteString("\n")

	// Add chat messages with modern bubbles
	for _, msg := range chatHistory {
		timestamp := msg.Timestamp.Format("15:04")
		modernBubble := ml.createModernChatBubble(msg.Role, msg.Content, timestamp)
		conversation.WriteString(modernBubble)
	}

	// Add bottom margin
	conversation.WriteString("\n\n")

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
			ml.app.clipboard = lastMsg.Content
		} else {
			ml.app.clipboard = conversationText
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
