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

func RegisterUser(w http.ResponseWriter, r *http.Request, s storage.Storage, tokenAuth *jwtauth.JWTAuth) {
	user := storage.User{}
	rawData, err := io.ReadAll(r.Body)
	err = json.Unmarshal(rawData, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	fmt.Println(user)
	err = s.SaveUser(user)
	var errLogin *storage.ErrLoginExist
	if errors.As(err, &errLogin) {
		w.WriteHeader(http.StatusConflict)

	}
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenString := app.GetToken(user.Login, tokenAuth)
	cookie := http.Cookie{Name: "jwt", Value: tokenString}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
	return

}
func Login(w http.ResponseWriter, r *http.Request, s storage.Storage, tokenAuth *jwtauth.JWTAuth) {

	user := storage.User{}
	rawData, err := io.ReadAll(r.Body)
	fmt.Println("user in login:", rawData)
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
		tokenString := app.GetToken(user.Login, tokenAuth)
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
func SaveOrder(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userName := fmt.Sprintf("%v", claims["user_login"])
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	order_num := string(rawData)
	fmt.Println(order_num)
	fmt.Println(userName)
	err = s.SaveOrder(userName, order_num)
	switch err.(type) {
	default:
		w.WriteHeader(http.StatusAccepted)
	case *storage.ErrAlreadyLoadedByThisUser:
		w.WriteHeader(http.StatusOK)
		return

	case *storage.ErrAlreadyLoadedByDifferentUser:
		w.WriteHeader(http.StatusConflict)
		return
	case *storage.ErrDBInteraction:
		w.WriteHeader(http.StatusInternalServerError)
		return
		//TODO поменять на case, еще 422 вставить
	}
}
