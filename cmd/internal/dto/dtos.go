package dto

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Accrual struct {
	Order   string
	Status  string
	Accrual int
}

type Withdraw struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}
