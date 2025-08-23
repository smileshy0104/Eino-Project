package model

import (
	"time"
)

// Document 文档模型
type Document struct {
	ID           string    `json:"id" gorm:"primarykey;type:varchar(64)"`
	Title        string    `json:"title" gorm:"type:varchar(255);not null;index"`
	Content      string    `json:"content" gorm:"type:longtext"`
	DocumentType string    `json:"document_type" gorm:"type:varchar(50);index"`
	Version      string    `json:"version" gorm:"type:varchar(50)"`
	Author       string    `json:"author" gorm:"type:varchar(100);index"`
	Department   string    `json:"department" gorm:"type:varchar(100);index"`
	FilePath     string    `json:"file_path" gorm:"type:varchar(500)"`
	FileSize     int64     `json:"file_size"`
	FileHash     string    `json:"file_hash" gorm:"type:varchar(128);index"`
	Tags         string    `json:"tags" gorm:"type:text"` // JSON数组字符串
	Metadata     string    `json:"metadata" gorm:"type:json"` // 额外元数据
	Status       string    `json:"status" gorm:"type:enum('active','archived','deleted');default:'active';index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
	
	// 关联关系
	Chunks []DocumentChunk `json:"chunks,omitempty" gorm:"foreignkey:DocumentID"`
}

// DocumentChunk 文档块模型
type DocumentChunk struct {
	ID         string    `json:"id" gorm:"primarykey;type:varchar(64)"`
	DocumentID string    `json:"document_id" gorm:"type:varchar(64);not null;index"`
	ChunkIndex int       `json:"chunk_index" gorm:"not null"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	VectorID   string    `json:"vector_id" gorm:"type:varchar(64);index"` // Milvus中的向量ID
	Metadata   string    `json:"metadata" gorm:"type:json"` // 块级元数据
	CreatedAt  time.Time `json:"created_at"`
	
	// 外键关系
	Document Document `json:"document,omitempty" gorm:"foreignkey:DocumentID"`
}

// User 用户模型
type User struct {
	ID         string    `json:"id" gorm:"primarykey;type:varchar(64)"`
	Username   string    `json:"username" gorm:"type:varchar(100);uniqueIndex;not null"`
	Email      string    `json:"email" gorm:"type:varchar(200);index"`
	Department string    `json:"department" gorm:"type:varchar(100);index"`
	Role       string    `json:"role" gorm:"type:enum('admin','user','readonly');default:'user'"`
	Avatar     string    `json:"avatar" gorm:"type:varchar(255)"`
	Status     string    `json:"status" gorm:"type:enum('active','inactive');default:'active';index"`
	LastLogin  *time.Time `json:"last_login,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// QueryHistory 查询历史模型
type QueryHistory struct {
	ID               string    `json:"id" gorm:"primarykey;type:varchar(64)"`
	UserID           string    `json:"user_id" gorm:"type:varchar(64);index"`
	Query            string    `json:"query" gorm:"type:text;not null"`
	Response         string    `json:"response" gorm:"type:longtext"`
	ResponseTimeMs   int       `json:"response_time_ms"` // 响应时间(毫秒)
	RetrievedDocs    string    `json:"retrieved_docs" gorm:"type:json"` // 检索到的文档ID列表
	SatisfactionScore *int     `json:"satisfaction_score,omitempty"` // 1-5分满意度
	Feedback         string    `json:"feedback" gorm:"type:text"` // 用户反馈
	CreatedAt        time.Time `json:"created_at"`
	
	// 外键关系
	User User `json:"user,omitempty" gorm:"foreignkey:UserID"`
}

// DocumentShare 文档分享模型
type DocumentShare struct {
	ID         string    `json:"id" gorm:"primarykey;type:varchar(64)"`
	UserID     string    `json:"user_id" gorm:"type:varchar(64);not null;index"`
	DocumentID string    `json:"document_id" gorm:"type:varchar(64);not null;index"`
	Permission string    `json:"permission" gorm:"type:enum('read','analyze','index','full');default:'read'"`
	ShareTime  time.Time `json:"share_time"`
	ExpireTime *time.Time `json:"expire_time,omitempty"`
	Status     string    `json:"status" gorm:"type:enum('active','expired','revoked');default:'active';index"`
	
	// 外键关系
	User     User     `json:"user,omitempty" gorm:"foreignkey:UserID"`
	Document Document `json:"document,omitempty" gorm:"foreignkey:DocumentID"`
}

// VectorCollection Milvus向量集合信息
type VectorCollection struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Dimension   int    `json:"dimension"`
	IndexType   string `json:"index_type"`
	MetricType  string `json:"metric_type"`
	Status      string `json:"status"`
}

// DocumentRequest 文档上传请求
type DocumentRequest struct {
	Title        string   `json:"title" binding:"required"`
	Content      string   `json:"content"`
	DocumentType string   `json:"document_type"`
	Version      string   `json:"version"`
	Author       string   `json:"author"`
	Department   string   `json:"department"`
	Tags         []string `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// QueryRequest 问答请求
type QueryRequest struct {
	Question string `json:"question" binding:"required"`
	UserID   string `json:"user_id"`
	Context  string `json:"context"` // 上下文信息
	TopK     int    `json:"top_k"`   // 检索文档数量，默认5
}

// QueryResponse 问答响应
type QueryResponse struct {
	Answer       string                 `json:"answer"`
	Sources      []DocumentSource       `json:"sources"`
	ResponseTime int                    `json:"response_time_ms"`
	QueryID      string                 `json:"query_id"`
	Confidence   float64               `json:"confidence"`
}

// DocumentSource 文档来源
type DocumentSource struct {
	DocumentID    string  `json:"document_id"`
	DocumentTitle string  `json:"document_title"`
	ChunkContent  string  `json:"chunk_content"`
	Relevance     float64 `json:"relevance"`
	Author        string  `json:"author"`
	Version       string  `json:"version"`
}

// FeedbackRequest 反馈请求
type FeedbackRequest struct {
	QueryID           string `json:"query_id" binding:"required"`
	SatisfactionScore int    `json:"satisfaction_score" binding:"required,min=1,max=5"`
	Feedback          string `json:"feedback"`
}

// StatsOverview 统计概览
type StatsOverview struct {
	TotalDocuments   int64   `json:"total_documents"`
	TotalUsers       int64   `json:"total_users"`
	TotalQueries     int64   `json:"total_queries"`
	AvgResponseTime  float64 `json:"avg_response_time_ms"`
	AvgSatisfaction  float64 `json:"avg_satisfaction"`
	ActiveUsers      int64   `json:"active_users_today"`
	QueriesLastHour  int64   `json:"queries_last_hour"`
}

// UsageStats 使用统计
type UsageStats struct {
	Date         string `json:"date"`
	QueryCount   int64  `json:"query_count"`
	UniqueUsers  int64  `json:"unique_users"`
	AvgResponse  float64 `json:"avg_response_time"`
	Satisfaction float64 `json:"satisfaction"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

func (DocumentChunk) TableName() string {
	return "document_chunks"
}

func (User) TableName() string {
	return "users"
}

func (QueryHistory) TableName() string {
	return "query_history"
}

func (DocumentShare) TableName() string {
	return "document_shares"
}