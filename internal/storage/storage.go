package storage

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type Order struct {
	Number     uint64 `json:"number"`
	Status     string `json:"status"`
	Accrual    uint32 `json:"accrual"`
	UploadedAt string `json:"uploaded_at"`
}
type Storage interface {
	SaveUser(user User) error
	GetUserPassword(user User) (string, error)
	SaveOrder(user string, order_num string) error
	GetUserOrders(user_login string) ([]Order, error)
}
