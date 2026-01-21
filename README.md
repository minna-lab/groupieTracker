# 🎸 Groupie Tracker GUI
Une application GUI développée en Go avec Fyne pour visualiser et explorer des informations sur des artistes et groupes musicaux à partir d'une API.
👥 Équipe de Développement
Projet réalisé par :

Raphaël
Minna
Berat

📋 Description
Groupie Tracker est une application graphique qui permet d'explorer des données sur des artistes et groupes musicaux. L'application récupère les informations depuis une API et les présente de manière interactive et conviviale.
API Utilisée
L'application utilise l'API Groupie Tracker : https://groupietrackers.herokuapp.com/api
L'API est composée de quatre sections principales :

Artists : Informations sur les groupes (nom, image, année de début, premier album, membres)
Locations : Lieux des concerts passés et à venir
Dates : Dates des concerts passés et à venir
Relations : Liens entre artistes, dates et lieux

## Fonctionnalités

- **Recherche** : par artiste, membres, lieux, dates (avec suggestions)
- **Filtres** : date de création, premier album, nombre de membres, lieux
- **Géolocalisation** : carte interactive des concerts
- **Interface** : design moderne et ergonomique
- **Bonus** : intégration Spotify, système de favoris

## Installation

### Prérequis
- Go 1.x+
- GCC (pour Fyne) : [tdm-gcc](https://jmeubank.github.io/tdm-gcc/download/)

### Lancement

```bash
git clone https://github.com/minna-lab/groupieTracker.git
cd groupie-tracker-gui
go mod download
go build -o groupie-tracker
./groupie-tracker
```

## Technologies

- Go
- Fyne (GUI)
- RESTful API



## Ressources

- [Documentation Fyne](https://developer.fyne.io/)
- [API Groupie Tracker](https://groupietrackers.herokuapp.com/api)

---

*Projet réalisé par Raphaël, Minna et Berat*