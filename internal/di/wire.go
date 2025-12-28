//go:build wireinject
// +build wireinject

// Package di provides 100% dependency injection using Google Wire.
package di

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"system/configs"
	"system/internal/app/router"
	ent "system/internal/ent/generated"
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
	"system/internal/platform/database"
	"system/internal/platform/i18n"
	"system/internal/platform/oauth2"
	fosite_storage "system/internal/platform/oauth2/storage"
	"system/internal/platform/resend"
	socket "system/internal/platform/socket"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/ory/fosite"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	// Register postgres driver for sql.Open("postgres", ...)
	_ "github.com/lib/pq"
)

// ===== Infrastructure Providers =====

// ProvideTransactionManager provides the Ent-based transaction manager
func ProvideTransactionManager(client *ent.Client) database.TransactionManager {
	return database.NewTransactionManager(client)
}

// ProvideRedisGoClient exports redis client from wrapper
func ProvideRedisGoClient(rdb *database.RedisClient) *redis.Client {
	return rdb.Client
}

func ProvideEmailConfig(cfg *configs.Config) *configs.EmailConfig {
	return &cfg.Email
}

func ProvideOAuthConfig(cfg *configs.Config) *configs.OAuthConfig {
	return &cfg.OAuth2
}

func ProvideWebAuthnConfig(cfg *configs.Config) configs.WebAuthnConfig {
	return cfg.WebAuthn
}

func ProvideViewTrackingConfig(cfg *configs.Config) *configs.ViewTrackingConfig {
	return &cfg.ViewTracking
}

func ProvideEmbeddingConfig(cfg *configs.Config) *configs.EmbeddingConfig {
	return &cfg.Embedding
}

// ProvideEntClient tạo Ent client từ Config
func ProvideEntClient(ctx context.Context, cfg *configs.Config, logger *zap.Logger) (*ent.Client, error) {
	entClient, err := database.NewEntClient(ctx, &cfg.DB, cfg.Log.DBLogQueries, logger)
	if err != nil {
		return nil, err
	}
	return entClient.Client, nil
}

// ProvideDBFromEnt creates *sql.DB from configs for repos needing raw SQL
func ProvideDBFromEnt(cfg *configs.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)
	return sql.Open("postgres", dsn)
}

var InfraSet = wire.NewSet(
	ProvideRedisGoClient,
	ProvideTransactionManager, // Added
	ProvideEmailConfig,
	ProvideOAuthConfig,
	ProvideWebAuthnConfig,
	ProvideViewTrackingConfig,
	ProvideEmbeddingConfig,
	ProvideEntClient,
	ProvideDBFromEnt,
)

// ===== Repository Providers =====

var RepositorySet = wire.NewSet(
	// OAuth2 (Ent)
	oauth2_module.NewOAuth2ClientEntRepository,
	oauth2_module.NewOAuth2SessionEntRepository,
	oauth2_module.NewConsentEntRepository,
	// OAuth2 (Redis - keep pgx)
	oauth2_module.NewAuthRequestRepository,

	// User (Ent)
	user_module.NewUserEntRepository,
	auth_module.NewUserRoleEntRepository,
	user_module.NewRoleEntRepository, // RBAC - now Ent
	// User Session (Redis - keep pgx)
	user_module.NewSessionRepository,

	// Auth (Ent)
	auth_module.NewEmailVerificationEntRepository,
	auth_module.NewPasswordResetEntRepository,
	auth_module.NewWebAuthnCredentialEntRepository,
	auth_module.NewWebAuthnSessionEntRepository,

	// Genre (Ent)
	genre_module.NewEntGenreRepository,

	// Author (Ent)
	author_module.NewEntAuthorRepository,

	// Artist (Ent)
	artist_module.NewEntArtistRepository,

	// Novel (Ent)
	novel_module.NewEntNovelRepository,
	novel_volume.NewEntVolumeRepository,
	novel_volume.NewVolumeHistoryEntRepository, // Now Ent
	novel_chapter.NewEntChapterRepository,
	novel_chapter.NewChapterHistoryEntRepository, // Now Ent

	// Analytics (ClickHouse/Redis - keep pgx)
	analytics_module.NewViewAnalyticsClickHouseRepository,
	analytics_module.NewViewTrackingRedisRepository,

	// Creator (Ent + raw SQL for ListCreators) & Organization (Ent)
	creator_module.NewEntRepository,
	organization_module.NewEntRepository,

	// Payment (Ent)
	payment_module.NewPaymentConfigEntRepository,
	payment_module.NewWalletEntRepository,
	payment_module.NewCoinPackageEntRepository,
	payment_module.NewTopupOrderEntRepository,
	payment_module.NewTransactionEntRepository,

	// Media Progress (Ent)
	media_progress_module.NewEntRepository,

	// Embedding (Ent + raw SQL for similarity)
	embedding_module.NewEntRepository,
)

