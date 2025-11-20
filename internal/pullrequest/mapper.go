package pullrequest

import "github.com/SeeXWH/pr-reviewer-service/internal/model"

func ToDomain(req CreatePRRequestDTO) model.PullRequest {
	return model.PullRequest{
		ID:       req.PRID,
		Name:     req.Name,
		AuthorID: req.AuthorID,
	}
}

func ToResponse(pr *model.PullRequest) PRResponseWrapper {
	if pr == nil {
		return PRResponseWrapper{}
	}

	reviewerIDs := make([]string, 0, len(pr.Reviewers))
	for _, r := range pr.Reviewers {
		reviewerIDs = append(reviewerIDs, r.ID)
	}

	return PRResponseWrapper{
		PR: PRInfoDTO{
			PRID:      pr.ID,
			Name:      pr.Name,
			AuthorID:  pr.AuthorID,
			Status:    pr.Status,
			Reviewers: reviewerIDs,
			CreatedAt: pr.CreatedAt,
			MergedAt:  pr.MergedAt,
		},
	}
}

func ToReassignResponse(pr *model.PullRequest, newUserID string) ReassignResponseWrapper {
	baseResp := ToResponse(pr)
	return ReassignResponseWrapper{
		PR:         baseResp.PR,
		ReplacedBy: newUserID,
	}
}
