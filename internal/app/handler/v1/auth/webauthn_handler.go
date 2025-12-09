package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/i18nkeys"
	"system/pkg/util/response"
	webauthnutil "system/pkg/util/webauthn"
)

// ==================== Registration Handlers ====================

// PasskeyRegisterBegin bắt đầu quá trình đăng ký passkey cho user đã authenticated
func (h *Handler) PasskeyRegisterBegin(c *gin.Context) {
	// Get authenticated user ID from context (set by middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", i18nkeys.AuthInvalidUserID, nil)
		return
	}

	zap.L().Info("PasskeyRegisterBegin called", zap.String("user_id", userID.String()))

	// Extract origin from Referer header
	// This is safe because the browser sends the Referer automatically
	referer := c.GetHeader("Referer")
	var clientOrigin string
	if referer != "" {
		// Import helper at top of file
		clientOrigin = webauthnutil.ExtractOrigin(referer)
		zap.L().Info("Extracted client origin from Referer", 
			zap.String("referer", referer), 
			zap.String("origin", clientOrigin))
	}

	// Begin registration
	options, err := h.webauthnService.BeginRegistration(c.Request.Context(), userID, clientOrigin)
	if err != nil {
		zap.L().Error("Failed to begin passkey registration", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, http.StatusInternalServerError, "REGISTRATION_BEGIN_FAILED", i18nkeys.WebAuthnRegistrationBeginFailed, nil)
		return
	}

	// Return options directly - protocol.CredentialCreation is already JSON-serializable
	c.JSON(http.StatusOK, options)
}

// PasskeyRegisterFinish hoàn thành quá trình đăng ký passkey
func (h *Handler) PasskeyRegisterFinish(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", i18nkeys.AuthInvalidUserID, nil)
		return
	}

	zap.L().Info("PasskeyRegisterFinish called", zap.String("user_id", userID.String()))

	// Parse credential creation response from client
	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(c.Request.Body)
	if err != nil {
		zap.L().Error("Failed to parse credential creation response", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "INVALID_CREDENTIAL_DATA", i18nkeys.WebAuthnInvalidCredentialData, nil)
		return
	}

	// Debug: log attestation format
	zap.L().Info("Parsed credential creation response", 
		zap.String("attestation_format", parsedResponse.Response.AttestationObject.Format))

	// Extract origin from Referer (same as BeginRegistration)
	referer := c.GetHeader("Referer")
	var clientOrigin string
	if referer != "" {
		clientOrigin = webauthnutil.ExtractOrigin(referer)
		zap.L().Info("Extracted client origin from Referer for finish", 
			zap.String("referer", referer), 
			zap.String("origin", clientOrigin))
	}

	// Get optional credential name from query or header
	var credentialName *string
	if name := c.Query("credential_name"); name != "" {
		credentialName = &name
	}

	// Finish registration
	credential, err := h.webauthnService.FinishRegistration(c.Request.Context(), userID, parsedResponse, credentialName, clientOrigin)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidOrExpiredSession) {
			response.Error(c, http.StatusBadRequest, "INVALID_SESSION", i18nkeys.WebAuthnInvalidOrExpiredSession, nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrCredentialAlreadyExists) {
			response.Error(c, http.StatusConflict, "CREDENTIAL_EXISTS", i18nkeys.WebAuthnCredentialAlreadyExists, nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrCredentialVerificationFailed) {
			response.Error(c, http.StatusBadRequest, "VERIFICATION_FAILED", i18nkeys.WebAuthnCredentialVerificationFailed, nil)
			return
		}
		zap.L().Error("Failed to finish passkey registration", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, http.StatusInternalServerError, "REGISTRATION_FAILED", i18nkeys.WebAuthnRegistrationFailed, nil)
		return
	}

	resp := RegistrationFinishResponse{
		UserID:       userID.String(),
		CredentialID: credential.ID.String(),
		Message:      "Passkey registered successfully",
	}

	response.Success(c, http.StatusCreated, i18nkeys.WebAuthnPasskeyRegistered, resp, nil)
}

