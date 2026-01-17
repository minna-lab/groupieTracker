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

	cache := service.NouveauCacheRelations()

	// Chargement initial
	w.SetContent(interfacegraphique.VueChargement("Chargement", "Récupération des artistes…", func() {}))
	w.Show()

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

		// Construction de l'accueil
		suggestions := service.ConstruireSuggestions(artistes, cache)
		gestionnaireFavoris := service.ObtenirGestionnaireFavoris()

		// Vue accueil artistes
		contenuAccueil := interfacegraphique.VueAccueil(
			artistes,
			imagesArtistes,

			// clic artiste => détails
			func(artiste modele.Artiste) {
				executerSurThreadPrincipal(func() {
					w.SetContent(interfacegraphique.VueChargement(
						"Chargement",
						"Récupération des concerts…",
						func() { w.SetContent(vueAccueil) },
					))
				})

				go func() {
					relation, err := service.RecupererRelationAvecCache(cache, artiste.ID)

					if err != nil {
						executerSurThreadPrincipal(func() {
							btnRetour := widget.NewButton("← Retour", func() { w.SetContent(vueAccueil) })
							w.SetContent(container.NewCenter(container.NewVBox(
								widget.NewLabel("Erreur : "+err.Error()),
								btnRetour,
							)))
						})
						return
					}

					executerSurThreadPrincipal(func() {
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
		)

		// Bouton favoris à droite
		btnFavoris := widget.NewButton("❤️ Mes Favoris", func() {
			vueFavoris := interfacegraphique.VueFavoris(
				// clic artiste favori => détails
				func(artiste modele.Artiste) {
					executerSurThreadPrincipal(func() {
						w.SetContent(interfacegraphique.VueChargement(
							"Chargement",
							"Récupération des concerts…",
							func() { w.SetContent(vueAccueil) },
						))
					})

					go func() {
						relation, err := service.RecupererRelationAvecCache(cache, artiste.ID)

						if err != nil {
							executerSurThreadPrincipal(func() {
								btnRetour := widget.NewButton("← Retour", func() { w.SetContent(vueAccueil) })
								w.SetContent(container.NewCenter(container.NewVBox(
									widget.NewLabel("Erreur : "+err.Error()),
									btnRetour,
								)))
							})
							return
						}

						executerSurThreadPrincipal(func() {
							w.SetContent(interfacegraphique.VueDetailsArtiste(
								w,
								artiste,
								relation,
								func() { w.SetContent(vueAccueil) },
							))
						})
					}()
				},
				// fonction pour rafraîchir la liste des favoris
				func() []modele.Artiste {
					return gestionnaireFavoris.ObtenirFavoris()
				},
				// fonction de retour
				func() { w.SetContent(vueAccueil) },
			)
			w.SetContent(vueFavoris)
		})

		// Créer une barre d'outils en haut à droite
		barreHaut := container.NewBorder(nil, nil, nil, btnFavoris)

		// Assembler la vue complète avec le bouton en haut
		vueAccueil = container.NewBorder(barreHaut, nil, nil, nil, contenuAccueil)

		executerSurThreadPrincipal(func() {
			w.SetContent(vueAccueil)
		})
	}()

	w.ShowAndRun()
}
