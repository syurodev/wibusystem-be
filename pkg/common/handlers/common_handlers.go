package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"wibusystem/pkg/common/response"
)

// NoRouteHandler returns a standardized 404 JSON response for non-existent routes
func NoRouteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, response.StandardResponse{
			Success: false,
			Message: "Route not found",
			Data:    nil,
			Error: &response.ErrorDetail{
				Code:        "route_not_found",
				Description: fmt.Sprintf("The requested route '%s %s' does not exist", c.Request.Method, c.Request.URL.Path),
			},
			Meta: map[string]interface{}{},
		})
	}
}