package storage

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    string `json:"accrual"`
	UploadedAt string `json:"uploaded_at"`
}
type Storage interface {
	SaveUser(user User) error
	GetUserPassword(user User) (string, error)
	SaveOrder(user string, order_num string) error
	GetUserOrders(user_login string) ([]Order, error)
}
