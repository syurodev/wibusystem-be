package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

// PasskeyListFragment returns HTML fragment for passkey list (HTMX)
func (h *Handler) PasskeyListFragment(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.String(http.StatusUnauthorized, "<p class='error'>Unauthorized</p>")
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		c.String(http.StatusBadRequest, "<p class='error'>Invalid user ID</p>")
		return
	}

	// Get credentials
	credentials, err := h.webauthnService.ListUserCredentials(c.Request.Context(), userID)
	if err != nil {
		zap.L().Error("Failed to list credentials", zap.Error(err), zap.String("user_id", userID.String()))
		c.String(http.StatusInternalServerError, "<p class='error'>Failed to load passkeys</p>")
		return
	}

	// Render partial template
	c.HTML(http.StatusOK, "auth/partials/passkey_list.html", gin.H{
		"credentials": credentials,
	})
}

// PasskeyDeleteFragment handles passkey deletion (HTMX)
func (h *Handler) PasskeyDeleteFragment(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.Header("HX-Trigger", `{"showError": "Unauthorized"}`)
		c.Status(http.StatusUnauthorized)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		c.Header("HX-Trigger", `{"showError": "Invalid user ID"}`)
		c.Status(http.StatusBadRequest)
		return
	}

	credentialIDStr := c.Param("id")
	if credentialIDStr == "" {
		c.Header("HX-Trigger", `{"showError": "Missing credential ID"}`)
		c.Status(http.StatusBadRequest)
		return
	}

	credentialID, err := uuid.FromString(credentialIDStr)
	if err != nil {
		c.Header("HX-Trigger", `{"showError": "Invalid credential ID"}`)
		c.Status(http.StatusBadRequest)
		return
	}

	// Delete credential
	if err := h.webauthnService.DeleteCredential(c.Request.Context(), userID, credentialID); err != nil {
		zap.L().Error("Failed to delete credential", zap.Error(err), zap.String("user_id", userID.String()))
		c.Header("HX-Trigger", `{"showError": "Failed to delete passkey"}`)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Return empty response - HTMX will remove the element
	c.Header("HX-Trigger", `{"showSuccess": "Passkey deleted successfully"}`)
	c.Status(http.StatusOK)
}

// PasskeyUpdateNameFragment handles passkey name update (HTMX)
func (h *Handler) PasskeyUpdateNameFragment(c *gin.Context) {
	// Get authenticated user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.Header("HX-Trigger", `{"showError": "Unauthorized"}`)
		c.Status(http.StatusUnauthorized)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		c.Header("HX-Trigger", `{"showError": "Invalid user ID"}`)
		c.Status(http.StatusBadRequest)
		return
	}

	credentialIDStr := c.Param("id")
	if credentialIDStr == "" {
		c.Header("HX-Trigger", `{"showError": "Missing credential ID"}`)
		c.Status(http.StatusBadRequest)
		return
	}

	credentialID, err := uuid.FromString(credentialIDStr)
	if err != nil {
		c.Header("HX-Trigger", `{"showError": "Invalid credential ID"}`)
		c.Status(http.StatusBadRequest)
		return
	}

	// Get new name from form
	newName := c.PostForm("credential_name")
	if newName == "" {
		c.Header("HX-Trigger", `{"showError": "Name cannot be empty"}`)
		c.Status(http.StatusBadRequest)
		return
	}

	// Update credential name
	if err := h.webauthnService.UpdateCredentialName(c.Request.Context(), userID, credentialID, newName); err != nil {
		zap.L().Error("Failed to update credential name", zap.Error(err), zap.String("user_id", userID.String()))
		c.Header("HX-Trigger", `{"showError": "Failed to update passkey name"}`)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Get updated credential to re-render
	credentials, err := h.webauthnService.ListUserCredentials(c.Request.Context(), userID)
	if err != nil {
		c.Header("HX-Trigger", `{"showError": "Failed to reload passkeys"}`)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Find the updated credential
	var updatedCred interface{}
	for _, cred := range credentials {
		if cred.ID == credentialID {
			updatedCred = cred
			break
		}
	}

	if updatedCred == nil {
		c.Header("HX-Trigger", `{"showError": "Credential not found"}`)
		c.Status(http.StatusNotFound)
		return
	}

	// Return updated card HTML
	c.Header("HX-Trigger", `{"showSuccess": "Passkey renamed successfully"}`)
	c.HTML(http.StatusOK, "auth/partials/passkey_card.html", gin.H{
		"credential": updatedCred,
	})
}

// Helper function to format time
func formatTimeAgo(t *time.Time) string {
	if t == nil {
		return "Never"
	}

	now := time.Now()
	diff := now.Sub(*t)

	days := int(diff.Hours() / 24)
	if days == 0 {
		return "Today"
	}
	if days == 1 {
		return "Yesterday"
	}
	if days < 7 {
		return formatPlural(days, "day")
	}
	if days < 30 {
		weeks := days / 7
		return formatPlural(weeks, "week")
	}

	return t.Format("Jan 2, 2006")
}

func formatPlural(count int, unit string) string {
	if count == 1 {
		return "1 " + unit + " ago"
	}
	return formatInt(count) + " " + unit + "s ago"
}

func formatInt(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
