package model

type User struct {
	ID       string `gorm:"primaryKey;column:user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
	TeamName string `json:"team_name"`
}
