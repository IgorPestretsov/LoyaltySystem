package storage

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type Storage interface {
	SaveUser(user User) error
}
