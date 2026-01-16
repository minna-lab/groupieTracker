package interfacegraphique

import (
	"fmt"
	"sort"
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
		parties := strings.FieldsFunc(texte, func(r rune) bool {
			return r == '-' || r == '/' || r == '.'
		})

		for _, p := range parties {
			if len(p) == 4 {
				if y, err := strconv.Atoi(p); err == nil {
					return y
				}
			}
		}

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
	// Chargement lieux : label + progressbar (sans changer de vue)
	// -------------------------
	etat := widget.NewLabel("")
	barre := widget.NewProgressBar()
	barre.Hide()

	var btnChargerLieux *widget.Button
	btnChargerLieux = widget.NewButton("Charger les lieux (recherche avancée)", func() {
		if onChargerLieux == nil {
			return
		}

		btnChargerLieux.Disable()
		etat.SetText("Chargement des lieux…")
		barre.SetValue(0)
		barre.Show()

		onChargerLieux(
			func(fait, total int) {
				etat.SetText(fmt.Sprintf("Indexation : %d/%d", fait, total))
				if total > 0 {
					barre.SetValue(float64(fait) / float64(total))
				}
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
	// TRI
	// -------------------------
	selectTrierPar := widget.NewSelect([]string{"Artiste", "Membres", "Lieux", "Premier album", "Date de création"}, nil)
	selectTrierPar.SetSelected("Artiste")

	selectOrdre := widget.NewSelect([]string{"Croissant", "Décroissant"}, nil)
	selectOrdre.SetSelected("Croissant")

	// -------------------------
	// FILTRES (range + checkbox)
	// -------------------------
	creationMin := widget.NewEntry()
	creationMin.SetPlaceHolder("Création min (ex: 1990)")
	creationMax := widget.NewEntry()
	creationMax.SetPlaceHolder("Création max (ex: 2015)")

	albumMin := widget.NewEntry()
	albumMin.SetPlaceHolder("Album min (année)")
	albumMax := widget.NewEntry()
	albumMax.SetPlaceHolder("Album max (année)")

	cb1 := widget.NewCheck("1", nil)
	cb2 := widget.NewCheck("2", nil)
	cb3 := widget.NewCheck("3", nil)
	cb4plus := widget.NewCheck("4+", nil)

	cbLieuxCharges := widget.NewCheck("Uniquement artistes avec lieux chargés", nil)

	// -------------------------
	// Fonction : applique TOUS les filtres + recherche + tri
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

		// Parse ranges (si vide => pas de filtre)
		toInt := func(s string) int {
			s = strings.TrimSpace(s)
			if s == "" {
				return 0
			}
			v, err := strconv.Atoi(s)
			if err != nil {
				return -1
			}
			return v
		}

		cMin := toInt(creationMin.Text)
		cMax := toInt(creationMax.Text)
		aMin := toInt(albumMin.Text)
		aMax := toInt(albumMax.Text)

		filtreMembresActif := cb1.Checked || cb2.Checked || cb3.Checked || cb4plus.Checked
		idsLieux := idsDepuisSuggestions(texte, "lieu")

		// filtrage
		artistesFiltres = artistesFiltres[:0]
		for _, a := range artistes {

			// filtre création
			if cMin > 0 && a.AnneeCreation < cMin {
				continue
			}
			if cMax > 0 && a.AnneeCreation > cMax {
				continue
			}

			// filtre album (année)
			anneeAlbum := extraireAnnee(a.PremierAlbum)
			if aMin > 0 && anneeAlbum > 0 && anneeAlbum < aMin {
				continue
			}
			if aMax > 0 && anneeAlbum > 0 && anneeAlbum > aMax {
				continue
			}

			// filtre membres
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

			// filtre "lieux chargés" (simple : si on a au moins une suggestion lieu pour cet artiste)
			if cbLieuxCharges.Checked {
				trouveLieu := false
				for _, s := range suggestions {
					if s.Type == "lieu" && s.ID == a.ID {
						trouveLieu = true
						break
					}
				}
				if !trouveLieu {
					continue
				}
			}

			// recherche texte
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
				if !trouve {
					continue
				}
			}

			artistesFiltres = append(artistesFiltres, a)
		}

		// tri
		trierPar := selectTrierPar.Selected
		ordre := selectOrdre.Selected

		// utilitaire : compare string/int selon ordre
		cmpStr := func(a, b string) bool {
			a, b = strings.ToLower(a), strings.ToLower(b)
			if ordre == "Croissant" {
				return a < b
			}
			return a > b
		}
		cmpInt := func(a, b int) bool {
			if ordre == "Croissant" {
				return a < b
			}
			return a > b
		}

		// récupère un "premier lieu" pour tri (simple)
		premierLieu := func(id int) string {
			for _, s := range suggestions {
				if s.Type == "lieu" && s.ID == id {
					return s.Texte
				}
			}
			return ""
		}

		sort.SliceStable(artistesFiltres, func(i, j int) bool {
			a1, a2 := artistesFiltres[i], artistesFiltres[j]
			switch trierPar {
			case "Artiste":
				return cmpStr(a1.Nom, a2.Nom)
			case "Membres":
				return cmpInt(len(a1.Membres), len(a2.Membres))
			case "Lieux":
				return cmpStr(premierLieu(a1.ID), premierLieu(a2.ID))
			case "Premier album":
				return cmpStr(a1.PremierAlbum, a2.PremierAlbum)
			case "Date de création":
				return cmpInt(a1.AnneeCreation, a2.AnneeCreation)
			default:
				return cmpStr(a1.Nom, a2.Nom)
			}
		})

		listeArtistes.Refresh()
	}

	// -------------------------
	// Events
	// -------------------------
	recherche.OnChanged = func(string) { appliquer() }
	selectTrierPar.OnChanged = func(string) { appliquer() }
	selectOrdre.OnChanged = func(string) { appliquer() }

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

	filtresTri := container.NewGridWithColumns(2,
		widget.NewLabel("Trier par :"), selectTrierPar,
		widget.NewLabel("Ordre :"), selectOrdre,
	)

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
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Tri", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		filtresTri,
		widget.NewSeparator(),
		btnChargerLieux,
		barre,
		etat,
	)

	appliquer()

	return container.NewBorder(haut, nil, nil, nil, listeArtistes)
}
