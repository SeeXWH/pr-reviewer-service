package pullrequest

import "errors"

var (
	ErrPRExists       = errors.New("PR id already exists")
	ErrAuthorNotFound = errors.New("author not found")
	ErrPRNotFound     = errors.New("PR not found")
	ErrPRMerged       = errors.New("cannot reassign on merged PR")
	ErrNotAssigned    = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate    = errors.New("no active replacement candidate in team")
)
