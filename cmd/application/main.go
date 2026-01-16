package main

import (
	"groupie-tracker/api"
	interfacegraphique "groupie-tracker/interface"
	"groupie-tracker/modele"
	"groupie-tracker/service"
	"io"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	a.Settings().SetTheme(interfacegraphique.ThemePerso{})

	w := a.NewWindow("Groupie Tracker")
	w.Resize(fyne.NewSize(900, 600))

	cache := service.NouveauCacheRelations()

	// Chargement initial
	w.SetContent(interfacegraphique.VueChargement("Chargement", "Récupération des artistes…", func() {}))

	// Chargement des artistes en arrière-plan
	go func() {
		artistes, err := api.RecupererArtistes()
		if err != nil {
			w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
			return
		}

		// Précharger les images des artistes
		imagesArtistes := make(map[int]fyne.Resource)
		for _, artiste := range artistes {
			if artiste.Image != "" {
				resp, err := http.Get(artiste.Image)
				if err == nil {
					imgData, err := io.ReadAll(resp.Body)
					resp.Body.Close()
					if err == nil {
						imagesArtistes[artiste.ID] = fyne.NewStaticResource("artiste", imgData)
					}
				}
			}
		}

		var vueAccueil fyne.CanvasObject
		var onglets *container.AppTabs

		// ✅ 1) On déclare d'abord la fonction de refresh (sinon scope error)
		var rafraichirAccueil func()

		// ✅ 2) Construction / reconstruction de l'accueil
		rafraichirAccueil = func() {
			suggestions := service.ConstruireSuggestions(artistes, cache)
			gestionnaireFavoris := service.ObtenirGestionnaireFavoris()

			// Vue accueil artistes
			vueAccueil = interfacegraphique.VueAccueil(
				artistes,
				imagesArtistes,

				// clic artiste => détails
				func(artiste modele.Artiste) {
					w.SetContent(interfacegraphique.VueChargement(
						"Chargement",
						"Récupération des concerts…",
						func() { w.SetContent(onglets) },
					))

					go func() {
						relation, err := service.RecupererRelationAvecCache(cache, artiste.ID)

						if err != nil {
							btnRetour := widget.NewButton("← Retour", func() { w.SetContent(onglets) })
							w.SetContent(container.NewCenter(container.NewVBox(
								widget.NewLabel("Erreur : "+err.Error()),
								btnRetour,
							)))
							return
						}

						w.SetContent(interfacegraphique.VueDetailsArtiste(
							w,
							artiste,
							relation,
							func() { w.SetContent(onglets) },
						))
					}()
				},

				// suggestions
				suggestions,
			)

			// Vue favoris
			vueFavoris := interfacegraphique.VueFavoris(
				// clic artiste favori => détails
				func(artiste modele.Artiste) {
					w.SetContent(interfacegraphique.VueChargement(
						"Chargement",
						"Récupération des concerts…",
						func() { w.SetContent(onglets) },
					))

					go func() {
						relation, err := service.RecupererRelationAvecCache(cache, artiste.ID)

						if err != nil {
							btnRetour := widget.NewButton("← Retour", func() { w.SetContent(onglets) })
							w.SetContent(container.NewCenter(container.NewVBox(
								widget.NewLabel("Erreur : "+err.Error()),
								btnRetour,
							)))
							return
						}

						w.SetContent(interfacegraphique.VueDetailsArtiste(
							w,
							artiste,
							relation,
							func() { w.SetContent(onglets) },
						))
					}()
				},
				// fonction pour rafraîchir la liste des favoris
				func() []modele.Artiste {
					return gestionnaireFavoris.ObtenirFavoris()
				},
			)

			// Onglets
			onglets = container.NewAppTabs(
				container.NewTabItem("Artistes", vueAccueil),
				container.NewTabItem("❤️ Favoris", vueFavoris),
			)

			w.SetContent(onglets)
		}

		// Premier affichage
		rafraichirAccueil()
	}()

	w.ShowAndRun()
}
