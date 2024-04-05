package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

// CreateUser implements UserI.
func (db DB) CreateUser(user User) error {
	result := db.db.Create(
		&user,
	)

	//エラー処理
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetUser implements UserI.
func (db DB) GetUser(discord_id string) (User, error) {
	return_data := User{}

	result := db.db.First(&return_data,User{Discord_id: discord_id})

	//エラー処理
	if result.Error != nil {
		return User{}, result.Error
	}

	return return_data, nil
}

// GetUsers implements UserI.
func (db DB) GetUsers() ([]User, error) {
	return_datas := []User{}

	result := db.db.Find(&return_datas)

	//エラー処理
	if result.Error != nil {
		return []User{}, result.Error
	}

	return return_datas, nil
}

// UpdateUser implements UserI.
func (db DB) UpdateUser(user User) error {
	result := db.db.Save(&user)

	//エラー処理
	if result.Error != nil {
		return result.Error
	}

	return nil
}

const dns = "test:%v@tcp(mysql:3306)/%v?charset=utf8mb4&parseTime=True&loc=Local"
//"root:%v@tcp(mysql:3306)/%v"

func Init() (UserI, error) {
	/*
	dburl := fmt.Sprintf(
		dns,
		os.Getenv("MYSQL_ROOT_PASSWORD"),
		os.Getenv("MYSQL_DATABASE"),
	)

	log.Println(dburl)

	db, err := gorm.Open(
		mysql.Open(dburl),
		&gorm.Config{})
	*/

	db,err := gorm.Open(sqlite.Open("test.db"),&gorm.Config{})

	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&User{})

	return (&DB{}).new(db), nil
}

func (db *DB) new(gormConnection *gorm.DB) DB {
	db.db = gormConnection
	return *db
}

/*
func (db *DB) GetUser(discord_id string) (User, error) {
	data := &User{}
	result := db.db.First(data, "discord_id = ?", discord_id)

	//エラー処理
	if result.Error != nil {
		return User{}, result.Error
	}

	return *data, nil
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
*/