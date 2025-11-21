package model

type User struct {
	ID       string `gorm:"primaryKey;column:user_id"`
	Username string
	IsActive bool
	TeamName string `gorm:"column:team_name;index"`
}
