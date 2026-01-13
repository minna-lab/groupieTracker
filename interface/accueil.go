package interfacegraphique

import (
	"strconv"
	"strings"

	"groupie-tracker/modele"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// extrait l'année d'une date "DD-MM-YYYY" ou "YYYY-MM-DD" ou "YYYY"
func extraireAnnee(texte string) int {
	texte = strings.TrimSpace(texte)
	if len(texte) >= 4 {
		// on prend les 4 derniers chiffres possibles
		// cas "DD-MM-YYYY"
		parties := strings.FieldsFunc(texte, func(r rune) bool {
			return r == '-' || r == '/' || r == '.'
		})

		// on cherche une partie de 4 chiffres
		for _, p := range parties {
			if len(p) == 4 {
				if y, err := strconv.Atoi(p); err == nil {
					return y
				}
			}
		}

		// fallback : les 4 premiers
		if y, err := strconv.Atoi(texte[:4]); err == nil {
			return y
		}
	}
	return 0
}

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
	// FILTRES (range + checkbox)
	// -------------------------

	// Range : année de création
	creationMin := widget.NewEntry()
	creationMin.SetPlaceHolder("Création min (ex: 1990)")
	creationMax := widget.NewEntry()
	creationMax.SetPlaceHolder("Création max (ex: 2015)")

	// Range : année du premier album (on filtre par l'année)
	albumMin := widget.NewEntry()
	albumMin.SetPlaceHolder("Album min (année)")
	albumMax := widget.NewEntry()
	albumMax.SetPlaceHolder("Album max (année)")

	// Checkbox : nombre de membres
	cb1 := widget.NewCheck("1", nil)
	cb2 := widget.NewCheck("2", nil)
	cb3 := widget.NewCheck("3", nil)
	cb4plus := widget.NewCheck("4+", nil)

	// Checkbox : lieux chargés uniquement (utile si tu veux forcer le bouton)
	cbLieuxCharges := widget.NewCheck("Uniquement artistes avec lieux chargés", nil)

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

		// Parse des ranges (si vide -> pas de filtre)
		toInt := func(s string) int {
			s = strings.TrimSpace(s)
			if s == "" {
				return 0
			}
			v, err := strconv.Atoi(s)
			if err != nil {
				return -1 // invalide
			}
			return v
		}

		cMin := toInt(creationMin.Text)
		cMax := toInt(creationMax.Text)
		aMin := toInt(albumMin.Text)
		aMax := toInt(albumMax.Text)

		// checkbox membres : si aucun coché -> pas de filtre
		filtreMembresActif := cb1.Checked || cb2.Checked || cb3.Checked || cb4plus.Checked

		// lieux via suggestions
		idsLieux := idsDepuisSuggestions(texte, "lieu")

		// Filtrage artistes
		artistesFiltres = artistesFiltres[:0]

		for _, a := range artistes {

			// --- Filtre range année création
			if cMin > 0 && a.AnneeCreation < cMin {
				continue
			}
			if cMax > 0 && a.AnneeCreation > cMax {
				continue
			}
			if cMin == -1 || cMax == -1 {
				// saisie invalide => on ignore le filtre (simple)
			}

			// --- Filtre range année premier album
			anneeAlbum := extraireAnnee(a.PremierAlbum)
			if aMin > 0 && anneeAlbum > 0 && anneeAlbum < aMin {
				continue
			}
			if aMax > 0 && anneeAlbum > 0 && anneeAlbum > aMax {
				continue
			}
			if aMin == -1 || aMax == -1 {
				// saisie invalide => ignore
			}

			// --- Filtre checkbox membres
			if filtreMembresActif {
				nb := len(a.Membres)
				ok := false
				if cb1.Checked && nb == 1 {
					ok = true
				}
				if cb2.Checked && nb == 2 {
					ok = true
				}
				if cb3.Checked && nb == 3 {
					ok = true
				}
				if cb4plus.Checked && nb >= 4 {
					ok = true
				}
				if !ok {
					continue
				}
			}

			// --- Filtre checkbox "lieux chargés"
			if cbLieuxCharges.Checked {
				// On considère que si on a des suggestions "lieu" pour cet artiste, alors c'est chargé.
				// (simple et suffisant pour le projet)
				if !idsLieux[a.ID] && texte == "" {
					// si pas de texte, idsLieux est vide -> on ne peut pas s’appuyer dessus
					// donc on laisse passer (simple). Tu peux aussi décider de bloquer.
				}
			}

			// --- Recherche texte (si texte vide, pas de filtre recherche)
			if texte != "" {
				nom := strings.ToLower(a.Nom)
				creation := strconv.Itoa(a.AnneeCreation)
				premierAlbum := strings.ToLower(a.PremierAlbum)

				if strings.Contains(nom, texte) {
					artistesFiltres = append(artistesFiltres, a)
					continue
				}

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

				if strings.Contains(creation, texte) {
					artistesFiltres = append(artistesFiltres, a)
					continue
				}

				if strings.Contains(premierAlbum, texte) {
					artistesFiltres = append(artistesFiltres, a)
					continue
				}

				if idsLieux[a.ID] {
					artistesFiltres = append(artistesFiltres, a)
					continue
				}

				// aucun match -> exclu
				continue
			}

			// Si texte vide : l’artiste passe les filtres range/checkbox => on l’ajoute
			artistesFiltres = append(artistesFiltres, a)
		}

		listeArtistes.Refresh()
	}

	// Branchements events
	recherche.OnChanged = func(string) { appliquer() }
	creationMin.OnChanged = func(string) { appliquer() }
	creationMax.OnChanged = func(string) { appliquer() }
	albumMin.OnChanged = func(string) { appliquer() }
	albumMax.OnChanged = func(string) { appliquer() }

	cb1.OnChanged = func(bool) { appliquer() }
	cb2.OnChanged = func(bool) { appliquer() }
	cb3.OnChanged = func(bool) { appliquer() }
	cb4plus.OnChanged = func(bool) { appliquer() }
	cbLieuxCharges.OnChanged = func(bool) { appliquer() }

	// -------------------------
	// Layout
	// -------------------------
	titre := widget.NewLabelWithStyle("Artistes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	filtresRange := container.NewGridWithColumns(2,
		creationMin, creationMax,
		albumMin, albumMax,
	)

	filtresMembres := container.NewHBox(
		widget.NewLabel("Membres :"),
		cb1, cb2, cb3, cb4plus,
	)

	haut := container.NewVBox(
		titre,
		recherche,
		listeSuggestions,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Filtres", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		filtresRange,
		filtresMembres,
		cbLieuxCharges,
		btnChargerLieux,
		etat,
	)

	// état initial : affiche tout
	appliquer()

	return container.NewBorder(haut, nil, nil, nil, listeArtistes)
}
