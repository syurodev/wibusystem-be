package novel_chapter

import (
	"context"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/translationcontribution"
)

// translationContributionEntRepository implements TranslationContributionRepository using Ent
type translationContributionEntRepository struct {
	client *ent.Client
}

// NewTranslationContributionEntRepository creates a new contribution repository using Ent
func NewTranslationContributionEntRepository(client *ent.Client) domain.TranslationContributionRepository {
	return &translationContributionEntRepository{client: client}
}

// GetByID gets a contribution by ID
func (r *translationContributionEntRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TranslationContribution, error) {
	c, err := database.GetClientFromContext(ctx, r.client).TranslationContribution.Query().
		Where(
			translationcontribution.IDEQ(id),
			translationcontribution.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entContributionToDomain(c), nil
}

// GetByChapterID gets contributions for a chapter
func (r *translationContributionEntRepository) GetByChapterID(ctx context.Context, chapterID uuid.UUID, filter domain.ContributionFilter) ([]*domain.TranslationContribution, error) {
	query := database.GetClientFromContext(ctx, r.client).TranslationContribution.Query().
		Where(
			translationcontribution.ChapterIDEQ(chapterID),
			translationcontribution.DeletedAtIsNil(),
		)

	if filter.Language != nil {
		query = query.Where(translationcontribution.LanguageEQ(*filter.Language))
	}
	if filter.Status != nil {
		query = query.Where(translationcontribution.StatusEQ(string(*filter.Status)))
	}
	if filter.ContributionType != nil {
		query = query.Where(translationcontribution.ContributionTypeEQ(string(*filter.ContributionType)))
	}

	orderField := translationcontribution.FieldCreatedAt
	if filter.SortBy == "upvotes" {
		orderField = translationcontribution.FieldUpvoteCount
	} else if filter.SortBy == "credit_points" {
		orderField = translationcontribution.FieldCreditPoints
	}

	if filter.SortOrder == "asc" {
		query = query.Order(ent.Asc(orderField))
	} else {
		query = query.Order(ent.Desc(orderField))
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	contributions, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.TranslationContribution, len(contributions))
	for i, c := range contributions {
		result[i] = entContributionToDomain(c)
	}
	return result, nil
}

// GetByContributorID gets contributions by contributor
func (r *translationContributionEntRepository) GetByContributorID(ctx context.Context, contributorID uuid.UUID, filter domain.ContributionFilter) ([]*domain.TranslationContribution, error) {
	query := database.GetClientFromContext(ctx, r.client).TranslationContribution.Query().
		Where(
			translationcontribution.ContributorIDEQ(contributorID),
			translationcontribution.DeletedAtIsNil(),
		)

	if filter.Language != nil {
		query = query.Where(translationcontribution.LanguageEQ(*filter.Language))
	}
	if filter.Status != nil {
		query = query.Where(translationcontribution.StatusEQ(string(*filter.Status)))
	}

	orderField := translationcontribution.FieldCreatedAt
	if filter.SortOrder == "asc" {
		query = query.Order(ent.Asc(orderField))
	} else {
		query = query.Order(ent.Desc(orderField))
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	contributions, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.TranslationContribution, len(contributions))
	for i, c := range contributions {
		result[i] = entContributionToDomain(c)
	}
	return result, nil
}

// GetPendingReview gets contributions pending review
func (r *translationContributionEntRepository) GetPendingReview(ctx context.Context, language *string, limit, offset int) ([]*domain.TranslationContribution, error) {
	query := database.GetClientFromContext(ctx, r.client).TranslationContribution.Query().
		Where(
			translationcontribution.StatusEQ(string(domain.TranslationStatusPendingReview)),
			translationcontribution.DeletedAtIsNil(),
		).
		Order(ent.Asc(translationcontribution.FieldCreatedAt))

	if language != nil {
		query = query.Where(translationcontribution.LanguageEQ(*language))
	}

	query = query.Limit(limit).Offset(offset)

	contributions, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.TranslationContribution, len(contributions))
	for i, c := range contributions {
		result[i] = entContributionToDomain(c)
	}
	return result, nil
}

// Create creates a new contribution
func (r *translationContributionEntRepository) Create(ctx context.Context, contribution *domain.TranslationContribution) error {
	builder := database.GetClientFromContext(ctx, r.client).TranslationContribution.Create().
		SetChapterID(contribution.ChapterID).
		SetContributorID(contribution.ContributorID).
		SetLanguage(contribution.Language).
		SetContributionType(string(contribution.ContributionType)).
		SetContent(contribution.Content). // string
		SetStatus(string(contribution.Status)).
		SetWordCount(contribution.WordCount).
		SetCharacterCount(contribution.CharacterCount).
		SetCreditPoints(contribution.CreditPoints).
		SetIsCredited(contribution.IsCredited)

	if contribution.ID != uuid.Nil {
		builder.SetID(contribution.ID)
	}
	if contribution.Title != nil {
		builder.SetTitle(*contribution.Title)
	}
	if contribution.ContributorNotes != nil {
		builder.SetContributorNotes(*contribution.ContributorNotes)
	}

	_, err := builder.Save(ctx)
	return err
}

