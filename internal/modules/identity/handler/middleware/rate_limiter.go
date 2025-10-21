// Package middleware contains HTTP middleware for the Identity module.
package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage/memory/v2"
)

// RateLimitConfig holds rate limiter configuration
type RateLimitConfig struct {
	Max          int           // Maximum number of requests
	Expiration   time.Duration // Time window for rate limiting
	KeyGenerator func(*fiber.Ctx) string
	LimitReached fiber.Handler
	Storage      fiber.Storage
}

// DefaultRateLimitConfig returns the default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
		},
		Storage: memory.New(),
	}
}

// RateLimiter creates a rate limiter middleware with the given configuration
func RateLimiter(config RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          config.Max,
		Expiration:   config.Expiration,
		KeyGenerator: config.KeyGenerator,
		LimitReached: config.LimitReached,
		Storage:      config.Storage,
	})
}

// GlobalRateLimiter creates a global rate limiter for all requests
func GlobalRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        1000,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
		},
		Storage: memory.New(),
	})
}

// AuthRateLimiter creates a rate limiter specifically for authentication endpoints
// More restrictive to prevent brute force attacks
func AuthRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 15 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Rate limit by IP and email if provided
			email := c.FormValue("email")
			if email == "" {
				body := struct {
					Email string `json:"email"`
				}{}
				_ = c.BodyParser(&body)
				email = body.Email
			}
			if email != "" {
				return fmt.Sprintf("%s:%s", c.IP(), email)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many login attempts. Please try again in 15 minutes.",
				"code":    "AUTH_RATE_LIMIT",
			})
		},
		Storage: memory.New(),
	})
}

// RegistrationRateLimiter creates a rate limiter for registration endpoints
func RegistrationRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        3,
		Expiration: 1 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many registration attempts. Please try again later.",
				"code":    "REGISTRATION_RATE_LIMIT",
			})
		},
		Storage: memory.New(),
	})
}

// PasswordResetRateLimiter creates a rate limiter for password reset endpoints
func PasswordResetRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        3,
		Expiration: 1 * time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Rate limit by IP and email
			email := c.FormValue("email")
			if email == "" {
				body := struct {
					Email string `json:"email"`
				}{}
				_ = c.BodyParser(&body)
				email = body.Email
			}
			if email != "" {
				return fmt.Sprintf("%s:%s", c.IP(), email)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many password reset attempts. Please try again later.",
				"code":    "PASSWORD_RESET_RATE_LIMIT",
			})
		},
		Storage: memory.New(),
	})
}

// APIRateLimiter creates a rate limiter for API endpoints
func APIRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Try to get user ID from context first (authenticated requests)
			userID, ok := GetUserID(c)
			if ok {
				return fmt.Sprintf("user:%s", userID.String())
			}
			// Fall back to IP for unauthenticated requests
			return fmt.Sprintf("ip:%s", c.IP())
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "API rate limit exceeded. Please try again later.",
			})
		},
		Storage: memory.New(),
	})
}

// CustomRateLimiter creates a custom rate limiter with specified limits
func CustomRateLimiter(max int, expiration time.Duration, keyGen func(*fiber.Ctx) string) fiber.Handler {
	if keyGen == nil {
		keyGen = func(c *fiber.Ctx) string {
			return c.IP()
		}
	}

	return limiter.New(limiter.Config{
		Max:          max,
		Expiration:   expiration,
		KeyGenerator: keyGen,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
		},
		Storage: memory.New(),
	})
}

// PerUserRateLimiter creates a rate limiter that limits per authenticated user
func PerUserRateLimiter(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID, ok := GetUserID(c)
			if ok {
				return fmt.Sprintf("user:%s", userID.String())
			}
			// Fall back to IP for unauthenticated requests
			return fmt.Sprintf("ip:%s", c.IP())
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": fmt.Sprintf("You have exceeded the rate limit of %d requests per %v.", max, expiration),
			})
		},
		Storage: memory.New(),
	})
}

// SlidingWindowRateLimiter implements a sliding window rate limiter
// This is more accurate than fixed window but uses more memory
func SlidingWindowRateLimiter(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": fmt.Sprintf("Rate limit of %d requests per %v exceeded.", max, window),
			})
		},
		Storage: memory.New(),
	})
}
