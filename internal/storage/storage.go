package storage

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type Storage interface {
	SaveUser(user User) error
	GetUserPassword(user User) (string, error)
	SaveOrder(user string, order_num string) error
}
