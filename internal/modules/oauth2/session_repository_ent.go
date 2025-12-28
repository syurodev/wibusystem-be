// ============================================================================
// OAuth2 Session Repository (Ent Implementation)
// ============================================================================
//
// Repository này triển khai OAuth2SessionRepository sử dụng Ent ORM.
// Dùng cho Fosite token store để lưu access tokens, refresh tokens, etc.
//
// ============================================================================

package oauth2

import (
	"context"
	"encoding/json"
	"system/internal/platform/database"
	"time"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/oauth2session"
	pkgerrors "system/pkg/errors"
)

// oauth2SessionEntRepository triển khai OAuth2SessionRepository sử dụng Ent
type oauth2SessionEntRepository struct {
	client *ent.Client
}

// NewOAuth2SessionEntRepository tạo instance mới
func NewOAuth2SessionEntRepository(client *ent.Client) domain.OAuth2SessionRepository {
	return &oauth2SessionEntRepository{client: client}
}

// CreateSession lưu một session mới vào database
func (r *oauth2SessionEntRepository) CreateSession(ctx context.Context, signature string, requestID string, sessionType string, sessionData []byte, expiresAt time.Time, clientID string, subject string) error {
	_, err := database.GetClientFromContext(ctx, r.client).OAuth2Session.Create().
		SetSignature(signature).
		SetRequestID(requestID).
		SetSessionType(sessionType).
		SetSessionData(json.RawMessage(sessionData)).
		SetExpiresAt(expiresAt).
		SetClientID(clientID).
		SetSubjectID(subject).
		SetActive(true).
		Save(ctx)
	return err
}

// GetSessionBySignature lấy một session từ database bằng signature
func (r *oauth2SessionEntRepository) GetSessionBySignature(ctx context.Context, signature string) ([]byte, error) {
	session, err := database.GetClientFromContext(ctx, r.client).OAuth2Session.Query().
		Where(
			oauth2session.SignatureEQ(signature),
			oauth2session.ActiveEQ(true),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerrors.ErrResourceNotFound
		}
		return nil, err
	}
	return session.SessionData, nil
}

// GetSessionWithClientBySignature lấy session data và client_id từ database
func (r *oauth2SessionEntRepository) GetSessionWithClientBySignature(ctx context.Context, signature string) ([]byte, string, error) {
	session, err := database.GetClientFromContext(ctx, r.client).OAuth2Session.Query().
		Where(
			oauth2session.SignatureEQ(signature),
			oauth2session.ActiveEQ(true),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, "", pkgerrors.ErrResourceNotFound
		}
		return nil, "", err
	}
	return session.SessionData, session.ClientID, nil
}

// DeleteSessionBySignature xóa một session khỏi database
func (r *oauth2SessionEntRepository) DeleteSessionBySignature(ctx context.Context, signature string) error {
	_, err := database.GetClientFromContext(ctx, r.client).OAuth2Session.Delete().
		Where(oauth2session.SignatureEQ(signature)).
		Exec(ctx)
	return err
}

// RevokeAllUserSessions đánh dấu inactive tất cả sessions của một user
func (r *oauth2SessionEntRepository) RevokeAllUserSessions(ctx context.Context, subjectID string) error {
	_, err := database.GetClientFromContext(ctx, r.client).OAuth2Session.Update().
		Where(
			oauth2session.SubjectIDEQ(subjectID),
			oauth2session.ActiveEQ(true),
		).
		SetActive(false).
		Save(ctx)
	return err
}

// GetActiveSessionsBySubject lấy tất cả signatures của sessions đang active của một user
func (r *oauth2SessionEntRepository) GetActiveSessionsBySubject(ctx context.Context, subjectID string) ([]string, error) {
	sessions, err := database.GetClientFromContext(ctx, r.client).OAuth2Session.Query().
		Where(
			oauth2session.SubjectIDEQ(subjectID),
			oauth2session.ActiveEQ(true),
		).
		Select(oauth2session.FieldSignature).
		All(ctx)
	if err != nil {
		return nil, err
	}

	signatures := make([]string, len(sessions))
	for i, s := range sessions {
		signatures[i] = s.Signature
	}
	return signatures, nil
}
