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
	rawData, err := io.ReadAll(r.Body)
	err = json.Unmarshal(rawData, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
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
	order_num_int, _ := strconv.Atoi(order_num)
	if luhn.Valid(order_num_int) == false {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
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
	case *storage.ErrFormat:
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
}

func GetAllUserOrders(w http.ResponseWriter, r *http.Request, s storage.Storage) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	w.Header().Add("Content-Type", "application/json")
	userName := fmt.Sprintf("%v", claims["user_login"])
	orders, err := s.GetUserOrders(userName)

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
		fmt.Println(err)
		return
	}
	b := balance{Current: accruals, Withdrawn: withdraws}
	output, err := json.Marshal(b)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(output)
	return
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

	if luhn.Valid(orderNumInt) == false {
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
	return
}