// ==================== Authentication Handlers ====================

// PasskeyAuthenticateBegin bắt đầu quá trình xác thực passkey
func (h *Handler) PasskeyAuthenticateBegin(c *gin.Context) {
	var req AuthenticationBeginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", i18nkeys.ValidationFailed, err.Error())
		return
	}

	// Get user by email or username
	identifier := strings.TrimSpace(req.Email)
	var user *domain.User
	var err error

	if strings.Contains(identifier, "@") {
		user, err = h.authService.GetUserByEmail(c.Request.Context(), strings.ToLower(identifier))
	} else {
		user, err = h.authService.GetUserByUsername(c.Request.Context(), identifier)
	}

	if err != nil {
		// Don't reveal if user exists or not
		response.Error(c, http.StatusUnauthorized, "AUTHENTICATION_FAILED", i18nkeys.WebAuthnAuthenticationFailed, nil)
		return
	}

	// Begin authentication
	options, err := h.webauthnService.BeginAuthentication(c.Request.Context(), user.ID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNoPasskeyRegistered) {
			response.Error(c, http.StatusNotFound, "NO_PASSKEY", i18nkeys.WebAuthnNoPasskeyRegistered, nil)
			return
		}
		zap.L().Error("Failed to begin passkey authentication", zap.Error(err), zap.String("user_id", user.ID.String()))
		response.Error(c, http.StatusInternalServerError, "AUTHENTICATION_BEGIN_FAILED", i18nkeys.WebAuthnAuthenticationBeginFailed, nil)
		return
	}

	// Store user ID in session for finish step
	// We use a temporary session cookie to carry user_id between begin and finish
	sessionID := uuid.Must(uuid.NewV7()).String()
	c.SetCookie("webauthn_session", sessionID, 300, "/", "", false, true) // 5 minutes, httpOnly

	// In production, you should store sessionID -> userID mapping in Redis
	// For now, we'll include user_id in the response (client must send it back)

	// Return options with user_id hint
	result := map[string]any{
		"publicKey": options,
		"user_id":   user.ID.String(), // Client must send this back
	}

	c.JSON(http.StatusOK, result)
}

// PasskeyAuthenticateFinish hoàn thành quá trình xác thực passkey
func (h *Handler) PasskeyAuthenticateFinish(c *gin.Context) {
	// Get user_id from request (sent by client from begin response)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		response.Error(c, http.StatusBadRequest, "MISSING_USER_ID", i18nkeys.WebAuthnMissingUserID, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", i18nkeys.AuthInvalidUserID, nil)
		return
	}

	// Parse assertion response from client
	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(c.Request.Body)
	if err != nil {
		zap.L().Error("Failed to parse credential assertion response", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "INVALID_CREDENTIAL_DATA", i18nkeys.WebAuthnInvalidCredentialData, nil)
		return
	}

	// Finish authentication
	user, err := h.webauthnService.FinishAuthentication(c.Request.Context(), userID, parsedResponse)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidOrExpiredSession) {
			response.Error(c, http.StatusBadRequest, "INVALID_SESSION", i18nkeys.WebAuthnInvalidOrExpiredSession, nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrAuthenticationFailed) {
			response.Error(c, http.StatusUnauthorized, "AUTHENTICATION_FAILED", i18nkeys.WebAuthnAuthenticationFailed, nil)
			return
		}
		zap.L().Error("Failed to finish passkey authentication", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, http.StatusInternalServerError, "AUTHENTICATION_FAILED", i18nkeys.WebAuthnAuthenticationFailed, nil)
		return
	}

	// Clear temporary session cookie
	c.SetCookie("webauthn_session", "", -1, "/", "", false, true)

	// Create OAuth2 session
	sessionID, err := h.oauth2Service.CreateUserSession(
		c.Request.Context(),
		user.ID,
		time.Hour,
		c.Request.UserAgent(),
		c.ClientIP(),
	)
	if err != nil {
		zap.L().Error("Failed to create OAuth2 session after passkey auth", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "SESSION_CREATION_FAILED", i18nkeys.WebAuthnSessionCreationFailed, nil)
		return
	}

	// Set secure session cookie
	c.SetCookie(
		"session_id",
		sessionID,
		3600,  // 1 hour
		"/",   // path
		"",    // domain
		false, // secure (set true in production with HTTPS)
		true,  // httpOnly
	)

	resp := AuthenticationFinishResponse{
		UserID:  user.ID.String(),
		Email:   user.Email,
		Message: "Authentication successful",
	}

	response.Success(c, http.StatusOK, i18nkeys.WebAuthnAuthenticationSuccess, resp, nil)
}

