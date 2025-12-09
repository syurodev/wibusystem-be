package router

import (
	"system/configs"
	"system/internal/app/handler/v1/artist"
	"system/internal/app/handler/v1/auth"
	"system/internal/app/handler/v1/author"
	"system/internal/app/handler/v1/creator"
	"system/internal/app/handler/v1/genre"
	"system/internal/app/handler/v1/media"
	"system/internal/app/handler/v1/novel"
	novel_volume "system/internal/app/handler/v1/novel/volume"
	volume_chapter "system/internal/app/handler/v1/novel/volume/chapter"
	oauth2_handler "system/internal/app/handler/v1/oauth2"
	"system/internal/app/handler/v1/oauth2_admin"
	"system/internal/app/handler/v1/public"
	"system/internal/app/handler/v1/user"
	"system/internal/app/worker"
	"system/internal/domain"
	"system/internal/oauth2"
	fosite_storage "system/internal/oauth2/storage"
	txdb "system/internal/pkg/db"
	"system/internal/pkg/repository"
	"system/internal/pkg/service"
	"system/internal/platform/database"

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
	Volume             domain.VolumeRepository
	VolumeHistory      service.VolumeHistoryRepository
	Chapter            domain.ChapterRepository
	ChapterHistory     service.ChapterHistoryRepository
	WebAuthnCredential domain.WebAuthnCredentialRepository
	WebAuthnSession    domain.WebAuthnSessionRepository
	ViewAnalytics      domain.ViewAnalyticsRepository
	ViewTracking       domain.ViewTrackingRepository
	Role               domain.RoleRepository
	Creator            domain.CreatorRepository
}

// Services chứa tất cả service instances
type Services struct {
	OAuth2       *service.OAuth2Service
	OAuth2Admin  *service.OAuth2AdminService
	Auth         *service.AuthService
	Email        *service.EmailService
	Genre        *service.GenreService
	Author       *service.AuthorService
	Artist       *service.ArtistService
	Novel        *service.NovelService
	Volume       *service.VolumeService
	Chapter      *service.ChapterService
	ViewTracking *service.ViewTrackingService
	Analytics    *service.AnalyticsService
	Creator      *service.CreatorService
	Cache        *service.CacheService
	Public       *service.PublicService
	WebAuthn     service.WebAuthnService
}

// Handlers chứa tất cả handler instances
type Handlers struct {
	OAuth2      *oauth2_handler.Handler
	OAuth2Admin *oauth2_admin.Handler
	Auth        *auth.Handler
	Genre       *genre.Handler
	Author      *author.Handler
	Artist      *artist.Handler
	Novel       *novel.Handler
	Volume      *novel_volume.Handler
	Chapter     *volume_chapter.Handler
	User        *user.Handler
	Media       *media.Handler
	Creator     *creator.Handler
	Public      *public.Handler
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
	oauth2ClientRepo := repository.NewOAuth2ClientRepository(db.Pool)

	return &Repositories{
		OAuth2Client:       oauth2ClientRepo,
		OAuth2Session:      repository.NewOAuth2SessionRepository(db.Pool),
		User:               repository.NewUserRepository(db.Pool),
		Session:            repository.NewSessionRepository(rdb),
		AuthRequest:        repository.NewAuthRequestRepository(rdb, oauth2ClientRepo),
		Consent:            repository.NewConsentRepository(db.Pool),
		EmailVerification:  repository.NewEmailVerificationRepository(db.Pool),
		PasswordReset:      repository.NewPasswordResetRepository(db.Pool),
		Genre:              repository.NewGenreRepository(db.Pool),
		Author:             repository.NewAuthorRepository(db.Pool),
		Artist:             repository.NewArtistRepository(db.Pool),
		Novel:              repository.NewNovelRepository(db.Pool),
		Volume:             repository.NewVolumeRepository(db.Pool),
		VolumeHistory:      repository.NewVolumeHistoryRepository(db.Pool),
		Chapter:            repository.NewChapterRepository(db.Pool),
		ChapterHistory:     repository.NewChapterHistoryRepository(db.Pool),
		WebAuthnCredential: repository.NewWebAuthnCredentialRepository(db.Pool),
		WebAuthnSession:    repository.NewWebAuthnSessionRepository(db.Pool),
		ViewAnalytics:      repository.NewViewAnalyticsClickHouseRepository(ch),
		ViewTracking:       repository.NewViewTrackingRedisRepository(rdb),
		Role:               repository.NewRoleRepository(db.Pool),
		Creator:            repository.NewCreatorRepository(db.Pool),
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

	oauth2Svc := service.NewOAuth2Service(
		repos.User,
		repos.Session,
		repos.AuthRequest,
		repos.Consent,
		repos.OAuth2Session,
		repos.OAuth2Client,
	)

	authSvc := service.NewAuthService(
		repos.User,
		repos.EmailVerification,
		repos.PasswordReset,
		repos.Role,
	)

	emailSvc := service.NewEmailService(&cfg.Email, zapLogger)
	oauth2AdminSvc := service.NewOAuth2AdminService(repos.OAuth2Client)

	genreSvc := service.NewGenreService(repos.Genre)
	authorSvc := service.NewAuthorService(repos.Author)
	artistSvc := service.NewArtistService(repos.Artist)
	novelSvc := service.NewNovelService(
		repos.Novel, repos.Volume, repos.Genre,
		repos.Author, repos.Artist, repos.Creator, txManager,
	)
	volumeSvc := service.NewVolumeService(repos.Volume, repos.VolumeHistory)
	chapterSvc := service.NewChapterService(
		repos.Chapter, repos.Volume, repos.ChapterHistory, repos.Creator,
	)
	viewTrackingSvc := service.NewViewTrackingService(
		repos.ViewTracking, repos.ViewAnalytics,
		repos.Chapter, repos.Novel, repos.Genre,
		zapLogger, &cfg.ViewTracking,
	)
	analyticsSvc := service.NewAnalyticsService(repos.ViewAnalytics, repos.Novel, zapLogger)
	creatorSvc := service.NewCreatorService(repos.Creator, repos.ViewAnalytics, repos.Novel, zapLogger)

	cacheSvc := service.NewCacheService(rdb, zapLogger)
	publicSvc := service.NewPublicService(analyticsSvc, creatorSvc, genreSvc, cacheSvc, zapLogger)

	webauthnSvc, err := service.NewWebAuthnService(
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
		Public:       publicSvc,
		WebAuthn:     webauthnSvc,
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
	oauth2Handler := oauth2_handler.NewHandler(
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
		OAuth2Admin: oauth2_admin.NewHandler(services.OAuth2Admin),
		Auth:        auth.NewHandler(services.Auth, services.Email, services.WebAuthn, services.OAuth2),
		Genre:       genre.NewHandler(services.Genre, zapLogger),
		Author:      author.NewHandler(services.Author, zapLogger),
		Artist:      artist.NewHandler(services.Artist),
		Novel:       novel.NewHandler(services.Novel, services.Volume, services.Chapter),
		Volume:      novel_volume.NewHandler(services.Volume, services.Novel),
		Chapter:     volume_chapter.NewHandler(services.Chapter, services.Volume, services.ViewTracking, services.Novel),
		User:        user.NewHandler(repos.User, repos.Session),
		Media:       media.NewHandler(services.Analytics),
		Creator:     creator.NewHandler(services.Creator),
		Public:      public.NewHandler(services.Public),
	}
}
