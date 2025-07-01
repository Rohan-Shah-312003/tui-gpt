package ui

import (
	"fmt"
	"time"

	"github.com/Rohan-Shah-312003/tui-gpt/internal/storage"
	"github.com/rivo/tview"
)

type App struct {
	app            *tview.Application
	pages          *tview.Pages
	storageManager *storage.Storage
	currentSession *storage.ChatSession
	chatHistory    []storage.ChatMessage

	// UI Components
	mainLayout     *MainLayout
	helpModal      *HelpModal
	chatListModal  *ChatListModal
	modelListModal *ModelListModal

	// State
	isShowingChatList  bool
	isShowingModelList bool
}

func NewApp() *App {
	return &App{
		app:         tview.NewApplication(),
		pages:       tview.NewPages(),
		chatHistory: []storage.ChatMessage{},
	}
}

func (a *App) Start() error {
	a.storageManager = storage.NewStorage()
	if err := a.storageManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize storage: %v", err)
	}

	a.startNewChat()
	a.setupUI()
	a.setupKeyBindings()

	return a.app.SetRoot(a.pages, true).EnableMouse(true).Run()
}

func (a *App) startNewChat() {
	a.currentSession = &storage.ChatSession{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []storage.ChatMessage{},
	}
	a.chatHistory = []storage.ChatMessage{}
}

func (a *App) setupUI() {
	a.mainLayout = NewMainLayout(a)
	a.helpModal = NewHelpModal(a)
	a.chatListModal = NewChatListModal(a)
	a.modelListModal = NewModelListModal(a)

	a.pages.AddPage("main", a.mainLayout.Create(), true, true)
	a.pages.AddPage("help", a.helpModal.Create(), true, false)
	a.pages.AddPage("chatlist", a.chatListModal.Create(), true, false)
	a.pages.AddPage("modellist", a.modelListModal.Create(), true, false)
}

func (a *App) GetApp() *tview.Application                     { return a.app }
func (a *App) GetPages() *tview.Pages                         { return a.pages }
func (a *App) GetStorageManager() *storage.Storage            { return a.storageManager }
func (a *App) GetCurrentSession() *storage.ChatSession        { return a.currentSession }
func (a *App) GetChatHistory() []storage.ChatMessage          { return a.chatHistory }
func (a *App) SetChatHistory(history []storage.ChatMessage)   { a.chatHistory = history }
func (a *App) SetCurrentSession(session *storage.ChatSession) { a.currentSession = session }

// StartApp - Entry point function for backward compatibility
func StartApp() {
	app := NewApp()
	if err := app.Start(); err != nil {
		panic(err)
	}
}
