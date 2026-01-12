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

const baseURL = "https://groupietrackers.herokuapp.com"

type runner interface {
	RunOnMain(func())
}

func executerSurThreadPrincipal(f func()) {
	// Certains drivers Fyne ont RunOnMain mais l'interface fyne.Driver ne l'expose pas.
	// On fait donc une assertion de type.
	if r, ok := fyne.CurrentApp().Driver().(runner); ok {
		r.RunOnMain(f)
		return
	}
	// Fallback : on exécute directement (souvent ça marche, mais c'est moins sûr)
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
		var artistes []modele.Artiste
		err := api.RecupererJSON(baseURL+"/api/artists", &artistes)

		executerSurThreadPrincipal(func() {
			if err != nil {
				w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error())))
				return
			}
			w.SetContent(interfacegraphique.VueAccueil(artistes))
		})
	}()

	w.ShowAndRun()
}
