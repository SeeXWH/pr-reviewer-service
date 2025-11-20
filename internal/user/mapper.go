package user

import "github.com/SeeXWH/pr-reviewer-service/internal/model"

func ToResponse(u *model.User) UserResponseWrapper {
	if u == nil {
		return UserResponseWrapper{}
	}

	return UserResponseWrapper{
		User: UserDTO{
			UserID:   u.ID,
			Username: u.Username,
			TeamName: u.TeamName,
			IsActive: u.IsActive,
		},
	}
}
