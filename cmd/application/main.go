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

	// Cache des concerts (relations)
	cache := service.NouveauCacheRelations()

	// File d’actions UI : toute modification d'UI passe par là
	ui := make(chan func(), 100)

	// Boucle UI
	go func() {
		for f := range ui {
			f()
		}
	}()

	// Chargement initial (pas de retour)
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

			// Accueil : liste des artistes
			vueAccueil = interfacegraphique.VueAccueil(artistes, func(artiste modele.Artiste) {

				// Chargement concerts (avec retour possible)
				w.SetContent(interfacegraphique.VueChargement(
					"Chargement",
					"Récupération des concerts…",
					func() { w.SetContent(vueAccueil) },
				))

				// Récupération concerts en arrière-plan
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

						// Affiche les détails (ta fonction attend w en 1er paramètre)
						w.SetContent(interfacegraphique.VueDetailsArtiste(
							w,
							artiste,
							relation,
							func() { w.SetContent(vueAccueil) },
						))
					}
				}()
			})

			w.SetContent(vueAccueil)
		}
	}()

	w.ShowAndRun()
}
