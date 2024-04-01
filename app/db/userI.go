package db

type UserI interface {
	GetUser(discord_id string) (User, error)
	GetUsers() ([]User, error)
	CreateUser(user User) error
	UpdateUser(use User) error
}
