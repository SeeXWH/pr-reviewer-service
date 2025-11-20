package pullrequest

import "time"

type CreatePRRequestDTO struct {
	PRID     string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type PRResponseWrapper struct {
	PR PRInfoDTO `json:"pr"`
}

type PRInfoDTO struct {
	PRID      string     `json:"pull_request_id"`
	Name      string     `json:"pull_request_name"`
	AuthorID  string     `json:"author_id"`
	Status    string     `json:"status"`
	Reviewers []string   `json:"assigned_reviewers"`
	CreatedAt time.Time  `json:"createdAt"`
	MergedAt  *time.Time `json:"mergedAt"`
}

type MergePRRequestDTO struct {
	PRID string `json:"pull_request_id"`
}

type ReassignPRRequestDTO struct {
	PRID      string `json:"pull_request_id"`
	OldUserID string `json:"old_user_id"`
}

type ReassignResponseWrapper struct {
	PR         PRInfoDTO `json:"pr"`
	ReplacedBy string    `json:"replaced_by"`
}
