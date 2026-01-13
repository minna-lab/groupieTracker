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

	// File d’actions UI
	ui := make(chan func(), 100)
	go func() {
		for f := range ui {
			f()
		}
	}()

	// Chargement initial
	w.SetContent(interfacegraphique.VueChargement("Chargement", "Récupération des artistes…", nil))
	w.Show()

	go func() {
		artistes, err := api.RecupererArtistes()

		ui <- func() {
			if err != nil {
				w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
				return
			}

			var vueAccueil fyne.CanvasObject

			// ✅ Construction des suggestions typées
			suggestions := service.ConstruireSuggestions(artistes, cache)

			vueAccueil = interfacegraphique.VueAccueil(
				artistes,
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

							// Mise à jour suggestions avec les lieux (maintenant en cache)
							suggestions = service.ConstruireSuggestions(artistes, cache)

							w.SetContent(interfacegraphique.VueDetailsArtiste(
								w,
								artiste,
								relation,
								func() { w.SetContent(vueAccueil) },
							))
						}
					}()
				},
				suggestions,
			)

			w.SetContent(vueAccueil)
		}
	}()

	w.ShowAndRun()
}
