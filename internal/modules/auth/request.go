package auth

// RegisterRequest là DTO cho việc đăng ký user mới
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	FullName string `json:"full_name" binding:"required,min=2,max=255"`
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
