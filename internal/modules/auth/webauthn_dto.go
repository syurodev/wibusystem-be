package auth

import "time"

// ==================== Registration DTOs ====================

// RegistrationBeginRequest là DTO cho việc bắt đầu registration flow
type RegistrationBeginRequest struct {
	Email    string  `json:"email" binding:"required,email,max=255"`
	FullName string  `json:"full_name" binding:"required,min=2,max=255"`
	Username *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"` // Optional username for WebAuthn
}

// RegistrationBeginResponse là DTO cho response của registration begin
// Chứa PublicKeyCredentialCreationOptions theo WebAuthn spec
type RegistrationBeginResponse struct {
	Challenge              string                         `json:"challenge"`
	RP                     RelyingParty                   `json:"rp"`
	User                   UserInfo                       `json:"user"`
	PubKeyCredParams       []PubKeyCredParam              `json:"pubKeyCredParams"`
	Timeout                int                            `json:"timeout,omitempty"`
	ExcludeCredentials     []CredentialDescriptor         `json:"excludeCredentials,omitempty"`
	AuthenticatorSelection AuthenticatorSelectionCriteria `json:"authenticatorSelection,omitempty"`
	Attestation            string                         `json:"attestation,omitempty"`
}

// RegistrationFinishRequest là DTO cho việc hoàn thành registration flow
type RegistrationFinishRequest struct {
	ID             string                           `json:"id" binding:"required"`
	RawID          string                           `json:"rawId" binding:"required"`
	Type           string                           `json:"type" binding:"required"`
	Response       AuthenticatorAttestationResponse `json:"response" binding:"required"`
	CredentialName *string                          `json:"credential_name,omitempty" binding:"omitempty,max=255"`
}

// RegistrationFinishResponse là DTO cho response sau khi hoàn thành registration
type RegistrationFinishResponse struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	CredentialID string `json:"credential_id"`
	Message      string `json:"message"`
}

// ==================== Authentication DTOs ====================

// AuthenticationBeginRequest là DTO cho việc bắt đầu authentication flow
type AuthenticationBeginRequest struct {
	Email string `json:"email" binding:"required"` // Can be email or username
}

// AuthenticationBeginResponse là DTO cho response của authentication begin
// Chứa PublicKeyCredentialRequestOptions theo WebAuthn spec
type AuthenticationBeginResponse struct {
	Challenge        string                 `json:"challenge"`
	Timeout          int                    `json:"timeout,omitempty"`
	RPID             string                 `json:"rpId"`
	AllowCredentials []CredentialDescriptor `json:"allowCredentials,omitempty"`
	UserVerification string                 `json:"userVerification,omitempty"`
}

// AuthenticationFinishRequest là DTO cho việc hoàn thành authentication flow
type AuthenticationFinishRequest struct {
	ID       string                         `json:"id" binding:"required"`
	RawID    string                         `json:"rawId" binding:"required"`
	Type     string                         `json:"type" binding:"required"`
	Response AuthenticatorAssertionResponse `json:"response" binding:"required"`
}

// AuthenticationFinishResponse là DTO cho response sau khi hoàn thành authentication
type AuthenticationFinishResponse struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`  // Optional OAuth2 access token
	RefreshToken string `json:"refresh_token,omitempty"` // Optional OAuth2 refresh token
}

// ==================== Credential Management DTOs ====================

// ListCredentialsResponse là DTO cho danh sách credentials của user
type ListCredentialsResponse struct {
	Credentials []CredentialInfo `json:"credentials"`
}

// CredentialInfo là thông tin về một credential
type CredentialInfo struct {
	ID             string     `json:"id"`
	CredentialID   string     `json:"credential_id"`
	CredentialName *string    `json:"credential_name,omitempty"`
	Transports     []string   `json:"transports,omitempty"`
	BackupEligible bool       `json:"backup_eligible"`
	BackupState    bool       `json:"backup_state"`
	CreatedAt      time.Time  `json:"created_at"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
}

// DeleteCredentialRequest là DTO cho việc xóa credential
type DeleteCredentialRequest struct {
	CredentialID string `json:"credential_id" binding:"required"`
}

// UpdateCredentialNameRequest là DTO cho việc cập nhật tên credential
type UpdateCredentialNameRequest struct {
	CredentialID   string `json:"credential_id" binding:"required"`
	CredentialName string `json:"credential_name" binding:"required,max=255"`
}

// ==================== WebAuthn Spec Types ====================

// RelyingParty represents the relying party (your application)
type RelyingParty struct {
	ID   string `json:"id"`   // e.g., "example.com"
	Name string `json:"name"` // e.g., "Example Corp"
}

// UserInfo represents user information for WebAuthn
type UserInfo struct {
	ID          string `json:"id"`          // Base64URL encoded user ID
	Name        string `json:"name"`        // e.g., "user@example.com"
	DisplayName string `json:"displayName"` // e.g., "John Doe"
}

// PubKeyCredParam represents a public key credential parameter
type PubKeyCredParam struct {
	Type string `json:"type"` // "public-key"
	Alg  int    `json:"alg"`  // COSE algorithm identifier (e.g., -7 for ES256, -257 for RS256)
}

// CredentialDescriptor describes a credential
type CredentialDescriptor struct {
	Type       string   `json:"type"`                 // "public-key"
	ID         string   `json:"id"`                   // Base64URL encoded credential ID
	Transports []string `json:"transports,omitempty"` // usb, nfc, ble, internal, hybrid
}

// AuthenticatorSelectionCriteria specifies requirements for the authenticator
type AuthenticatorSelectionCriteria struct {
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"` // platform, cross-platform
	RequireResidentKey      bool   `json:"requireResidentKey,omitempty"`
	ResidentKey             string `json:"residentKey,omitempty"`      // discouraged, preferred, required
	UserVerification        string `json:"userVerification,omitempty"` // required, preferred, discouraged
}

// AuthenticatorAttestationResponse represents the response from authenticator during registration
type AuthenticatorAttestationResponse struct {
	ClientDataJSON    string   `json:"clientDataJSON" binding:"required"`
	AttestationObject string   `json:"attestationObject" binding:"required"`
	Transports        []string `json:"transports,omitempty"`
}

// AuthenticatorAssertionResponse represents the response from authenticator during authentication
type AuthenticatorAssertionResponse struct {
	ClientDataJSON    string `json:"clientDataJSON" binding:"required"`
	AuthenticatorData string `json:"authenticatorData" binding:"required"`
	Signature         string `json:"signature" binding:"required"`
	UserHandle        string `json:"userHandle,omitempty"`
}
