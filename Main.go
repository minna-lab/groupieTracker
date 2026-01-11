package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Groupie Tracker")
	w.SetContent(widget.NewLabel("Fyne fonctionne ✅"))
	w.Resize(fyne.NewSize(420, 240))
	w.ShowAndRun()
}
