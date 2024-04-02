package db

type User struct {
	Discord_id  string `gorm:"primary_key"`
	Students_id string
	Name        string
	Class       string
	Is_paid     bool
	Is_agreed   bool
}
