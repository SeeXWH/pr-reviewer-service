package team

import "github.com/SeeXWH/pr-reviewer-service/internal/model"

func ToDomain(req TeamCreateRequestDTO) model.Team {
	members := make([]model.User, len(req.Members))

	for i, m := range req.Members {
		members[i] = model.User{
			ID:       m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
			TeamName: req.TeamName,
		}
	}

	return model.Team{
		Name:    req.TeamName,
		Members: members,
	}
}
func ToResponse(t *model.Team) CreateTeamResponseDTO {
	if t == nil {
		return CreateTeamResponseDTO{}
	}

	members := make([]TeamMemberDTO, len(t.Members))
	for i, m := range t.Members {
		members[i] = TeamMemberDTO{
			UserID:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	teamInfo := TeamInfoDTO{
		TeamName: t.Name,
		Members:  members,
	}

	return CreateTeamResponseDTO{
		Team: teamInfo,
	}
}
