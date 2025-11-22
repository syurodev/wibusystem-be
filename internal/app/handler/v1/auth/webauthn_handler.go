package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/pkg/service"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
)

// WebAuthnHandler handles WebAuthn/Passkey authentication endpoints
type WebAuthnHandler struct {
	webAuthnService *service.WebAuthnService
}

// NewWebAuthnHandler creates a new instance of WebAuthnHandler
func NewWebAuthnHandler(webAuthnService *service.WebAuthnService) *WebAuthnHandler {
	return &WebAuthnHandler{
		webAuthnService: webAuthnService,
	}
}

// ==================== Registration Endpoints ====================

// BeginRegistration initiates the passkey registration process
// @Summary Begin passkey registration
// @Description Starts the WebAuthn registration ceremony for a new passkey
// @Tags WebAuthn
// @Accept json
// @Produce json
// @Param request body RegistrationBeginRequest true "Registration begin request"
// @Success 200 {object} RegistrationBeginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/passkey/register/begin [post]
func (h *WebAuthnHandler) BeginRegistration(c *gin.Context) {
	var req RegistrationBeginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Begin registration ceremony
	options, err := h.webAuthnService.BeginRegistration(c.Request.Context(), req.Email, req.FullName)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "REGISTRATION_BEGIN_FAILED", "webauthn.registration_begin_failed", nil)
		zap.L().Error("Failed to begin passkey registration", zap.Error(err), zap.String("email", req.Email))
		return
	}

	// Convert to response DTO
	resp := convertToRegistrationBeginResponse(options)

	response.JSON(c, http.StatusOK, resp)
}

// FinishRegistration completes the passkey registration process
// @Summary Finish passkey registration
// @Description Completes the WebAuthn registration ceremony and stores the new credential
// @Tags WebAuthn
// @Accept json
// @Produce json
// @Param request body RegistrationFinishRequest true "Registration finish request"
// @Success 200 {object} RegistrationFinishResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/passkey/register/finish [post]
func (h *WebAuthnHandler) FinishRegistration(c *gin.Context) {
	var req RegistrationFinishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Parse credential creation response
	parsedResponse, err := parseCredentialCreationResponse(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CREDENTIAL_DATA", "webauthn.invalid_credential_data", err.Error())
		zap.L().Error("Failed to parse credential creation response", zap.Error(err))
		return
	}

	// Get email from session/context (you may want to get this from a temporary session)
	// For now, we'll expect it to be passed in the request or extracted from the credential
	email := c.GetString("registration_email")
	if email == "" {
		// Try to extract from user handle or require it in the request
		response.Error(c, http.StatusBadRequest, "MISSING_EMAIL", "webauthn.missing_email", nil)
		return
	}

	// Finish registration
	user, credential, err := h.webAuthnService.FinishRegistration(c.Request.Context(), email, req.CredentialName, parsedResponse)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidOrExpiredSession) {
			response.Error(c, http.StatusUnauthorized, "INVALID_SESSION", "webauthn.invalid_or_expired_session", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "REGISTRATION_FAILED", "webauthn.registration_failed", nil)
		zap.L().Error("Failed to finish passkey registration", zap.Error(err))
		return
	}

	resp := RegistrationFinishResponse{
		UserID:       user.ID.String(),
		Email:        user.Email,
		CredentialID: credential.CredentialID,
		Message:      "Passkey registered successfully",
	}

	response.JSON(c, http.StatusOK, resp)
}

// ==================== Authentication Endpoints ====================

// BeginAuthentication initiates the passkey authentication process
// @Summary Begin passkey authentication
// @Description Starts the WebAuthn authentication ceremony
// @Tags WebAuthn
// @Accept json
// @Produce json
// @Param request body AuthenticationBeginRequest true "Authentication begin request"
// @Success 200 {object} AuthenticationBeginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/passkey/authenticate/begin [post]
func (h *WebAuthnHandler) BeginAuthentication(c *gin.Context) {
	var req AuthenticationBeginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Begin authentication ceremony
	options, err := h.webAuthnService.BeginAuthentication(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrNoPasskeyRegistered) {
			response.Error(c, http.StatusNotFound, "NO_PASSKEY", "webauthn.no_passkey_registered", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "AUTHENTICATION_BEGIN_FAILED", "webauthn.authentication_begin_failed", nil)
		zap.L().Error("Failed to begin passkey authentication", zap.Error(err), zap.String("email", req.Email))
		return
	}

	// Convert to response DTO
	resp := convertToAuthenticationBeginResponse(options)

	response.JSON(c, http.StatusOK, resp)
}

