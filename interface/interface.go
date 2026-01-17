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

// Theme personnalisé
type ThemePerso struct{}

func (ThemePerso) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	colors := map[fyne.ThemeColorName]color.Color{
		theme.ColorNamePrimary:   color.NRGBA{34, 139, 230, 255},
		theme.ColorNameSelection: color.NRGBA{34, 139, 230, 55},
	}
	if c, ok := colors[name]; ok {
		return c
	}
	if name == theme.ColorNameBackground {
		return map[bool]color.Color{true: color.NRGBA{16, 16, 18, 255}, false: color.NRGBA{246, 247, 250, 255}}[variant == theme.VariantDark]
	}
	if name == theme.ColorNameInputBackground {
		return map[bool]color.Color{true: color.NRGBA{28, 28, 33, 255}, false: color.NRGBA{255, 255, 255, 255}}[variant == theme.VariantDark]
	}
	if name == theme.ColorNameForeground {
		return map[bool]color.Color{true: color.NRGBA{235, 235, 235, 255}, false: color.NRGBA{25, 25, 28, 255}}[variant == theme.VariantDark]
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (ThemePerso) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (ThemePerso) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (ThemePerso) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// coupe un texte trop long pour éviter qu'il déborde
func tronquer(texte string, max int) string {
	texte = strings.TrimSpace(texte)
	if max <= 0 || len(texte) <= max {
		return texte
	}
	return texte[:max-1] + "…"
}

// Composants réutilisables
func Carte(contenu fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 14
	ombre := canvas.NewRectangle(color.NRGBA{0, 0, 0, 18})
	ombre.CornerRadius = 14
	return container.NewPadded(container.NewStack(container.NewPadded(ombre), bg, container.NewPadded(contenu)))
}

func TitreSection(texte string) fyne.CanvasObject {
	t := canvas.NewText(texte, theme.Color(theme.ColorNameForeground))
	t.TextStyle = fyne.TextStyle{Bold: true}
	t.TextSize = 18
	return t
}

// Vue de chargement
func VueChargement(titre, message string, onRetour func()) fyne.CanvasObject {
	barre := widget.NewProgressBarInfinite()
	barre.Start()
	var retour fyne.CanvasObject = widget.NewLabel("")
	if onRetour != nil {
		retour = widget.NewButton("← Retour", onRetour)
	}
	return container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle(titre, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel(message), barre, retour))
}

// ============================================================================
// VUE ACCUEIL - LAYOUT 2 COLONNES
// ============================================================================

