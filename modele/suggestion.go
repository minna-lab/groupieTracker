package modele

type Suggestion struct {
	Texte string // ex: "Phil Collins"
	Type  string // ex: "membre", "artiste", "lieu", "premier album", "année création"
	ID    int    // id artiste associé (si applicable)
}
