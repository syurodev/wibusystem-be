package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// AppHandler is a custom handler type that returns data and an error.
// This allows the wrapper to handle response formatting.
type AppHandler func(c *fiber.Ctx) (any, *Pagination, error)

// WithStandardResponse is a middleware that wraps an AppHandler to provide a standard JSON response.
func WithStandardResponse(handler AppHandler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, pagination, err := handler(c)

		if err != nil {
			// Basic error handling. This can be expanded to handle different error types.
			// For example, check for a custom error type that includes a status code.
			var e *fiber.Error
			statusCode := fiber.StatusInternalServerError
			if errors.As(err, &e) {
				statusCode = e.Code
			}

			return c.Status(statusCode).JSON(NewErrorResponse("error", err.Error(), "An error occurred"))
		}

		// Determine success status code based on HTTP method
		statusCode := fiber.StatusOK
		if c.Method() == fiber.MethodPost {
			statusCode = fiber.StatusCreated
		}

		return c.Status(statusCode).JSON(NewSuccessResponse(data, "Success", pagination))
	}
}
