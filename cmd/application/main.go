package main

import (
	"groupie-tracker/api"
	interfacegraphique "groupie-tracker/interface"
	"groupie-tracker/modele"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

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
	w := a.NewWindow("Groupie Tracker")
	w.Resize(fyne.NewSize(900, 600))

	loading := widget.NewLabel("Chargement des artistes…")
	w.SetContent(container.NewCenter(loading))
	w.Show()

	go func() {
		artistes, err := api.RecupererArtistes()

		executerSurThreadPrincipal(func() {
			if err != nil {
				w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
				return
			}

			var vueAccueil fyne.CanvasObject
			vueAccueil = interfacegraphique.VueAccueil(artistes, func(artiste modele.Artiste) {
				w.SetContent(container.NewCenter(widget.NewLabel("Chargement des détails…")))

				go func() {
					relation, err := api.RecupererRelation(artiste.ID)

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
			})

			w.SetContent(vueAccueil)
		})
	}()

	w.ShowAndRun()
}
