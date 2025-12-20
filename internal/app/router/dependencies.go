package router

import (
	"context"
	"system/configs"
	"system/internal/app/worker"
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
	novel_module "system/internal/modules/novel"
	novel_chapter "system/internal/modules/novel_chapter"
	novel_volume "system/internal/modules/novel_volume"
	oauth2_module "system/internal/modules/oauth2"
	organization_module "system/internal/modules/organization"
	payment_module "system/internal/modules/payment"
	user_module "system/internal/modules/user"
	"system/internal/platform/cache"
	"system/internal/platform/database"
	txdb "system/internal/platform/database"
	"system/internal/platform/oauth2"
	fosite_storage "system/internal/platform/oauth2/storage"
	"system/internal/platform/resend"
	socket "system/internal/platform/socket"

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
	PaymentConfig      domain.PaymentConfigurationRepository
	Wallet             domain.WalletRepository
	CoinPackage        domain.CoinPackageRepository
	TopupOrder         domain.TopupOrderRepository
	Transaction        domain.TransactionRepository
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
	Embedding    *embedding_module.Service
	PaymentConfig payment_module.ConfigUseCase
	Wallet        payment_module.WalletUseCase
	Topup         payment_module.TopupUseCase
	TransactionSvc payment_module.TransactionUseCase
	SocketHub     *socket.Hub
}

// Handlers chứa tất cả handler instances
type Handlers struct {
	OAuth2       *oauth2_module.Handler
	OAuth2Admin  *oauth2_module.AdminHandler
	Auth         *auth_module.Handler
	Genre        *genre_module.Handler
	Author       *author_module.Handler
	Artist       *artist_module.Handler
	Novel        *novel_module.Handler
	Volume       *novel_volume.Handler
	Chapter      *novel_chapter.Handler
	User         *user_module.Handler
	Media        *media_module.Handler
	Creator      *creator_module.Handler
	Organization *organization_module.Handler
	Embedding    *embedding_module.Handler
	PaymentConfig *payment_module.Handler
	Wallet        *payment_module.WalletHandler
	Webhook       *payment_module.WebhookHandler
	Socket        *socket.Handler
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

	// Start Embedding Worker
	embeddingWorker := worker.NewEmbeddingWorker(services.Embedding, services.Novel, zapLogger, &cfg.Embedding)
	embeddingWorker.Start(context.Background())

	// --- Transaction Manager for Use Cases ---
	txManager := txdb.NewTransactionManager(db.Pool)

	// --- Handlers ---
	handlers := newHandlers(cfg, repos, services, oauth2Provider, txManager, zapLogger)

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
		PaymentConfig:      payment_module.NewConfigurationRepository(db.Pool),
		Wallet:             payment_module.NewWalletRepository(db.Pool),
		CoinPackage:        payment_module.NewCoinPackageRepository(db.Pool),
		TopupOrder:         payment_module.NewTopupOrderRepository(db.Pool),
		Transaction:        payment_module.NewTransactionRepository(db.Pool),
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
		repos.Chapter, repos.Volume, repos.ChapterHistory, repos.Creator, repos.ViewTracking, repos.User,
	)
	
	// Analytics & Tracking
	viewTrackingSvc := analytics_module.NewViewTrackingService(
		repos.ViewTracking, repos.ViewAnalytics,
		repos.Chapter, repos.Novel, repos.Genre,
		zapLogger, &cfg.ViewTracking,
	)
	analyticsSvc := analytics_module.NewService(
		repos.ViewAnalytics,
		novelSvc,
		repos.Creator,
		repos.Organization,
		repos.Chapter,
		repos.User,
		genreRepo,
		rdb,
		zapLogger,
	)
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



	// Embedding Service
	embeddingRepo := embedding_module.NewRepository(db.Pool)
	embedder := embedding_module.NewNoopEmbedder(cfg.Embedding.Dimensions)
	embeddingSvc := embedding_module.NewService(embeddingRepo, embedder, rdb.Client, zapLogger, &cfg.Embedding)

	// Payment Config Service
	paymentConfigUC := payment_module.NewConfigUseCase(repos.PaymentConfig)

	// Socket Hub
	socketHub := socket.NewHub(zapLogger)
	go socketHub.Run()

	// Wallet Services
	walletUC := payment_module.NewWalletUseCase(repos.Wallet, zapLogger)
	transactionUC := payment_module.NewTransactionUseCase(repos.Transaction, zapLogger)
	topupUC := payment_module.NewTopupUseCase(
		db.Pool,
		repos.Wallet,
		repos.CoinPackage,
		repos.TopupOrder,
		repos.Transaction,
		paymentConfigUC,
		socketHub,
		zapLogger,
	)

	// Start Payment Workers
	worker.StartPaymentWorkers(topupUC, zapLogger)

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
		Embedding:    embeddingSvc,
		PaymentConfig: paymentConfigUC,
		Wallet:        walletUC,
		Topup:         topupUC,
		TransactionSvc: transactionUC,
		SocketHub:     socketHub,
	}, nil
}