// ===== OAuth2/Fosite Providers =====

var OAuth2Set = wire.NewSet(
	fosite_storage.NewSQLStore,
	fosite_storage.NewRedisStore,
	fosite_storage.NewHybridStore,
	oauth2.NewOAuth2Provider,
)

// ===== Service Providers =====

var ServiceSet = wire.NewSet(
	// User
	user_module.NewService,

	// OAuth2
	oauth2_module.NewService,
	oauth2_module.NewAdminService,

	// Auth
	auth_module.NewService,
	auth_module.NewWebAuthnService,

	// Email
	resend.NewClient,
	wire.Bind(new(email.EmailSender), new(*resend.Client)),
	email.NewService,
	wire.Bind(new(email.EmailService), new(*email.Service)),

	// Bind auth.WebAuthnService to oauth2.WebAuthnService and auth.OAuth2SessionCreator
	wire.Bind(new(oauth2_module.WebAuthnService), new(auth_module.WebAuthnService)),
	wire.Bind(new(auth_module.OAuth2SessionCreator), new(oauth2_module.OAuth2Service)),

	// Genre
	genre_module.NewService,
	// Bind to novel module's UCGenreService interface
	wire.Bind(new(novel_module.UCGenreService), new(genre_module.GenreService)),

	// Author
	author_module.NewService,
	// Bind to novel module's UCAuthorService interface
	wire.Bind(new(novel_module.UCAuthorService), new(author_module.AuthorService)),

	// Artist
	artist_module.NewService,
	// Bind to novel module's UCArtistService interface
	wire.Bind(new(novel_module.UCArtistService), new(artist_module.ArtistService)),

	// Novel
	novel_module.NewService,
	// Bind to novel module's UCNovelService interface
	wire.Bind(new(novel_module.UCNovelService), new(novel_module.NovelService)),

	// Volume
	novel_volume.NewService,

	// Chapter
	novel_chapter.NewService,

	// Analytics
	analytics_module.NewViewTrackingService,
	wire.Bind(new(novel_chapter.ViewTracker), new(*analytics_module.ViewTrackingService)),
	analytics_module.NewService,

	// Creator
	creator_module.NewService,
	// Bind to novel module's UCCreatorService interface
	wire.Bind(new(novel_module.UCCreatorService), new(creator_module.CreatorService)),

	// Organization
	organization_module.NewService,

	// Embedding
	ProvideNoopEmbedder,
	embedding_module.NewService,
	wire.Bind(new(embedding_module.ContentFetcher), new(novel_module.NovelService)),
	// Bind to novel module's UCEmbeddingService interface
	wire.Bind(new(novel_module.UCEmbeddingService), new(*embedding_module.Service)),

	// Payment
	payment_module.NewConfigUseCase,
	payment_module.NewWalletUseCase,
	payment_module.NewTransactionUseCase,

	// Socket Hub (also provides TopupNotifier)
	socket.NewHub,
	wire.Bind(new(payment_module.TopupNotifier), new(*socket.Hub)),

	// Topup (needs TopupNotifier)
	payment_module.NewTopupUseCase,

	// Media Progress
	media_progress_module.NewService,

	// Cache
	cache.NewCacheService,

	// TopNovelAdapter (provides TopNovelService)
	router.NewTopNovelAdapter,
	wire.Bind(new(novel_module.TopNovelService), new(*router.TopNovelAdapter)),
)

// ProvideNoopEmbedder creates a NoopEmbedder that implements Embedder interface
func ProvideNoopEmbedder(cfg *configs.EmbeddingConfig) embedding_module.Embedder {
	return embedding_module.NewNoopEmbedder(cfg.Dimensions)
}

// ===== UseCase Providers =====

