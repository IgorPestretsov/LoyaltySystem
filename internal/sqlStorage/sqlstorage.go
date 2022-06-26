package sqlStorage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/IgorPestretsov/LoyaltySystem/internal/app"
	"github.com/IgorPestretsov/LoyaltySystem/internal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type SQLStorage struct {
	db *sql.DB
}

func NewSQLStorage(dsn string) *SQLStorage {
	s := SQLStorage{}
	var err error
	s.db, err = sql.Open("postgres", dsn)

	if err != nil {
		panic(err)
	}
	err = s.createTables()
	if err != nil {
		panic(err)
	}
	return &s
}

func (s *SQLStorage) createTables() error {
	_, err := s.db.Query("CREATE TABLE IF NOT EXISTS users (" +
		"uid SERIAL PRIMARY KEY," +
		"login VARCHAR(30) UNIQUE, " +
		"password VARCHAR(100))" +
		";")
	if err != nil {
		panic(err)
	}

	_, err = s.db.Query("CREATE TABLE IF NOT EXISTS orders (" +
		"order_num BIGSERIAL PRIMARY KEY," +
		"status VARCHAR(30) DEFAULT 'NEW'," +
		"accrual double precision DEFAULT 0," +
		"uploaded_at timestamp with time zone NOT NULL DEFAULT NOW()," +
		"uid VARCHAR(30))" +
		";")
	if err != nil {
		panic(err)
	}
	return nil
}

func (s *SQLStorage) SaveUser(user storage.User) error {
	hashedPass, err := app.HashPassword(user.Password)
	var pqErr *pq.Error
	_, err = s.db.Exec("insert into users(login,password) values ($1,$2);",
		user.Login, hashedPass)
	if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
		var errLoginExist *storage.ErrLoginExist
		return errLoginExist
	}
	return err

}
func (s *SQLStorage) DropTableUsers() {
	_, _ = s.db.Exec("drop table users;")

}
func (s *SQLStorage) GetUserPassword(user storage.User) (string, error) {
	var actualPasswordHash string
	_ = s.db.QueryRow("select password from users where login=$1", user.Login).Scan(&actualPasswordHash)
	return actualPasswordHash, nil
}
func (s *SQLStorage) SaveOrder(user string, order_num string) error {
	var pqErr *pq.Error
	uid, _ := s.getUIDbyUserLogin(user)
	_, err := s.db.Exec("insert into orders(order_num, uid) values ($1,$2);",
		order_num, uid)
	fmt.Println(err)
	if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
		uidThatLoaded, _ := s.getUIDbyOrderNum(order_num)
		if uidThatLoaded == uid {
			var errAlreadyLoadedByThisUser *storage.ErrAlreadyLoadedByThisUser
			return errAlreadyLoadedByThisUser
		} else {
			var errAlreadyLoadedByDifferentUser *storage.ErrAlreadyLoadedByDifferentUser
			return errAlreadyLoadedByDifferentUser
		}
	}
	if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.InvalidTextRepresentation {
		fmt.Println("here:", pqErr.Code)
		var errFormat *storage.ErrFormat
		return errFormat
	}
	if err != nil {
		var errDBInteraction *storage.ErrDBInteraction
		return errDBInteraction
	}
	return nil

}
func (s *SQLStorage) getUIDbyUserLogin(user_login string) (string, error) {

	var uid string
	err := s.db.QueryRow("select uid from users where login=$1;", user_login).Scan(&uid)
	return uid, err
}
func (s *SQLStorage) getUIDbyOrderNum(orderNum string) (string, error) {

	var uid string
	err := s.db.QueryRow("select uid from orders where order_num=$1;", orderNum).Scan(&uid)
	return uid, err
}

func (s *SQLStorage) GetUserOrders(user_login string) ([]storage.Order, error) {
	uid, _ := s.getUIDbyUserLogin(user_login)
	q := "select order_num, status, accrual, uploaded_at  from orders where uid=$1"
	rows, err := s.db.Query(q, uid)
	fmt.Println(err)
	if err != nil {
		panic(err)
	}
	//err := rows.Err()

	//if err != nil {
	//	var errDBInteraction *storage.ErrDBInteraction
	//	return nil, errDBInteraction
	//}
	var orders []storage.Order
	for rows.Next() {
		var order storage.Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			panic(err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (s *SQLStorage) GetRequiringToBeProcessed() ([]storage.OrderForProcessing, error) {
	q := "select order_num, status from orders where status = 'NEW'"
	rows, err := s.db.Query(q)
	if err != nil {
		panic(err)
	}
	//err := rows.Err()

	//if err != nil {
	//	var errDBInteraction *storage.ErrDBInteraction
	//	return nil, errDBInteraction
	//}
	var orders []storage.OrderForProcessing
	for rows.Next() {
		var order storage.OrderForProcessing
		err := rows.Scan(&order.Number, &order.Status)
		if err != nil {
			panic(err)
		}
		orders = append(orders, order)
	}
	return orders, nil

}
func (s *SQLStorage) ChangeStatus(uid string, status string) error {
	_, err := s.db.Exec("update orders set status=$2 where order_num=$1", uid, status)
	if err != nil {
		var errDBInteraction *storage.ErrDBInteraction
		return errDBInteraction
	}
	return nil
}

func (s *SQLStorage) ChangeStatusAndAcc(uid string, status string, accrual float32) error {
	_, err := s.db.Exec("update orders set status=$2, accrual=$3 where order_num=$1", uid, status, accrual)
	if err != nil {
		var errDBInteraction *storage.ErrDBInteraction
		fmt.Println(err)
		return errDBInteraction
	}
	return nil
}
