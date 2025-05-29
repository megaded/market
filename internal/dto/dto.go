package dto

import (
	"encoding/json"
	"time"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Accrual struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual uint   `json:"accrual"`
}

type Withdraw struct {
	Order       string    `json:"order"`
	Sum         uint      `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    uint      `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (o Withdraw) MarshalJSON() ([]byte, error) {
	type Dup Withdraw
	tmp := struct {
		ProcessedAt string `json:"processed_at"`
		Dup
	}{
		Dup: (Dup)(o),
	}
	tmp.ProcessedAt = o.ProcessedAt.Format(time.RFC3339)
	b, err := json.Marshal(tmp)
	return b, err
}

func (o Order) MarshalJSON() ([]byte, error) {
	type Dup Order
	tmp := struct {
		UploadedAt string `json:"uploaded_at"`
		Dup
	}{
		Dup: (Dup)(o),
	}
	tmp.UploadedAt = o.UploadedAt.Format(time.RFC3339)
	b, err := json.Marshal(tmp)
	return b, err
}

type Balance struct {
	Current   uint `json:"current"`
	Withdrawn uint `json:"withdrawn"`
}
