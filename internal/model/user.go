package model

type User struct {
	ID       string `gorm:"primaryKey;column:user_id" json:"user_id"`
	Username string
	IsActive bool
	TeamName string
}
