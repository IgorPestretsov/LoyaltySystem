package orderbroker

import (
	"encoding/json"
	"github.com/IgorPestretsov/LoyaltySystem/internal/storage"
	"log"
	"net/http"
	"time"
)

type accResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}
type Broker struct {
	s                storage.Storage
	queue            chan storage.OrderForProcessing
	accrualGetRecURL string
	client           *http.Client
}

func New(s storage.Storage, accURL string) *Broker {
	var b = Broker{}
	b.s = s
	b.queue = make(chan storage.OrderForProcessing, 20)
	b.accrualGetRecURL = accURL + "/api/orders/"
	b.client = &http.Client{Timeout: time.Second * 10}
	return &b
}

func (b *Broker) Start() {
	go b.getNewRecs()
	go b.GetStatusFromAccrual()

}
func (b *Broker) getNewRecs() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in getNewRecs", r)
		}
	}()
	for {
		output, err := b.s.GetRequiringToBeProcessed()
		if err != nil {
			log.Println("Broker err: DB interaction error")
		}
		for _, r := range output {
			b.queue <- r
			b.s.ChangeStatus(r.Number, storage.StatusProcessing)
		}
		time.Sleep(time.Second * 1)

	}
}

func (b *Broker) GetStatusFromAccrual() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in GetStatusFromAccrual", r)
		}
	}()
	var resp accResponse
	for r := range b.queue {
		resp = accResponse{}
		err := b.getJSON(b.accrualGetRecURL+r.Number, &resp)
		if err != nil {
			b.queue <- r
			time.Sleep(time.Second)
			continue
		}
		if (resp.Status != storage.StatusProcessed) && (resp.Status != storage.StatusInvalid) {
			b.queue <- r
			time.Sleep(time.Second)
			continue
		} else {
			_ = b.s.ChangeStatusAndAcc(resp.Order, resp.Status, resp.Accrual)
		}

	}
}
func (b *Broker) getJSON(url string, target interface{}) error {
	r, err := b.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
