package ui

import (
	"github.com/rivo/tview"

	"github.com/Rohan-Shah-312003/tui-gpt/internal/groq"
)

func StartApp() {
	app := tview.NewApplication()
	input := tview.NewInputField().
		SetLabel("Prompt: ").
		SetFieldWidth(0)

	output := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetChangedFunc(func() { app.Draw() }).
		SetScrollable(true)

	form := tview.NewForm().
		AddFormItem(input).
		AddButton("Send", func() {
			prompt := input.GetText()
			if prompt == "" {
				output.SetText("[red]Empty Prompt!")
				return
			}

			output.SetText("[yellow]Loading...")
			go func() {
				reply, err := groq.SendPrompt(prompt)
				if err != nil {
					output.SetText("[red]Error:" + err.Error())
					return
				}
				output.SetText("[green]Response: \n\n" + reply)
			}()
		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(form, 5, 1, true).
		AddItem(output, 0, 4, false)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
