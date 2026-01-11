package modele

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
