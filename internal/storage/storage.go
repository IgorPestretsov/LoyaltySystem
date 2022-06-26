package storage

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    uint32 `json:"accrual"`
	UploadedAt string `json:"uploaded_at"`
}

type OrderForProcessing struct {
	Number string
	Status string
}

type Storage interface {
	SaveUser(user User) error
	GetUserPassword(user User) (string, error)
	SaveOrder(user string, order_num string) error
	GetUserOrders(user_login string) ([]Order, error)
	GetRequiringToBeProcessed() ([]OrderForProcessing, error)
	ChangeStatus(ui string, status string) error
	ChangeStatusAndAcc(uid string, status string, accrual int32) error
}

const StatusProcessing = "PROCESSING"
const StatusNew = "NEW"
const StatusProcessed = "PROCESSED"
const StatusInvalid = "INVALID"
