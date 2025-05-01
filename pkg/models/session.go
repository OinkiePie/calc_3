package models

type Session struct {
	ID      string
	UserID  int64
	Expires int64
}
