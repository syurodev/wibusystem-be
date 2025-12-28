// Package router provides the application router setup and type definitions.
package router

import (
	"context"
	"system/internal/domain"
	analytics_module "system/internal/modules/analytics"
	artist_module "system/internal/modules/artist"
	auth_module "system/internal/modules/auth"
	author_module "system/internal/modules/author"
	creator_module "system/internal/modules/creator"
	"system/internal/modules/email"
	embedding_module "system/internal/modules/embedding"
	genre_module "system/internal/modules/genre"
	media_module "system/internal/modules/media"
	media_progress_module "system/internal/modules/media_progress"
	novel_module "system/internal/modules/novel"
	novel_chapter "system/internal/modules/novel_chapter"
	novel_volume "system/internal/modules/novel_volume"
	oauth2_module "system/internal/modules/oauth2"
	organization_module "system/internal/modules/organization"
	payment_module "system/internal/modules/payment"
	user_module "system/internal/modules/user"
	"system/internal/platform/cache"
	socket "system/internal/platform/socket"

	"github.com/ory/fosite"
)

// Repositories holds all repository instances
type Repositories struct {
	OAuth2Client       domain.OAuth2ClientRepository
	OAuth2Session      domain.OAuth2SessionRepository
	User               domain.UserRepository
	Session            domain.SessionRepository
	AuthRequest        domain.AuthRequestRepository
	Consent            domain.ConsentRepository
	EmailVerification  domain.EmailVerificationRepository
	PasswordReset      domain.PasswordResetRepository
	Genre              domain.GenreRepository
	Author             domain.AuthorRepository
	Artist             domain.ArtistRepository
	Novel              domain.NovelRepository
	Volume             domain.NovelVolumeRepository
	VolumeHistory      novel_volume.VolumeHistoryRepository
	Chapter            domain.NovelChapterRepository
	ChapterHistory     novel_chapter.ChapterHistoryRepository
	WebAuthnCredential domain.WebAuthnCredentialRepository
	WebAuthnSession    domain.WebAuthnSessionRepository
	ViewAnalytics      domain.ViewAnalyticsRepository
	ViewTracking       domain.ViewTrackingRepository
	Role               domain.RoleRepository
	Creator            domain.CreatorRepository
	Organization       domain.OrganizationRepository
	PaymentConfig      domain.PaymentConfigurationRepository
	Wallet             domain.WalletRepository
	CoinPackage        domain.CoinPackageRepository
	TopupOrder         domain.TopupOrderRepository
	Transaction        domain.TransactionRepository
	MediaProgress      domain.MediaProgressRepository
}

// Services holds all service instances
type Services struct {
	OAuth2         oauth2_module.OAuth2Service
	OAuth2Admin    oauth2_module.OAuth2AdminService
	Auth           auth_module.AuthService
	Email          email.EmailService
	Genre          genre_module.GenreService
	Author         author_module.AuthorService
	Artist         artist_module.ArtistService
	Novel          novel_module.NovelService
	Volume         novel_volume.VolumeService
	Chapter        novel_chapter.ChapterService
	ViewTracking   *analytics_module.ViewTrackingService
	Analytics      analytics_module.AnalyticsService
	Creator        creator_module.CreatorService
	WebAuthn       auth_module.WebAuthnService
	User           user_module.UserService
	Organization   organization_module.OrganizationService
	Embedding      *embedding_module.Service
	Cache          *cache.CacheService
	PaymentConfig  payment_module.ConfigUseCase
	Wallet         payment_module.WalletUseCase
	Topup          payment_module.TopupUseCase
	TransactionSvc payment_module.TransactionUseCase
	MediaProgress  media_progress_module.Service
	SocketHub      *socket.Hub
}

// Handlers holds all HTTP handler instances
type Handlers struct {
	OAuth2        *oauth2_module.Handler
	OAuth2Admin   *oauth2_module.AdminHandler
	Auth          *auth_module.Handler
	Genre         *genre_module.Handler
	Author        *author_module.Handler
	Artist        *artist_module.Handler
	Novel         *novel_module.Handler
	Volume        *novel_volume.Handler
	Chapter       *novel_chapter.Handler
	User          *user_module.Handler
	Creator       *creator_module.Handler
	Organization  *organization_module.Handler
	Embedding     *embedding_module.Handler
	PaymentConfig *payment_module.Handler
	Wallet        *payment_module.WalletHandler
	Webhook       *payment_module.WebhookHandler
	Socket        *socket.Handler
	MediaProgress *media_progress_module.Handler
	Media         *media_module.Handler
}

// Dependencies holds all application dependencies (used by router)
type Dependencies struct {
	Repositories   *Repositories
	Services       *Services
	Handlers       *Handlers
	OAuth2Provider fosite.OAuth2Provider
}

// TopNovelAdapter adapts AnalyticsService to novel.TopNovelService interface
type TopNovelAdapter struct {
	analytics analytics_module.AnalyticsService
}

// NewTopNovelAdapter creates a new TopNovelAdapter
func NewTopNovelAdapter(analytics analytics_module.AnalyticsService) *TopNovelAdapter {
	return &TopNovelAdapter{analytics: analytics}
}

// GetTopNovelsWithRank implements novel.TopNovelService interface
func (a *TopNovelAdapter) GetTopNovelsWithRank(ctx context.Context, period string, offset int, limit int) ([]novel_module.MediaRankResult, error) {
	// Get top novels using analytics service
	results, err := a.analytics.GetTopMediaWithRankComparison(ctx, period, "novel", offset, limit)
	if err != nil {
		return nil, err
	}

	mediaResults := make([]novel_module.MediaRankResult, 0, len(results))
	for _, r := range results {
		mediaResults = append(mediaResults, novel_module.MediaRankResult{
			Novel: r.Novel,
			Stats: novel_module.MediaRankStat{
				EntityID:     r.Stats.EntityID,
				TotalViews:   int(r.Stats.TotalViews),
				CurrentRank:  r.Stats.CurrentRank,
				PreviousRank: r.Stats.PreviousRank,
				RankChange:   r.Stats.RankChange,
			},
		})
	}
	return mediaResults, nil
}
