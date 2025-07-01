package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rohan-Shah-312003/tui-gpt/internal/groq"
	"github.com/Rohan-Shah-312003/tui-gpt/internal/storage"
)

func (a *App) sendMessage() {
	prompt := strings.TrimSpace(a.mainLayout.inputField.GetText())
	if prompt == "" {
		a.mainLayout.updateStatus("[red]⚠️  Empty message!")
		return
	}

	userMsg := storage.ChatMessage{
		Role:      "user",
		Content:   prompt,
		Timestamp: time.Now(),
	}
	a.chatHistory = append(a.chatHistory, userMsg)
	a.currentSession.Messages = a.chatHistory

	a.mainLayout.inputField.SetText("")
	a.mainLayout.updateConversationView()
	a.mainLayout.updateStatus("[yellow]🤔 AI is thinking...")
	a.mainLayout.updateSidebar()

	go func() {
		reply, err := groq.SendPrompt(prompt)
		a.app.QueueUpdateDraw(func() {
			if err != nil {
				errorMsg := storage.ChatMessage{
					Role:      "error",
					Content:   fmt.Sprintf("Error: %v", err),
					Timestamp: time.Now(),
				}
				a.chatHistory = append(a.chatHistory, errorMsg)
				a.mainLayout.updateStatus("[red]❌ Error occurred!")
			} else {
				aiMsg := storage.ChatMessage{
					Role:      "assistant",
					Content:   reply,
					Timestamp: time.Now(),
				}
				a.chatHistory = append(a.chatHistory, aiMsg)
				a.mainLayout.updateStatus("[green]✅ Response received!")
			}
			a.currentSession.Messages = a.chatHistory
			a.mainLayout.updateConversationView()
			a.mainLayout.updateSidebar()
			go a.saveCurrentChat()
			go func() {
				time.Sleep(3 * time.Second)
				a.app.QueueUpdateDraw(func() {
					a.mainLayout.updateStatus("[green]Ready 🟢")
				})
			}()
		})
	}()
}

func (a *App) clearChat() {
	a.chatHistory = []storage.ChatMessage{}
	a.currentSession.Messages = a.chatHistory
	a.mainLayout.updateConversationView()
	a.mainLayout.updateSidebar()
	a.mainLayout.updateStatus("[blue]🧹 Chat cleared!")
	go func() {
		time.Sleep(2 * time.Second)
		a.app.QueueUpdateDraw(func() {
			a.mainLayout.updateStatus("[green]Ready 🟢")
		})
	}()
}

func (a *App) saveCurrentChat() {
	if len(a.chatHistory) == 0 {
		return
	}
	a.currentSession.Messages = a.chatHistory
	if err := a.storageManager.SaveChat(a.currentSession); err != nil {
		a.app.QueueUpdateDraw(func() {
			a.mainLayout.updateStatus(fmt.Sprintf("[red]❌ Save failed: %v", err))
		})
	}
}
