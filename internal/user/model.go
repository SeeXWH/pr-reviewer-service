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
