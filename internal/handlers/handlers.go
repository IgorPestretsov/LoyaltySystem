package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IgorPestretsov/LoyaltySystem/internal/storage"
	"io"
	"net/http"
)

func RegisterUser(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	user := storage.User{}
	rawData, err := io.ReadAll(r.Body)
	err = json.Unmarshal(rawData, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	err = s.SaveUser(user)
	if errors.Is(err, storage.ErrNotFound) {
		w.WriteHeader(http.StatusConflict)

	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return

}
