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

// Permet d'exécuter une fonction sur le thread UI quand c'est supporté
type runner interface{ RunOnMain(func()) }

func executerSurThreadPrincipal(f func()) {
	if r, ok := fyne.CurrentApp().Driver().(runner); ok {
		r.RunOnMain(f)
		return
	}
	f()
}

func main() {
	a := app.New()
	a.Settings().SetTheme(interfacegraphique.ThemePerso{})

	w := a.NewWindow("Groupie Tracker")
	w.Resize(fyne.NewSize(1000, 700))

	cacheRelations := service.NouveauCacheRelations()

<<<<<<< HEAD
	// ✅ On affiche DIRECTEMENT une "page d'accueil vide" (pas un écran de chargement)
	// Accueil vide = artistes vide + suggestions vide + boutons/filtres visibles
	vueAccueilVide := interfacegraphique.VueAccueil(
		[]modele.Artiste{},
		func(modele.Artiste) {}, // clic artiste -> rien tant que pas chargé
		[]modele.Suggestion{},
		nil, // onChargerLieux désactivé tant que pas chargé
	)

	// Petit bandeau discret de chargement (dans la page, pas un écran)
	bandeau := container.NewVBox(
		widget.NewLabel("Récupération des artistes…"),
		widget.NewProgressBarInfinite(),
	)

	// On met le bandeau en haut, et l'accueil vide en dessous
	w.SetContent(container.NewBorder(bandeau, nil, nil, nil, vueAccueilVide))
	w.Show()
=======
	// Chargement initial
	w.SetContent(interfacegraphique.VueChargement("Chargement", "Récupération des artistes…", func() {}))
>>>>>>> 3f20571393c554d76fd3c15ca70e65f28920b991

	// ✅ Chargement des artistes en arrière-plan
	go func() {
		artistes, err := api.RecupererArtistes()
		if err != nil {
			w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
			return
		}

<<<<<<< HEAD
		executerSurThreadPrincipal(func() {
			if err != nil {
				// On reste dans la même page, mais on affiche l’erreur proprement
				w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
				return
=======
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
>>>>>>> 3f20571393c554d76fd3c15ca70e65f28920b991
			}
		}

<<<<<<< HEAD
			suggestions := []modele.Suggestion{} // branche ici si tu en génères ailleurs

			onChargerLieux := func(progress func(fait, total int), fin func(err error)) {
				go func() {
					total := len(artistes)
					fait := 0

					for _, a := range artistes {
						_, _ = service.RecupererRelationAvecCache(cacheRelations, a.ID)
						fait++
						if progress != nil {
							executerSurThreadPrincipal(func() { progress(fait, total) })
=======
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
>>>>>>> 3f20571393c554d76fd3c15ca70e65f28920b991
						}

<<<<<<< HEAD
					if fin != nil {
						executerSurThreadPrincipal(func() { fin(nil) })
					}
				}()
			}

			var vueAccueil fyne.CanvasObject

			vueAccueil = interfacegraphique.VueAccueil(
				artistes,
				func(artiste modele.Artiste) {
					// ✅ pas de vue de chargement concerts
					go func() {
						relation, err := service.RecupererRelationAvecCache(cacheRelations, artiste.ID)
						executerSurThreadPrincipal(func() {
							if err != nil {
								w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
								return
							}

							w.SetContent(interfacegraphique.VueDetailsArtiste(
								w,
								artiste,
								relation,
								func() { w.SetContent(vueAccueil) },
							))
						})
					}()
				},
				suggestions,
				onChargerLieux,
			)

			// ✅ Remplace l’accueil “vide” par le vrai accueil
			w.SetContent(vueAccueil)
		})
=======
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
>>>>>>> 3f20571393c554d76fd3c15ca70e65f28920b991
	}()

	w.ShowAndRun()
}
