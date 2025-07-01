package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// wrapText wraps the input text to the specified maxWidth, breaking on word boundaries
func wrapText(text string, maxWidth int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len() > 0 && currentLine.Len()+1+len(word) > maxWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// createUserBubble creates a styled text view for user messages
func createUserBubble(text, timestamp string, width int) *tview.TextView {
	bubble := tview.NewTextView()
	bubble.SetText(text)
	bubble.SetTextColor(tcell.ColorWhite)
	bubble.SetDynamicColors(true)
	bubble.SetTextAlign(tview.AlignRight)
	bubble.SetBorder(true)
	bubble.SetBorderColor(tcell.ColorBlue)
	bubble.SetTitle(" You (" + timestamp + ")")
	bubble.SetTitleColor(tcell.ColorBlue)
	bubble.SetWrap(true)
	bubble.SetWordWrap(true)
	return bubble
}

// createAIBubble creates a styled text view for AI assistant messages
func createAIBubble(text, timestamp string, width int) *tview.TextView {
	bubble := tview.NewTextView()
	bubble.SetText(text)
	bubble.SetTextColor(tcell.ColorWhite)
	bubble.SetDynamicColors(true)
	bubble.SetTextAlign(tview.AlignLeft)
	bubble.SetBorder(true)
	bubble.SetBorderColor(tcell.ColorGreen)
	bubble.SetTitle(" AI (" + timestamp + ")")
	bubble.SetTitleColor(tcell.ColorGreen)
	bubble.SetWrap(true)
	bubble.SetWordWrap(true)
	return bubble
}

// createErrorBubble creates a styled text view for error messages
func createErrorBubble(text, timestamp string, width int) *tview.TextView {
	bubble := tview.NewTextView()
	bubble.SetText(text)
	bubble.SetTextColor(tcell.ColorWhite)
	bubble.SetDynamicColors(true)
	bubble.SetTextAlign(tview.AlignCenter)
	bubble.SetBorder(true)
	bubble.SetBorderColor(tcell.ColorRed)
	bubble.SetTitle(" Error (" + timestamp + ")")
	bubble.SetTitleColor(tcell.ColorRed)
	bubble.SetWrap(true)
	bubble.SetWordWrap(true)
	return bubble
}
