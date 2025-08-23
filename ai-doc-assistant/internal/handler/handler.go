package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ai-doc-assistant/internal/service"
)

// Handlers 处理器集合
type Handlers struct {
	einoService *service.EinoService
}

// NewHandlers 创建处理器集合
func NewHandlers(einoService *service.EinoService) *Handlers {
	return &Handlers{
		einoService: einoService,
	}
}

// Home 首页
func (h *Handlers) Home(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "AI文档助手",
		"version": "1.0.0",
	})
}

// HealthCheck 健康检查
func (h *Handlers) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.einoService.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// 其他处理器方法的占位符实现
func (h *Handlers) UploadDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) ListDocuments(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) GetDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) UpdateDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) DeleteDocument(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) BatchUpload(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) AskQuestion(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) QueryHistory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) SubmitFeedback(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) CreateUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) ListUsers(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) GetUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) UpdateUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) GetOverview(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) GetUsageStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) GetPerformanceStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func (h *Handlers) Metrics(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}