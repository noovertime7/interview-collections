package config

import (
	"github.com/shenyisyn/goft-gin/goft"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type DbConfig struct {
}

func NewDbConfig() *DbConfig {
	return &DbConfig{}
}

func (this *DbConfig) InitGorm() *gorm.DB {
	dsn := "root:123456@tcp(localhost:3306)/k8s?charset=utf8mb4&parseTime=True&loc=Local"
	gormdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	goft.Error(err)
	db, err := gormdb.DB()
	goft.Error(err)
	db.SetConnMaxLifetime(time.Minute * 10)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)
	return gormdb
}
