package db

type User struct {
	Discord_id  string `gorm:"primary_key"`
	Students_id string
	Name        string
	Class       string
	Signature   string
	NowTime     int64
	Is_paid     bool
	Is_agreed   bool
}
