package db

type User struct {
	discord_id  string `gorm:"primary_key"`
	students_id string
	name        string
	class       string
	year        string
	is_paid     bool
	is_agreed   bool
}
