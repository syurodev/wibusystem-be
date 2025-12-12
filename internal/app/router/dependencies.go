package router

import (
	"system/configs"
	"system/internal/app/worker"
	"system/internal/domain"
	analytics_module "system/internal/modules/analytics"
	artist_module "system/internal/modules/artist"
	auth_module "system/internal/modules/auth"
	author_module "system/internal/modules/author"
	creator_module "system/internal/modules/creator"
	"system/internal/modules/email"
	genre_module "system/internal/modules/genre"
	media_module "system/internal/modules/media"
	novel_module "system/internal/modules/novel"
	novel_chapter "system/internal/modules/novel_chapter"
	novel_volume "system/internal/modules/novel_volume"
	oauth2_module "system/internal/modules/oauth2"
	organization_module "system/internal/modules/organization"
	user_module "system/internal/modules/user"
	"system/internal/platform/cache"
	"system/internal/platform/database"
	txdb "system/internal/platform/database"
	"system/internal/platform/oauth2"
	fosite_storage "system/internal/platform/oauth2/storage"
	"system/internal/platform/resend"

	"github.com/ory/fosite"
	"go.uber.org/zap"
)

// Repositories chứa tất cả repository instances
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
}

// Services chứa tất cả service instances
type Services struct {
	OAuth2       oauth2_module.OAuth2Service
	OAuth2Admin  oauth2_module.OAuth2AdminService
	Auth         auth_module.AuthService
	Email        email.EmailService // Change from *resend.Service to interface
	Genre        genre_module.GenreService

	Author       author_module.AuthorService
	Artist       artist_module.ArtistService
	Novel        novel_module.NovelService
	Volume       novel_volume.VolumeService
	Chapter      novel_chapter.ChapterService
	ViewTracking *analytics_module.ViewTrackingService
	Analytics    analytics_module.AnalyticsService
	Creator      creator_module.CreatorService
	Cache        *cache.CacheService
	Media        media_module.MediaService
	WebAuthn     auth_module.WebAuthnService
	User         user_module.UserService
	Organization organization_module.OrganizationService
}

// Handlers chứa tất cả handler instances
type Handlers struct {
	OAuth2      *oauth2_module.Handler
	OAuth2Admin *oauth2_module.AdminHandler
	Auth        *auth_module.Handler
	Genre       *genre_module.Handler
	Author      *author_module.Handler
	Artist      *artist_module.Handler
	Novel       *novel_module.Handler
	Volume      *novel_volume.Handler
	Chapter     *novel_chapter.Handler
	User        *user_module.Handler
	Media       *media_module.Handler
	Creator     *creator_module.Handler
	Organization *organization_module.Handler
}

// Dependencies chứa tất cả dependencies của ứng dụng
type Dependencies struct {
	Repos          *Repositories
	Services       *Services
	Handlers       *Handlers
	OAuth2Provider fosite.OAuth2Provider
}

// NewDependencies khởi tạo tất cả dependencies
func NewDependencies(
	cfg *configs.Config,
	db *database.PostgresDB,
	rdb *database.RedisClient,
	ch *database.ClickHouseClient,
	zapLogger *zap.Logger,
) (*Dependencies, error) {
	// --- Repositories ---
	repos := newRepositories(db, rdb, ch)

	// --- Fosite Storage & Provider ---
	sqlStore := fosite_storage.NewSQLStore(repos.OAuth2Client, repos.OAuth2Session)
	redisStore := fosite_storage.NewRedisStore(rdb, repos.OAuth2Client, zapLogger)
	hybridStore := fosite_storage.NewHybridStore(sqlStore, redisStore)
	oauth2Provider := oauth2.NewOAuth2Provider(hybridStore, &cfg.OAuth2)

	// --- Services ---
	services, err := newServices(cfg, repos, rdb, db, zapLogger)
	if err != nil {
		return nil, err
	}

	// Start View Tracking Workers
	worker.StartViewTrackingWorkers(&cfg.ViewTracking, services.ViewTracking, zapLogger)

	// --- Handlers ---
	handlers := newHandlers(cfg, repos, services, oauth2Provider, zapLogger)

	return &Dependencies{
		Repos:          repos,
		Services:       services,
		Handlers:       handlers,
		OAuth2Provider: oauth2Provider,
	}, nil
}

// newRepositories khởi tạo tất cả repositories
func newRepositories(db *database.PostgresDB, rdb *database.RedisClient, ch *database.ClickHouseClient) *Repositories {
	oauth2ClientRepo := oauth2_module.NewOAuth2ClientRepository(db.Pool)

	return &Repositories{
		OAuth2Client:       oauth2ClientRepo,
		OAuth2Session:      oauth2_module.NewOAuth2SessionRepository(db.Pool),
		User:               user_module.NewUserRepository(db.Pool),
		Session:            user_module.NewSessionRepository(rdb),
		AuthRequest:        oauth2_module.NewAuthRequestRepository(rdb, oauth2ClientRepo),
		Consent:            oauth2_module.NewConsentRepository(db.Pool),
		EmailVerification:  auth_module.NewEmailVerificationRepository(db.Pool),
		PasswordReset:      auth_module.NewPasswordResetRepository(db.Pool),
		Genre:              genre_module.NewGenreRepository(db.Pool),
		Author:             author_module.NewAuthorRepository(db.Pool),
		Artist:             artist_module.NewArtistRepository(db.Pool),
		Novel:              novel_module.NewNovelRepository(db.Pool),
		Volume:             novel_volume.NewVolumeRepository(db.Pool),
		VolumeHistory:      novel_volume.NewVolumeHistoryRepository(db.Pool),
		Chapter:            novel_chapter.NewChapterRepository(db.Pool),
		ChapterHistory:     novel_chapter.NewChapterHistoryRepository(db.Pool),
		WebAuthnCredential: auth_module.NewWebAuthnCredentialRepository(db.Pool),
		WebAuthnSession:    auth_module.NewWebAuthnSessionRepository(db.Pool),
		ViewAnalytics:      analytics_module.NewViewAnalyticsClickHouseRepository(ch),
		ViewTracking:       analytics_module.NewViewTrackingRedisRepository(rdb),
		Role:               user_module.NewRoleRepository(db.Pool),
		Creator:            creator_module.NewCreatorRepository(db.Pool),
		Organization:       organization_module.NewRepository(db.Pool),
	}
}

