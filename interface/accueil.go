package interfacegraphique

import (
	"strings"

	"groupie-tracker/modele"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func VueAccueil(artistes []modele.Artiste, onSelection func(modele.Artiste)) fyne.CanvasObject {
	// Liste filtrée (au début = tout)
	artistesFiltres := make([]modele.Artiste, len(artistes))
	copy(artistesFiltres, artistes)

	// Composant liste
	liste := widget.NewList(
		func() int { return len(artistesFiltres) },
		func() fyne.CanvasObject { return widget.NewLabel("...") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(artistesFiltres[i].Nom)
		},
	)

	liste.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(artistesFiltres) {
			onSelection(artistesFiltres[id])
		}
	}

	// Champ recherche
	recherche := widget.NewEntry()
	recherche.SetPlaceHolder("Rechercher (nom ou membre)…")

	// Fonction de filtrage
	filtrer := func(texte string) {
		texte = strings.ToLower(strings.TrimSpace(texte))

		artistesFiltres = artistesFiltres[:0] // on vide sans réallouer

		if texte == "" {
			artistesFiltres = make([]modele.Artiste, len(artistes))
			copy(artistesFiltres, artistes)
			liste.Refresh()
			return
		}

		for _, a := range artistes {
			// match sur nom
			if strings.Contains(strings.ToLower(a.Nom), texte) {
				artistesFiltres = append(artistesFiltres, a)
				continue
			}
			// match sur membres
			for _, m := range a.Membres {
				if strings.Contains(strings.ToLower(m), texte) {
					artistesFiltres = append(artistesFiltres, a)
					break
				}
			}
		}

		liste.Refresh()
	}

	// Filtrage en temps réel
	recherche.OnChanged = filtrer

	titre := widget.NewLabelWithStyle("Artistes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	return container.NewBorder(
		container.NewVBox(titre, recherche),
		nil, nil, nil,
		liste,
	)
}