var UseCaseSet = wire.NewSet(
	// Novel
	novel_module.NewCreateNovelUseCase,
	novel_module.NewUpdateNovelUseCase,
	novel_module.NewDeleteNovelUseCase,
	novel_module.NewGetNovelUseCase,
	novel_module.NewListNovelsUseCase,
	novel_module.NewIncrementViewCountUseCase,
	novel_module.NewGetNovelFullUseCase,

	// Genre
	genre_module.NewCreateGenreUseCase,
	genre_module.NewUpdateGenreUseCase,
	genre_module.NewDeleteGenreUseCase,
	genre_module.NewGetGenreUseCase,
	genre_module.NewListGenresUseCase,
	genre_module.NewListSelectionUseCase,
	genre_module.NewMergeGenresUseCase,
	genre_module.NewPreviewMergeUseCase,

	// Author
	author_module.NewCreateAuthorUseCase,
	author_module.NewUpdateAuthorUseCase,
	author_module.NewDeleteAuthorUseCase,
	author_module.NewGetAuthorUseCase,
	author_module.NewListAuthorsUseCase,
	author_module.NewListSelectionUseCase,
	author_module.NewMergeAuthorsUseCase,
	author_module.NewPreviewMergeUseCase,

	// Artist
	artist_module.NewCreateArtistUseCase,
	artist_module.NewUpdateArtistUseCase,
	artist_module.NewDeleteArtistUseCase,
	artist_module.NewGetArtistUseCase,
	artist_module.NewListArtistsUseCase,
	artist_module.NewListSelectionUseCase,
	artist_module.NewMergeArtistsUseCase,
	artist_module.NewPreviewMergeUseCase,

	// Volume
	novel_volume.NewCreateVolumeUseCase,
	novel_volume.NewUpdateVolumeUseCase,
	novel_volume.NewDeleteVolumeUseCase,
	novel_volume.NewGetVolumeUseCase,
	novel_volume.NewListVolumesByNovelUseCase,
	novel_volume.NewUpdateDisplayOrderUseCase,
	novel_volume.NewPublishVolumeUseCase,
	novel_volume.NewUnpublishVolumeUseCase,

	// Chapter
	novel_chapter.NewCreateChapterUseCase,
	novel_chapter.NewUpdateChapterUseCase,
	novel_chapter.NewDeleteChapterUseCase,
	novel_chapter.NewGetChapterUseCase,
	novel_chapter.NewListChaptersByNovelUseCase,
	novel_chapter.NewListChaptersByVolumeUseCase,
	novel_chapter.NewPublishChapterUseCase,
	novel_chapter.NewScheduleChapterUseCase,
	novel_chapter.NewUpdateStatisticsUseCase,

	// Auth
	auth_module.NewRegisterUserUseCase,
	auth_module.NewVerifyEmailUseCase,
	auth_module.NewForgotPasswordUseCase,
	auth_module.NewResetPasswordUseCase,

	// Creator
	creator_module.NewListCreatorsUseCase,
)

// ===== Handler Providers =====

var HandlerSet = wire.NewSet(
	oauth2_module.NewHandler,
	oauth2_module.NewAdminHandler,
	auth_module.NewHandler,
	genre_module.NewHandler,
	author_module.NewHandler,
	artist_module.NewHandler,
	novel_module.NewHandler,
	novel_volume.NewHandler,
	novel_chapter.NewHandler,
	user_module.NewHandler,
	creator_module.NewHandler,
	organization_module.NewHandler,
	embedding_module.NewHandler,
	payment_module.NewHandler,
	payment_module.NewWalletHandler,
	payment_module.NewWebhookHandler,
	media_progress_module.NewHandler,
	media_module.NewHandler,
	socket.NewHandler,
)

// ===== Aggregated Structs =====

var StructSet = wire.NewSet(
	wire.Struct(new(router.Repositories), "*"),
	wire.Struct(new(router.Services), "*"),
	wire.Struct(new(router.Handlers), "*"),
)

// ===== Application =====

// Application holds all initialized components
type Application struct {
	Config     *configs.Config
	Logger     *zap.Logger
	Router     *gin.Engine
	HTTPServer *http.Server
	Redis      *database.RedisClient
	ClickHouse *database.ClickHouseClient
	I18n       *i18n.I18n
	Repos      *router.Repositories
	Services   *router.Services
	Handlers   *router.Handlers
	OAuth2     fosite.OAuth2Provider
	EntClient  *ent.Client // Ent ORM client
}

// ProvideHTTPServer creates the HTTP server
func ProvideHTTPServer(cfg *configs.Config, r *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}
}

// ProvideRouter creates and configures the Gin router
func ProvideRouter(
	cfg *configs.Config,
	i18nInstance *i18n.I18n,
	logger *zap.Logger,
	handlers *router.Handlers,
	services *router.Services,
	oauth2Provider fosite.OAuth2Provider,
) *gin.Engine {
	return router.NewRouterWithHandlers(cfg, i18nInstance, logger, handlers, services, oauth2Provider)
}

// MasterSet contains all providers
var MasterSet = wire.NewSet(
	InfraSet,
	RepositorySet,
	OAuth2Set,
	ServiceSet,
	UseCaseSet,
	HandlerSet,
	StructSet,
	ProvideRouter,
	ProvideHTTPServer,
	wire.Struct(new(Application), "*"),
)

// InitializeApplication wires all dependencies
func InitializeApplication(
	ctx context.Context,
	cfg *configs.Config,
	logger *zap.Logger,
	i18nInstance *i18n.I18n,
	rdb *database.RedisClient,
	ch *database.ClickHouseClient,
) (*Application, error) {
	wire.Build(MasterSet)
	return nil, nil
}
