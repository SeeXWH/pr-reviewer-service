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

func ToReviewsResponse(userID string, prs []model.PullRequest) UserReviewsResponseDTO {
	prDTOs := make([]PullRequestShortDTO, 0, len(prs))

	for _, pr := range prs {
		prDTOs = append(prDTOs, PullRequestShortDTO{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		})
	}

	return UserReviewsResponseDTO{
		UserID:       userID,
		PullRequests: prDTOs,
	}
}
