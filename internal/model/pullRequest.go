package model

import "time"

type PullRequest struct {
	ID        string  `gorm:"primaryKey;column:pull_request_id" json:"pull_request_id"`
	Name      string  `json:"pull_request_name"`
	Status    string  `json:"status"`
	AuthorID  string  `json:"author_id"`
	Reviewers []*User `gorm:"many2many:pr_reviewers;" json:"assigned_reviewers"`
	CreatedAt time.Time
	MergedAt  *time.Time
}
