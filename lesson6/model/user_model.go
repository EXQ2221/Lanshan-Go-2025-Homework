package model

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Changepw struct {
	Username string `json:"username"`
	OldPass  string `json:"oldpass""`
	Newpass  string `json:"newpass"`
}
