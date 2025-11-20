package team

type CreateRequestDTO struct {
	TeamName string                 `json:"team_name"`
	Members  []UserCreateRequestDTO `json:"members"`
}

type UserCreateRequestDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type CreateTeamResponseDTO struct {
	Team InfoDTO `json:"team"`
}

type InfoDTO struct {
	TeamName string      `json:"team_name"`
	Members  []MemberDTO `json:"members"`
}

type MemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
