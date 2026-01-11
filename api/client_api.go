package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func RecupererJSON(url string, cible interface{}) error {
	reponse, err := http.Get(url)
	if err != nil {
		return err
	}
	defer reponse.Body.Close()

	if reponse.StatusCode != http.StatusOK {
		return fmt.Errorf("erreur HTTP : %s", reponse.Status)
	}

	err = json.NewDecoder(reponse.Body).Decode(cible)
	if err != nil {
		return err
	}

	return nil
}
