package interfacegraphique

import (
	"fmt"
	"sort"

	"groupie-tracker/modele"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func VueDetailsArtiste(
	fenetre fyne.Window,
	artiste modele.Artiste,
	relation modele.Relation,
	retour func(),
) fyne.CanvasObject {

	// Trier les lieux pour un affichage propre
	lieux := make([]string, 0, len(relation.DatesParLieu))
	for lieu := range relation.DatesParLieu {
		lieux = append(lieux, lieu)
	}
	sort.Strings(lieux)

	// Construire le texte des concerts
	texteConcerts := ""
	for _, lieu := range lieux {
		texteConcerts += fmt.Sprintf("📍 %s\n", lieu)
		for _, date := range relation.DatesParLieu[lieu] {
			texteConcerts += fmt.Sprintf("   - %s\n", date)
		}
		texteConcerts += "\n"
	}

	btnRetour := widget.NewButton("← Retour", retour)

	titre := widget.NewLabelWithStyle(
		artiste.Nom,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	infos := widget.NewLabel(
		fmt.Sprintf("Création : %d | Premier album : %s",
			artiste.AnneeCreation,
			artiste.PremierAlbum,
		),
	)

	membres := widget.NewLabel("Membres : " + fmt.Sprint(artiste.Membres))

	concerts := widget.NewMultiLineEntry()
	concerts.SetText(texteConcerts)
	concerts.Disable()

	return container.NewBorder(
		container.NewVBox(btnRetour, titre, infos, membres),
		nil, nil, nil,
		container.NewScroll(concerts),
	)
}