func VueAccueil(
	artistes []modele.Artiste,
	imagesArtistes map[int]fyne.Resource,
	onSelection func(modele.Artiste),
	suggestions []modele.Suggestion,
) fyne.CanvasObject {

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

		// Carte
		carte := container.NewVBox(
			img,
			nomLabel,
			anneeLabel,
		)

		// Bouton transparent pour le clic (sans effet de survol)
		btn := widget.NewButton("", func() {
			onSelection(artiste)
		})
		btn.Importance = widget.LowImportance

		// Stack simple sans carte qui crée l'ombre grise
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

	selectTrierPar := widget.NewSelect([]string{"Artiste", "Lieux", "Premier album", "Date de création"}, nil)
	selectTrierPar.SetSelected("Artiste")
	selectOrdre := widget.NewSelect([]string{"Croissant", "Décroissant"}, nil)
	selectOrdre.SetSelected("Croissant")

	// Nouveau filtre pour le nombre de membres
	selectNombreMembres := widget.NewSelect([]string{"Tous", "1", "2", "3", "4", "5+"}, nil)
	selectNombreMembres.SetSelected("Tous")

	var appliquer func()

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

		// Filtrer artistes
		idsLieux := idsDepuisSuggestions(texte, "lieu")
		artistesFiltres = artistesFiltres[:0]
		nombreMembres := selectNombreMembres.Selected

		for _, a := range artistes {
			// Filtre par nombre de membres
			if nombreMembres != "Tous" {
				nbMembres := len(a.Membres)
				if nombreMembres == "5+" {
					if nbMembres < 5 {
						continue
					}
				} else {
					nbVoulu, _ := strconv.Atoi(nombreMembres)
					if nbMembres != nbVoulu {
						continue
					}
				}
			}

			// Filtre par texte
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
		trierPar, ordre := selectTrierPar.Selected, selectOrdre.Selected
		for i := 0; i < len(artistesFiltres)-1; i++ {
			for j := 0; j < len(artistesFiltres)-i-1; j++ {
				echange := false
				switch trierPar {
				case "Artiste":
					a1, a2 := strings.ToLower(artistesFiltres[j].Nom), strings.ToLower(artistesFiltres[j+1].Nom)
					echange = (ordre == "Croissant" && a1 > a2) || (ordre == "Décroissant" && a1 < a2)
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
		rafraichirGrille()
	}

	recherche.OnChanged = func(string) { appliquer() }
	selectTrierPar.OnChanged = func(string) { appliquer() }
	selectOrdre.OnChanged = func(string) { appliquer() }
	selectNombreMembres.OnChanged = func(string) { appliquer() }

	// =========================================================================
	// COLONNE DE GAUCHE - FILTRES & RECHERCHE
	// =========================================================================
	titreRecherche := widget.NewLabelWithStyle("🔍 Recherche", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titreFiltres := widget.NewLabelWithStyle("⚙️ Filtres", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	filtresTri := container.NewVBox(
		widget.NewLabel("Trier par :"),
		selectTrierPar,
		widget.NewLabel("Ordre :"),
		selectOrdre,
		widget.NewLabel("Nombre de membres :"),
		selectNombreMembres,
	)

	colonneGauche := container.NewVBox(
		titreRecherche,
		recherche,
		listeSuggestions,
		titreFiltres,
		filtresTri,
	)

	// Mettre la colonne gauche dans un scroll pour éviter débordement
	colonneGaucheScroll := container.NewVScroll(colonneGauche)
	colonneGaucheScroll.SetMinSize(fyne.NewSize(300, 0))

	// Fond sombre pour la colonne de gauche
	bgColonneGauche := canvas.NewRectangle(color.NRGBA{240, 240, 240, 255})
	colonneGaucheAvecBg := container.NewStack(bgColonneGauche, colonneGaucheScroll)

	// =========================================================================
	// COLONNE DE DROITE - LISTE DES ARTISTES
	// =========================================================================
	titreEcouter := canvas.NewText("🎵 ÉCOUTER", theme.Color(theme.ColorNameForeground))
	titreEcouter.TextStyle = fyne.TextStyle{Bold: true}
	titreEcouter.TextSize = 24

	sousTitre := canvas.NewText("Pour vous", theme.Color(theme.ColorNameForeground))
	sousTitre.TextSize = 16

	enTeteDroite := container.NewVBox(
		titreEcouter,
		sousTitre,
	)

	// Scroll pour la grille d'artistes
	scrollGrille := container.NewVScroll(grilleArtistes)

	colonneDroite := container.NewBorder(enTeteDroite, nil, nil, nil, scrollGrille)

	// =========================================================================
	// ASSEMBLAGE FINAL - 2 COLONNES SANS BARRE DE SÉPARATION
	// =========================================================================
	appliquer()

	// Container simple sans barre de séparation
	layoutFinal := container.NewBorder(
		nil, nil,
		colonneGaucheAvecBg,
		nil,
		colonneDroite,
	)

	return layoutFinal
}

// Vue détails d'un artiste
func VueDetailsArtiste(
	fenetre fyne.Window,
	artiste modele.Artiste,
	relation modele.Relation,
	retour func(),
) fyne.CanvasObject {

	// -------------------------
	// Boutons du haut
	// -------------------------
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
	btnFavori.OnTapped = func() {
		ajoute := gestionnaireFavoris.Basculer(artiste)
		estFavori = ajoute
		majTexteFavori()
	}

	btnSpotify := widget.NewButton("🎵 Spotify", func() {
		query := url.QueryEscape(artiste.Nom)
		u, err := url.Parse("spotify:search:" + query)
		if err != nil {
			dialog.ShowError(err, fenetre)
			return
		}
		_ = fyne.CurrentApp().OpenURL(u)
	})

	barreBoutons := container.NewHBox(btnRetour, btnFavori, btnSpotify)

	// -------------------------
	// Blocs (cartes)
	// -------------------------
	card := func(titre string, contenu fyne.CanvasObject) fyne.CanvasObject {
		return widget.NewCard(titre, "", contenu)
	}

	// Image de l'artiste
	var imageArtiste fyne.CanvasObject
	if artiste.Image != "" {
		img := canvas.NewImageFromURI(storage.NewURI(artiste.Image))
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(350, 350))
		imageArtiste = container.NewCenter(img)
	} else {
		imageArtiste = widget.NewLabel("Image non disponible")
	}

	// Bloc Artiste avec image
	lblNom := widget.NewLabelWithStyle(artiste.Nom, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	blocArtiste := card("Artiste", container.NewVBox(imageArtiste, lblNom))

	// Bloc Informations
	info1 := widget.NewLabel(fmt.Sprintf("📅 Année de création : %d", artiste.AnneeCreation))
	info2 := widget.NewLabel(fmt.Sprintf("💿 Premier album : %s", artiste.PremierAlbum))
	blocInfos := card("Informations", container.NewVBox(info1, info2))

	// Bloc Membres (en 2 colonnes pour être plus propre)
	membresTexte := strings.Join(artiste.Membres, ", ")
	lblMembres := widget.NewLabel(membresTexte)
	lblMembres.Wrapping = fyne.TextWrapWord
	blocMembres := card("Membres", lblMembres)

	// -------------------------
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

		// Titre contenu dans la carte (tronqué + wrap)
		titreLieu := widget.NewLabelWithStyle("📍 "+tronquer(lieu, maxTitreLieu), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		titreLieu.Wrapping = fyne.TextWrapWord

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
			titreLieu,
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

	// -------------------------
	// Mise en page globale
	// -------------------------
	ligneHaut := container.NewGridWithColumns(2, blocArtiste, blocInfos)

	page := container.NewVBox(
		barreBoutons,
		ligneHaut,
		blocMembres,
		blocConcerts,
	)

	// Scroll global (pas de scroll interne dans les concerts)
	return container.NewVScroll(page)
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
