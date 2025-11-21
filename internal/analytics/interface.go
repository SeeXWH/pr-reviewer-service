package analytics

import "context"

type Provider interface {
	GetStats(context.Context) ([]ReviewerStat, error)
}

type Storer interface {
	GetReviewerStats(context.Context) ([]ReviewerStat, error)
}
