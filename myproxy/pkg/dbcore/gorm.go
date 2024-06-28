package dbcore

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var DB *MyDB

type MyDB struct {
	*gorm.DB
}

func NewMyDB() *MyDB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/k8s-user?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal()
	}
	return &MyDB{db}
}

func init() {
	DB = NewMyDB()
}
