package db

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

const dns = "root:%v@tcp(mysql:3306)/%v"

func Init() (UserI, error) {
	db, err := gorm.Open(
		mysql.Open(
			fmt.Sprintf(
				dns,
				os.Getenv("MYSQL_ROOT_PASSWORD"),
				os.Getenv("MYSQL_DATABASE"))),
		&gorm.Config{})

	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&User{})

	return (&DB{}).new(db), nil
}

func (db *DB) new(gormConnection *gorm.DB) UserI {
	return &DB{db: gormConnection}

}

func (db *DB) GetUser(discord_id string) (User, error) {
	return db.db.First(&User{}, "discord_id = ?", discord_id)
}

func (db *DB) GetUsers() ([]User, error) {
	return []User{}, nil
}

func (db *DB) CreateUser(user User) error {
	return nil
}

func (db *DB) Init() error {
	return nil
}
