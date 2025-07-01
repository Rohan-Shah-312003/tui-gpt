package ui

import (
	"github.com/gdamore/tcell/v2"
)

func (a *App) setupKeyBindings() {
	a.mainLayout.inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			a.sendMessage()
			return nil
		case tcell.KeyCtrlU:
			a.mainLayout.inputField.SetText("")
			return nil
		}
		return event
	})

	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlH:
			if !a.isShowingChatList && !a.isShowingModelList {
				a.toggleHelp()
			}
			return nil
		case tcell.KeyCtrlL:
			if !a.isShowingChatList && !a.isShowingModelList {
				a.clearChat()
			}
			return nil
		case tcell.KeyCtrlN:
			if !a.isShowingChatList && !a.isShowingModelList {
				a.newChat()
			}
			return nil
		case tcell.KeyCtrlS:
			if !a.isShowingChatList && !a.isShowingModelList {
				a.saveCurrentChat()
				a.mainLayout.updateStatus("[green]💾 Chat saved!")
			}
			return nil
		case tcell.KeyCtrlO:
			if !a.isShowingChatList && !a.isShowingModelList {
				a.chatListModal.Show()
			}
			return nil
		case tcell.KeyCtrlUnderscore:
			if !a.isShowingChatList && !a.isShowingModelList {
				a.modelListModal.Show()
			}
			return nil
		}
		return event
	})
}

func (a *App) toggleHelp() {
	if a.pages.HasPage("help") {
		name, _ := a.pages.GetFrontPage()
		if name == "help" {
			a.pages.HidePage("help")
		} else {
			a.pages.ShowPage("help")
		}
	}
}

func (a *App) newChat() {
	a.saveCurrentChat()
	a.startNewChat()
	a.mainLayout.updateConversationView()
	a.mainLayout.updateSidebar()
	a.mainLayout.updateStatus("[blue]🆕 Started new chat!")
}
