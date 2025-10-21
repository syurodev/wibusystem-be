# Catalog Module - Migration Plan

## Status: 🚧 TO BE MIGRATED

The Catalog module is currently in the old microservices structure at `services/catalog/` and needs to be migrated to the modular monolith structure.

## Current Location

```
services/catalog/
├── api-design/
├── config/
├── grpc/
├── handlers/
├── middleware/
├── repositories/
├── routes/
├── services/
├── main.go
└── go.mod
```

**Statistics:**
- Total Go files: 45
- Already implemented with repositories, handlers, routes
- Has gRPC support
- Has API documentation

## Target Structure

```
internal/modules/catalog/
├── domain/              # Domain entities (Novel, Chapter, Volume, Creator, etc.)
│   ├── novel.go
│   ├── chapter.go
│   ├── volume.go
│   ├── creator.go
│   ├── character.go
│   └── genre.go
├── repository/          # Repository interfaces
│   ├── novel_repository.go
│   ├── chapter_repository.go
│   ├── volume_repository.go
│   ├── creator_repository.go
│   ├── character_repository.go
│   ├── genre_repository.go
│   └── postgres/       # PostgreSQL implementations
│       ├── novel_repository.go
│       ├── chapter_repository.go
│       └── ...
├── service/             # Business logic
│   ├── novel_service.go
│   ├── chapter_service.go
│   ├── volume_service.go
│   ├── creator_service.go
│   └── ...
├── handler/             # HTTP handlers
│   ├── http/
│   │   ├── novel_handler.go
│   │   ├── chapter_handler.go
│   │   ├── volume_handler.go
│   │   ├── creator_handler.go
│   │   └── router.go
│   └── middleware/
│       ├── auth_middleware.go
│       └── ...
└── dto/                 # Data transfer objects
    ├── novel.go
    ├── chapter.go
    └── ...
```

## Migration Steps

### Phase 1: Domain Layer (Days 1-2)
- [ ] Extract domain entities from existing code
- [ ] Define Novel entity with validation
- [ ] Define Chapter entity
- [ ] Define Volume entity
- [ ] Define Creator entity
- [ ] Define Character entity
- [ ] Define Genre entity
- [ ] Domain validation and business rules

### Phase 2: Repository Layer (Days 3-4)
- [ ] Define repository interfaces
- [ ] Migrate existing repository code to new structure
- [ ] Update to use domain entities
- [ ] Add missing CRUD operations
- [ ] Add filtering and search
- [ ] PostgreSQL implementations

### Phase 3: Service Layer (Days 5-7)
- [ ] Migrate business logic from old handlers
- [ ] NovelService implementation
- [ ] ChapterService implementation
- [ ] VolumeService implementation
- [ ] CreatorService implementation
- [ ] CharacterService implementation
- [ ] GenreService implementation
- [ ] Cross-module integration (with Identity for auth)

### Phase 4: Handler Layer (Days 8-9)
- [ ] Migrate HTTP handlers
- [ ] Update to use Fiber framework (consistent with Identity)
- [ ] Add validation middleware
- [ ] Add authentication middleware (integrate with Identity)
- [ ] Route configuration
- [ ] Error handling

### Phase 5: Testing (Day 10)
- [ ] Unit tests
- [ ] Integration tests
- [ ] API tests
- [ ] Performance tests

## Key Entities

### Novel
- ID, Title, Description, Status
- Cover images, tags
- Author/Creator references
- Genres, ratings
- Publish dates
- Volumes and chapters

### Chapter
- ID, Title, Content
- Chapter number, volume reference
- Publish date
- Word count
- Status (draft, published)

### Volume
- ID, Title, Description
- Volume number
- Novel reference
- Chapters list

### Creator
- ID, Name, Bio
- Profile images
- Social links
- Novels created

### Character
- ID, Name, Description
- Images
- Novel references
- Character traits

### Genre
- ID, Name, Description
- Novel references

## Dependencies

Catalog module will depend on:
- **Identity module** for authentication/authorization
- **Shared infrastructure** (database, config)
- **Common types** (pagination, errors)

## API Endpoints (Estimated)

### Novels
- GET /api/v1/novels - List novels
- GET /api/v1/novels/:id - Get novel details
- POST /api/v1/novels - Create novel (auth)
- PUT /api/v1/novels/:id - Update novel (auth)
- DELETE /api/v1/novels/:id - Delete novel (auth)
- GET /api/v1/novels/:id/chapters - List chapters
- GET /api/v1/novels/:id/volumes - List volumes

### Chapters
- GET /api/v1/chapters/:id - Get chapter
- POST /api/v1/chapters - Create chapter (auth)
- PUT /api/v1/chapters/:id - Update chapter (auth)
- DELETE /api/v1/chapters/:id - Delete chapter (auth)

### Volumes
- GET /api/v1/volumes/:id - Get volume
- POST /api/v1/volumes - Create volume (auth)
- PUT /api/v1/volumes/:id - Update volume (auth)
- DELETE /api/v1/volumes/:id - Delete volume (auth)

### Creators
- GET /api/v1/creators - List creators
- GET /api/v1/creators/:id - Get creator
- POST /api/v1/creators - Create creator (auth)
- PUT /api/v1/creators/:id - Update creator (auth)

### Characters
- GET /api/v1/characters - List characters
- GET /api/v1/characters/:id - Get character
- POST /api/v1/characters - Create character (auth)
- PUT /api/v1/characters/:id - Update character (auth)

### Genres
- GET /api/v1/genres - List genres
- GET /api/v1/genres/:id - Get genre
- POST /api/v1/genres - Create genre (auth)
- PUT /api/v1/genres/:id - Update genre (auth)

**Estimated Total: 30-40 endpoints**

## Notes

1. **Reuse existing code**: The old catalog service has working repositories and handlers. We can migrate and adapt them rather than rewrite from scratch.

2. **Database schema**: The catalog database schema should be reviewed and potentially updated to align with the new domain models.

3. **Multi-tenancy**: Consider if novels should be tenant-scoped (e.g., each organization has its own novels).

4. **Authentication**: Integrate with Identity module's session-based authentication.

5. **File storage**: Novels may have cover images, chapter images - need file storage strategy.

6. **Search**: Consider full-text search for novels, chapters (PostgreSQL full-text or Elasticsearch).

7. **Caching**: High-read operations (novel lists, chapter content) should be cached.

## Existing Assets to Preserve

From `services/catalog/`:
- ✅ Repository implementations
- ✅ Handler logic
- ✅ API design documentation
- ✅ Route definitions
- ✅ Middleware (adapt for Fiber)
- ✅ gRPC clients (optional, evaluate if still needed)

## Estimated Timeline

- **Migration**: 10 days (following Identity module pattern)
- **Testing**: 2-3 days
- **Integration**: 1-2 days
- **Total**: ~2-3 weeks

## Priority

**Priority: MEDIUM**

Catalog is an important module but not blocking. Should be migrated after:
1. ✅ Identity module (complete)
2. Testing and stabilization of Identity
3. Then Catalog migration

## References

- Old service: `services/catalog/`
- API design docs: `services/catalog/api-design/`
- Database design: `catalog-db-design.md` (root)
- Identity module: `internal/modules/identity/` (reference pattern)

---

**Last Updated:** January 2025  
**Status:** Planning phase  
**Next Action:** Wait for Identity module testing completion