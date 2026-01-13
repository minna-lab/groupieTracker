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
	// Liste filtrée des artistes
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
	// Recherche + suggestions
	// -------------------------
	recherche := widget.NewEntry()
	recherche.SetPlaceHolder("Rechercher (artiste, membre, lieu, dates)…")

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
		recherche.SetText(suggestionsFiltrees[id].Texte)
	}

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
	// Bouton charger lieux
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
				etat.SetText("Lieux chargés ✅")
			},
		)
	})

	// -------------------------
	// FILTRES DE TRI
	// -------------------------
	selectTrierPar := widget.NewSelect([]string{"Artiste", "Membres", "Lieux", "Premier album", "Date de création"}, nil)
	selectTrierPar.SetSelected("Artiste")

	selectOrdre := widget.NewSelect([]string{"Croissant", "Décroissant"}, nil)
	selectOrdre.SetSelected("Croissant")

	// -------------------------
	// Fonction : applique TOUS les filtres + recherche
	// -------------------------
	var appliquer func()

	appliquer = func() {
		texte := strings.ToLower(strings.TrimSpace(recherche.Text))

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

		// lieux via suggestions
		idsLieux := idsDepuisSuggestions(texte, "lieu")

		// Filtrage artistes par recherche
		artistesFiltres = artistesFiltres[:0]
		for _, a := range artistes {
			if texte != "" {
				nom := strings.ToLower(a.Nom)
				creation := strconv.Itoa(a.AnneeCreation)
				premierAlbum := strings.ToLower(a.PremierAlbum)

				trouve := false
				if strings.Contains(nom, texte) {
					trouve = true
				}

				if !trouve {
					for _, m := range a.Membres {
						if strings.Contains(strings.ToLower(m), texte) {
							trouve = true
							break
						}
					}
				}

				if !trouve && strings.Contains(creation, texte) {
					trouve = true
				}

				if !trouve && strings.Contains(premierAlbum, texte) {
					trouve = true
				}

				if !trouve && idsLieux[a.ID] {
					trouve = true
				}

				if trouve {
					artistesFiltres = append(artistesFiltres, a)
				}
			} else {
				artistesFiltres = append(artistesFiltres, a)
			}
		}

		// Tri (bubble sort)
		trierPar := selectTrierPar.Selected
		ordre := selectOrdre.Selected

		for i := 0; i < len(artistesFiltres)-1; i++ {
			for j := 0; j < len(artistesFiltres)-i-1; j++ {
				echange := false

				switch trierPar {
				case "Artiste":
					a1, a2 := strings.ToLower(artistesFiltres[j].Nom), strings.ToLower(artistesFiltres[j+1].Nom)
					echange = (ordre == "Croissant" && a1 > a2) || (ordre == "Décroissant" && a1 < a2)
				case "Membres":
					n1, n2 := len(artistesFiltres[j].Membres), len(artistesFiltres[j+1].Membres)
					echange = (ordre == "Croissant" && n1 > n2) || (ordre == "Décroissant" && n1 < n2)
				case "Lieux":
					l1 := ""
					l2 := ""
					for _, s := range suggestions {
						if s.Type == "lieu" && s.ID == artistesFiltres[j].ID {
							l1 = strings.ToLower(s.Texte)
							break
						}
					}
					for _, s := range suggestions {
						if s.Type == "lieu" && s.ID == artistesFiltres[j+1].ID {
							l2 = strings.ToLower(s.Texte)
							break
						}
					}
					echange = (ordre == "Croissant" && l1 > l2) || (ordre == "Décroissant" && l1 < l2)
				case "Premier album":
					a1, a2 := strings.ToLower(artistesFiltres[j].PremierAlbum), strings.ToLower(artistesFiltres[j+1].PremierAlbum)
					echange = (ordre == "Croissant" && a1 > a2) || (ordre == "Décroissant" && a1 < a2)
				case "Date de création":
					d1, d2 := artistesFiltres[j].AnneeCreation, artistesFiltres[j+1].AnneeCreation
					echange = (ordre == "Croissant" && d1 > d2) || (ordre == "Décroissant" && d1 < d2)
				}

				if echange {
					artistesFiltres[j], artistesFiltres[j+1] = artistesFiltres[j+1], artistesFiltres[j]
				}
			}
		}

		listeArtistes.Refresh()
	}

	// Branchements events
	recherche.OnChanged = func(string) { appliquer() }
	selectTrierPar.OnChanged = func(string) { appliquer() }
	selectOrdre.OnChanged = func(string) { appliquer() }

	// -------------------------
	// Layout
	// -------------------------
	titre := widget.NewLabelWithStyle("Artistes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	filtresTri := container.NewGridWithColumns(2,
		widget.NewLabel("Trier par :"),
		selectTrierPar,
		widget.NewLabel("Ordre :"),
		selectOrdre,
	)

	haut := container.NewVBox(
		titre,
		recherche,
		listeSuggestions,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Tri", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		filtresTri,
		widget.NewSeparator(),
		btnChargerLieux,
		etat,
	)

	// état initial : affiche tout
	appliquer()

	return container.NewBorder(haut, nil, nil, nil, listeArtistes)
}
