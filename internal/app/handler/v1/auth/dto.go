package auth

import "time"

// RegisterRequest là DTO cho việc đăng ký user mới
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	FullName string `json:"full_name" binding:"required,min=2,max=255"`
}

// RegisterResponse là DTO cho response sau khi đăng ký
type RegisterResponse struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// VerifyEmailRequest là DTO cho việc verify email
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// ForgotPasswordRequest là DTO cho việc quên mật khẩu
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest là DTO cho việc reset mật khẩu
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=128"`
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
