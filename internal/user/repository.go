package user

import (
	"context"

	"github.com/SeeXWH/pr-reviewer-service/internal/model"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *db.PostgresDB
}

func NewRepository(db *db.PostgresDB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) UpdateActiveStatus(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	var user model.User
	err := r.db.PostgresDB.WithContext(ctx).First(&user, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	err = r.db.PostgresDB.WithContext(ctx).Model(&user).Update("is_active", isActive).Error
	if err != nil {
		return nil, err
	}
	user.IsActive = isActive
	return &user, nil
}

func (r *Repository) GetUserReviews(ctx context.Context, userID string) ([]model.PullRequest, error) {
	var count int64
	err := r.db.PostgresDB.WithContext(ctx).Model(&model.User{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var prs []model.PullRequest
	err = r.db.PostgresDB.WithContext(ctx).
		Table("pull_requests").
		Joins("JOIN pr_reviewers ON pr_reviewers.pull_request_id = pull_requests.pull_request_id").
		Where("pr_reviewers.user_id = ?", userID).
		Find(&prs).Error
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (r *Repository) GetReviewCandidates(ctx context.Context, teamName string, authorID string) ([]model.User, error) {
	var candidates []model.User
	err := r.db.PostgresDB.WithContext(ctx).
		Where("team_name = ? AND is_active = ? AND user_id != ?", teamName, true, authorID).
		Order("RANDOM()").
		Limit(2).
		Find(&candidates).Error

	if err != nil {
		return nil, err
	}
	return candidates, nil
}

func (r *Repository) GetByID(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := r.db.PostgresDB.WithContext(ctx).First(&user, "user_id = ?", userID).Error
	return &user, err
}

func (r *Repository) GetReplacementCandidate(
	ctx context.Context,
	teamName string,
	excludeUserIDs []string,
) (*model.User, error) {
	var candidate model.User
	err := r.db.PostgresDB.WithContext(ctx).
		Where("team_name = ? AND is_active = ?", teamName, true).
		Not("user_id", excludeUserIDs).
		Order("RANDOM()").
		First(&candidate).Error

	if err != nil {
		return nil, err
	}
	return &candidate, nil
}

func (r *Repository) MassDeactivateAndReassign(
	ctx context.Context,
	teamName string,
	userIDs []string,
) (MassDeactivateResult, error) {
	var result MassDeactivateResult

	err := r.db.PostgresDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		count, err := r.deactivateUsers(tx, teamName, userIDs)
		if err != nil {
			return err
		}
		result.DeactivatedCount = count
		if count == 0 {
			return nil
		}

		candidates, err := r.getActiveCandidates(tx, teamName)
		if err != nil {
			return err
		}

		affectedPRs, err := r.getAffectedPRs(tx, userIDs)
		if err != nil {
			return err
		}
		if len(affectedPRs) == 0 {
			return nil
		}

		newRelations, count := r.calculateReplacements(affectedPRs, candidates)
		result.ReassignedCount = count

		return r.applyReviewerChanges(tx, userIDs, affectedPRs, newRelations)
	})

	return result, err
}

func (r *Repository) deactivateUsers(tx *gorm.DB, teamName string, userIDs []string) (int, error) {
	res := tx.Model(&model.User{}).
		Where("team_name = ? AND user_id IN ?", teamName, userIDs).
		Update("is_active", false)
	return int(res.RowsAffected), res.Error
}

func (r *Repository) getActiveCandidates(tx *gorm.DB, teamName string) ([]model.User, error) {
	var candidates []model.User
	err := tx.Where("team_name = ? AND is_active = ?", teamName, true).Find(&candidates).Error
	return candidates, err
}

func (r *Repository) getAffectedPRs(tx *gorm.DB, userIDs []string) ([]affectedPR, error) {
	var rows []affectedPR
	err := tx.Table("pr_reviewers").
		Select("pr_reviewers.pull_request_id as pr_id, pull_requests.author_id").
		Joins("JOIN pull_requests ON pull_requests.pull_request_id = pr_reviewers.pull_request_id").
		Where("pr_reviewers.user_id IN ? AND pull_requests.status = ?", userIDs, "OPEN").
		Scan(&rows).Error
	return rows, err
}

func (r *Repository) calculateReplacements(rows []affectedPR, candidates []model.User) ([]prReviewer, int) {
	var newRelations []prReviewer
	reassignedCount := 0

	if len(candidates) == 0 {
		return newRelations, 0
	}

	for _, row := range rows {
		candidate := pickRandomCandidate(candidates, row.AuthorID)
		if candidate != nil {
			newRelations = append(newRelations, prReviewer{
				PullRequestID: row.PRID,
				UserID:        candidate.ID,
			})
			reassignedCount++
		}
	}
	return newRelations, reassignedCount
}

func (r *Repository) applyReviewerChanges(
	tx *gorm.DB,
	oldUserIDs []string,
	affected []affectedPR,
	newRelations []prReviewer,
) error {
	affectedPRIDs := make([]string, 0, len(affected))
	for _, r := range affected {
		affectedPRIDs = append(affectedPRIDs, r.PRID)
	}

	if err := tx.Table("pr_reviewers").
		Where("user_id IN ? AND pull_request_id IN ?", oldUserIDs, affectedPRIDs).
		Delete(nil).Error; err != nil {
		return err
	}

	if len(newRelations) > 0 {
		return tx.Table("pr_reviewers").
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&newRelations).Error
	}

	return nil
}

func pickRandomCandidate(candidates []model.User, excludeAuthorID string) *model.User {
	for _, c := range candidates {
		if c.ID != excludeAuthorID {
			return &c
		}
	}
	return nil
}
