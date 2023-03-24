package model

type User struct {
	Id       string
	Name     string
	Password string
}
type Customer interface {
	getId() string
	getName() string
}
