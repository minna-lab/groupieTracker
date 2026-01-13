package main

import (
	"groupie-tracker/api"
	interfacegraphique "groupie-tracker/interface"
	"groupie-tracker/modele"
	"groupie-tracker/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Groupie Tracker")
	w.Resize(fyne.NewSize(900, 600))

	cache := service.NouveauCacheRelations()

	// File d’actions UI : toute modification UI passe par là
	ui := make(chan func(), 100)
	go func() {
		for f := range ui {
			f()
		}
	}()

	// Chargement initial
	w.SetContent(interfacegraphique.VueChargement("Chargement", "Récupération des artistes…", nil))
	w.Show()

	// Chargement des artistes en arrière-plan
	go func() {
		artistes, err := api.RecupererArtistes()

		ui <- func() {
			if err != nil {
				w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
				return
			}

			var vueAccueil fyne.CanvasObject

			// ✅ 1) On déclare d'abord la fonction de refresh (sinon scope error)
			var rafraichirAccueil func()

			// ✅ 2) Callback "Charger les lieux" (utilise rafraichirAccueil)
			onChargerLieux := func(progress func(fait, total int), fin func(err error)) {
				go func() {
					total := len(artistes)
					fait := 0

					for _, ar := range artistes {
						_, err := service.RecupererRelationAvecCache(cache, ar.ID)

						fait++
						ui <- func() { progress(fait, total) }

						if err != nil {
							ui <- func() { fin(err) }
							return
						}
					}

					// Tout est en cache -> refresh accueil pour inclure les lieux dans les suggestions
					ui <- func() {
						fin(nil)
						rafraichirAccueil()
					}
				}()
			}

			// ✅ 3) Construction / reconstruction de l'accueil
			rafraichirAccueil = func() {
				suggestions := service.ConstruireSuggestions(artistes, cache)

				vueAccueil = interfacegraphique.VueAccueil(
					artistes,

					// clic artiste => détails
					func(artiste modele.Artiste) {
						w.SetContent(interfacegraphique.VueChargement(
							"Chargement",
							"Récupération des concerts…",
							func() { w.SetContent(vueAccueil) },
						))

						go func() {
							relation, err := service.RecupererRelationAvecCache(cache, artiste.ID)

							ui <- func() {
								if err != nil {
									btnRetour := widget.NewButton("← Retour", func() { w.SetContent(vueAccueil) })
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
									func() { w.SetContent(vueAccueil) },
								))
							}
						}()
					},

					// suggestions
					suggestions,

					// bouton "Charger les lieux"
					onChargerLieux,
				)

				w.SetContent(vueAccueil)
			}

			// Premier affichage
			rafraichirAccueil()
		}
	}()

	w.ShowAndRun()
}
