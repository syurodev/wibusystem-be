package http

import (
	"wibusystem/internal/modules/catalog/service"
	"wibusystem/internal/platform/response"

	"github.com/gofiber/fiber/v2"
)

// SetupCatalogRoutes sets up the routes for the catalog module.
func SetupCatalogRoutes(
	app *fiber.App,
	novelService service.NovelService,
	volumeService service.VolumeService,
	chapterService service.ChapterService,
	creatorService service.CreatorService,
	genreService service.GenreService,
	characterService service.CharacterService,
) {
	novelHandler := NewNovelHandler(novelService)
	volumeHandler := NewVolumeHandler(volumeService)
	chapterHandler := NewChapterHandler(chapterService)
	creatorHandler := NewCreatorHandler(creatorService)
	genreHandler := NewGenreHandler(genreService)
	characterHandler := NewCharacterHandler(characterService)

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Novel routes
	novels := v1.Group("/novels")
	novels.Post("/", response.WithStandardResponse(novelHandler.CreateNovel))
	novels.Get("/", response.WithStandardResponse(novelHandler.ListNovels))
	novels.Get("/:id", response.WithStandardResponse(novelHandler.GetNovel))

	// Volume routes
	volumes := v1.Group("/volumes")
	volumes.Post("/", response.WithStandardResponse(volumeHandler.CreateVolume))
	volumes.Get("/:id", response.WithStandardResponse(volumeHandler.GetVolume))
	novels.Get("/:novelId/volumes", response.WithStandardResponse(volumeHandler.ListVolumesByNovel)) // Nested under novels

	// Chapter routes
	chapters := v1.Group("/chapters")
	chapters.Post("/", response.WithStandardResponse(chapterHandler.CreateChapter))
	chapters.Get("/:id", response.WithStandardResponse(chapterHandler.GetChapter))
	volumes.Get("/:volumeId/chapters", response.WithStandardResponse(chapterHandler.ListChaptersByVolume)) // Nested under volumes

	// Creator routes
	creators := v1.Group("/creators")
	creators.Post("/", response.WithStandardResponse(creatorHandler.CreateCreator))
	creators.Get("/", response.WithStandardResponse(creatorHandler.ListCreators))
	creators.Get("/:id", response.WithStandardResponse(creatorHandler.GetCreator))

	// Genre routes
	genres := v1.Group("/genres")
	genres.Post("/", response.WithStandardResponse(genreHandler.CreateGenre))
	genres.Get("/", response.WithStandardResponse(genreHandler.ListGenres))
	genres.Get("/:id", response.WithStandardResponse(genreHandler.GetGenre))

	// Character routes
	characters := v1.Group("/characters")
	characters.Post("/", response.WithStandardResponse(characterHandler.CreateCharacter))
	characters.Get("/:id", response.WithStandardResponse(characterHandler.GetCharacter))
	novels.Get("/:novelId/characters", response.WithStandardResponse(characterHandler.ListCharactersByNovel)) // Nested under novels
}
