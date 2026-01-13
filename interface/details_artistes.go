package interfacegraphique

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"

	"groupie-tracker/modele"
	"groupie-tracker/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	// Bouton like/unlike
	gestionnaireFavoris := service.ObtenirGestionnaireFavoris()
	estFavori := gestionnaireFavoris.EstFavori(artiste.ID)

	var btnLike *widget.Button
	btnLike = widget.NewButton(map[bool]string{true: "💔 Retirer des favoris", false: "❤️ Ajouter aux favoris"}[estFavori], func() {
		ajoute := gestionnaireFavoris.Basculer(artiste)
		if ajoute {
			btnLike.SetText("💔 Retirer des favoris")
		} else {
			btnLike.SetText("❤️ Ajouter aux favoris")
		}
	})

	// ✅ Bouton carte (corrigé)
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

	// Bouton Spotify
	btnSpotify := widget.NewButton("🎵 Écouter sur Spotify", func() {
		// Utiliser le protocole spotify:// qui ouvre l'application Spotify
		searchQuery := url.QueryEscape(artiste.Nom)
		spotifyURL := "spotify:search:" + searchQuery

		u, err := url.Parse(spotifyURL)
		if err != nil {
			dialog.ShowError(err, fenetre)
			return
		}

		_ = fyne.CurrentApp().OpenURL(u)
	})

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

	haut := container.NewVBox(
		container.NewHBox(btnRetour, btnLike, btnCarte, btnSpotify),
		titre,
		infos,
		membres,
	)

	return container.NewBorder(
		haut,
		nil, nil, nil,
		container.NewScroll(concerts),
	)
}
