package main

import (
	"groupie-tracker/api"
	interfacegraphique "groupie-tracker/interface"
	"groupie-tracker/modele"
	"groupie-tracker/service"
	"io"
	"net/http"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

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
			fyne.Do(func() {
				w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
			})
			return
		}

		// Map partagée pour les images (chargées progressivement)
		imagesArtistes := make(map[int]fyne.Resource)
		var mu sync.Mutex

		var vueAccueil *fyne.Container

		// Construction de l'accueil
		suggestions := service.ConstruireSuggestions(artistes, cache)
		gestionnaireFavoris := service.ObtenirGestionnaireFavoris()

		// Variable pour stocker la fonction de rafraîchissement
		var rafraichirGrilleAccueil func()

		// Vue accueil artistes
		contenuAccueil, rafraichirGrilleAccueil := interfacegraphique.VueAccueil(
			artistes,
			imagesArtistes,

			// clic artiste => détails
			func(artiste modele.Artiste) {
				w.SetContent(interfacegraphique.VueChargement(
					"Chargement",
					"Récupération des concerts…",
					func() { w.SetContent(vueAccueil) },
				))

				go func() {
					relation, err := service.RecupererRelationAvecCache(cache, artiste.ID)

					if err != nil {
						fyne.Do(func() {
							btnRetour := widget.NewButton("← Retour", func() { w.SetContent(vueAccueil) })
							w.SetContent(container.NewCenter(container.NewVBox(
								widget.NewLabel("Erreur : "+err.Error()),
								btnRetour,
							)))
						})
						return
					}

					fyne.Do(func() {
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
					w.SetContent(interfacegraphique.VueChargement(
						"Chargement",
						"Récupération des concerts…",
						func() { w.SetContent(vueAccueil) },
					))

					go func() {
						relation, err := service.RecupererRelationAvecCache(cache, artiste.ID)

						if err != nil {
							fyne.Do(func() {
								btnRetour := widget.NewButton("← Retour", func() { w.SetContent(vueAccueil) })
								w.SetContent(container.NewCenter(container.NewVBox(
									widget.NewLabel("Erreur : "+err.Error()),
									btnRetour,
								)))
							})
							return
						}

						fyne.Do(func() {
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

		fyne.Do(func() {
			w.SetContent(vueAccueil)
		})

		// Charger les images en arrière-plan APRÈS avoir affiché l'interface
		go func() {
			// Limiter les téléchargements simultanés à 10
			semaphore := make(chan struct{}, 10)
			var wg sync.WaitGroup

			for _, artiste := range artistes {
				if artiste.Image == "" {
					continue
				}

				wg.Add(1)
				go func(id int, imageURL string) {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					resp, err := http.Get(imageURL)
					if err == nil {
						imgData, err := io.ReadAll(resp.Body)
						resp.Body.Close()
						if err == nil {
							mu.Lock()
							imagesArtistes[id] = fyne.NewStaticResource("artiste", imgData)
							mu.Unlock()
							// Rafraîchir l'interface pour afficher la nouvelle image (sur le thread UI)
							fyne.Do(func() {
								if rafraichirGrilleAccueil != nil {
									rafraichirGrilleAccueil()
								}
							})
						}
					}
				}(artiste.ID, artiste.Image)
			}

			wg.Wait()
		}()
	}()

	w.ShowAndRun()
}
