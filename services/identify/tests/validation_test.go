package tests

import (
	"testing"

	"wibusystem/services/identify/services"
)

func TestValidationService_ValidatePassword(t *testing.T) {
	vs := services.NewValidationService()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid strong password",
			password: "MyStrongPass123!",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Pass1!",
			wantErr:  true,
		},
		{
			name:     "no uppercase",
			password: "mystrongpass123!",
			wantErr:  true,
		},
		{
			name:     "no lowercase",
			password: "MYSTRONGPASS123!",
			wantErr:  true,
		},
		{
			name:     "no numbers",
			password: "MyStrongPass!",
			wantErr:  true,
		},
		{
			name:     "no special characters",
			password: "MyStrongPass123",
			wantErr:  true,
		},
		{
			name:     "common weak password",
			password: "password123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := vs.ValidatePassword(tt.password)
			hasErrors := len(errors) > 0

			if hasErrors != tt.wantErr {
				t.Errorf("ValidatePassword() errors = %v, wantErr %v", errors, tt.wantErr)
			}
		})
	}
}

func TestValidationService_ValidateUsername(t *testing.T) {
	vs := services.NewValidationService()

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "valid username",
			username: "john_doe123",
			wantErr:  false,
		},
		{
			name:     "valid username with hyphen",
			username: "john-doe",
			wantErr:  false,
		},
		{
			name:     "too short",
			username: "jo",
			wantErr:  true,
		},
		{
			name:     "too long",
			username: "this_is_a_very_long_username_that_exceeds_thirty_characters",
			wantErr:  true,
		},
		{
			name:     "starts with number",
			username: "123john",
			wantErr:  true,
		},
		{
			name:     "contains invalid characters",
			username: "john@doe",
			wantErr:  true,
		},
		{
			name:     "reserved username",
			username: "admin",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := vs.ValidateUsername(tt.username)
			hasErrors := len(errors) > 0

			if hasErrors != tt.wantErr {
				t.Errorf("ValidateUsername() errors = %v, wantErr %v", errors, tt.wantErr)
			}
		})
	}
}

func TestValidationService_SanitizeEmail(t *testing.T) {
	vs := services.NewValidationService()

	tests := []struct {
		name    string
		email   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "john.doe@example.com",
			want:    "john.doe@example.com",
			wantErr: false,
		},
		{
			name:    "email with uppercase",
			email:   "John.Doe@EXAMPLE.COM",
			want:    "john.doe@example.com",
			wantErr: false,
		},
		{
			name:    "email with spaces",
			email:   "  john.doe@example.com  ",
			want:    "john.doe@example.com",
			wantErr: false,
		},
		{
			name:    "invalid email format",
			email:   "invalid-email",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty email",
			email:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vs.SanitizeEmail(tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("SanitizeEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInputSanitizer_SanitizeHTML(t *testing.T) {
	is := services.NewInputSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "clean text",
			input: "Hello World",
			want:  "Hello World",
		},
		{
			name:  "text with HTML tags",
			input: "<p>Hello <b>World</b></p>",
			want:  "Hello World",
		},
		{
			name:  "text with script tag",
			input: "<script>alert('xss')</script>Hello",
			want:  "Hello",
		},
		{
			name:  "text with javascript",
			input: "javascript:alert('xss')",
			want:  "alert('xss')",
		},
		{
			name:  "text with data URI",
			input: "data:text/html,<script>alert('xss')</script>",
			want:  "text/html,alert('xss')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := is.SanitizeHTML(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInputSanitizer_SanitizeFileName(t *testing.T) {
	is := services.NewInputSanitizer()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "clean filename",
			input: "document.pdf",
			want:  "document.pdf",
		},
		{
			name:  "filename with path traversal",
			input: "../../../etc/passwd",
			want:  "etcpasswd",
		},
		{
			name:  "filename with dangerous characters",
			input: "file<>:\"|*?.txt",
			want:  "file.txt",
		},
		{
			name:  "filename with spaces and dots",
			input: "  .my file. .txt.  ",
			want:  "my file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := is.SanitizeFileName(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
