package modele

// Types de base pour les artistes
type Artiste struct {
	ID            int      `json:"id"`
	Image         string   `json:"image"`
	Nom           string   `json:"name"`
	Membres       []string `json:"members"`
	AnneeCreation int      `json:"creationDate"`
	PremierAlbum  string   `json:"firstAlbum"`
}

type Lieux struct {
	ID    int      `json:"id"`
	Lieux []string `json:"locations"`
}

type Dates struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

type Relation struct {
	ID           int                 `json:"id"`
	DatesParLieu map[string][]string `json:"datesLocations"`
}

// Coordonnées géographiques
type Coordonnees struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// Suggestion pour l'autocomplétion
type Suggestion struct {
	Texte string // ex: "Phil Collins"
	Type  string // ex: "membre", "artiste", "lieu", "premier album", "année création"
	ID    int    // id artiste associé (si applicable)
}
