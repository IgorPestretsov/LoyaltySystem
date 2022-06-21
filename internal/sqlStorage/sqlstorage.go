package sqlStorage

import (
	"database/sql"
	"errors"
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
		"order_num SERIAL PRIMARY KEY," +
		"uid VARCHAR(30) UNIQUE)" +
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
		return storage.ErrLoginExist
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
	if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
		uidThatLoaded, _ := s.getUIDbyOrderNum(order_num)
		if uidThatLoaded == uid {
			return storage.ErrAlreadyLoadedByThisUser
		} else {
			return storage.ErrAlreadyLoadedByDifferentUser
		}
	}
	if err != nil {
		return storage.ErrDBInteraction
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
