package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IgorPestretsov/LoyaltySystem/internal/app"
	"github.com/IgorPestretsov/LoyaltySystem/internal/storage"
	"github.com/go-chi/jwtauth/v5"
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
func Login(w http.ResponseWriter, r *http.Request, s storage.Storage, tokenAuth *jwtauth.JWTAuth) {

	user := storage.User{}
	rawData, err := io.ReadAll(r.Body)
	err = json.Unmarshal(rawData, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	var actualPasswordHash string
	actualPasswordHash, err = s.GetUserPassword(user)
	fmt.Println(actualPasswordHash)
	passIsOk := app.CheckPasswordHash(user.Password, actualPasswordHash)
	if passIsOk {
		fmt.Println("pass is ok")
		_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_login": user.Login})
		cookie := http.Cookie{Name: "jwt", Value: tokenString}
		http.SetCookie(w, &cookie)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		fmt.Println("wrong Password")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

}
