package interfacegraphique

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemePerso : surcharge quelques couleurs du thème par défaut.
type ThemePerso struct{}

func (ThemePerso) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {

	case theme.ColorNamePrimary:
		return color.NRGBA{R: 34, G: 139, B: 230, A: 255} // Bleu

	case theme.ColorNameBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 16, G: 16, B: 18, A: 255}
		}
		return color.NRGBA{R: 246, G: 247, B: 250, A: 255}

	case theme.ColorNameInputBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 28, G: 28, B: 33, A: 255}
		}
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	case theme.ColorNameForeground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 235, G: 235, B: 235, A: 255}
		}
		return color.NRGBA{R: 25, G: 25, B: 28, A: 255}

	case theme.ColorNameSelection:
		return color.NRGBA{R: 34, G: 139, B: 230, A: 55}
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (ThemePerso) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (ThemePerso) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (ThemePerso) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
