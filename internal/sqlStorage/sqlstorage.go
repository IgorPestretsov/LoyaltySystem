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
	return nil
}

func (s *SQLStorage) SaveUser(user storage.User) error {
	hashedPass, err := app.HashPassword(user.Password)
	var pqErr *pq.Error
	_, err = s.db.Exec("insert into users(login,password) values ($1,$2);",
		user.Login, hashedPass)
	if errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
		return storage.ErrNotFound
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
