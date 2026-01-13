package interfacegraphique

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

// Carte : encadre joliment un contenu (fond + coins arrondis + padding)
func Carte(contenu fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	bg.CornerRadius = 14

	ombre := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 18})
	ombre.CornerRadius = 14

	return container.NewPadded(
		container.NewStack(
			container.NewPadded(ombre), // ombre légère
			bg,
			container.NewPadded(contenu),
		),
	)
}

// TitreSection : titre plus visible (gras + taille)
func TitreSection(texte string) fyne.CanvasObject {
	t := canvas.NewText(texte, theme.Color(theme.ColorNameForeground))
	t.TextStyle = fyne.TextStyle{Bold: true}
	t.TextSize = 18
	return t
}
