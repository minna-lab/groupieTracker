package interfacegraphique

import (
	"fmt"
	"image/color"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"groupie-tracker/modele"
	"groupie-tracker/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Theme personnalisé en mode dark
type ThemePerso struct{}

func (ThemePerso) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	colors := map[fyne.ThemeColorName]color.Color{
		theme.ColorNamePrimary: color.NRGBA{34, 139, 230, 255}, theme.ColorNameSelection: color.NRGBA{34, 139, 230, 55},
		theme.ColorNameBackground: color.NRGBA{16, 16, 18, 255}, theme.ColorNameInputBackground: color.NRGBA{28, 28, 33, 255},
		theme.ColorNameForeground: color.NRGBA{235, 235, 235, 255},
	}
	if c, ok := colors[name]; ok {
		return c
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
}
func (ThemePerso) Font(style fyne.TextStyle) fyne.Resource    { return theme.DefaultTheme().Font(style) }
func (ThemePerso) Icon(name fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(name) }
func (ThemePerso) Size(name fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(name) }

func tronquer(texte string, max int) string {
	texte = strings.TrimSpace(texte)
	if max <= 0 || len(texte) <= max {
		return texte
	}
	return texte[:max-1] + "…"
}

// genererCarteGoogleMaps crée une URL OpenStreetMap interactive pour un lieu
func genererCarteGoogleMaps(lieu string) string {
	// URL OpenStreetMap avec le lieu comme marqueur
	// Format: https://www.openstreetmap.org/search?query=lieu
	mapsURL := "https://www.openstreetmap.org/search?query=" + url.QueryEscape(lieu)

	return mapsURL
}

func Carte(contenu fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 14
	ombre := canvas.NewRectangle(color.NRGBA{0, 0, 0, 18})
	ombre.CornerRadius = 14
	return container.NewPadded(container.NewStack(container.NewPadded(ombre), bg, container.NewPadded(contenu)))
}

func TitreSection(texte string) fyne.CanvasObject {
	t := canvas.NewText(texte, theme.Color(theme.ColorNameForeground))
	t.TextStyle, t.TextSize = fyne.TextStyle{Bold: true}, 18
	return t
}

func VueChargement(titre, message string, onRetour func()) fyne.CanvasObject {
	barre := widget.NewProgressBarInfinite()
	barre.Start()
	var retour fyne.CanvasObject = widget.NewLabel("")
	if onRetour != nil {
		retour = widget.NewButton("← Retour", onRetour)
	}
	return container.NewCenter(container.NewVBox(widget.NewLabelWithStyle(titre, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), widget.NewLabel(message), barre, retour))
}

// ============================================================================
// VUE ACCUEIL - LAYOUT 2 COLONNES
// ============================================================================

func VueAccueil(
	artistes []modele.Artiste,
	imagesArtistes map[int]fyne.Resource,
	onSelection func(modele.Artiste),
	suggestions []modele.Suggestion,
) (fyne.CanvasObject, func()) {

	artistesFiltres := make([]modele.Artiste, len(artistes))
	copy(artistesFiltres, artistes)

	// Conteneur pour la grille d'artistes
	grilleArtistes := container.NewVBox()

	// Fonction pour créer une carte artiste
	creerCarteArtiste := func(artiste modele.Artiste) fyne.CanvasObject {
		// Image
		var img *canvas.Image
		if imgResource, ok := imagesArtistes[artiste.ID]; ok {
			img = canvas.NewImageFromResource(imgResource)
		} else {
			img = canvas.NewImageFromResource(theme.AccountIcon())
		}
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(120, 120))

		// Nom de l'artiste (tronqué)
		nomLabel := widget.NewLabel(tronquer(artiste.Nom, 18))
		nomLabel.Alignment = fyne.TextAlignCenter
		nomLabel.TextStyle = fyne.TextStyle{Bold: true}

		// Année
		anneeLabel := widget.NewLabel(fmt.Sprintf("(%d)", artiste.AnneeCreation))
		anneeLabel.Alignment = fyne.TextAlignCenter

		// Contenu de la carte
		contenuCarte := container.NewVBox(
			img,
			nomLabel,
			anneeLabel,
		)

		// Fond sombre pour la carte
		bgCarte := canvas.NewRectangle(color.NRGBA{40, 40, 45, 255})
		bgCarte.CornerRadius = 10

		// Carte avec fond sombre
		carte := container.NewStack(
			bgCarte,
			container.NewPadded(contenuCarte),
		)

		// Bouton transparent pour le clic (sans effet de survol)
		btn := widget.NewButton("", func() {
			onSelection(artiste)
		})
		btn.Importance = widget.LowImportance

		// Stack avec bouton
		return container.NewStack(
			carte,
			btn,
		)
	}

	// Fonction pour rafraîchir la grille
	var rafraichirGrille func()
	rafraichirGrille = func() {
		grilleArtistes.Objects = nil

		// Créer des lignes de 4 artistes
		const artistesParLigne = 4
		for i := 0; i < len(artistesFiltres); i += artistesParLigne {
			fin := i + artistesParLigne
			if fin > len(artistesFiltres) {
				fin = len(artistesFiltres)
			}

			ligne := make([]fyne.CanvasObject, 0, artistesParLigne)
			for j := i; j < fin; j++ {
				ligne = append(ligne, creerCarteArtiste(artistesFiltres[j]))
			}

			grilleArtistes.Add(container.NewGridWithColumns(artistesParLigne, ligne...))
		}

		grilleArtistes.Refresh()
	}

	recherche := widget.NewEntry()
	recherche.SetPlaceHolder("Rechercher (artiste, membre, lieu, dates)…")

	suggestionsFiltrees := []modele.Suggestion{}
	listeSuggestions := widget.NewList(
		func() int { return len(suggestionsFiltrees) },
		func() fyne.CanvasObject {
			return widget.NewLabel("...")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(suggestionsFiltrees) {
				return
			}
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
	// Limiter la hauteur de la liste de suggestions
	listeSuggestions.Resize(fyne.NewSize(0, 150))

	idsDepuisSuggestions := func(texte, typ string) map[int]bool {
		res := make(map[int]bool)
		for _, s := range suggestions {
			if s.Type == typ && strings.Contains(strings.ToLower(s.Texte), texte) {
				res[s.ID] = true
			}
		}
		return res
	}

	check1Membre, check2Membres, check3Membres, check4Membres, check5PlusMembres := widget.NewCheck("1", nil), widget.NewCheck("2", nil), widget.NewCheck("3", nil), widget.NewCheck("4", nil), widget.NewCheck("5+", nil)
	filtreNbMembres := container.NewVBox(widget.NewLabel("Membres :"), container.NewGridWithColumns(5, check1Membre, check2Membres, check3Membres, check4Membres, check5PlusMembres))

	entryAnneeCreationMin, entryAnneeCreationMax := widget.NewEntry(), widget.NewEntry()
	entryAnneeCreationMin.SetPlaceHolder("Min (ex: 1990)")
	entryAnneeCreationMax.SetPlaceHolder("Max (ex: 2020)")
	filtrePlageCreation := container.NewVBox(widget.NewLabel("Année de création :"), container.NewGridWithColumns(2, entryAnneeCreationMin, entryAnneeCreationMax))

	entryLieux := widget.NewEntry()
	entryLieux.SetPlaceHolder("Chercher un lieu...")
	filtreRechercheLieux := container.NewVBox(widget.NewLabel("Lieux de concerts :"), entryLieux)
	selectTrierPar := widget.NewSelect([]string{"Artiste", "Date de premier album"}, nil)
	selectTrierPar.SetSelected("Artiste")
	filtreTri := container.NewVBox(widget.NewLabel("Trier par :"), selectTrierPar)

	var appliquer func()

	btnReinitialiser := widget.NewButton("🔄 Réinitialiser les filtres", func() {
		check1Membre.SetChecked(false)
		check2Membres.SetChecked(false)
		check3Membres.SetChecked(false)
		check4Membres.SetChecked(false)
		check5PlusMembres.SetChecked(false)
		entryAnneeCreationMin.SetText("")
		entryAnneeCreationMax.SetText("")
		entryLieux.SetText("")
		recherche.SetText("")
		selectTrierPar.SetSelected("Artiste")
	})

	tousLesFiltres := container.NewVBox(filtreTri, widget.NewSeparator(), filtreNbMembres, widget.NewSeparator(), filtrePlageCreation, widget.NewSeparator(), filtreRechercheLieux, widget.NewSeparator(), btnReinitialiser)

	appliquer = func() {
		texte := strings.ToLower(strings.TrimSpace(recherche.Text))

		// Filtrer suggestions
		suggestionsFiltrees = suggestionsFiltrees[:0]
		if texte != "" {
			for _, s := range suggestions {
				if strings.Contains(strings.ToLower(s.Texte), texte) {
					suggestionsFiltrees = append(suggestionsFiltrees, s)
					if len(suggestionsFiltrees) == 5 {
						break
					}
				}
			}
		}
		listeSuggestions.Refresh()

		anneeCreationMin, anneeCreationMax := 0, 9999
		if entryAnneeCreationMin.Text != "" {
			if val, err := strconv.Atoi(strings.TrimSpace(entryAnneeCreationMin.Text)); err == nil {
				anneeCreationMin = val
			}
		}
		if entryAnneeCreationMax.Text != "" {
			if val, err := strconv.Atoi(strings.TrimSpace(entryAnneeCreationMax.Text)); err == nil {
				anneeCreationMax = val
			}
		}

		lieutexte := strings.ToLower(strings.TrimSpace(entryLieux.Text))
		idsLieuxFiltre := make(map[int]bool)
		if lieutexte != "" {
			for _, s := range suggestions {
				if s.Type == "lieu" && strings.Contains(strings.ToLower(s.Texte), lieutexte) {
					idsLieuxFiltre[s.ID] = true
				}
			}
		}

		membresVoulus := make(map[int]bool)
		if check1Membre.Checked {
			membresVoulus[1] = true
		}
		if check2Membres.Checked {
			membresVoulus[2] = true
		}
		if check3Membres.Checked {
			membresVoulus[3] = true
		}
		if check4Membres.Checked {
			membresVoulus[4] = true
		}
		if check5PlusMembres.Checked {
			membresVoulus[5] = true
		}
		accepterTousMembres := len(membresVoulus) == 0

		// Filtrer artistes
		idsLieux := idsDepuisSuggestions(texte, "lieu")
		artistesFiltres = artistesFiltres[:0]

		for _, a := range artistes {
			if !accepterTousMembres {
				nbMembres := len(a.Membres)
				if !((nbMembres >= 5 && membresVoulus[5]) || membresVoulus[nbMembres]) {
					continue
				}
			}

			if a.AnneeCreation < anneeCreationMin || a.AnneeCreation > anneeCreationMax {
				continue
			}
			if lieutexte != "" && !idsLieuxFiltre[a.ID] {
				continue
			}

			// --- Filtre texte recherche générale ---
			if texte == "" {
				artistesFiltres = append(artistesFiltres, a)
				continue
			}
			trouve := strings.Contains(strings.ToLower(a.Nom), texte) ||
				strings.Contains(strconv.Itoa(a.AnneeCreation), texte) ||
				strings.Contains(strings.ToLower(a.PremierAlbum), texte) ||
				idsLieux[a.ID]
			if !trouve {
				for _, m := range a.Membres {
					if strings.Contains(strings.ToLower(m), texte) {
						trouve = true
						break
					}
				}
			}
			if trouve {
				artistesFiltres = append(artistesFiltres, a)
			}
		}

		// Tri bubble sort
		trierPar := selectTrierPar.Selected
		for i := 0; i < len(artistesFiltres)-1; i++ {
			for j := 0; j < len(artistesFiltres)-i-1; j++ {
				echange := false
				switch trierPar {
				case "Artiste":
					a1, a2 := strings.ToLower(artistesFiltres[j].Nom), strings.ToLower(artistesFiltres[j+1].Nom)
					echange = a1 > a2
				case "Date de premier album":
					a1, a2 := strings.ToLower(artistesFiltres[j].PremierAlbum), strings.ToLower(artistesFiltres[j+1].PremierAlbum)
					echange = a1 > a2
				}
				if echange {
					artistesFiltres[j], artistesFiltres[j+1] = artistesFiltres[j+1], artistesFiltres[j]
				}
			}
		}
		rafraichirGrille()
	}

	recherche.OnChanged, selectTrierPar.OnChanged = func(string) { appliquer() }, func(string) { appliquer() }
	check1Membre.OnChanged, check2Membres.OnChanged, check3Membres.OnChanged, check4Membres.OnChanged, check5PlusMembres.OnChanged = func(bool) { appliquer() }, func(bool) { appliquer() }, func(bool) { appliquer() }, func(bool) { appliquer() }, func(bool) { appliquer() }
	entryAnneeCreationMin.OnChanged, entryAnneeCreationMax.OnChanged, entryLieux.OnChanged = func(string) { appliquer() }, func(string) { appliquer() }, func(string) { appliquer() }

	colonneGauche := container.NewVBox(widget.NewLabelWithStyle("🔍 Recherche", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), recherche, listeSuggestions, widget.NewSeparator(), widget.NewLabelWithStyle("⚙️ Filtres", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), tousLesFiltres)

	colonneGaucheScroll := container.NewVScroll(colonneGauche)
	colonneGaucheScroll.SetMinSize(fyne.NewSize(300, 0))
	colonneGaucheAvecBg := container.NewStack(canvas.NewRectangle(color.NRGBA{25, 25, 30, 255}), colonneGaucheScroll)
	titreGroupieTracker := canvas.NewText("🎸 GROUPIE TRACKER", theme.Color(theme.ColorNameForeground))
	titreGroupieTracker.TextStyle, titreGroupieTracker.TextSize, titreGroupieTracker.Alignment = fyne.TextStyle{Bold: true}, 32, fyne.TextAlignCenter
	enTeteGlobal := container.NewCenter(titreGroupieTracker)
	titreEcouter := canvas.NewText("🎵 ÉCOUTER", theme.Color(theme.ColorNameForeground))
	titreEcouter.TextStyle, titreEcouter.TextSize = fyne.TextStyle{Bold: true}, 24
	sousTitre := canvas.NewText("Pour vous", theme.Color(theme.ColorNameForeground))
	sousTitre.TextSize = 16
	enTeteDroite := container.NewVBox(titreEcouter, sousTitre)

	appliquer()
	colonneDroite := container.NewBorder(enTeteDroite, nil, nil, nil, container.NewVScroll(grilleArtistes))
	contenuPrincipal := container.NewBorder(nil, nil, colonneGaucheAvecBg, nil, colonneDroite)
	return container.NewBorder(enTeteGlobal, nil, nil, nil, contenuPrincipal), rafraichirGrille
}

// Vue détails d'un artiste
func VueDetailsArtiste(
	fenetre fyne.Window,
	artiste modele.Artiste,
	relation modele.Relation,
	retour func(),
) fyne.CanvasObject {

	btnRetour := widget.NewButton("← Retour", retour)
	gestionnaireFavoris := service.ObtenirGestionnaireFavoris()
	estFavori := gestionnaireFavoris.EstFavori(artiste.ID)
	btnFavori := widget.NewButton("", nil)
	majTexteFavori := func() {
		if estFavori {
			btnFavori.SetText("💔 Retirer des favoris")
		} else {
			btnFavori.SetText("❤️ Ajouter aux favoris")
		}
	}
	majTexteFavori()
	btnFavori.OnTapped = func() { estFavori = gestionnaireFavoris.Basculer(artiste); majTexteFavori() }
	btnSpotify := widget.NewButton("🎵 Spotify", func() {
		if u, err := url.Parse("spotify:search:" + url.QueryEscape(artiste.Nom)); err == nil {
			fyne.CurrentApp().OpenURL(u)
		} else {
			dialog.ShowError(err, fenetre)
		}
	})
	barreBoutons := container.NewHBox(btnRetour, btnFavori, btnSpotify)

	card := func(titre string, contenu fyne.CanvasObject) fyne.CanvasObject {
		return widget.NewCard(titre, "", contenu)
	}
	var imageArtiste fyne.CanvasObject
	if artiste.Image != "" {
		img := canvas.NewImageFromURI(storage.NewURI(artiste.Image))
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(350, 350))
		imageArtiste = container.NewCenter(img)
	} else {
		imageArtiste = widget.NewLabel("Image non disponible")
	}
	blocArtiste := card("Artiste", container.NewVBox(imageArtiste, widget.NewLabelWithStyle(artiste.Nom, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})))
	blocInfos := card("Informations", container.NewVBox(widget.NewLabel(fmt.Sprintf("📅 Année de création : %d", artiste.AnneeCreation)), widget.NewLabel(fmt.Sprintf("💿 Premier album : %s", artiste.PremierAlbum))))
	lblMembres := widget.NewLabel(strings.Join(artiste.Membres, ", "))
	lblMembres.Wrapping = fyne.TextWrapWord
	blocMembres := card("Membres", lblMembres)

	// Bloc Concerts (cartes par lieu)
	// -------------------------
	// Trier les lieux
	lieux := make([]string, 0, len(relation.DatesParLieu))
	for lieu := range relation.DatesParLieu {
		lieux = append(lieux, lieu)
	}
	sort.Strings(lieux)

	// Création des "cartes concert"
	cartes := make([]fyne.CanvasObject, 0, len(lieux))

	const maxTitreLieu = 22     // limite pour éviter débordement
	const maxDatesAffichees = 6 // limite pour éviter une carte gigantesque

	for _, lieu := range lieux {
		dates := relation.DatesParLieu[lieu]

		// Créer un texte coloré pour le lieu
		texteLieu := canvas.NewText("📍 "+tronquer(lieu, maxTitreLieu), color.NRGBA{34, 139, 230, 255})
		texteLieu.TextStyle = fyne.TextStyle{Bold: true}
		texteLieu.TextSize = 14

		// Bouton pour le lieu qui ouvre OpenStreetMap interactive
		btnLieu := widget.NewButton("", func(lieuCopie string) func() {
			return func() {
				// Générer l'URL OpenStreetMap interactive
				mapsURL := genererCarteGoogleMaps(lieuCopie)

				// Ouvrir l'URL dans le navigateur par défaut
				if u, err := url.Parse(mapsURL); err == nil {
					_ = fyne.CurrentApp().OpenURL(u)
				}
			}
		}(lieu))
		btnLieu.Importance = widget.LowImportance

		// Créer un conteneur avec le texte coloré et le bouton
		btnWithText := container.NewStack(
			texteLieu,
			btnLieu,
		)

		nb := widget.NewLabel(fmt.Sprintf("%d date(s)", len(dates)))

		// Afficher seulement quelques dates, sinon le bloc explose
		limite := dates
		if len(dates) > maxDatesAffichees {
			limite = dates[:maxDatesAffichees]
		}

		listeDates := container.NewVBox()
		for _, d := range limite {
			listeDates.Add(widget.NewLabel("• " + d))
		}

		if len(dates) > maxDatesAffichees {
			restant := len(dates) - maxDatesAffichees
			listeDates.Add(widget.NewLabel(fmt.Sprintf("… +%d autre(s)", restant)))
		}

		contenu := container.NewVBox(
			btnWithText,
			widget.NewSeparator(),
			nb,
			widget.NewSeparator(),
			listeDates,
		)

		// La carte elle-même
		cartes = append(cartes, widget.NewCard("", "", contenu))
	}

	var blocConcerts fyne.CanvasObject
	if len(cartes) == 0 {
		blocConcerts = card("Concerts", widget.NewLabel("Aucun concert trouvé."))
	} else {
		// Grid 3 colonnes (propre sur écran large)
		grille := container.NewGridWithColumns(3, cartes...)
		blocConcerts = card("Concerts", grille)
	}

	return container.NewVScroll(container.NewVBox(barreBoutons, container.NewGridWithColumns(2, blocArtiste, blocInfos), blocMembres, blocConcerts))
}

