package interfacegraphique

import (
	"groupie-tracker/modele"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func VueAccueil(artistes []modele.Artiste, onSelection func(modele.Artiste)) fyne.CanvasObject {
	liste := widget.NewList(
		func() int { return len(artistes) },
		func() fyne.CanvasObject { return widget.NewLabel("...") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(artistes[i].Nom)
		},
	)

	liste.OnSelected = func(id widget.ListItemID) {
		onSelection(artistes[id])
	}

	titre := widget.NewLabelWithStyle("Artistes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewBorder(titre, nil, nil, nil, liste)
}
