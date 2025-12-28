package novel

import "system/internal/domain"

// NovelFullData chứa toàn bộ dữ liệu cần thiết cho trang chi tiết novel
type NovelFullData struct {
	Novel               *domain.Novel
	Genres              []*domain.Genre
	Authors             []*domain.NovelAuthor
	Artists             []*domain.NovelArtist
	Volumes             []*domain.NovelVolume
	Chapters            []*domain.NovelChapter // All published chapters
	ChaptersWithoutVol  []*domain.NovelChapter // Chapters không thuộc volume nào
	VolumesWithChapters []*domain.NovelVolumeWithChapters
}
