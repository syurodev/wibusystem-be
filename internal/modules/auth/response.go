package auth

import "time"

// RegisterResponse là DTO cho response sau khi đăng ký
type RegisterResponse struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// UserProfileResponse là DTO cho thông tin user
type UserProfileResponse struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	EmailVerified bool       `json:"email_verified"`
	FullName      *string    `json:"full_name,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
}
