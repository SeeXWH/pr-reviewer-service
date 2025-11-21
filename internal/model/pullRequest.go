package model

import "time"

type PullRequest struct {
	ID        string `gorm:"primaryKey;column:pull_request_id"`
	Name      string
	Status    string
	AuthorID  string
	Author    User    `gorm:"foreignKey:AuthorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Reviewers []*User `gorm:"many2many:pr_reviewers;"`
	CreatedAt time.Time
	MergedAt  *time.Time
}
