package interfacegraphique

import (
	"strconv"
	"strings"

	"groupie-tracker/modele"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func VueAccueil(
	artistes []modele.Artiste,
	onSelection func(modele.Artiste),
	suggestions []modele.Suggestion,
	onChargerLieux func(progress func(fait, total int), fin func(err error)),
) fyne.CanvasObject {

	// -------------------------
	// 1) Liste filtrée des artistes
	// -------------------------
	artistesFiltres := make([]modele.Artiste, len(artistes))
	copy(artistesFiltres, artistes)

	listeArtistes := widget.NewList(
		func() int { return len(artistesFiltres) },
		func() fyne.CanvasObject { return widget.NewLabel("...") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(artistesFiltres[i].Nom)
		},
	)

	listeArtistes.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(artistesFiltres) {
			onSelection(artistesFiltres[id])
		}
	}

	// -------------------------
	// 2) Champ de recherche
	// -------------------------
	recherche := widget.NewEntry()
	recherche.SetPlaceHolder("Rechercher (artiste, membre, lieu, dates)…")

	// -------------------------
	// 3) Liste de suggestions (max 8)
	// -------------------------
	suggestionsFiltrees := []modele.Suggestion{}

	listeSuggestions := widget.NewList(
		func() int { return len(suggestionsFiltrees) },
		func() fyne.CanvasObject { return widget.NewLabel("...") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			s := suggestionsFiltrees[i]
			o.(*widget.Label).SetText(s.Texte + " — " + s.Type)
		},
	)

	listeSuggestions.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(suggestionsFiltrees) {
			return
		}
		recherche.SetText(suggestionsFiltrees[id].Texte) // déclenche OnChanged
	}

	// Aide : IDs d'artistes qui matchent les suggestions d’un type
	idsDepuisSuggestions := func(texte string, typ string) map[int]bool {
		res := make(map[int]bool)
		for _, s := range suggestions {
			if s.Type != typ {
				continue
			}
			if strings.Contains(strings.ToLower(s.Texte), texte) {
				res[s.ID] = true
			}
		}
		return res
	}

	// -------------------------
	// 4) Etat + bouton chargement lieux (corrigé)
	// -------------------------
	etat := widget.NewLabel("")

	var btnChargerLieux *widget.Button
	btnChargerLieux = widget.NewButton("Charger les lieux (recherche avancée)", func() {
		if onChargerLieux == nil {
			return
		}

		btnChargerLieux.Disable()
		etat.SetText("Chargement des lieux…")

		onChargerLieux(
			func(fait, total int) {
				etat.SetText("Indexation : " + strconv.Itoa(fait) + "/" + strconv.Itoa(total))
			},
			func(err error) {
				btnChargerLieux.Enable()
				if err != nil {
					etat.SetText("Erreur : " + err.Error())
					return
				}
				etat.SetText("Lieux chargés ✅ (recherche par lieu active)")
			},
		)
	})

	// -------------------------
	// 5) Filtrage principal
	// -------------------------
	var filtrer func(string)

	filtrer = func(texte string) {
		texte = strings.ToLower(strings.TrimSpace(texte))

		// Suggestions (max 8)
		suggestionsFiltrees = suggestionsFiltrees[:0]
		if texte != "" {
			for _, s := range suggestions {
				if strings.Contains(strings.ToLower(s.Texte), texte) {
					suggestionsFiltrees = append(suggestionsFiltrees, s)
					if len(suggestionsFiltrees) == 8 {
						break
					}
				}
			}
		}
		listeSuggestions.Refresh()

		// Artistes
		artistesFiltres = artistesFiltres[:0]
		if texte == "" {
			artistesFiltres = make([]modele.Artiste, len(artistes))
			copy(artistesFiltres, artistes)
			listeArtistes.Refresh()
			return
		}

		idsLieux := idsDepuisSuggestions(texte, "lieu")

		for _, a := range artistes {
			nom := strings.ToLower(a.Nom)
			creation := strconv.Itoa(a.AnneeCreation)
			premierAlbum := strings.ToLower(a.PremierAlbum)

			// artiste
			if strings.Contains(nom, texte) {
				artistesFiltres = append(artistesFiltres, a)
				continue
			}

			// membres
			okMembre := false
			for _, m := range a.Membres {
				if strings.Contains(strings.ToLower(m), texte) {
					okMembre = true
					break
				}
			}
			if okMembre {
				artistesFiltres = append(artistesFiltres, a)
				continue
			}

			// année création
			if strings.Contains(creation, texte) {
				artistesFiltres = append(artistesFiltres, a)
				continue
			}

			// premier album
			if strings.Contains(premierAlbum, texte) {
				artistesFiltres = append(artistesFiltres, a)
				continue
			}

			// lieux (via suggestions)
			if idsLieux[a.ID] {
				artistesFiltres = append(artistesFiltres, a)
				continue
			}
		}

		listeArtistes.Refresh()
	}

	recherche.OnChanged = filtrer

	// -------------------------
	// 6) Layout
	// -------------------------
	titre := widget.NewLabelWithStyle("Artistes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	haut := container.NewVBox(
		titre,
		recherche,
		listeSuggestions,
		btnChargerLieux,
		etat,
	)

	return container.NewBorder(haut, nil, nil, nil, listeArtistes)
}
