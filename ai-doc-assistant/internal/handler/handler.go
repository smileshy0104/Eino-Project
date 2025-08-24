package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ai-doc-assistant/internal/model"
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
	var req model.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
			"details": err.Error(),
		})
		return
	}

	// 设置默认值
	if req.TopK == 0 {
		req.TopK = 5
	}
	if req.UserID == "" {
		req.UserID = "anonymous" // 匿名用户
	}

	ctx := c.Request.Context()
	
	// 调用Eino服务进行问答并保存历史
	response, err := h.einoService.QueryKnowledgeWithHistory(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) QueryHistory(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		userID = "anonymous"
	}
	
	limit := 20 // 默认返回最近20条
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	histories, err := h.einoService.GetQueryHistory(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取查询历史失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": histories,
		"count": len(histories),
	})
}

func (h *Handlers) SubmitFeedback(c *gin.Context) {
	var req model.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数",
			"details": err.Error(),
		})
		return
	}

	err := h.einoService.UpdateQueryFeedback(req.QueryID, req.SatisfactionScore, req.Feedback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "提交反馈失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "反馈提交成功",
	})
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