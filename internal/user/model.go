package user

type affectedPR struct {
	PRID          string
	AuthorID      string
	OldReviewerID string
}

type MassDeactivateResult struct {
	DeactivatedCount int
	ReassignedCount  int
}

type prReviewer struct {
	PullRequestID string `gorm:"column:pull_request_id"`
	UserID        string `gorm:"column:user_id"`
}
