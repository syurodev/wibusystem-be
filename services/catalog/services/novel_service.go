package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	"wibusystem/services/catalog/grpc"
	"wibusystem/services/catalog/repositories"
	"wibusystem/services/catalog/services/interfaces"
)

// NovelService implements genre-related business logic
type NovelService struct {
	repos       *repositories.Repositories
	grpcClients *grpc.ClientManager
}

func NewNovelService(repos *repositories.Repositories, grpcClients *grpc.ClientManager) interfaces.NovelServiceInterface {
	return &NovelService{
		repos:       repos,
		grpcClients: grpcClients,
	}
}

func (n NovelService) CreateNovel(ctx context.Context, req d.CreateNovelRequest) (*m.Novel, error) {
	return n.repos.Novel.CreateNovel(ctx, req)
}

func (n NovelService) ListNovels(ctx context.Context, req d.ListNovelsRequest) (*d.PaginatedNovelsResponse, error) {
	// Get novels from repository
	response, err := n.repos.Novel.ListNovels(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list novels from repository: %w", err)
	}

	// Collect unique user IDs and tenant IDs for batch fetching
	userIDs := make([]string, 0)
	tenantIDs := make([]string, 0)
	userIDSet := make(map[string]bool)
	tenantIDSet := make(map[string]bool)

	for _, novel := range response.Novels {
		// Determine primary owner IDs based on ownership type
		if novel.PrimaryOwnerID != "" {
			switch novel.OwnershipType {
			case "TENANT":
				if !tenantIDSet[novel.PrimaryOwnerID] {
					tenantIDs = append(tenantIDs, novel.PrimaryOwnerID)
					tenantIDSet[novel.PrimaryOwnerID] = true
				}
			default:
				if !userIDSet[novel.PrimaryOwnerID] {
					userIDs = append(userIDs, novel.PrimaryOwnerID)
					userIDSet[novel.PrimaryOwnerID] = true
				}
			}
		}

		// Original creator is always a user
		if novel.OriginalCreatorID != "" && !userIDSet[novel.OriginalCreatorID] {
			userIDs = append(userIDs, novel.OriginalCreatorID)
			userIDSet[novel.OriginalCreatorID] = true
		}
	}

	// Fetch user and tenant data via gRPC if there are IDs to fetch
	var users map[string]*d.UserSummary
	var tenants map[string]*d.TenantSummary

	if len(userIDs) > 0 {
		users, err = n.grpcClients.GetUsers(ctx, userIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch users via gRPC: %w", err)
		}
	}

	if len(tenantIDs) > 0 {
		tenants, err = n.grpcClients.GetTenants(ctx, tenantIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tenants via gRPC: %w", err)
		}
	}

	// Update response with populated user and tenant data
	for i := range response.Novels {
		novel := &response.Novels[i]

		// Populate primary owner info based on ownership type
		if novel.PrimaryOwnerID != "" {
			switch novel.OwnershipType {
			case "TENANT":
				if tenant, exists := tenants[novel.PrimaryOwnerID]; exists {
					novel.PrimaryOwner = tenant
				}
			default:
				if user, exists := users[novel.PrimaryOwnerID]; exists {
					novel.PrimaryOwner = user
				}
			}
		}

		// Populate original creator info (always a user)
		if novel.OriginalCreatorID != "" {
			if user, exists := users[novel.OriginalCreatorID]; exists {
				novel.OriginalCreator = user
			}
		}
	}

	return response, nil
}

// GetNovelByID implements detailed novel retrieval with optional translations and stats
// Uses NovelQueryRepository (CQRS pattern) for optimized single-query data fetching
func (n NovelService) GetNovelByID(ctx context.Context, id string, includeTranslations, includeStats bool, language string) (*d.NovelDetailResponse, error) {
	// Parse UUID
	novelUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid novel ID format: %w", err)
	}

	// Use NovelQueryRepository for optimized query with all relations in single DB call
	response, err := n.repos.NovelQuery.GetNovelWithFullDetails(ctx, novelUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get novel with full details: %w", err)
	}

	// Set current language from request parameter
	response.CurrentLanguage = language

	// Load translations if includeTranslations is true
	if includeTranslations {
		translations, err := n.repos.Novel.GetNovelTranslations(ctx, novelUUID, language)
		if err != nil {
			return nil, fmt.Errorf("failed to load novel translations: %w", err)
		}
		response.Translations = translations
	}

	// Load stats if includeStats is true
	if includeStats {
		stats, err := n.repos.Novel.GetNovelStats(ctx, novelUUID)
		if err != nil {
			return nil, fmt.Errorf("failed to load novel stats: %w", err)
		}
		response.Stats = stats
	}

	return response, nil
}

// UpdateNovel implements novel update with validation
func (n NovelService) UpdateNovel(ctx context.Context, id string, req d.UpdateNovelRequest) (*d.UpdateNovelResponse, error) {
	// Parse UUID
	novelUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid novel ID format: %w", err)
	}

	// Update novel through repository
	novel, err := n.repos.Novel.UpdateNovel(ctx, novelUUID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update novel: %w", err)
	}

	// Create response
	response := &d.UpdateNovelResponse{
		ID:    novel.ID.String(),
		Title: "",
	}

	// Set title from name field
	if novel.Name != nil {
		response.Title = *novel.Name
	}

	return response, nil
}

// DeleteNovel implements novel deletion with purchase checks
func (n NovelService) DeleteNovel(ctx context.Context, id string, deletedByUserID string) error {
	// Parse UUIDs
	novelUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid novel ID format: %w", err)
	}

	deletedByUUID, err := uuid.Parse(deletedByUserID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// Delete novel through repository (includes purchase checks)
	err = n.repos.Novel.DeleteNovel(ctx, novelUUID, deletedByUUID)
	if err != nil {
		return fmt.Errorf("failed to delete novel: %w", err)
	}

	return nil
}
