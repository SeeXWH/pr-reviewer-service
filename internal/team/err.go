package team

import "errors"

var (
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("resource not found")
)
