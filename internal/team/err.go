package team

import "errors"

var (
	ErrTeamExists   = errors.New("team_name already exists")
	ErrTeamNotFound = errors.New("resource not found")
)