// newHandlers khởi tạo tất cả handlers
func newHandlers(
	cfg *configs.Config,
	repos *Repositories,
	services *Services,
	oauth2Provider fosite.OAuth2Provider,
	txManager txdb.TransactionManager,
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

	// Create Novel UseCase
	createNovelUC := novel_module.NewCreateNovelUseCase(
		txManager,
		services.Novel,
		services.Genre,
		services.Author,
		services.Artist,
		services.Creator,
		services.Embedding,
	)

	// Create Genre UseCases
	createGenreUC := genre_module.NewCreateGenreUseCase(services.Genre)
	updateGenreUC := genre_module.NewUpdateGenreUseCase(services.Genre)
	deleteGenreUC := genre_module.NewDeleteGenreUseCase(services.Genre)
	getGenreUC := genre_module.NewGetGenreUseCase(services.Genre)
	listGenresUC := genre_module.NewListGenresUseCase(services.Genre)
	listSelectUC := genre_module.NewListSelectionUseCase(services.Genre)
	mergeGenresUC := genre_module.NewMergeGenresUseCase(services.Genre)
	previewMergeUC := genre_module.NewPreviewMergeUseCase(services.Genre)

	// Create Artist UseCases
	createArtistUC := artist_module.NewCreateArtistUseCase(services.Artist)
	updateArtistUC := artist_module.NewUpdateArtistUseCase(services.Artist)
	deleteArtistUC := artist_module.NewDeleteArtistUseCase(services.Artist)
	getArtistUC := artist_module.NewGetArtistUseCase(services.Artist)
	listArtistsUC := artist_module.NewListArtistsUseCase(services.Artist)
	listSelectArtistUC := artist_module.NewListSelectionUseCase(services.Artist)
	mergeArtistsUC := artist_module.NewMergeArtistsUseCase(services.Artist)
	previewMergeArtistUC := artist_module.NewPreviewMergeUseCase(services.Artist)

	return &Handlers{
		OAuth2:      oauth2Handler,
		OAuth2Admin: oauth2_module.NewAdminHandler(services.OAuth2Admin),
		Auth: func() *auth_module.Handler {
			registerUC := auth_module.NewRegisterUserUseCase(services.Auth)
			verifyUC := auth_module.NewVerifyEmailUseCase(services.Auth)
			forgotUC := auth_module.NewForgotPasswordUseCase(services.Auth)
			resetUC := auth_module.NewResetPasswordUseCase(services.Auth)

			return auth_module.NewHandler(
				services.Auth,
				services.Email,
				services.WebAuthn,
				services.OAuth2,
				registerUC, verifyUC, forgotUC, resetUC,
			)
		}(),
		Genre:       genre_module.NewHandler(
			createGenreUC, updateGenreUC, deleteGenreUC, getGenreUC,
			listGenresUC, listSelectUC, mergeGenresUC, previewMergeUC,
			services.Analytics,
			zapLogger,
		),
		Author: func() *author_module.Handler {
			createAuthorUC := author_module.NewCreateAuthorUseCase(services.Author)
			updateAuthorUC := author_module.NewUpdateAuthorUseCase(services.Author)
			deleteAuthorUC := author_module.NewDeleteAuthorUseCase(services.Author)
			getAuthorUC := author_module.NewGetAuthorUseCase(services.Author)
			listAuthorsUC := author_module.NewListAuthorsUseCase(services.Author)
			listSelectAuthorUC := author_module.NewListSelectionUseCase(services.Author)
			mergeAuthorsUC := author_module.NewMergeAuthorsUseCase(services.Author)
			previewMergeAuthorUC := author_module.NewPreviewMergeUseCase(services.Author)
			return author_module.NewHandler(
				createAuthorUC, updateAuthorUC, deleteAuthorUC, getAuthorUC,
				listAuthorsUC, listSelectAuthorUC, mergeAuthorsUC, previewMergeAuthorUC,
				zapLogger,
			)
		}(),
		Artist:      artist_module.NewHandler(
			createArtistUC, updateArtistUC, deleteArtistUC, getArtistUC,
			listArtistsUC, listSelectArtistUC, mergeArtistsUC, previewMergeArtistUC,
			zapLogger,
		),
		Novel: func() *novel_module.Handler {
			updateNovelUC := novel_module.NewUpdateNovelUseCase(services.Novel, services.Embedding)
			deleteNovelUC := novel_module.NewDeleteNovelUseCase(
				txManager,
				services.Novel,
				services.Genre,
				services.Creator,
			)
			getNovelUC := novel_module.NewGetNovelUseCase(services.Novel)
			listNovelsUC := novel_module.NewListNovelsUseCase(services.Novel)
			viewCountUC := novel_module.NewIncrementViewCountUseCase(services.Novel)
			getNovelFullUC := novel_module.NewGetNovelFullUseCase(services.Novel)
			return novel_module.NewHandler(
				services.Novel, createNovelUC,
				updateNovelUC, deleteNovelUC, getNovelUC,
				listNovelsUC, viewCountUC, getNovelFullUC,
				services.Volume, services.Chapter,
			)
		}(),
		Volume: func() *novel_volume.Handler {
			createUC := novel_volume.NewCreateVolumeUseCase(services.Volume)
			updateUC := novel_volume.NewUpdateVolumeUseCase(services.Volume)
			deleteUC := novel_volume.NewDeleteVolumeUseCase(services.Volume)
			getUC := novel_volume.NewGetVolumeUseCase(services.Volume)
			listUC := novel_volume.NewListVolumesByNovelUseCase(services.Volume)
			orderUC := novel_volume.NewUpdateDisplayOrderUseCase(services.Volume)
			pubUC := novel_volume.NewPublishVolumeUseCase(services.Volume)
			unpubUC := novel_volume.NewUnpublishVolumeUseCase(services.Volume)
			return novel_volume.NewHandler(
				services.Volume, createUC, updateUC, deleteUC,
				getUC, listUC, orderUC, pubUC, unpubUC,
			)
		}(),
		Chapter: func() *novel_chapter.Handler {
			createUC := novel_chapter.NewCreateChapterUseCase(services.Chapter)
			updateUC := novel_chapter.NewUpdateChapterUseCase(services.Chapter)
			deleteUC := novel_chapter.NewDeleteChapterUseCase(services.Chapter)
			getUC := novel_chapter.NewGetChapterUseCase(services.Chapter)
			listNovelUC := novel_chapter.NewListChaptersByNovelUseCase(services.Chapter)
			listVolumeUC := novel_chapter.NewListChaptersByVolumeUseCase(services.Chapter)
			pubUC := novel_chapter.NewPublishChapterUseCase(services.Chapter)
			schedUC := novel_chapter.NewScheduleChapterUseCase(services.Chapter)
			viewUC := novel_chapter.NewIncrementViewCountUseCase(services.Chapter)
			statsUC := novel_chapter.NewUpdateStatisticsUseCase(services.Chapter)

			return novel_chapter.NewHandler(
				services.Chapter,
				createUC, updateUC, deleteUC, getUC,
				listNovelUC, listVolumeUC,
				pubUC, schedUC,
				viewUC, statsUC,
			)
		}(),
		User:        user_module.NewHandler(services.User),
		Media: func() *media_module.Handler {
			trendingUC := media_module.NewGetTrendingUseCase(services.Analytics)
			homeUC := media_module.NewGetHomeDataUseCase(services.Media)
			return media_module.NewHandler(trendingUC, homeUC)
		}(),
		Creator: func() *creator_module.Handler {
			uc := creator_module.NewListCreatorsUseCase(services.Creator)
			return creator_module.NewHandler(uc, services.Analytics)
		}(),
		Organization: organization_module.NewHandler(services.Organization, services.Analytics, zapLogger),
		Embedding:    embedding_module.NewHandler(services.Embedding, services.Novel, services.Cache),
		PaymentConfig: payment_module.NewHandler(services.PaymentConfig, zapLogger),
		Wallet:        payment_module.NewWalletHandler(services.Wallet, services.Topup, services.TransactionSvc, zapLogger),
		Webhook:       payment_module.NewWebhookHandler(services.Topup, services.PaymentConfig, zapLogger),
		Socket:        socket.NewHandler(services.SocketHub, zapLogger),
	}
}
