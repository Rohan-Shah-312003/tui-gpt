package ui

import (
	"fmt"

	"github.com/Rohan-Shah-312003/tui-gpt/internal/groq"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ModelListModal struct {
	app       *App
	modelList *tview.List
}

func NewModelListModal(app *App) *ModelListModal {
	return &ModelListModal{
		app:       app,
		modelList: tview.NewList(),
	}
}

func (mlm *ModelListModal) Create() *tview.Flex {
	mlm.modelList.ShowSecondaryText(true).
		SetHighlightFullLine(true).
		SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
			mlm.selectModel(index)
		})
	mlm.modelList.SetBorder(true).SetTitle(" AI Models ").SetBorderColor(tcell.ColorDarkCyan)

	instructions := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]🤖 Model Selection\n\n[white]• Use ↑/↓ to navigate\n• Press Enter to select model\n• Press Escape to close").
		SetTextAlign(tview.AlignLeft)
	instructions.SetBorder(true).SetTitle(" Instructions ").SetBorderColor(tcell.ColorGreen)

	modelButtonFlex := mlm.createButtonFlex()

	modelListLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(instructions, 6, 1, false).
		AddItem(mlm.modelList, 0, 1, true).
		AddItem(modelButtonFlex, 3, 1, false)

	mlm.setupInputCapture()

	return modelListLayout
}

func (mlm *ModelListModal) createButtonFlex() *tview.Flex {
	modelButtonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	selectButton := tview.NewButton("✅ Select").SetSelectedFunc(func() {
		index := mlm.modelList.GetCurrentItem()
		if index >= 0 {
			mlm.selectModel(index)
		}
	})

	closeButton := tview.NewButton("❌ Close").SetSelectedFunc(func() {
		mlm.Hide()
	})

	modelButtonFlex.AddItem(selectButton, 0, 1, false).AddItem(closeButton, 0, 1, false)

	return modelButtonFlex
}

func (mlm *ModelListModal) setupInputCapture() {
	mlm.modelList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			mlm.Hide()
			return nil
		}
		return event
	})
}

func (mlm *ModelListModal) Show() {
	models := groq.GetAvailableModels()
	currentModel := groq.GetCurrentModel()
	mlm.modelList.Clear()

	modelKeys := []string{
		"llama3-70b-8192",
		"llama3-8b-8192",
		"mixtral-8x7b-32768",
		"gemma-7b-it",
		"llama3-groq-70b-8192-tool-use-preview",
		"llama3-groq-8b-8192-tool-use-preview",
	}

	for _, key := range modelKeys {
		if name, exists := models[key]; exists {
			mainText := name
			if key == currentModel {
				mainText = "✅ " + name + " (Current)"
			}
			secondaryText := key
			mlm.modelList.AddItem(mainText, secondaryText, 0, nil)
		}
	}

	mlm.app.pages.ShowPage("modellist")
	mlm.app.isShowingModelList = true
	mlm.app.app.SetFocus(mlm.modelList)
}

func (mlm *ModelListModal) Hide() {
	mlm.app.pages.HidePage("modellist")
	mlm.app.isShowingModelList = false
}

func (mlm *ModelListModal) selectModel(index int) {
	modelKeys := []string{
		"llama3-70b-8192",
		"llama3-8b-8192",
		"mixtral-8x7b-32768",
		"gemma-7b-it",
		"llama3-groq-70b-8192-tool-use-preview",
		"llama3-groq-8b-8192-tool-use-preview",
	}

	if index < 0 || index >= len(modelKeys) {
		mlm.app.mainLayout.updateStatus("[red]❌ Invalid model selection")
		return
	}

	selectedModel := modelKeys[index]
	models := groq.GetAvailableModels()

	if err := groq.SetModel(selectedModel); err != nil {
		mlm.app.mainLayout.updateStatus(fmt.Sprintf("[red]❌ Failed to set model: %v", err))
		return
	}

	mlm.app.mainLayout.updateStatus(fmt.Sprintf("[green]🤖 Model changed to: %s", models[selectedModel]))
	mlm.app.mainLayout.updateSidebar()

	mlm.Hide()
	mlm.app.app.SetFocus(mlm.app.mainLayout.inputField)
}
