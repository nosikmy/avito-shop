package model

type InfoOutput struct {
	Coins       int    `json:"coins"`
	Inventory   []Item `json:"inventory"`
	CoinHistory `json:"coinHistory"`
}

type Item struct {
	Type     string `json:"type" db:"item"`
	Quantity int    `json:"quantity" db:"quantity"`
}

type CoinHistory struct {
	Received []Receive `json:"received"`
	Sent     []Send    `json:"sent"`
}

type Receive struct {
	FromUser string `json:"fromUser" db:"sender"`
	Amount   int    `json:"amount" db:"amount"`
}

type Send struct {
	ToUser string `json:"toUser" db:"receiver"`
	Amount int    `json:"amount" db:"amount"`
}

type User struct {
	Username string `db:"username"`
	Balance  int    `db:"balance"`
}
