package model

// Catalog domain enums mirrored from the catalog database schema. Keeping them
// in pkg/common allows multiple services to share the same typed constants.

type ContentStatus string

type SeasonName string

type CreatorRole string

type ContentType string

type ContentRelationType string

type PurchaseItemType string

type RentalItemType string

type SubscriptionTier string

const (
	ContentStatusOngoing   ContentStatus = "ONGOING"
	ContentStatusCompleted ContentStatus = "COMPLETED"
	ContentStatusHiatus    ContentStatus = "HIATUS"

	SeasonNameSpring SeasonName = "SPRING"
	SeasonNameSummer SeasonName = "SUMMER"
	SeasonNameFall   SeasonName = "FALL"
	SeasonNameWinter SeasonName = "WINTER"

	CreatorRoleAuthor      CreatorRole = "AUTHOR"
	CreatorRoleIllustrator CreatorRole = "ILLUSTRATOR"
	CreatorRoleArtist      CreatorRole = "ARTIST"
	CreatorRoleStudio      CreatorRole = "STUDIO"
	CreatorRoleVoiceActor  CreatorRole = "VOICE_ACTOR"

	ContentTypeAnime ContentType = "ANIME"
	ContentTypeManga ContentType = "MANGA"
	ContentTypeNovel ContentType = "NOVEL"

	ContentRelationAdaptation ContentRelationType = "ADAPTATION"
	ContentRelationSequel     ContentRelationType = "SEQUEL"
	ContentRelationSpinoff    ContentRelationType = "SPINOFF"
	ContentRelationRelated    ContentRelationType = "RELATED"

	PurchaseItemAnimeEpisode PurchaseItemType = "ANIME_EPISODE"
	PurchaseItemAnimeSeason  PurchaseItemType = "ANIME_SEASON"
	PurchaseItemMangaChapter PurchaseItemType = "MANGA_CHAPTER"
	PurchaseItemMangaVolume  PurchaseItemType = "MANGA_VOLUME"
	PurchaseItemMangaSeries  PurchaseItemType = "MANGA_SERIES"
	PurchaseItemNovelChapter PurchaseItemType = "NOVEL_CHAPTER"
	PurchaseItemNovelVolume  PurchaseItemType = "NOVEL_VOLUME"
	PurchaseItemNovelSeries  PurchaseItemType = "NOVEL_SERIES"

	RentalItemAnimeSeason RentalItemType = "ANIME_SEASON"
	RentalItemAnimeSeries RentalItemType = "ANIME_SERIES"
	RentalItemMangaVolume RentalItemType = "MANGA_VOLUME"
	RentalItemMangaSeries RentalItemType = "MANGA_SERIES"
	RentalItemNovelVolume RentalItemType = "NOVEL_VOLUME"
	RentalItemNovelSeries RentalItemType = "NOVEL_SERIES"

	SubscriptionTierFree    SubscriptionTier = "FREE"
	SubscriptionTierPremium SubscriptionTier = "PREMIUM"
	SubscriptionTierVIP     SubscriptionTier = "VIP"
)
