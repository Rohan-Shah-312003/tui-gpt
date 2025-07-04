package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ChatListModal struct {
	app      *App
	chatList *tview.List
}

func NewChatListModal(app *App) *ChatListModal {
	return &ChatListModal{
		app:      app,
		chatList: tview.NewList(),
	}
}

func (clm *ChatListModal) Create() *tview.Flex {
	clm.chatList.ShowSecondaryText(true).
		SetHighlightFullLine(true).
		SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			clm.loadChatFromList(index)
		})
	clm.chatList.SetBorder(true).SetTitle(" Chat History ").SetBorderColor(tcell.ColorDarkCyan)

	instructions := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]📚 Chat History\n\n[white]• Use ↑/↓ to navigate\n• Press Enter to load chat\n• Press 'd' to delete selected\n• Press Escape to close").
		SetTextAlign(tview.AlignLeft)
	instructions.SetBorder(true).SetTitle(" Instructions ").SetBorderColor(tcell.ColorGreen)

	chatButtonFlex := clm.createButtonFlex()

	chatListLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(instructions, 8, 1, false).
		AddItem(clm.chatList, 0, 1, true).
		AddItem(chatButtonFlex, 3, 1, false)

	clm.setupInputCapture()

	return chatListLayout
}

func (clm *ChatListModal) createButtonFlex() *tview.Flex {
	chatButtonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	loadButton := tview.NewButton("📂Load").SetSelectedFunc(func() {
		index := clm.chatList.GetCurrentItem()
		if index >= 0 {
			clm.loadChatFromList(index)
		}
	}).SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack))

	deleteButton := tview.NewButton("️🗑️Delete").SetSelectedFunc(func() {
		index := clm.chatList.GetCurrentItem()
		if index >= 0 {
			clm.deleteChatFromList(index)
		}
	}).SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorDarkRed).Foreground(tcell.ColorBlack))

	closeButton := tview.NewButton("❌Close").SetSelectedFunc(func() {
		clm.Hide()
	}).SetLabelColor(tcell.ColorBlack).SetStyle(tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorBlack))

	chatButtonFlex.AddItem(loadButton, 0, 1, false).
		AddItem(deleteButton, 0, 1, false).
		AddItem(closeButton, 0, 1, false)

	return chatButtonFlex
}

func (clm *ChatListModal) setupInputCapture() {
	clm.chatList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			clm.Hide()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'd', 'D':
				index := clm.chatList.GetCurrentItem()
				if index >= 0 {
					clm.deleteChatFromList(index)
				}
				return nil
			}
		}
		return event
	})
}

func (clm *ChatListModal) Show() {
	summaries, err := clm.app.storageManager.GetChatSummaries()
	if err != nil {
		clm.app.mainLayout.updateStatus(fmt.Sprintf("[red]❌ Failed to load chats: %v", err))
		return
	}

	clm.chatList.Clear()
	if len(summaries) == 0 {
		clm.chatList.AddItem("No saved chats", "Start a conversation to create your first chat!", 0, nil)
	} else {
		for _, summary := range summaries {
			mainText := summary.Title
			secondaryText := fmt.Sprintf("%d messages • Updated: %s",
				summary.MessageCount,
				summary.UpdatedAt.Format("Jan 2, 15:04"))
			clm.chatList.AddItem(mainText, secondaryText, 0, nil)
		}
	}

	clm.app.pages.ShowPage("chatlist")
	clm.app.isShowingChatList = true
	clm.app.app.SetFocus(clm.chatList)
}

func (clm *ChatListModal) Hide() {
	clm.app.pages.HidePage("chatlist")
	clm.app.isShowingChatList = false
}

func (clm *ChatListModal) loadChatFromList(index int) {
	summaries, err := clm.app.storageManager.GetChatSummaries()
	if err != nil || index >= len(summaries) {
		clm.app.mainLayout.updateStatus("[red]❌ Failed to load chat")
		return
	}

	clm.app.saveCurrentChat()
	session, err := clm.app.storageManager.LoadChat(summaries[index].ID)
	if err != nil {
		clm.app.mainLayout.updateStatus(fmt.Sprintf("[red]❌ Failed to load chat: %v", err))
		return
	}

	clm.app.SetCurrentSession(session)
	clm.app.SetChatHistory(session.Messages)
	clm.app.mainLayout.updateConversationView()
	clm.app.mainLayout.updateSidebar()
	clm.app.mainLayout.updateStatus("[green]📂 Chat loaded successfully!")
	clm.Hide()
	clm.app.app.SetFocus(clm.app.mainLayout.inputField)
}

func (clm *ChatListModal) deleteChatFromList(index int) {
	summaries, err := clm.app.storageManager.GetChatSummaries()
	if err != nil || index >= len(summaries) {
		clm.app.mainLayout.updateStatus("[red]❌ Failed to delete chat")
		return
	}

	if err := clm.app.storageManager.DeleteChat(summaries[index].ID); err != nil {
		clm.app.mainLayout.updateStatus(fmt.Sprintf("[red]❌ Failed to delete chat: %v", err))
		return
	}

	clm.app.mainLayout.updateStatus("[yellow]🗑️ Chat deleted!")
	clm.Show() // Refresh the list
}
