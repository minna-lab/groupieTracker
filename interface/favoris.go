package interfacegraphique

import (
	"groupie-tracker/modele"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// VueFavoris affiche la liste des artistes favoris
func VueFavoris(
	onSelection func(modele.Artiste),
	rafraichir func() []modele.Artiste,
) fyne.CanvasObject {

	favoris := rafraichir()

	// Liste des favoris
	listeFavoris := widget.NewList(
		func() int { return len(favoris) },
		func() fyne.CanvasObject { return widget.NewLabel("...") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(favoris) {
				o.(*widget.Label).SetText(favoris[i].Nom)
			}
		},
	)

	listeFavoris.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(favoris) {
			onSelection(favoris[id])
		}
	}

	// Bouton rafraîchir
	btnRafraichir := widget.NewButton("🔄 Rafraîchir", func() {
		favoris = rafraichir()
		listeFavoris.Refresh()
	})

	titre := widget.NewLabelWithStyle(
		"❤️ Mes Favoris",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	// Message si aucun favori
	var contenu fyne.CanvasObject
	if len(favoris) == 0 {
		message := widget.NewLabel("Aucun artiste favori pour le moment.\nCliquez sur ❤️ dans la page d'un artiste pour l'ajouter !")
		contenu = container.NewCenter(message)
	} else {
		contenu = listeFavoris
	}

	haut := container.NewVBox(
		titre,
		btnRafraichir,
	)

	return container.NewBorder(haut, nil, nil, nil, contenu)
}
