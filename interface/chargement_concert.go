package interfacegraphique

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func VueChargement(titre string, message string, onRetour func()) fyne.CanvasObject {
	barre := widget.NewProgressBarInfinite()
	barre.Start()

	var boutonRetour fyne.CanvasObject
	if onRetour != nil {
		boutonRetour = widget.NewButton("← Retour", onRetour)
	} else {
		boutonRetour = widget.NewLabel("") // espace vide si pas de retour
	}

	return container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithStyle(titre, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel(message),
			barre,
			boutonRetour,
		),
	)
}
