package interfacegraphique

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func VueChargement(titre string, message string) fyne.CanvasObject {
	spinner := widget.NewProgressBarInfinite()
	icone := widget.NewIcon(theme.InfoIcon())

	return container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle(titre, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		container.NewCenter(icone),
		widget.NewLabel(message),
		spinner,
	))
}
