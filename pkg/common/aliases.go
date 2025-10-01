package common

import (
	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	r "wibusystem/pkg/common/response"
)

// Models
type User = m.User
type Tenant = m.Tenant
type Membership = m.Membership
type Permission = m.Permission
type Role = m.Role
type RolePermission = m.RolePermission
type RoleAssignment = m.RoleAssignment
type AuthType = m.AuthType

const (
	AuthTypePassword = m.AuthTypePassword
	AuthTypeOAuth    = m.AuthTypeOAuth
	AuthTypeOIDC     = m.AuthTypeOIDC
	AuthTypeSAML     = m.AuthTypeSAML
	AuthTypeWebAuthn = m.AuthTypeWebAuthn
	AuthTypeTOTP     = m.AuthTypeTOTP
	AuthTypePasskey  = m.AuthTypePasskey
)

type Credential = m.Credential
type Device = m.Device
type Session = m.Session
type OAuth2Client = m.OAuth2Client
type StringArray = m.StringArray

type ContentStatus = m.ContentStatus

const (
	ContentStatusOngoing   = m.ContentStatusOngoing
	ContentStatusCompleted = m.ContentStatusCompleted
	ContentStatusHiatus    = m.ContentStatusHiatus
)

type SeasonName = m.SeasonName

const (
	SeasonNameSpring = m.SeasonNameSpring
	SeasonNameSummer = m.SeasonNameSummer
	SeasonNameFall   = m.SeasonNameFall
	SeasonNameWinter = m.SeasonNameWinter
)

type CreatorRole = m.CreatorRole

const (
	CreatorRoleAuthor      = m.CreatorRoleAuthor
	CreatorRoleIllustrator = m.CreatorRoleIllustrator
	CreatorRoleArtist      = m.CreatorRoleArtist
	CreatorRoleStudio      = m.CreatorRoleStudio
	CreatorRoleVoiceActor  = m.CreatorRoleVoiceActor
)

type ContentType = m.ContentType

const (
	ContentTypeAnime = m.ContentTypeAnime
	ContentTypeManga = m.ContentTypeManga
	ContentTypeNovel = m.ContentTypeNovel
)

type ContentRelationType = m.ContentRelationType

const (
	ContentRelationAdaptation = m.ContentRelationAdaptation
	ContentRelationSequel     = m.ContentRelationSequel
	ContentRelationSpinoff    = m.ContentRelationSpinoff
	ContentRelationRelated    = m.ContentRelationRelated
)

type PurchaseItemType = m.PurchaseItemType

const (
	PurchaseItemAnimeEpisode = m.PurchaseItemAnimeEpisode
	PurchaseItemAnimeSeason  = m.PurchaseItemAnimeSeason
	PurchaseItemMangaChapter = m.PurchaseItemMangaChapter
	PurchaseItemMangaVolume  = m.PurchaseItemMangaVolume
	PurchaseItemMangaSeries  = m.PurchaseItemMangaSeries
	PurchaseItemNovelChapter = m.PurchaseItemNovelChapter
	PurchaseItemNovelVolume  = m.PurchaseItemNovelVolume
	PurchaseItemNovelSeries  = m.PurchaseItemNovelSeries
)

type RentalItemType = m.RentalItemType

const (
	RentalItemAnimeSeason = m.RentalItemAnimeSeason
	RentalItemAnimeSeries = m.RentalItemAnimeSeries
	RentalItemMangaVolume = m.RentalItemMangaVolume
	RentalItemMangaSeries = m.RentalItemMangaSeries
	RentalItemNovelVolume = m.RentalItemNovelVolume
	RentalItemNovelSeries = m.RentalItemNovelSeries
)

type SubscriptionTier = m.SubscriptionTier

const (
	SubscriptionTierFree    = m.SubscriptionTierFree
	SubscriptionTierPremium = m.SubscriptionTierPremium
	SubscriptionTierVIP     = m.SubscriptionTierVIP
)

// DTOs
type CreateUserRequest = d.CreateUserRequest
type UpdateUserRequest = d.UpdateUserRequest
type CreateTenantRequest = d.CreateTenantRequest
type LoginRequest = d.LoginRequest
type LoginResponse = d.LoginResponse
type UserInfo = d.UserInfo

// Responses
type ErrorDetail = r.ErrorDetail
type StandardResponse = r.StandardResponse
