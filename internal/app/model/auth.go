package model

type AuthInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type AuthOutput struct {
	Token string `json:"token"`
}

type AuthDB struct {
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
}
