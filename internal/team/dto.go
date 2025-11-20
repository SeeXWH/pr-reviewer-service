package team

type TeamCreateRequestDTO struct {
	TeamName string                 `json:"team_name"`
	Members  []UserCreateRequestDTO `json:"members"`
}

type UserCreateRequestDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type CreateTeamResponseDTO struct {
	Team TeamInfoDTO `json:"team"`
}

type TeamInfoDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
