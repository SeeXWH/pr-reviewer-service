package model

type Team struct {
	Name    string `gorm:"primaryKey;column:team_name"`
	Members []User `gorm:"foreignKey:TeamName;references:Name"`
}
