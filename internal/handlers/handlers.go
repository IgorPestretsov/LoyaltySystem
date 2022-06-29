package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IgorPestretsov/LoyaltySystem/internal/app"
	"github.com/IgorPestretsov/LoyaltySystem/internal/storage"
	"github.com/go-chi/jwtauth/v5"
	"github.com/theplant/luhn"
	"io"
	"net/http"
	"strconv"
)

type balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

func RegisterUser(w http.ResponseWriter, r *http.Request, s storage.Storage, tokenAuth *jwtauth.JWTAuth) {
	user := storage.User{}
	rawData, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(rawData, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = s.SaveUser(user)
	var errLogin *storage.ErrLoginExist
	if errors.As(err, &errLogin) {
		w.WriteHeader(http.StatusConflict)

	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenString := app.GetToken(user.Login, tokenAuth)
	cookie := http.Cookie{Name: "jwt", Value: tokenString}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)

}
func Login(w http.ResponseWriter, r *http.Request, s storage.Storage, tokenAuth *jwtauth.JWTAuth) {

	user := storage.User{}
	rawData, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(rawData, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var actualPasswordHash string
	actualPasswordHash, _ = s.GetUserPassword(user)
	passIsOk := app.CheckPasswordHash(user.Password, actualPasswordHash)
	if passIsOk {
		tokenString := app.GetToken(user.Login, tokenAuth)
		cookie := http.Cookie{Name: "jwt", Value: tokenString}
		http.SetCookie(w, &cookie)
		w.WriteHeader(http.StatusOK)
		return
	} else {
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
		return
	}
	orderNum := string(rawData)
	orderNumInt, _ := strconv.Atoi(orderNum)
	if !luhn.Valid(orderNumInt) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	err = s.SaveOrder(userName, orderNum)

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
	case *storage.ErrFormat:
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
}

func GetAllUserOrders(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	w.Header().Add("Content-Type", "application/json")
	userName := fmt.Sprintf("%v", claims["user_login"])
	orders, _ := s.GetUserOrders(userName)

	if orders == nil {
		w.WriteHeader(http.StatusNoContent)
	}

	output, err := json.Marshal(orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(output)
	if err != nil {
		return
	}
}
func GetBalance(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	w.Header().Add("Content-Type", "application/json")
	userName := fmt.Sprintf("%v", claims["user_login"])
	accruals, withdraws, err := s.GetBalance(userName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b := balance{Current: accruals, Withdrawn: withdraws}
	output, err := json.Marshal(b)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}
func Withdraw(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	request := withdrawRequest{}
	rawData, err := io.ReadAll(r.Body)
	err = json.Unmarshal(rawData, &request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	orderNumInt, err := strconv.Atoi(request.Order)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !luhn.Valid(orderNumInt) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	_, claims, _ := jwtauth.FromContext(r.Context())
	w.Header().Add("Content-Type", "application/json")
	userName := fmt.Sprintf("%v", claims["user_login"])
	err = s.Withdraw(userName, request.Order, request.Sum)
	var errNotEnough *storage.ErrNotEnoughPoints
	if errors.As(err, &errNotEnough) {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func GeWithdrawals(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	w.Header().Add("Content-Type", "application/json")
	userName := fmt.Sprintf("%v", claims["user_login"])
	witdrawals, err := s.GetWithdrawals(userName)
	var errNotEnough *storage.ErrNotEnoughPoints
	if errors.As(err, &errNotEnough) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if witdrawals == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	output, err := json.Marshal(witdrawals)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(output)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