// FinishAuthentication completes the passkey authentication process
// @Summary Finish passkey authentication
// @Description Completes the WebAuthn authentication ceremony and authenticates the user
// @Tags WebAuthn
// @Accept json
// @Produce json
// @Param request body AuthenticationFinishRequest true "Authentication finish request"
// @Success 200 {object} AuthenticationFinishResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/passkey/authenticate/finish [post]
func (h *WebAuthnHandler) FinishAuthentication(c *gin.Context) {
	var req AuthenticationFinishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Parse credential assertion response
	parsedResponse, err := parseCredentialAssertionResponse(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CREDENTIAL_DATA", "webauthn.invalid_credential_data", err.Error())
		zap.L().Error("Failed to parse credential assertion response", zap.Error(err))
		return
	}

	// Get email from session (you may want to get this from a temporary session)
	email := c.GetString("authentication_email")
	if email == "" {
		response.Error(c, http.StatusBadRequest, "MISSING_EMAIL", "webauthn.missing_email", nil)
		return
	}

	// Finish authentication
	user, _, err := h.webAuthnService.FinishAuthentication(c.Request.Context(), email, parsedResponse)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "user.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidOrExpiredSession) {
			response.Error(c, http.StatusUnauthorized, "INVALID_SESSION", "webauthn.invalid_or_expired_session", nil)
			return
		}
		response.Error(c, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "webauthn.authentication_failed", nil)
		zap.L().Error("Failed to finish passkey authentication", zap.Error(err))
		return
	}

	resp := AuthenticationFinishResponse{
		UserID:  user.ID.String(),
		Email:   user.Email,
		Message: "Authentication successful",
		// You may want to generate OAuth2 tokens here and include them in the response
	}

	response.JSON(c, http.StatusOK, resp)
}

// ==================== Credential Management Endpoints ====================

// ListCredentials returns all passkeys for the authenticated user
// @Summary List user's passkeys
// @Description Get all registered passkeys for the authenticated user
// @Tags WebAuthn
// @Produce json
// @Success 200 {object} ListCredentialsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/passkey/credentials [get]
// @Security BearerAuth
func (h *WebAuthnHandler) ListCredentials(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	// Parse user ID
	uid, err := parseUUID(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	// Get credentials
	credentials, err := h.webAuthnService.GetUserCredentials(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "FAILED_TO_GET_CREDENTIALS", "webauthn.failed_to_get_credentials", nil)
		zap.L().Error("Failed to get user credentials", zap.Error(err), zap.String("user_id", userIDStr))
		return
	}

	// Convert to response
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

	response.JSON(c, http.StatusOK, resp)
}