// ==================== Credential Management Handlers ====================

// PasskeyListCredentials trả về danh sách tất cả passkeys của user
func (h *Handler) PasskeyListCredentials(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", i18nkeys.AuthInvalidUserID, nil)
		return
	}

	// Get credentials
	credentials, err := h.webauthnService.ListUserCredentials(c.Request.Context(), userID)
	if err != nil {
		zap.L().Error("Failed to list credentials", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", i18nkeys.WebAuthnListCredentialsFailed, nil)
		return
	}

	// Convert to response format
	credInfos := make([]CredentialInfo, len(credentials))
	for i, cred := range credentials {
		credInfos[i] = CredentialInfo{
			ID:             cred.ID.String(),
			CredentialID:   cred.CredentialID,
			CredentialName: cred.CredentialName,
			Transports:     cred.Transports,
			BackupEligible: cred.BackupEligible,
			BackupState:    cred.BackupState,
			CreatedAt:      cred.CreatedAt,
			LastUsedAt:     cred.LastUsedAt,
		}
	}

	resp := ListCredentialsResponse{
		Credentials: credInfos,
	}

	response.Success(c, http.StatusOK, i18nkeys.WebAuthnCredentialsListed, resp, nil)
}

// PasskeyDeleteCredential xóa một passkey
func (h *Handler) PasskeyDeleteCredential(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", i18nkeys.AuthInvalidUserID, nil)
		return
	}

	var req DeleteCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", i18nkeys.ValidationFailed, err.Error())
		return
	}

	credentialID, err := uuid.FromString(req.CredentialID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CREDENTIAL_ID", i18nkeys.WebAuthnInvalidCredentialID, nil)
		return
	}

	// Delete credential
	if err := h.webauthnService.DeleteCredential(c.Request.Context(), userID, credentialID); err != nil {
		if errors.Is(err, pkgerrors.ErrCredentialNotFound) {
			response.Error(c, http.StatusNotFound, "CREDENTIAL_NOT_FOUND", i18nkeys.WebAuthnCredentialNotFound, nil)
			return
		}
		zap.L().Error("Failed to delete credential", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", i18nkeys.WebAuthnDeleteCredentialFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, i18nkeys.WebAuthnCredentialDeleted, nil, nil)
}

// PasskeyUpdateCredentialName cập nhật tên của passkey
func (h *Handler) PasskeyUpdateCredentialName(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", i18nkeys.AuthInvalidUserID, nil)
		return
	}

	var req UpdateCredentialNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", i18nkeys.ValidationFailed, err.Error())
		return
	}

	credentialID, err := uuid.FromString(req.CredentialID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CREDENTIAL_ID", i18nkeys.WebAuthnInvalidCredentialID, nil)
		return
	}

	// Update credential name
	if err := h.webauthnService.UpdateCredentialName(c.Request.Context(), userID, credentialID, req.CredentialName); err != nil {
		if errors.Is(err, pkgerrors.ErrCredentialNotFound) {
			response.Error(c, http.StatusNotFound, "CREDENTIAL_NOT_FOUND", i18nkeys.WebAuthnCredentialNotFound, nil)
			return
		}
		zap.L().Error("Failed to update credential name", zap.Error(err), zap.String("user_id", userID.String()))
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", i18nkeys.WebAuthnUpdateCredentialFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, i18nkeys.WebAuthnCredentialUpdated, nil, nil)
}
