package interfacegraphique

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"groupie-tracker/modele"
	"groupie-tracker/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// coupe un texte trop long pour éviter qu'il déborde
func tronquer(texte string, max int) string {
	texte = strings.TrimSpace(texte)
	if max <= 0 || len(texte) <= max {
		return texte
	}
	return texte[:max-1] + "…"
}

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

	btnCarte := widget.NewButton("🗺️ Voir sur la carte", func() {
		markers, err := service.ConstruireMarkers(relation)
		if err != nil {
			dialog.ShowError(err, fenetre)
			return
		}

		chemin, err := service.GenererFichierCarteHTML(artiste.Nom, markers)
		if err != nil {
			dialog.ShowError(err, fenetre)
			return
		}

		abs, err := filepath.Abs(chemin)
		if err != nil {
			dialog.ShowError(err, fenetre)
			return
		}

		u, err := url.Parse("file://" + abs)
		if err != nil {
			dialog.ShowError(err, fenetre)
			return
		}

		_ = fyne.CurrentApp().OpenURL(u)
	})

	btnSpotify := widget.NewButton("🎵 Spotify", func() {
		query := url.QueryEscape(artiste.Nom)
		u, err := url.Parse("spotify:search:" + query)
		if err != nil {
			dialog.ShowError(err, fenetre)
			return
		}
		_ = fyne.CurrentApp().OpenURL(u)
	})

	barreBoutons := container.NewHBox(btnRetour, btnFavori, btnCarte, btnSpotify)

	// -------------------------
	// Blocs (cartes)
	// -------------------------
	card := func(titre string, contenu fyne.CanvasObject) fyne.CanvasObject {
		return widget.NewCard(titre, "", contenu)
	}

	// Bloc Artiste
	lblNom := widget.NewLabelWithStyle(artiste.Nom, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	blocArtiste := card("Artiste", container.NewVBox(lblNom))

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
