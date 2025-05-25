package dto

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
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Balance struct {
	Current  float64 `json:"current"`
	Withdraw float64 `json:"withdraw"`
}
