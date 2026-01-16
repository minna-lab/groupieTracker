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

	// ✅ Chargement des artistes en arrière-plan
	go func() {
		artistes, err := api.RecupererArtistes()

		executerSurThreadPrincipal(func() {
			if err != nil {
				// On reste dans la même page, mais on affiche l’erreur proprement
				w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
				return
			}

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
						}
					}

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
	}()

	w.ShowAndRun()
}