// DeleteCredential deletes a passkey
// @Summary Delete a passkey
// @Description Delete a registered passkey for the authenticated user
// @Tags WebAuthn
// @Accept json
// @Produce json
// @Param request body DeleteCredentialRequest true "Delete credential request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/passkey/credentials [delete]
// @Security BearerAuth
func (h *WebAuthnHandler) DeleteCredential(c *gin.Context) {
	var req DeleteCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	uid, err := parseUUID(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	// Delete credential
	err = h.webAuthnService.DeleteCredential(c.Request.Context(), uid, req.CredentialID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrCredentialNotFound) {
			response.Error(c, http.StatusNotFound, "CREDENTIAL_NOT_FOUND", "webauthn.credential_not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrUnauthorized) {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "webauthn.delete_failed", nil)
		zap.L().Error("Failed to delete credential", zap.Error(err), zap.String("credential_id", req.CredentialID))
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"message": "Credential deleted successfully"})
}

// UpdateCredentialName updates the name of a passkey
// @Summary Update passkey name
// @Description Update the friendly name of a registered passkey
// @Tags WebAuthn
// @Accept json
// @Produce json
// @Param request body UpdateCredentialNameRequest true "Update credential name request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/passkey/credentials/name [put]
// @Security BearerAuth
func (h *WebAuthnHandler) UpdateCredentialName(c *gin.Context) {
	var req UpdateCredentialNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	uid, err := parseUUID(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	// Update credential name
	err = h.webAuthnService.UpdateCredentialName(c.Request.Context(), uid, req.CredentialID, req.CredentialName)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrCredentialNotFound) {
			response.Error(c, http.StatusNotFound, "CREDENTIAL_NOT_FOUND", "webauthn.credential_not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrUnauthorized) {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "webauthn.update_failed", nil)
		zap.L().Error("Failed to update credential name", zap.Error(err), zap.String("credential_id", req.CredentialID))
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"message": "Credential name updated successfully"})
}

// ==================== Helper Functions ====================

// convertToRegistrationBeginResponse converts WebAuthn options to DTO
func convertToRegistrationBeginResponse(options *protocol.CredentialCreation) RegistrationBeginResponse {
	return RegistrationBeginResponse{
		Challenge: protocol.URLEncodedBase64(options.Response.Challenge),
		RP: RelyingParty{
			ID:   options.Response.RelyingParty.ID,
			Name: options.Response.RelyingParty.Name,
		},
		User: UserInfo{
			ID:          protocol.URLEncodedBase64(options.Response.User.ID),
			Name:        options.Response.User.Name,
			DisplayName: options.Response.User.DisplayName,
		},
		PubKeyCredParams:       convertPubKeyCredParams(options.Response.Parameters),
		Timeout:                int(options.Response.Timeout),
		ExcludeCredentials:     convertCredentialDescriptors(options.Response.CredentialExcludeList),
		AuthenticatorSelection: convertAuthenticatorSelection(options.Response.AuthenticatorSelection),
		Attestation:            string(options.Response.Attestation),
	}
}

// convertToAuthenticationBeginResponse converts WebAuthn options to DTO
func convertToAuthenticationBeginResponse(options *protocol.CredentialAssertion) AuthenticationBeginResponse {
	return AuthenticationBeginResponse{
		Challenge:        protocol.URLEncodedBase64(options.Response.Challenge),
		Timeout:          int(options.Response.Timeout),
		RPID:             options.Response.RelyingPartyID,
		AllowCredentials: convertCredentialDescriptors(options.Response.AllowedCredentials),
		UserVerification: string(options.Response.UserVerification),
	}
}

// parseCredentialCreationResponse parses the credential creation response from client
func parseCredentialCreationResponse(req *RegistrationFinishRequest) (*protocol.ParsedCredentialCreationData, error) {
	// Create the protocol credential creation response
	response := &protocol.CredentialCreationResponse{
		PublicKeyCredential: protocol.PublicKeyCredential{
			Credential: protocol.Credential{
				ID:   req.ID,
				Type: req.Type,
			},
			RawID: []byte(req.RawID),
		},
	}

	// Parse the response
	parsedResponse, err := response.Parse()
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}

// parseCredentialAssertionResponse parses the credential assertion response from client
func parseCredentialAssertionResponse(req *AuthenticationFinishRequest) (*protocol.ParsedCredentialAssertionData, error) {
	// Create the protocol credential assertion response
	response := &protocol.CredentialAssertionResponse{
		PublicKeyCredential: protocol.PublicKeyCredential{
			Credential: protocol.Credential{
				ID:   req.ID,
				Type: req.Type,
			},
			RawID: []byte(req.RawID),
		},
	}

	// Parse the response
	parsedResponse, err := response.Parse()
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}

// Helper conversion functions
func convertPubKeyCredParams(params []protocol.CredentialParameter) []PubKeyCredParam {
	result := make([]PubKeyCredParam, len(params))
	for i, p := range params {
		result[i] = PubKeyCredParam{
			Type: string(p.Type),
			Alg:  int(p.Algorithm),
		}
	}
	return result
}

func convertCredentialDescriptors(descriptors []protocol.CredentialDescriptor) []CredentialDescriptor {
	result := make([]CredentialDescriptor, len(descriptors))
	for i, d := range descriptors {
		transports := make([]string, len(d.Transport))
		for j, t := range d.Transport {
			transports[j] = string(t)
		}
		result[i] = CredentialDescriptor{
			Type:       string(d.Type),
			ID:         protocol.URLEncodedBase64(d.CredentialID),
			Transports: transports,
		}
	}
	return result
}

func convertAuthenticatorSelection(selection protocol.AuthenticatorSelection) AuthenticatorSelectionCriteria {
	return AuthenticatorSelectionCriteria{
		AuthenticatorAttachment: string(selection.AuthenticatorAttachment),
		RequireResidentKey:      selection.RequireResidentKey != nil && *selection.RequireResidentKey,
		ResidentKey:             string(selection.ResidentKey),
		UserVerification:        string(selection.UserVerification),
	}
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.FromString(s)
}