// newServices khởi tạo tất cả services
func newServices(
	cfg *configs.Config,
	repos *Repositories,
	rdb *database.RedisClient,
	db *database.PostgresDB,
	zapLogger *zap.Logger,
) (*Services, error) {
	txManager := txdb.NewTransactionManager(db.Pool)

	// User Service (needed by Auth and OAuth2)
	userSvc := user_module.NewService(repos.User, repos.Session)


	oauth2Svc := oauth2_module.NewService(
		userSvc,
		repos.AuthRequest,
		repos.Consent,
		repos.OAuth2Session,
		repos.OAuth2Client,
	)

	authSvc := auth_module.NewService(
		userSvc,
		repos.EmailVerification,
		repos.PasswordReset,
		repos.Role,
	)

	// Email chain
	resendClient := resend.NewClient(&cfg.Email, zapLogger)
	emailSvc := email.NewService(resendClient, &cfg.Email)

	oauth2AdminSvc := oauth2_module.NewAdminService(repos.OAuth2Client)

	// Genre module - fully migrated to modules/genre
	genreRepo := genre_module.NewGenreRepository(db.Pool)
	genreSvc := genre_module.NewService(genreRepo)

	authorSvc := author_module.NewService(repos.Author)
	artistSvc := artist_module.NewService(repos.Artist)
	novelSvc := novel_module.NewService(
		repos.Novel, repos.Volume, repos.Genre,
		repos.Author, repos.Artist, repos.Creator, txManager,
	)
	volumeSvc := novel_volume.NewService(repos.Volume, repos.VolumeHistory)
	chapterSvc := novel_chapter.NewService(
		repos.Chapter, repos.Volume, repos.ChapterHistory, repos.Creator,
	)
	
	// Analytics & Tracking
	viewTrackingSvc := analytics_module.NewViewTrackingService(
		repos.ViewTracking, repos.ViewAnalytics,
		repos.Chapter, repos.Novel, repos.Genre,
		zapLogger, &cfg.ViewTracking,
	)
	analyticsSvc := analytics_module.NewService(repos.ViewAnalytics, novelSvc, zapLogger)
	// Creator
	creatorSvc := creator_module.NewService(
		repos.Creator,
		analyticsSvc,
		novelSvc,
		zapLogger,
	)

	cacheSvc := cache.NewCacheService(rdb, zapLogger)
	mediaSvc := media_module.NewMediaService(analyticsSvc, creatorSvc, cacheSvc, zapLogger)

	webauthnSvc, err := auth_module.NewWebAuthnService(
		cfg.WebAuthn,
		repos.WebAuthnCredential,
		repos.WebAuthnSession,
		repos.User,
		zapLogger,
	)
	if err != nil {
		return nil, err
	}



	return &Services{
		OAuth2:       oauth2Svc,
		OAuth2Admin:  oauth2AdminSvc,
		Auth:         authSvc,
		Email:        emailSvc,
		Genre:        genreSvc,
		Author:       authorSvc,
		Artist:       artistSvc,
		Novel:        novelSvc,
		Volume:       volumeSvc,
		Chapter:      chapterSvc,
		ViewTracking: viewTrackingSvc,
		Analytics:    analyticsSvc,
		Creator:      creatorSvc,
		Cache:        cacheSvc,
		Media:        mediaSvc,
		WebAuthn:     webauthnSvc,
		User:         userSvc,
		Organization: organization_module.NewService(repos.Organization),
	}, nil
}

// newHandlers khởi tạo tất cả handlers
func newHandlers(
	cfg *configs.Config,
	repos *Repositories,
	services *Services,
	oauth2Provider fosite.OAuth2Provider,
	zapLogger *zap.Logger,
) *Handlers {
	oauth2Handler := oauth2_module.NewHandler(
		&cfg.OAuth2,
		oauth2Provider,
		services.OAuth2,
		repos.AuthRequest,
		repos.OAuth2Client,
		services.WebAuthn,
		zapLogger,
	)

	return &Handlers{
		OAuth2:      oauth2Handler,
		OAuth2Admin: oauth2_module.NewAdminHandler(services.OAuth2Admin),
		Auth:        auth_module.NewHandler(services.Auth, services.Email, services.WebAuthn, services.OAuth2),
		Genre:       genre_module.NewHandler(services.Genre, zapLogger),
		Author:      author_module.NewHandler(services.Author, zapLogger),
		Artist:      artist_module.NewHandler(services.Artist),
		Novel:       novel_module.NewHandler(services.Novel, services.Volume, services.Chapter),
		Volume:      novel_volume.NewHandler(services.Volume),
		Chapter:     novel_chapter.NewHandler(services.Chapter),
		User:        user_module.NewHandler(services.User),
		Media:       media_module.NewHandler(services.Analytics, services.Media),
		Creator:     creator_module.NewHandler(services.Creator),
		Organization: organization_module.NewHandler(services.Organization, zapLogger),
	}
}
