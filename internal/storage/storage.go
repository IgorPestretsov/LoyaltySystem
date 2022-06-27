package storage

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

type OrderForProcessing struct {
	Number string
	Status string
}

type Storage interface {
	SaveUser(user User) error
	GetUserPassword(user User) (string, error)
	SaveOrder(user string, orderNum string) error
	GetUserOrders(userLogin string) ([]Order, error)
	GetRequiringToBeProcessed() ([]OrderForProcessing, error)
	ChangeStatus(ui string, status string) error
	ChangeStatusAndAcc(uid string, status string, accrual float32) error
	GetBalance(uid string) (float32, float32, error)
	Withdraw(uid string, orderNum string, sum float32) error
}

const StatusProcessing = "PROCESSING"
const StatusNew = "NEW"
const StatusProcessed = "PROCESSED"
const StatusInvalid = "INVALID"
