// Package auth provides translation-specific authorization helpers
package auth

// TranslationAction defines actions that can be performed on translations
type TranslationAction string

const (
	ActionTranslationSubmit     TranslationAction = "submit"
	ActionTranslationUpdate     TranslationAction = "update"
	ActionTranslationVote       TranslationAction = "vote"
	ActionTranslationReview     TranslationAction = "review"
	ActionTranslationApprove    TranslationAction = "approve"
	ActionTranslationReject     TranslationAction = "reject"
	ActionTranslationDelete     TranslationAction = "delete"
)

// TranslationStatus represents the status of a translation contribution
type TranslationStatus string

const (
	StatusPending  TranslationStatus = "pending"
	StatusApproved TranslationStatus = "approved"
	StatusRejected TranslationStatus = "rejected"
)

// ReferenceType defines the type of content being translated
type ReferenceType string

const (
	RefTypeNovel        ReferenceType = "novel"
	RefTypeNovelChapter ReferenceType = "novel_chapter"
)

// VoteType defines the type of vote on a translation
type VoteType string

const (
	VoteUpvote   VoteType = "upvote"
	VoteDownvote VoteType = "downvote"
)

// Translation permission checks

// CanSubmitTranslation checks if user can submit translations
func CanSubmitTranslation(userPermissions []GlobalPermission) bool {
	return HasPermission(userPermissions, PermTranslationSubmit)
}

// CanUpdateOwnTranslation checks if user can update their own translations
func CanUpdateOwnTranslation(userPermissions []GlobalPermission) bool {
	return HasPermission(userPermissions, PermTranslationUpdateSelf)
}

// CanVoteOnTranslation checks if user can vote on translations
func CanVoteOnTranslation(userPermissions []GlobalPermission) bool {
	return HasPermission(userPermissions, PermTranslationVote)
}

// CanReviewTranslations checks if user can review translations (moderator action)
func CanReviewTranslations(userRole GlobalRole) bool {
	return IsAdminRole(userRole)
}

// CanApproveTranslations checks if user can approve translations
func CanApproveTranslations(userRole GlobalRole) bool {
	return userRole == RoleSuperAdmin || userRole == RoleAdmin || userRole == RoleModerator
}

// CanManageTranslations checks if user can manage all translations
func CanManageTranslations(userRole GlobalRole) bool {
	return userRole == RoleSuperAdmin || userRole == RoleAdmin
}

// Validation helpers

// IsValidTranslationStatus checks if a status is valid
func IsValidTranslationStatus(status string) bool {
	validStatuses := []TranslationStatus{StatusPending, StatusApproved, StatusRejected}
	for _, validStatus := range validStatuses {
		if string(validStatus) == status {
			return true
		}
	}
	return false
}

// IsValidReferenceType checks if a reference type is valid
func IsValidReferenceType(refType string) bool {
	validTypes := []ReferenceType{RefTypeNovel, RefTypeNovelChapter}
	for _, validType := range validTypes {
		if string(validType) == refType {
			return true
		}
	}
	return false
}

// IsValidVoteType checks if a vote type is valid
func IsValidVoteType(voteType string) bool {
	validTypes := []VoteType{VoteUpvote, VoteDownvote}
	for _, validType := range validTypes {
		if string(validType) == voteType {
			return true
		}
	}
	return false
}

// GetAllValidStatuses returns all valid translation statuses
func GetAllValidStatuses() []string {
	return []string{
		string(StatusPending),
		string(StatusApproved),
		string(StatusRejected),
	}
}

// GetAllValidReferenceTypes returns all valid reference types
func GetAllValidReferenceTypes() []string {
	return []string{
		string(RefTypeNovel),
		string(RefTypeNovelChapter),
	}
}

// GetAllValidVoteTypes returns all valid vote types
func GetAllValidVoteTypes() []string {
	return []string{
		string(VoteUpvote),
		string(VoteDownvote),
	}
}