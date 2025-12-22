/*
Media Progress Routes - Route Registration
==========================================

Đăng ký routes cho media progress module.
Tất cả routes đều require authentication.

ROUTES:
───────
History endpoints (frontend compatible):
  GET    /api/v1/history           → Lấy danh sách progress
  GET    /api/v1/history/recent    → Lấy N mục gần nhất
  POST   /api/v1/history           → Cập nhật progress
  DELETE /api/v1/history/:id       → Xóa progress
  POST   /api/v1/history/clear     → Xóa toàn bộ

Progress endpoints (for chapter list view):
  GET    /api/v1/progress/:type/:id/units           → Lấy chapter status
  POST   /api/v1/progress/:type/:id/units/:uid/complete → Đánh dấu đã đọc
*/

package media_progress

import (
	"github.com/gin-gonic/gin"
)

// RegisterHistoryRoutes đăng ký routes cho /api/v1/history
// Được gọi từ router.go với authenticated route group
func (h *Handler) RegisterHistoryRoutes(rg *gin.RouterGroup) {
	// GET /api/v1/history - Lấy danh sách progress (paginated)
	rg.GET("", h.GetProgressList)

	// GET /api/v1/history/recent - Lấy N mục gần nhất cho "Continue" section
	rg.GET("/recent", h.GetRecentProgress)

	// POST /api/v1/history - Tạo/cập nhật progress
	rg.POST("", h.UpdateProgress)

	// DELETE /api/v1/history/:id - Xóa progress
	rg.DELETE("/:id", h.DeleteProgress)

	// POST /api/v1/history/clear - Xóa toàn bộ progress
	rg.POST("/clear", h.ClearAllProgress)
}

// RegisterProgressRoutes đăng ký routes cho /api/v1/progress
// Được gọi từ router.go với authenticated route group
func (h *Handler) RegisterProgressRoutes(rg *gin.RouterGroup) {
	// GET /api/v1/progress/:media_type/:media_id/units - Lấy chapter read status
	rg.GET("/:media_type/:media_id/units", h.GetUnitProgress)

	// POST /api/v1/progress/:media_type/:media_id/units/:unit_id/complete - Đánh dấu đã đọc
	rg.POST("/:media_type/:media_id/units/:unit_id/complete", h.MarkUnitComplete)
}
