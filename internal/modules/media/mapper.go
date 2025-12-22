package media

import (
	"system/internal/domain"
	res "system/internal/dto/media"
	analytics_module "system/internal/modules/analytics"
)

// mapNovelToMediaSeriesResponse converts domain.Novel to MediaSeriesResponse
func mapNovelToMediaSeriesResponse(n *domain.Novel) *res.MediaSeriesResponse {
	if n == nil {
		return nil
	}

	// Build owner info
	owner := res.OwnerInfo{
		ID:          n.OwnerID.String(),
		DisplayName: "",
		Username:    "",
	}
	if n.OwnerDisplayName != nil {
		owner.DisplayName = *n.OwnerDisplayName
	}
	if n.OwnerUsername != nil {
		owner.Username = *n.OwnerUsername
	}
	if n.OwnerAvatarURL != nil {
		owner.AvatarURL = n.OwnerAvatarURL
	}

	// Build genres
	genres := make([]res.GenreInfo, 0, len(n.Genres))
	for _, g := range n.Genres {
		genres = append(genres, res.GenreInfo{
			ID:   g.ID.String(),
			Name: g.Name,
			Slug: g.Slug,
		})
	}

	// Get cover URL
	var coverURL *string
	if n.CoverImageURL != nil {
		coverURL = n.CoverImageURL
	}

	return &res.MediaSeriesResponse{
		ID:               n.ID.String(),
		Title:            n.Title,
		OriginalTitle:    n.OriginalTitle,
		Slug:             n.Slug,
		OriginalLanguage: n.OriginalLanguage,
		Synopsis:         n.Synopsis,
		CoverURL:         coverURL,
		Type:             domain.MediaTypeNovel,
		Status:           string(n.Status),
		Genres:           genres,
		Owner:            owner,
		Rating:           n.RatingAverage,
		Views:            n.ViewCount,
		Favorites:        n.FavoriteCount,
		LatestChapter:    nil, // Can be populated if needed
		CreatedAt:        n.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        n.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// mapMediaRankToResponse converts MediaRankResponse to MediaSeriesResponse
// Uses the full Novel data if available, otherwise falls back to basic fields
func mapMediaRankToResponse(rank *analytics_module.MediaRankResponse) *res.MediaSeriesResponse {
	if rank == nil {
		return nil
	}

	// If we have full Novel data, use it for complete response
	if rank.Novel != nil {
		resp := mapNovelToMediaSeriesResponse(rank.Novel)
		if resp != nil {
			// Override views with analytics data (more accurate for the period)
			resp.Views = int64(rank.Stats.TotalViews)
			
			// Add rank info
			currentRank := rank.Stats.CurrentRank
			resp.CurrentRank = &currentRank
			resp.PreviousRank = rank.Stats.PreviousRank
			resp.RankChange = rank.Stats.RankChange
		}
		return resp
	}

	// Fallback: Build basic response from rank data
	owner := res.OwnerInfo{
		ID:          "",
		DisplayName: "",
		Username:    "",
	}

	genres := make([]res.GenreInfo, 0)

	var coverURL *string
	if rank.Cover != "" {
		coverURL = &rank.Cover
	}

	currentRank := rank.Stats.CurrentRank

	return &res.MediaSeriesResponse{
		ID:            rank.ID.String(),
		Title:         rank.Title,
		Slug:          rank.Slug,
		CoverURL:      coverURL,
		Type:          rank.Type,
		Status:        "",
		Genres:        genres,
		Owner:         owner,
		Rating:        0,
		Views:         int64(rank.Stats.TotalViews),
		Favorites:     0,
		LatestChapter: nil,
		CreatedAt:     "",
		UpdatedAt:     "",
		CurrentRank:   &currentRank,
		PreviousRank:  rank.Stats.PreviousRank,
		RankChange:    rank.Stats.RankChange,
	}
}
