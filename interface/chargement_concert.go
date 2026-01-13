package interfacegraphique

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func VueChargement(titre string, message string) fyne.CanvasObject {
	barre := widget.NewProgressBarInfinite()
	barre.Start()

	return container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithStyle(titre, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel(message),
			barre,
		),
	)
}
