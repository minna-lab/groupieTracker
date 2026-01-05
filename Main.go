package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Bienvenue sur Groupie Tracker !\nChargement des données en cours...")
	})

	fmt.Println("Serveur démarré sur http://localhost:8080")
	//Démarrage serveur sur le port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