// Update updates a contribution
func (r *translationContributionEntRepository) Update(ctx context.Context, contribution *domain.TranslationContribution) error {
	builder := database.GetClientFromContext(ctx, r.client).TranslationContribution.UpdateOneID(contribution.ID).
		SetContent(contribution.Content). // string
		SetStatus(string(contribution.Status)).
		SetWordCount(contribution.WordCount).
		SetCharacterCount(contribution.CharacterCount)

	if contribution.Title != nil {
		builder.SetTitle(*contribution.Title)
	} else {
		builder.ClearTitle()
	}
	if contribution.ContributorNotes != nil {
		builder.SetContributorNotes(*contribution.ContributorNotes)
	} else {
		builder.ClearContributorNotes()
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete soft deletes a contribution
func (r *translationContributionEntRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := database.GetClientFromContext(ctx, r.client).TranslationContribution.UpdateOneID(id).
		SetDeletedAt(now).
		Save(ctx)
	return err
}

// Approve approves a contribution
func (r *translationContributionEntRepository) Approve(ctx context.Context, id, reviewerID uuid.UUID, reviewNotes string, creditPoints int) error {
	now := time.Now()
	_, err := database.GetClientFromContext(ctx, r.client).TranslationContribution.UpdateOneID(id).
		SetStatus(string(domain.TranslationStatusApproved)).
		SetReviewedBy(reviewerID).
		SetReviewedAt(now).
		SetReviewNotes(reviewNotes).
		SetCreditPoints(creditPoints).
		SetIsCredited(true).
		Save(ctx)
	return err
}

// Reject rejects a contribution
func (r *translationContributionEntRepository) Reject(ctx context.Context, id, reviewerID uuid.UUID, reviewNotes string) error {
	now := time.Now()
	_, err := database.GetClientFromContext(ctx, r.client).TranslationContribution.UpdateOneID(id).
		SetStatus(string(domain.TranslationStatusRejected)).
		SetReviewedBy(reviewerID).
		SetReviewedAt(now).
		SetReviewNotes(reviewNotes).
		Save(ctx)
	return err
}

// Vote votes on a contribution
func (r *translationContributionEntRepository) Vote(ctx context.Context, contributionID, userID uuid.UUID, isUpvote bool) error {
	builder := database.GetClientFromContext(ctx, r.client).TranslationContribution.UpdateOneID(contributionID)
	if isUpvote {
		builder.AddUpvoteCount(1)
	} else {
		builder.AddDownvoteCount(1)
	}

	_, err := builder.Save(ctx)
	return err
}

// UpdateStatistics updates statistics
func (r *translationContributionEntRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.ContributionStatisticsUpdate) error {
	builder := database.GetClientFromContext(ctx, r.client).TranslationContribution.UpdateOneID(id)

	if stats.UpvoteCount != nil {
		builder.SetUpvoteCount(*stats.UpvoteCount)
	}
	if stats.DownvoteCount != nil {
		builder.SetDownvoteCount(*stats.DownvoteCount)
	}
	if stats.CreditPoints != nil {
		builder.SetCreditPoints(*stats.CreditPoints)
	}

	_, err := builder.Save(ctx)
	return err
}

// Helper to convert Ent entity to domain model
func entContributionToDomain(c *ent.TranslationContribution) *domain.TranslationContribution {
	return &domain.TranslationContribution{
		ID:                    c.ID,
		ChapterID:             c.ChapterID,
		ContributorID:         c.ContributorID,
		Language:              c.Language,
		ContributionType:      domain.ContributionType(c.ContributionType),
		Title:                 c.Title,
		Content:               c.Content, // string
		ContributorNotes:      c.ContributorNotes,
		Status:                domain.TranslationStatus(c.Status),
		ReviewedBy:            c.ReviewedBy,
		ReviewedAt:            c.ReviewedAt,
		ReviewNotes:           c.ReviewNotes,
		OfficialTranslationID: c.OfficialTranslationID,
		CreditPoints:          c.CreditPoints,
		IsCredited:            c.IsCredited,
		WordCount:             c.WordCount,
		CharacterCount:        c.CharacterCount,
		UpvoteCount:           c.UpvoteCount,
		DownvoteCount:         c.DownvoteCount,
		CreatedAt:             c.CreatedAt,
		UpdatedAt:             c.UpdatedAt,
		DeletedAt:             c.DeletedAt,
	}
}
