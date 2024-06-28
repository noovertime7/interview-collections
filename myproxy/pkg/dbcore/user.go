package dbcore

type User struct {
	Id     uint `gorm:"primaryKey"`
	User   string
	Passwd string
}
