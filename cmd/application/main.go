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

	go func() {
		artistes, err := api.RecupererArtistes()
		if err != nil {
			fyne.Do(func() { w.SetContent(container.NewCenter(widget.NewLabel("Erreur : " + err.Error()))) })
			return
		}

		imagesArtistes := make(map[int]fyne.Resource)
		var mu sync.Mutex
		var vueAccueil *fyne.Container
		suggestions := service.ConstruireSuggestions(artistes, cache)
		gestionnaireFavoris := service.ObtenirGestionnaireFavoris()
		var rafraichirGrilleAccueil func()
		chargerDetails := func(artiste modele.Artiste) {
			w.SetContent(interfacegraphique.VueChargement("Chargement", "Récupération des concerts…", func() { w.SetContent(vueAccueil) }))
			go func() {
				relation, err := service.RecupererRelationAvecCache(cache, artiste.ID)
				if err != nil {
					fyne.Do(func() {
						w.SetContent(container.NewCenter(container.NewVBox(widget.NewLabel("Erreur : "+err.Error()), widget.NewButton("← Retour", func() { w.SetContent(vueAccueil) }))))
					})
					return
				}
				fyne.Do(func() {
					w.SetContent(interfacegraphique.VueDetailsArtiste(w, artiste, relation, func() { w.SetContent(vueAccueil) }))
				})
			}()
		}

		contenuAccueil, rafraichirGrilleAccueil := interfacegraphique.VueAccueil(artistes, imagesArtistes, chargerDetails, suggestions)

		btnFavoris := widget.NewButton("❤️ Mes Favoris", func() {
			w.SetContent(interfacegraphique.VueFavoris(chargerDetails, func() []modele.Artiste { return gestionnaireFavoris.ObtenirFavoris() }, func() { w.SetContent(vueAccueil) }))
		})
		vueAccueil = container.NewBorder(container.NewBorder(nil, nil, nil, btnFavoris), nil, nil, nil, contenuAccueil)
		fyne.Do(func() { w.SetContent(vueAccueil) })

		go func() {
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
					if resp, err := http.Get(imageURL); err == nil {
						if imgData, err := io.ReadAll(resp.Body); err == nil {
							resp.Body.Close()
							mu.Lock()
							imagesArtistes[id] = fyne.NewStaticResource("artiste", imgData)
							mu.Unlock()
							fyne.Do(func() {
								if rafraichirGrilleAccueil != nil {
									rafraichirGrilleAccueil()
								}
							})
						} else {
							resp.Body.Close()
						}
					}
				}(artiste.ID, artiste.Image)
			}
			wg.Wait()
		}()
	}()

	w.ShowAndRun()
}
