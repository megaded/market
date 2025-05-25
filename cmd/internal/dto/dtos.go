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
	Order   string
	Status  string
	Accrual float64
}

type Withdraw struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (o *Withdraw) UnmarshalJSON(b []byte) error {
	type Dup Withdraw
	tmp := struct {
		ProcessedAt string `json:"processed_at"`
		*Dup
	}{
		Dup: (*Dup)(o),
	}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	o.ProcessedAt, err = time.Parse(time.RFC3339, tmp.ProcessedAt)
	if err != nil {
		return err
	}
	return nil
}

func (o *Order) UnmarshalJSON(b []byte) error {
	type Dup Order
	tmp := struct {
		UploadedAt string `json:"uploaded_at"`
		*Dup
	}{
		Dup: (*Dup)(o),
	}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	o.UploadedAt, err = time.Parse(time.RFC3339, tmp.UploadedAt)
	if err != nil {
		return err
	}
	return nil
}

type Balance struct {
	Current  float64 `json:"current"`
	Withdraw float64 `json:"withdraw"`
}