// Vue favoris
func VueFavoris(onSelection func(modele.Artiste), rafraichir func() []modele.Artiste, retour func()) fyne.CanvasObject {
	favoris := rafraichir()
	listeFavoris := widget.NewList(
		func() int { return len(favoris) },
		func() fyne.CanvasObject {
			img := canvas.NewImageFromResource(theme.AccountIcon())
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(50, 50))
			label := widget.NewLabel("...")
			return container.NewHBox(img, label)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= len(favoris) {
				return
			}
			artiste := favoris[i]
			box := o.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)

			if artiste.Image != "" {
				newImg := canvas.NewImageFromURI(storage.NewURI(artiste.Image))
				newImg.FillMode = canvas.ImageFillContain
				newImg.SetMinSize(fyne.NewSize(50, 50))
				box.Objects[0] = newImg
			} else {
				img := canvas.NewImageFromResource(theme.AccountIcon())
				img.FillMode = canvas.ImageFillContain
				img.SetMinSize(fyne.NewSize(50, 50))
				box.Objects[0] = img
			}
			label.SetText(artiste.Nom)
		},
	)
	listeFavoris.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(favoris) {
			onSelection(favoris[id])
		}
	}

	btnRetour := widget.NewButton("← Retour", retour)

	titre := widget.NewLabelWithStyle("❤️ Mes Favoris", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	var contenu fyne.CanvasObject = listeFavoris
	if len(favoris) == 0 {
		contenu = container.NewCenter(widget.NewLabel("Aucun artiste favori.\nCliquez sur ❤️ dans la page d'un artiste !"))
	}
	return container.NewBorder(container.NewVBox(titre, btnRetour), nil, nil, nil, contenu)
}
