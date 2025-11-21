package user

import "github.com/SeeXWH/pr-reviewer-service/internal/model"

func ToResponse(u *model.User) ResponseWrapper {
	if u == nil {
		return ResponseWrapper{}
	}

	return ResponseWrapper{
		User: DTO{
			UserID:   u.ID,
			Username: u.Username,
			TeamName: u.TeamName,
			IsActive: u.IsActive,
		},
	}
}

func ToReviewsResponse(userID string, prs []model.PullRequest) ReviewsResponseDTO {
	prDTOs := make([]PullRequestShortDTO, 0, len(prs))

	for _, pr := range prs {
		prDTOs = append(prDTOs, PullRequestShortDTO{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		})
	}

	return ReviewsResponseDTO{
		UserID:       userID,
		PullRequests: prDTOs,
	}
}

func ToMassDeactivateResponse(res MassDeactivateResult) MassDeactivateResponseDTO {
	return MassDeactivateResponseDTO{
		DeactivatedCount: res.DeactivatedCount,
		ReassignedPRs:    res.ReassignedCount,
	}
}
