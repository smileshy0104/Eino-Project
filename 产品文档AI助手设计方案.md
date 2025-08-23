# 产品文档AI助手设计方案

基于 Eino 框架的智能文档检索系统

---

## 📋 项目概述

### 🎯 项目背景
在敏捷开发模式下，产品需求文档更新频繁，历史版本众多，团队成员在查找特定需求信息时面临以下痛点：
- 📁 文档分散在不同版本和文件夹中
- 🔍 关键词搜索效果不佳，无法理解语义
- ⏰ 查找历史需求耗时过长，影响开发效率
- 🔄 文档更新后团队同步困难

### 💡 解决方案
构建一个基于 Eino 框架的智能文档助手，具备：
- 🤖 **智能问答**：自然语言查询历史需求
- 📚 **语义检索**：理解需求意图，精准匹配文档
- 🔄 **实时同步**：自动更新向量库，保持数据最新
- 📢 **团队通知**：通过飞书机器人推送文档变更

---

## 🎨 产品功能设计

### 核心功能模块

#### 1️⃣ 智能文档问答
```
用户输入：
"之前用户登录功能的验证码有效期是多少？"

AI助手回答：
📄 根据 v2.3 产品需求文档 (2024-03-15)：
用户登录验证码有效期设定为 5分钟
相关截图：[验证码流程图]
负责人：@张三 
关联需求：PRD-001-用户认证模块

🔗 相关文档链接：
- 用户认证详细设计 v2.3.pdf
- 登录流程优化方案 v1.2.docx
```

#### 2️⃣ 文档版本对比
```
用户输入：
"支付模块在最近几个版本有什么变化？"

AI助手回答：
📈 支付模块版本变更历史：

v3.1 (2024-04-10):
✅ 新增微信支付
✅ 优化支付超时处理

v3.0 (2024-03-20): 
✅ 支持支付宝
❌ 移除银行卡直连

v2.8 (2024-02-15):
✅ 添加支付状态回调机制
```

#### 3️⃣ 需求关联分析
```
用户输入：
"这个功能会影响到哪些其他模块？"

AI助手分析：
🔗 关联影响分析：

直接影响：
- 用户中心模块（用户状态同步）
- 通知模块（支付成功通知）

间接影响：
- 订单模块（状态更新逻辑）
- 报表模块（财务统计）

建议评估：
⚠️  需要同步更新用户中心API
⚠️  通知模板需要添加新字段
```

#### 4️⃣ 实时文档监控
- 📁 监控指定目录的文档变更
- 🔄 自动解析新增/修改的文档
- 📊 增量更新向量库
- 📢 通过飞书推送变更通知

---

## 🏗️ 技术架构设计

### 系统架构图
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  飞书企业文档源  │    │   Eino处理引擎   │    │   用户交互层     │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ • 飞书文档API   │───▶│ • Transformer   │───▶│ • Web控制台     │
│ • 多维表格API   │    │ • Embedder      │    │ • 飞书机器人     │
│ • 知识库API     │    │ • Indexer       │    │ • 飞书小程序     │
│ • Webhook监听   │    │ • Retriever     │    │ • API接口       │
│ • 权限同步      │    │ • ChatModel     │    │ • 移动端应用     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   存储层        │
                    ├─────────────────┤
                    │ • Milvus向量库  │
                    │ • MySQL元数据   │
                    │ • Redis缓存     │
                    │ • 飞书Token缓存 │
                    └─────────────────┘
```

### 核心技术组件

#### 🚀 飞书文档集成层
```go
package feishu

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/cloudwego/eino/chain"
    "github.com/cloudwego/eino/compose"
)

// 飞书API客户端
type FeishuDocClient struct {
    AppID       string
    AppSecret   string
    AccessToken string
    BaseURL     string
    HTTPClient  *http.Client
}

// 文档信息结构体
type DocumentInfo struct {
    Token      string    `json:"token"`
    Type       string    `json:"type"`
    Title      string    `json:"title"`
    URL        string    `json:"url"`
    Owner      string    `json:"owner"`
    CreateTime time.Time `json:"create_time"`
    UpdateTime time.Time `json:"update_time"`
}

// 文档内容结构体
type DocumentContent struct {
    Content  string                 `json:"content"`
    Metadata map[string]interface{} `json:"metadata"`
}

// 创建飞书API客户端
func NewFeishuDocClient(appID, appSecret string) *FeishuDocClient {
    client := &FeishuDocClient{
        AppID:     appID,
        AppSecret: appSecret,
        BaseURL:   "https://open.feishu.cn",
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
    
    // 获取访问令牌
    if err := client.refreshAccessToken(); err != nil {
        log.Printf("获取访问令牌失败: %v", err)
    }
    
    return client
}

// 获取所有可访问的文档
func (c *FeishuDocClient) GetAllDocuments(ctx context.Context, folderToken string) ([]*DocumentInfo, error) {
    docTypes := []string{"doc", "sheet", "bitable", "wiki"}
    var allDocs []*DocumentInfo
    
    for _, docType := range docTypes {
        docs, err := c.fetchDocumentsByType(ctx, docType, folderToken)
        if err != nil {
            log.Printf("获取 %s 类型文档失败: %v", docType, err)
            continue
        }
        allDocs = append(allDocs, docs...)
    }
    
    return allDocs, nil
}

// 获取文档具体内容
func (c *FeishuDocClient) GetDocumentContent(ctx context.Context, docToken, docType string) (*DocumentContent, error) {
    switch docType {
    case "doc":
        return c.getDocContent(ctx, docToken)
    case "sheet":
        return c.getSheetContent(ctx, docToken)
    case "bitable":
        return c.getBitableContent(ctx, docToken)
    case "wiki":
        return c.getWikiContent(ctx, docToken)
    default:
        return nil, fmt.Errorf("不支持的文档类型: %s", docType)
    }
}

// 设置Webhook监听
func (c *FeishuDocClient) SetupWebhook(ctx context.Context, callbackURL string) error {
    events := []string{
        "drive.file.created_in_folder_v1",
        "drive.file.edit_v1", 
        "drive.file.title_updated_v1",
        "drive.file.trashed_v1",
    }
    
    for _, event := range events {
        if err := c.subscribeEvent(ctx, event, callbackURL); err != nil {
            log.Printf("订阅事件 %s 失败: %v", event, err)
        }
    }
    
    return nil
}

// 飞书文档同步器
type FeishuDocumentSyncer struct {
    client   *FeishuDocClient
    pipeline compose.Chain
}

// 创建文档同步器
func NewFeishuDocumentSyncer(client *FeishuDocClient, pipeline compose.Chain) *FeishuDocumentSyncer {
    return &FeishuDocumentSyncer{
        client:   client,
        pipeline: pipeline,
    }
}

// 全量同步所有文档
func (s *FeishuDocumentSyncer) SyncAllDocuments(ctx context.Context) error {
    documents, err := s.client.GetAllDocuments(ctx, "")
    if err != nil {
        return fmt.Errorf("获取文档列表失败: %w", err)
    }
    
    for _, doc := range documents {
        if err := s.syncSingleDocument(ctx, doc); err != nil {
            log.Printf("同步文档失败 %s: %v", doc.Title, err)
            continue
        }
        log.Printf("同步文档成功: %s", doc.Title)
    }
    
    return nil
}

// 同步单个文档
func (s *FeishuDocumentSyncer) syncSingleDocument(ctx context.Context, doc *DocumentInfo) error {
    content, err := s.client.GetDocumentContent(ctx, doc.Token, doc.Type)
    if err != nil {
        return fmt.Errorf("获取文档内容失败: %w", err)
    }
    
    // 构建处理输入
    input := map[string]interface{}{
        "content": content.Content,
        "metadata": map[string]interface{}{
            "doc_token":   doc.Token,
            "doc_type":    doc.Type,
            "title":       doc.Title,
            "url":         doc.URL,
            "owner":       doc.Owner,
            "create_time": doc.CreateTime,
            "update_time": doc.UpdateTime,
            "source":      "feishu",
        },
    }
    
    // 使用Eino处理文档
    _, err = s.pipeline.Invoke(ctx, input)
    return err
}
```

#### 📄 文档处理流水线
```go
package pipeline

import (
    "context"
    
    "github.com/cloudwego/eino/chain"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/components/document"
    "github.com/cloudwego/eino/components/embedding" 
    "github.com/cloudwego/eino/components/indexing"
)

// 创建专门适配飞书文档的处理流水线
func NewFeishuDocumentPipeline() compose.Chain {
    // 文档加载器
    docLoader := document.NewDocumentLoader(document.Config{
        ChunkSize:    512,
        ChunkOverlap: 50,
        PreserveStructure: true, // 保持飞书文档结构
    })
    
    // 内容分割器
    splitter := document.NewTextSplitter(document.SplitterConfig{
        ChunkSize:    512,
        ChunkOverlap: 50,
        Separators:   []string{"\n\n", "\n", "。", ".", " "},
    })
    
    // 向量嵌入器
    embedder := embedding.NewEmbedding(embedding.Config{
        Provider: "openai",
        Model:    "text-embedding-ada-002",
        APIKey:   os.Getenv("OPENAI_API_KEY"),
    })
    
    // 向量索引器
    indexer := indexing.NewMilvusIndexer(indexing.MilvusConfig{
        Host:           "localhost:19530",
        CollectionName: "feishu_docs",
        Dimension:      1536,
        IndexType:      "IVF_FLAT",
        MetricType:     "L2",
    })
    
    // 构建处理链
    pipeline := compose.NewChain().
        AppendRunnable(docLoader).
        AppendRunnable(splitter).
        AppendRunnable(embedder).
        AppendRunnable(indexer)
    
    return pipeline
}

// 飞书文档内容解析器
type FeishuContentParser struct{}

func NewFeishuContentParser() *FeishuContentParser {
    return &FeishuContentParser{}
}

func (p *FeishuContentParser) Invoke(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    content, ok := input["content"].(string)
    if !ok {
        return input, fmt.Errorf("invalid content type")
    }
    
    docType, _ := input["metadata"].(map[string]interface{})["doc_type"].(string)
    
    var parsedContent string
    var err error
    
    switch docType {
    case "doc":
        parsedContent, err = p.parseFeishuDoc(content)
    case "sheet":
        parsedContent, err = p.parseFeishuSheet(content)
    case "bitable":
        parsedContent, err = p.parseFeishuBitable(content)
    case "wiki":
        parsedContent, err = p.parseFeishuWiki(content)
    default:
        parsedContent = content
    }
    
    if err != nil {
        return input, err
    }
    
    input["parsed_content"] = parsedContent
    return input, nil
}

// 解析飞书文档
func (p *FeishuContentParser) parseFeishuDoc(content string) (string, error) {
    // 解析飞书文档的特殊格式
    // 处理标题、段落、表格、代码块等
    return content, nil
}

// 支持的飞书文档类型
/*
- 飞书文档 (Doc) - 富文本文档，支持标题、段落、表格、图片等
- 飞书表格 (Sheet) - 数据表格，自动提取单元格内容 
- 多维表格 (Bitable) - 结构化数据，支持字段类型识别
- 飞书知识库 (Wiki) - 知识库页面，保持页面层级结构
- 思维导图 (MindMap) - 流程图表，提取节点和关系信息
*/
```

#### 🔍 智能检索引擎
```go
package retrieval

import (
    "context"
    
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/components/retriever"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/prompt"
)

// 创建RAG检索链
func NewDocumentQAChain() compose.Chain {
    // 查询分析器
    queryAnalyzer := NewQueryAnalyzer()
    
    // 语义检索器
    semanticRetriever := retriever.NewVectorRetriever(retriever.Config{
        Collection:      "feishu_docs",
        TopK:           5,
        ScoreThreshold: 0.7,
        SearchParams: map[string]interface{}{
            "nprobe": 10,
        },
    })
    
    // 相关性过滤器
    relevanceFilter := NewRelevanceFilter()
    
    // 上下文聚合器
    contextAggregator := NewContextAggregator()
    
    // 对话模型
    chatModel := model.NewChatModel(model.Config{
        Provider:    "openai",
        Model:       "gpt-4",
        Temperature: 0.1,
        MaxTokens:   2000,
    })
    
    // 构建检索链
    chain := compose.NewChain().
        AppendRunnable(queryAnalyzer).
        AppendRunnable(semanticRetriever).
        AppendRunnable(relevanceFilter).
        AppendRunnable(contextAggregator).
        AppendRunnable(chatModel)
    
    return chain
}

// 查询分析器
type QueryAnalyzer struct{}

func NewQueryAnalyzer() *QueryAnalyzer {
    return &QueryAnalyzer{}
}

func (q *QueryAnalyzer) Invoke(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
    query := input["query"].(string)
    
    // 分析查询意图
    intent := q.analyzeIntent(query)
    
    // 提取关键词
    keywords := q.extractKeywords(query)
    
    // 检测时间范围
    timeRange := q.detectTimeRange(query)
    
    // 识别文档类型偏好
    docTypePreference := q.detectDocTypePreference(query)
    
    input["analyzed_query"] = map[string]interface{}{
        "original_query":      query,
        "intent":              intent,
        "keywords":            keywords,
        "time_range":          timeRange,
        "doc_type_preference": docTypePreference,
    }
    
    return input, nil
}

// 文档问答系统
type DocumentQASystem struct {
    qaChain compose.Chain
}

func NewDocumentQASystem() *DocumentQASystem {
    return &DocumentQASystem{
        qaChain: NewDocumentQAChain(),
    }
}

func (qa *DocumentQASystem) Ask(ctx context.Context, question string) (string, error) {
    input := map[string]interface{}{
        "query": question,
    }
    
    result, err := qa.qaChain.Invoke(ctx, input)
    if err != nil {
        return "", err
    }
    
    answer, ok := result["answer"].(string)
    if !ok {
        return "", fmt.Errorf("invalid answer format")
    }
    
    return answer, nil
}

// 检索策略配置
/*
检索策略：
- 语义相似度检索（主要）: 使用向量相似度匹配
- 关键词匹配检索（辅助）: 补充传统文本匹配  
- 时间范围过滤: 支持"最近一个月"等时间查询
- 文档类型过滤: 可指定查询特定类型文档
- 作者/负责人过滤: 按文档创建者或负责人筛选
- 权限检查: 确保用户只能访问有权限的文档
*/
```

#### 🔄 飞书实时更新机制
```go
package webhook

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
)

// 飞书事件数据结构
type FeishuEvent struct {
    Schema string `json:"schema"`
    Header struct {
        EventID    string `json:"event_id"`
        EventType  string `json:"event_type"`
        CreateTime string `json:"create_time"`
        Token      string `json:"token"`
        AppID      string `json:"app_id"`
        TenantKey  string `json:"tenant_key"`
    } `json:"header"`
    Event map[string]interface{} `json:"event"`
}

// 飞书Webhook监听器
type FeishuWebhookListener struct {
    syncer   *FeishuDocumentSyncer
    notifier *FeishuNotifier
    server   *gin.Engine
}

func NewFeishuWebhookListener(syncer *FeishuDocumentSyncer, notifier *FeishuNotifier) *FeishuWebhookListener {
    listener := &FeishuWebhookListener{
        syncer:   syncer,
        notifier: notifier,
        server:   gin.Default(),
    }
    
    // 设置路由
    listener.setupRoutes()
    
    return listener
}

func (l *FeishuWebhookListener) setupRoutes() {
    l.server.POST("/webhook/feishu", l.handleWebhook)
}

func (l *FeishuWebhookListener) handleWebhook(c *gin.Context) {
    var event FeishuEvent
    if err := c.ShouldBindJSON(&event); err != nil {
        log.Printf("解析webhook数据失败: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 验证事件合法性
    if !l.validateEvent(&event) {
        log.Printf("事件验证失败: %s", event.Header.EventID)
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid event"})
        return
    }
    
    // 处理文档事件
    go l.handleDocumentEvent(context.Background(), &event)
    
    c.JSON(http.StatusOK, gin.H{"challenge": event.Event["challenge"]})
}

func (l *FeishuWebhookListener) handleDocumentEvent(ctx context.Context, event *FeishuEvent) {
    switch event.Header.EventType {
    case "drive.file.edit_v1", "drive.file.title_updated_v1":
        l.handleDocumentUpdate(ctx, event)
    case "drive.file.created_in_folder_v1":
        l.handleDocumentCreate(ctx, event)
    case "drive.file.trashed_v1":
        l.handleDocumentDelete(ctx, event)
    default:
        log.Printf("未处理的事件类型: %s", event.Header.EventType)
    }
}

func (l *FeishuWebhookListener) handleDocumentUpdate(ctx context.Context, event *FeishuEvent) {
    docToken, ok := event.Event["file_token"].(string)
    if !ok {
        log.Printf("无效的文档token")
        return
    }
    
    docType, _ := event.Event["file_type"].(string)
    
    // 重新同步文档
    doc := &DocumentInfo{
        Token: docToken,
        Type:  docType,
    }
    
    if err := l.syncer.syncSingleDocument(ctx, doc); err != nil {
        log.Printf("同步文档失败 %s: %v", docToken, err)
        return
    }
    
    // 发送更新通知
    l.sendUpdateNotification(event)
}

func (l *FeishuWebhookListener) handleDocumentCreate(ctx context.Context, event *FeishuEvent) {
    // 处理新文档创建
    l.handleDocumentUpdate(ctx, event)
}

func (l *FeishuWebhookListener) handleDocumentDelete(ctx context.Context, event *FeishuEvent) {
    docToken, ok := event.Event["file_token"].(string)
    if !ok {
        return
    }
    
    // 从向量库中删除相关文档
    // 这里需要实现删除逻辑
    log.Printf("文档已删除: %s", docToken)
}

func (l *FeishuWebhookListener) sendUpdateNotification(event *FeishuEvent) {
    docInfo := map[string]interface{}{
        "title":       event.Event["file_name"],
        "update_time": time.Now().Format("2006-01-02 15:04:05"),
        "operator":    event.Event["operator_id"],
        "doc_token":   event.Event["file_token"],
        "doc_url":     fmt.Sprintf("https://bytedance.feishu.cn/docs/%s", event.Event["file_token"]),
    }
    
    l.notifier.SendDocumentUpdateNotification(docInfo)
}

func (l *FeishuWebhookListener) Start(addr string) error {
    log.Printf("启动Webhook监听器，地址: %s", addr)
    return l.server.Run(addr)
}

// 权限同步器
type FeishuPermissionSyncer struct {
    client *FeishuDocClient
}

func NewFeishuPermissionSyncer(client *FeishuDocClient) *FeishuPermissionSyncer {
    return &FeishuPermissionSyncer{
        client: client,
    }
}

type DocumentPermission struct {
    DocToken  string    `json:"doc_token"`
    Viewers   []string  `json:"viewers"`
    Editors   []string  `json:"editors"`
    Owners    []string  `json:"owners"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (s *FeishuPermissionSyncer) SyncDocumentPermissions(ctx context.Context, docToken string) (*DocumentPermission, error) {
    permissions, err := s.client.getDocumentPermissions(ctx, docToken)
    if err != nil {
        return nil, fmt.Errorf("获取文档权限失败: %w", err)
    }
    
    permissionCache := &DocumentPermission{
        DocToken:  docToken,
        Viewers:   permissions["viewers"].([]string),
        Editors:   permissions["editors"].([]string),
        Owners:    permissions["owners"].([]string),
        UpdatedAt: time.Now(),
    }
    
    // 保存到本地缓存
    if err := s.savePermissionCache(permissionCache); err != nil {
        log.Printf("保存权限缓存失败: %v", err)
    }
    
    return permissionCache, nil
}

func (s *FeishuPermissionSyncer) savePermissionCache(permission *DocumentPermission) error {
    // 实现权限缓存保存逻辑
    // 可以保存到 Redis 或数据库中
    return nil
}
```

---

## 🛠️ 实施方案

### Phase 1: 核心功能开发（4周）

**Week 1-2: 基础架构搭建**
- [x] Eino环境配置
- [x] Milvus向量数据库部署
- [x] 文档处理流水线开发
- [x] 基础Web界面

**Week 3-4: 智能检索功能**
- [x] RAG检索链实现
- [x] 多模态文档解析
- [x] 基础问答功能
- [x] 检索结果优化

### Phase 2: 高级功能开发（3周）

**Week 5-6: 实时更新机制**
- [x] 文件监控系统
- [x] 增量更新逻辑
- [x] 版本管理功能

**Week 7: 飞书集成**
- [x] 飞书机器人开发
- [x] 消息推送功能
- [x] 用户权限管理

### Phase 3: 优化与部署（2周）

**Week 8: 系统优化**
- [x] 性能调优
- [x] 缓存策略优化
- [x] 错误处理完善

**Week 9: 上线部署**
- [x] 生产环境部署
- [x] 用户培训
- [x] 运维监控

---

## 💻 技术实现详解

### 完整应用示例

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/gin-gonic/gin"
    "your-project/pkg/feishu"
    "your-project/pkg/pipeline"
    "your-project/pkg/retrieval"
    "your-project/pkg/webhook"
)

func main() {
    // 读取配置
    config := loadConfig()
    
    // 初始化飞书客户端
    feishuClient := feishu.NewFeishuDocClient(
        config.Feishu.AppID,
        config.Feishu.AppSecret,
    )
    
    // 创建文档处理流水线
    docPipeline := pipeline.NewFeishuDocumentPipeline()
    
    // 创建文档同步器
    syncer := feishu.NewFeishuDocumentSyncer(feishuClient, docPipeline)
    
    // 创建问答系统
    qaSystem := retrieval.NewDocumentQASystem()
    
    // 创建通知器
    notifier := feishu.NewFeishuNotifier(config.Feishu.Bot.WebhookURL)
    
    // 创建Webhook监听器
    webhookListener := webhook.NewFeishuWebhookListener(syncer, notifier)
    
    // 设置API路由
    router := gin.Default()
    setupAPIRoutes(router, qaSystem)
    
    // 启动服务
    go func() {
        log.Println("启动API服务器 :8080")
        if err := router.Run(":8080"); err != nil {
            log.Fatalf("API服务器启动失败: %v", err)
        }
    }()
    
    go func() {
        log.Println("启动Webhook监听器 :8081")
        if err := webhookListener.Start(":8081"); err != nil {
            log.Fatalf("Webhook监听器启动失败: %v", err)
        }
    }()
    
    // 首次全量同步
    log.Println("开始全量同步文档...")
    if err := syncer.SyncAllDocuments(context.Background()); err != nil {
        log.Printf("全量同步失败: %v", err)
    } else {
        log.Println("全量同步完成")
    }
    
    // 设置Webhook
    if err := feishuClient.SetupWebhook(context.Background(), "http://your-domain.com:8081/webhook/feishu"); err != nil {
        log.Printf("设置Webhook失败: %v", err)
    }
    
    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("服务器正在关闭...")
}

func setupAPIRoutes(router *gin.Engine, qaSystem *retrieval.DocumentQASystem) {
    api := router.Group("/api/v1")
    
    // 文档问答接口
    api.POST("/ask", func(c *gin.Context) {
        var request struct {
            Question string `json:"question" binding:"required"`
        }
        
        if err := c.ShouldBindJSON(&request); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        answer, err := qaSystem.Ask(c.Request.Context(), request.Question)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(200, gin.H{
            "answer": answer,
            "timestamp": time.Now(),
        })
    })
    
    // 健康检查接口
    api.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
}

type Config struct {
    Feishu struct {
        AppID     string `yaml:"app_id"`
        AppSecret string `yaml:"app_secret"`
        Bot       struct {
            WebhookURL string `yaml:"webhook_url"`
        } `yaml:"bot"`
    } `yaml:"feishu"`
}

func loadConfig() *Config {
    // 实现配置文件加载逻辑
    return &Config{}
}
```

### 项目结构

```
feishu-doc-ai/
├── cmd/
│   └── server/
│       └── main.go                 # 主程序入口
├── pkg/
│   ├── feishu/                     # 飞书API集成
│   │   ├── client.go              # 飞书API客户端  
│   │   ├── syncer.go              # 文档同步器
│   │   └── notifier.go            # 飞书通知器
│   ├── pipeline/                   # 文档处理流水线
│   │   ├── processor.go           # 文档处理器
│   │   └── parser.go              # 内容解析器
│   ├── retrieval/                  # 检索和问答
│   │   ├── qa.go                  # 问答系统
│   │   └── analyzer.go            # 查询分析器
│   ├── webhook/                    # Webhook处理
│   │   ├── listener.go            # 事件监听器
│   │   └── handler.go             # 事件处理器
│   └── config/                     # 配置管理
│       └── config.go              # 配置结构体
├── configs/
│   └── config.yaml                 # 配置文件
├── docker/
│   ├── Dockerfile                  # Docker构建文件
│   └── docker-compose.yml         # 容器编排
├── docs/                           # 项目文档
├── go.mod                          # Go模块文件
└── go.sum                          # 依赖版本锁定
```

### Go依赖管理

```go
// go.mod
module feishu-doc-ai

go 1.21

require (
    github.com/cloudwego/eino v0.0.10
    github.com/gin-gonic/gin v1.9.1
    github.com/go-redis/redis/v8 v8.11.5
    github.com/milvus-io/milvus-sdk-go/v2 v2.3.1
    github.com/spf13/viper v1.16.0
    github.com/stretchr/testify v1.8.4
    gopkg.in/yaml.v3 v3.0.1
)
```

### 核心配置文件

```yaml
# configs/config.yaml
app:
  name: "飞书文档AI助手"
  version: "1.0.0"
  host: "0.0.0.0"
  port: 8080
  webhook_port: 8081

# 飞书配置 (已在前面提供)
feishu:
  app_id: "${FEISHU_APP_ID}"
  app_secret: "${FEISHU_APP_SECRET}"
  # ... (其他配置)

# Eino框架配置
eino:
  models:
    embedding:
      provider: "openai"
      model: "text-embedding-ada-002"
    chat:
      provider: "openai" 
      model: "gpt-4"
      
  # Milvus向量数据库
  milvus:
    host: "localhost"
    port: 19530
    collection: "feishu_docs"
    dimension: 1536
    
  # 文档处理配置
  document:
    chunk_size: 512
    chunk_overlap: 50
    batch_size: 100
```

---

## 🏗️ 详细系统设计

### 26. 系统架构图

#### 整体架构概览
```
                    🌐 Internet
                         │
                    ┌────▼────┐
                    │  Nginx  │ (负载均衡 + SSL)
                    │ Gateway │
                    └────┬────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
   ┌────▼────┐     ┌────▼────┐     ┌────▼────┐
   │API 服务1│     │API 服务2│     │API 服务3│
   │:8080    │     │:8080    │     │:8080    │
   └────┬────┘     └────┬────┘     └────┬────┘
        │                │                │
        └────────────────┼────────────────┘
                         │
            ┌────────────┴────────────┐
            │                         │
       ┌────▼────┐              ┌────▼────┐
       │Webhook  │              │定时任务  │
       │监听器   │              │调度器   │
       │:8081    │              │Cron     │
       └────┬────┘              └────┬────┘
            │                        │
    ┌───────┴────────────────────────┴───────┐
    │             Eino 处理引擎               │
    ├─────────────────┬─────────────────────┤
    │ Document        │ Embedding           │
    │ Processor       │ Generator           │
    ├─────────────────┼─────────────────────┤
    │ Content         │ Vector              │
    │ Splitter        │ Indexer             │
    └─────────────────┴─────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
   ┌────▼────┐     ┌────▼────┐     ┌────▼────┐
   │ Milvus  │     │ MySQL   │     │ Redis   │
   │向量数据库│     │元数据库  │     │缓存层   │
   │:19530   │     │:3306    │     │:6379    │
   └─────────┘     └─────────┘     └─────────┘
                                        │
                    ┌──────────────────┴──────────────────┐
                    │                                     │
               ┌────▼────┐                          ┌────▼────┐
               │飞书API   │                          │OpenAI   │
               │集成层    │                          │API      │
               └─────────┘                          └─────────┘
```

#### 数据流架构
```
┌─────────────┐    HTTP/HTTPS    ┌─────────────┐
│   前端应用   │◄────────────────▶│  API网关    │
│ React/Vue   │                  │   Nginx     │
└─────────────┘                  └──────┬──────┘
                                        │
                                        ▼
                                ┌─────────────┐
                                │  业务服务层  │
                                │  Go Server  │
                                └──────┬──────┘
                                       │
                    ┌──────────────────┼──────────────────┐
                    │                  │                  │
                    ▼                  ▼                  ▼
            ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
            │  数据处理层  │    │   存储层    │    │  外部服务层  │
            │ Eino Engine │    │MySQL/Milvus│    │飞书/OpenAI  │
            └─────────────┘    └─────────────┘    └─────────────┘
```

### 27. 数据库Schema设计

#### MySQL 元数据存储
```sql
-- 文档信息表
CREATE TABLE documents (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    doc_token VARCHAR(64) NOT NULL UNIQUE COMMENT '飞书文档token',
    doc_type ENUM('doc', 'sheet', 'bitable', 'wiki') NOT NULL COMMENT '文档类型',
    title VARCHAR(512) NOT NULL COMMENT '文档标题',
    url VARCHAR(1024) NOT NULL COMMENT '文档链接',
    owner_id VARCHAR(64) NOT NULL COMMENT '文档所有者ID',
    owner_name VARCHAR(128) NOT NULL COMMENT '文档所有者姓名',
    folder_token VARCHAR(64) COMMENT '所属文件夹token',
    content_hash VARCHAR(64) COMMENT '内容hash，用于检测变更',
    word_count INT DEFAULT 0 COMMENT '文档字数',
    chunk_count INT DEFAULT 0 COMMENT '分块数量',
    status ENUM('active', 'deleted', 'processing', 'failed') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    synced_at TIMESTAMP NULL COMMENT '最后同步时间',
    
    INDEX idx_doc_token (doc_token),
    INDEX idx_doc_type (doc_type),
    INDEX idx_owner_id (owner_id),
    INDEX idx_updated_at (updated_at),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文档基本信息表';

-- 文档内容块表
CREATE TABLE document_chunks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    document_id BIGINT NOT NULL COMMENT '文档ID',
    chunk_index INT NOT NULL COMMENT '块序号',
    content TEXT NOT NULL COMMENT '文本内容',
    content_type ENUM('text', 'table', 'image', 'code') DEFAULT 'text',
    char_count INT NOT NULL DEFAULT 0,
    vector_id VARCHAR(64) COMMENT 'Milvus中的向量ID',
    metadata JSON COMMENT '额外元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    INDEX idx_document_id (document_id),
    INDEX idx_vector_id (vector_id),
    UNIQUE KEY uk_doc_chunk (document_id, chunk_index)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文档内容分块表';

-- 用户表
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    feishu_user_id VARCHAR(64) NOT NULL UNIQUE COMMENT '飞书用户ID',
    name VARCHAR(128) NOT NULL COMMENT '用户姓名',
    avatar_url VARCHAR(512) COMMENT '头像链接',
    email VARCHAR(256) COMMENT '邮箱',
    department VARCHAR(256) COMMENT '部门',
    role ENUM('admin', 'user', 'guest') DEFAULT 'user',
    status ENUM('active', 'inactive') DEFAULT 'active',
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_feishu_user_id (feishu_user_id),
    INDEX idx_email (email),
    INDEX idx_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息表';

-- 文档权限表
CREATE TABLE document_permissions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    document_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    permission ENUM('owner', 'editor', 'viewer') NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    granted_by BIGINT COMMENT '授权者ID',
    
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uk_doc_user (document_id, user_id),
    INDEX idx_user_id (user_id),
    INDEX idx_permission (permission)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文档权限表';

-- 问答记录表
CREATE TABLE qa_sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id VARCHAR(64) NOT NULL COMMENT '会话ID',
    user_id BIGINT NOT NULL,
    question TEXT NOT NULL,
    answer TEXT,
    context_docs JSON COMMENT '参考的文档列表',
    satisfaction_score TINYINT COMMENT '满意度评分 1-5',
    response_time_ms INT COMMENT '响应时间(毫秒)',
    model_used VARCHAR(64) COMMENT '使用的AI模型',
    tokens_used INT COMMENT '消耗的token数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_session_id (session_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='问答记录表';

-- 系统配置表
CREATE TABLE system_configs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    config_key VARCHAR(128) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    config_type ENUM('string', 'int', 'float', 'bool', 'json') DEFAULT 'string',
    description VARCHAR(512),
    updated_by BIGINT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';

-- 同步日志表
CREATE TABLE sync_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    sync_type ENUM('full', 'incremental', 'single') NOT NULL,
    document_id BIGINT COMMENT '单文档同步时的文档ID',
    status ENUM('running', 'success', 'failed') NOT NULL,
    total_docs INT DEFAULT 0,
    processed_docs INT DEFAULT 0,
    failed_docs INT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    
    INDEX idx_sync_type (sync_type),
    INDEX idx_status (status),
    INDEX idx_started_at (started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='同步日志表';
```

#### Milvus 向量数据库Schema
```go
// Milvus Collection Schema
type DocumentVector struct {
    ID          int64     `milvus:"id,primary_key,auto_id"`           // 主键，自动生成
    DocumentID  int64     `milvus:"document_id"`                      // 文档ID，对应MySQL
    ChunkIndex  int32     `milvus:"chunk_index"`                      // 块索引
    Vector      []float32 `milvus:"vector,dim:1536"`                 // 1536维向量
    ContentHash string    `milvus:"content_hash,varchar:64"`         // 内容hash
    DocType     string    `milvus:"doc_type,varchar:16"`             // 文档类型
    UpdateTime  int64     `milvus:"update_time"`                     // 更新时间戳
}

// 创建Collection的参数
CollectionSchema: {
    Name: "feishu_docs",
    Fields: [
        {Name: "id", DataType: Int64, PrimaryKey: true, AutoID: true},
        {Name: "document_id", DataType: Int64},
        {Name: "chunk_index", DataType: Int32},
        {Name: "vector", DataType: FloatVector, Dim: 1536},
        {Name: "content_hash", DataType: VarChar, MaxLength: 64},
        {Name: "doc_type", DataType: VarChar, MaxLength: 16},
        {Name: "update_time", DataType: Int64},
    ],
    Indexes: [
        {FieldName: "vector", IndexType: "IVF_FLAT", MetricType: "L2", Params: {"nlist": 1024}},
        {FieldName: "document_id", IndexType: "STL_SORT"},
        {FieldName: "doc_type", IndexType: "STL_SORT"},
    ]
}
```

### 28. API接口设计

#### RESTful API规范
```go
// API Base URL: https://doc-ai.yourcompany.com/api/v1

// 1. 认证相关接口
type AuthAPI struct{}

// POST /api/v1/auth/login - 用户登录
type LoginRequest struct {
    Code string `json:"code" binding:"required"` // 飞书授权码
}

type LoginResponse struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
    User      UserInfo  `json:"user"`
}

// GET /api/v1/auth/profile - 获取用户信息
type UserInfo struct {
    ID         int64  `json:"id"`
    Name       string `json:"name"`
    Avatar     string `json:"avatar"`
    Email      string `json:"email"`
    Department string `json:"department"`
    Role       string `json:"role"`
}

// 2. 文档相关接口
type DocumentAPI struct{}

// GET /api/v1/documents - 获取文档列表
type GetDocumentsRequest struct {
    Page     int    `form:"page,default=1"`
    PageSize int    `form:"page_size,default=20"`
    Type     string `form:"type"`        // 文档类型过滤
    Owner    string `form:"owner"`       // 所有者过滤
    Keyword  string `form:"keyword"`     // 关键词搜索
    StartDate string `form:"start_date"` // 开始日期
    EndDate   string `form:"end_date"`   // 结束日期
}

type GetDocumentsResponse struct {
    Documents []DocumentInfo `json:"documents"`
    Total     int64          `json:"total"`
    Page      int            `json:"page"`
    PageSize  int            `json:"page_size"`
}

type DocumentInfo struct {
    ID          int64     `json:"id"`
    DocToken    string    `json:"doc_token"`
    Type        string    `json:"type"`
    Title       string    `json:"title"`
    URL         string    `json:"url"`
    Owner       string    `json:"owner"`
    WordCount   int       `json:"word_count"`
    ChunkCount  int       `json:"chunk_count"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    SyncedAt    *time.Time `json:"synced_at"`
}

// POST /api/v1/documents/sync - 手动同步文档
type SyncDocumentsRequest struct {
    DocTokens []string `json:"doc_tokens"` // 指定文档，空则全量同步
    Force     bool     `json:"force"`      // 强制重新同步
}

type SyncDocumentsResponse struct {
    TaskID string `json:"task_id"`
    Status string `json:"status"`
}

// 3. 问答相关接口
type QAAPI struct{}

// POST /api/v1/qa/ask - 提问接口
type AskRequest struct {
    Question   string            `json:"question" binding:"required"`
    SessionID  string            `json:"session_id"`
    Context    map[string]interface{} `json:"context"`
    Filters    map[string]interface{} `json:"filters"` // 文档类型、时间范围等过滤
}

type AskResponse struct {
    Answer       string           `json:"answer"`
    SessionID    string           `json:"session_id"`
    Sources      []SourceDocument `json:"sources"`      // 参考文档
    Confidence   float64          `json:"confidence"`   // 置信度
    ResponseTime int              `json:"response_time"` // 响应时间ms
    TokensUsed   int              `json:"tokens_used"`   // token消耗
}

type SourceDocument struct {
    DocID     int64   `json:"doc_id"`
    Title     string  `json:"title"`
    URL       string  `json:"url"`
    Excerpt   string  `json:"excerpt"`   // 相关片段
    Score     float64 `json:"score"`     // 相关度分数
    ChunkIndex int    `json:"chunk_index"`
}

// GET /api/v1/qa/sessions/{session_id}/history - 获取对话历史
type GetHistoryResponse struct {
    SessionID string    `json:"session_id"`
    Messages  []Message `json:"messages"`
}

type Message struct {
    Type      string    `json:"type"`      // "question" | "answer"
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
    Sources   []SourceDocument `json:"sources,omitempty"`
}

// POST /api/v1/qa/feedback - 反馈评价
type FeedbackRequest struct {
    SessionID string `json:"session_id" binding:"required"`
    MessageID string `json:"message_id" binding:"required"`
    Score     int    `json:"score" binding:"required,min=1,max=5"` // 1-5分
    Comment   string `json:"comment"`
}

// 4. 统计分析接口
type AnalyticsAPI struct{}

// GET /api/v1/analytics/dashboard - 仪表板数据
type DashboardResponse struct {
    DocumentStats DocumentStats `json:"document_stats"`
    QAStats       QAStats       `json:"qa_stats"`
    UserStats     UserStats     `json:"user_stats"`
    SyncStats     SyncStats     `json:"sync_stats"`
}

type DocumentStats struct {
    TotalDocs     int64 `json:"total_docs"`
    DocsToday     int64 `json:"docs_today"`
    DocsThisWeek  int64 `json:"docs_this_week"`
    DocsThisMonth int64 `json:"docs_this_month"`
    TypeDistribution map[string]int64 `json:"type_distribution"`
}

type QAStats struct {
    TotalQuestions int64   `json:"total_questions"`
    QuestionsToday int64   `json:"questions_today"`
    AvgResponseTime int    `json:"avg_response_time"`
    AvgSatisfaction float64 `json:"avg_satisfaction"`
    PopularQuestions []string `json:"popular_questions"`
}

// 5. 管理员接口
type AdminAPI struct{}

// GET /api/v1/admin/sync/logs - 获取同步日志
type GetSyncLogsResponse struct {
    Logs []SyncLog `json:"logs"`
    Total int64    `json:"total"`
}

type SyncLog struct {
    ID            int64     `json:"id"`
    SyncType      string    `json:"sync_type"`
    Status        string    `json:"status"`
    TotalDocs     int       `json:"total_docs"`
    ProcessedDocs int       `json:"processed_docs"`
    FailedDocs    int       `json:"failed_docs"`
    ErrorMessage  string    `json:"error_message,omitempty"`
    StartedAt     time.Time `json:"started_at"`
    CompletedAt   *time.Time `json:"completed_at"`
}

// POST /api/v1/admin/config - 更新系统配置
type UpdateConfigRequest struct {
    Configs map[string]interface{} `json:"configs"`
}
```

### 29. 前端组件结构

#### React/TypeScript 组件架构
```typescript
// 项目结构
src/
├── components/          // 通用组件
│   ├── common/
│   │   ├── Header.tsx
│   │   ├── Sidebar.tsx
│   │   ├── Loading.tsx
│   │   ├── Toast.tsx
│   │   └── ErrorBoundary.tsx
│   ├── forms/
│   │   ├── SearchBox.tsx
│   │   ├── FilterPanel.tsx
│   │   └── DateRangePicker.tsx
│   └── charts/
│       ├── DocumentChart.tsx
│       └── UsageChart.tsx
├── pages/               // 页面组件
│   ├── Dashboard/
│   │   ├── index.tsx
│   │   ├── DocumentStats.tsx
│   │   └── QAStats.tsx
│   ├── Chat/
│   │   ├── index.tsx
│   │   ├── ChatWindow.tsx
│   │   ├── MessageList.tsx
│   │   ├── MessageItem.tsx
│   │   ├── InputBox.tsx
│   │   └── SourcePanel.tsx
│   ├── Documents/
│   │   ├── index.tsx
│   │   ├── DocumentList.tsx
│   │   ├── DocumentCard.tsx
│   │   └── DocumentDetail.tsx
│   └── Admin/
│       ├── index.tsx
│       ├── SyncManager.tsx
│       ├── UserManager.tsx
│       └── ConfigManager.tsx
├── hooks/               // 自定义Hooks
│   ├── useAuth.ts
│   ├── useChat.ts
│   ├── useDocuments.ts
│   └── useWebSocket.ts
├── services/            // API服务层
│   ├── api.ts
│   ├── auth.ts
│   ├── chat.ts
│   └── documents.ts
├── store/               // 状态管理
│   ├── index.ts
│   ├── authSlice.ts
│   ├── chatSlice.ts
│   └── documentsSlice.ts
├── types/               // 类型定义
│   ├── api.ts
│   ├── chat.ts
│   └── document.ts
├── utils/               // 工具函数
│   ├── request.ts
│   ├── format.ts
│   └── constants.ts
└── styles/              // 样式文件
    ├── globals.css
    ├── components.css
    └── variables.css
```

#### 核心组件实现
```typescript
// 1. 主聊天组件
interface ChatWindowProps {
  sessionId?: string;
  onNewSession: (sessionId: string) => void;
}

const ChatWindow: React.FC<ChatWindowProps> = ({ sessionId, onNewSession }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const { sendMessage, history } = useChat(sessionId);

  const handleSendMessage = async (question: string) => {
    setIsLoading(true);
    try {
      const response = await sendMessage(question);
      setMessages(prev => [...prev, 
        { type: 'question', content: question, timestamp: new Date() },
        { type: 'answer', content: response.answer, sources: response.sources, timestamp: new Date() }
      ]);
    } catch (error) {
      // 错误处理
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="chat-window">
      <MessageList messages={messages} />
      <InputBox onSend={handleSendMessage} disabled={isLoading} />
      {isLoading && <Loading />}
    </div>
  );
};

// 2. 消息列表组件
const MessageList: React.FC<{ messages: Message[] }> = ({ messages }) => {
  return (
    <div className="message-list">
      {messages.map((message, index) => (
        <MessageItem 
          key={index} 
          message={message}
          showSources={message.type === 'answer'}
        />
      ))}
    </div>
  );
};

// 3. 文档搜索组件
const DocumentSearch: React.FC = () => {
  const [filters, setFilters] = useState<DocumentFilters>({});
  const { documents, loading, searchDocuments } = useDocuments();

  return (
    <div className="document-search">
      <SearchBox onSearch={searchDocuments} />
      <FilterPanel filters={filters} onChange={setFilters} />
      <DocumentList documents={documents} loading={loading} />
    </div>
  );
};

// 4. 仪表板组件
const Dashboard: React.FC = () => {
  const { stats, loading } = useDashboardStats();

  if (loading) return <Loading />;

  return (
    <div className="dashboard">
      <div className="stats-grid">
        <DocumentStats stats={stats.documentStats} />
        <QAStats stats={stats.qaStats} />
        <UserStats stats={stats.userStats} />
        <SyncStats stats={stats.syncStats} />
      </div>
      <div className="charts-section">
        <DocumentChart data={stats.documentTrends} />
        <UsageChart data={stats.usageTrends} />
      </div>
    </div>
  );
};

// 5. 自定义Hooks
const useChat = (sessionId?: string) => {
  const [history, setHistory] = useState<Message[]>([]);
  
  const sendMessage = async (question: string) => {
    const response = await fetch('/api/v1/qa/ask', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question, session_id: sessionId })
    });
    return response.json();
  };

  const loadHistory = async () => {
    if (sessionId) {
      const response = await fetch(`/api/v1/qa/sessions/${sessionId}/history`);
      const data = await response.json();
      setHistory(data.messages);
    }
  };

  useEffect(() => {
    loadHistory();
  }, [sessionId]);

  return { sendMessage, history, loadHistory };
};
```

### 30. 部署架构

#### 容器化部署方案
```yaml
# docker-compose.yml
version: '3.8'

services:
  # Nginx 网关
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
      - ./nginx/logs:/var/log/nginx
    depends_on:
      - api-server
      - web-app
    networks:
      - feishu-network

  # API服务 (多实例)
  api-server:
    build:
      context: .
      dockerfile: docker/Dockerfile.api
    environment:
      - GO_ENV=production
      - MYSQL_HOST=mysql
      - REDIS_HOST=redis
      - MILVUS_HOST=milvus-standalone
    volumes:
      - ./configs:/app/configs
      - ./logs:/app/logs
    depends_on:
      - mysql
      - redis
      - milvus-standalone
    deploy:
      replicas: 3
    networks:
      - feishu-network

  # Web前端
  web-app:
    build:
      context: ./web
      dockerfile: Dockerfile
    environment:
      - NODE_ENV=production
      - REACT_APP_API_BASE_URL=https://doc-ai.yourcompany.com/api
    networks:
      - feishu-network

  # MySQL数据库
  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
      - MYSQL_DATABASE=feishu_doc_ai
      - MYSQL_USER=app_user
      - MYSQL_PASSWORD=${MYSQL_PASSWORD}
    volumes:
      - mysql-data:/var/lib/mysql
      - ./database/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "3306:3306"
    networks:
      - feishu-network

  # Redis缓存
  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis-data:/data
    ports:
      - "6379:6379"
    networks:
      - feishu-network

  # Milvus向量数据库
  etcd:
    container_name: milvus-etcd
    image: quay.io/coreos/etcd:v3.5.0
    environment:
      - ETCD_AUTO_COMPACTION_MODE=revision
      - ETCD_AUTO_COMPACTION_RETENTION=1000
      - ETCD_QUOTA_BACKEND_BYTES=4294967296
    volumes:
      - etcd-data:/etcd
    command: etcd -advertise-client-urls=http://127.0.0.1:2379 -listen-client-urls http://0.0.0.0:2379 --data-dir /etcd
    networks:
      - feishu-network

  minio:
    container_name: milvus-minio
    image: minio/minio:RELEASE.2022-03-17T06-34-49Z
    environment:
      MINIO_ACCESS_KEY: minioadmin
      MINIO_SECRET_KEY: minioadmin
    volumes:
      - minio-data:/data
    command: minio server /data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
    networks:
      - feishu-network

  milvus-standalone:
    container_name: milvus-standalone
    image: milvusdb/milvus:v2.3.1
    command: ["milvus", "run", "standalone"]
    environment:
      ETCD_ENDPOINTS: etcd:2379
      MINIO_ADDRESS: minio:9000
    volumes:
      - milvus-data:/var/lib/milvus
    ports:
      - "19530:19530"
      - "9091:9091"
    depends_on:
      - "etcd"
      - "minio"
    networks:
      - feishu-network

  # 监控服务
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    networks:
      - feishu-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana:/etc/grafana/provisioning
    networks:
      - feishu-network

volumes:
  mysql-data:
  redis-data:
  milvus-data:
  etcd-data:
  minio-data:
  prometheus-data:
  grafana-data:

networks:
  feishu-network:
    driver: bridge
```

#### Kubernetes部署配置
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: feishu-doc-ai

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: feishu-doc-ai
data:
  config.yaml: |
    app:
      name: "飞书文档AI助手"
      port: 8080
    feishu:
      app_id: "${FEISHU_APP_ID}"
      app_secret: "${FEISHU_APP_SECRET}"
    # ... 其他配置

---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
  namespace: feishu-doc-ai
type: Opaque
data:
  mysql-password: <base64-encoded-password>
  redis-password: <base64-encoded-password>
  openai-api-key: <base64-encoded-key>

---
# k8s/api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: feishu-doc-ai
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      containers:
      - name: api-server
        image: feishu-doc-ai/api-server:latest
        ports:
        - containerPort: 8080
        - containerPort: 8081  # webhook port
        env:
        - name: MYSQL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secret
              key: mysql-password
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secret
              key: redis-password
        volumeMounts:
        - name: config-volume
          mountPath: /app/configs
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi" 
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config-volume
        configMap:
          name: app-config

---
# k8s/api-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: api-server-service
  namespace: feishu-doc-ai
spec:
  selector:
    app: api-server
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: webhook
    port: 8081
    targetPort: 8081
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-ingress
  namespace: feishu-doc-ai
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - doc-ai.yourcompany.com
    secretName: app-tls
  rules:
  - host: doc-ai.yourcompany.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-server-service
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-app-service
            port:
              number: 80

---
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-server-hpa
  namespace: feishu-doc-ai
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-server
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## 🚀 部署方案

### 基础设施需求

**服务器配置:**
- **CPU**: 4核心以上
- **内存**: 16GB以上  
- **存储**: 500GB SSD
- **网络**: 10Mbps以上带宽

**依赖服务:**
```yaml
version: '3.8'
services:
  # Milvus向量数据库
  milvus:
    image: milvusdb/milvus:latest
    ports:
      - "19530:19530"
    volumes:
      - milvus_data:/var/lib/milvus
    environment:
      - MILVUS_CONFIG_PATH=/milvus/configs/milvus.yaml
  
  # MySQL元数据存储
  mysql:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      - MYSQL_ROOT_PASSWORD=your_password
      - MYSQL_DATABASE=document_ai
    volumes:
      - mysql_data:/var/lib/mysql
  
  # Redis缓存
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
  
  # 应用主服务
  document-ai:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - milvus
      - mysql
      - redis
    environment:
      - MILVUS_HOST=milvus
      - MYSQL_HOST=mysql
      - REDIS_HOST=redis
    volumes:
      - ./documents:/app/documents
      - ./logs:/app/logs

volumes:
  milvus_data:
  mysql_data:
  redis_data:
```

### 环境配置

**应用配置文件 (`config.yaml`):**
```yaml
# 应用配置
app:
  name: "产品文档AI助手"
  version: "1.0.0"
  host: "0.0.0.0"
  port: 8080
  debug: false

# 数据库配置
database:
  milvus:
    host: "localhost"
    port: 19530
    collection_name: "product_docs"
    
  mysql:
    host: "localhost"
    port: 3306
    username: "root"
    password: "your_password"
    database: "document_ai"
    
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0

# AI模型配置
models:
  embedding:
    provider: "openai"
    model: "text-embedding-ada-002"
    api_key: "${OPENAI_API_KEY}"
    
  chat:
    provider: "openai"
    model: "gpt-4"
    api_key: "${OPENAI_API_KEY}"
    temperature: 0.1
    max_tokens: 2000

# 文档监控配置
monitoring:
  watch_paths:
    - "/data/product_docs"
    - "/data/requirements"
  file_types: [".docx", ".pdf", ".md", ".txt"]
  update_delay: 5  # 秒

# 飞书集成配置
feishu:
  # 飞书应用配置
  app_id: "${FEISHU_APP_ID}"
  app_secret: "${FEISHU_APP_SECRET}"
  
  # API配置
  api_base_url: "https://open.feishu.cn"
  
  # 机器人配置
  bot:
    webhook_url: "${FEISHU_WEBHOOK_URL}"
    encrypt_key: "${FEISHU_ENCRYPT_KEY}"
    verification_token: "${FEISHU_VERIFICATION_TOKEN}"
  
  # 文档同步配置
  sync:
    batch_size: 50  # 批量同步文档数量
    rate_limit: 100  # 每分钟API调用限制
    retry_times: 3   # 失败重试次数
    sync_interval: 3600  # 全量同步间隔(秒)
    
  # 监听的文档文件夹
  watch_folders:
    - "产品需求文档"
    - "技术设计文档" 
    - "项目管理文档"
    - "用户研究报告"
  
  # Webhook事件订阅
  events:
    - "drive.file.created_in_folder_v1"
    - "drive.file.edit_v1"
    - "drive.file.title_updated_v1"
    - "drive.file.trashed_v1"

# 日志配置
logging:
  level: "INFO"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  file: "/app/logs/app.log"
  max_size: "100MB"
  backup_count: 5
```

---

## 🔐 飞书API权限配置

### 必需的API权限

**应用权限申请清单:**
```yaml
# 文档相关权限
document_permissions:
  - "drive:drive"              # 云文档基础权限
  - "docs:doc"                # 飞书文档权限  
  - "sheets:spreadsheet"      # 飞书表格权限
  - "wiki:wiki"               # 飞书知识库权限
  - "bitable:app"             # 多维表格权限

# 用户和组织权限
user_permissions:
  - "contact:user.base:readonly"     # 读取用户基本信息
  - "contact:department.base:readonly" # 读取部门信息

# 机器人权限
bot_permissions:
  - "im:message"              # 发送消息权限
  - "im:message.group_at_msg" # 群聊@消息权限
```

### 安全配置要点

**1. Token安全管理**
- 使用环境变量存储敏感信息
- 定期刷新Access Token
- 实现Token失效自动重试机制

**2. 权限最小化原则**
- 只申请必需的最小权限
- 按用户角色控制文档访问
- 定期审计权限使用情况

**3. 数据安全**
- 本地向量库加密存储
- API调用日志脱敏处理
- 实现数据备份和恢复机制

### 飞书应用创建步骤

**Step 1: 创建飞书企业应用**
```bash
1. 访问 https://open.feishu.cn/app
2. 点击"创建企业自建应用"
3. 填写应用名称："产品文档AI助手"
4. 选择应用类型："机器人"
5. 上传应用图标和描述
```

**Step 2: 配置权限范围**
```bash
1. 进入"权限管理"页面
2. 按照上述权限清单申请权限
3. 提交审核（通常需要1-3个工作日）
4. 获得管理员批准后生效
```

**Step 3: 获取应用凭证**
```bash
1. 记录 App ID 和 App Secret
2. 配置服务器回调地址
3. 设置事件订阅Webhook
4. 测试API调用是否正常
```

---

## 📊 效果预期与成本分析

### 预期效果

**效率提升:**
- 📈 文档查找时间减少 **80%** (从平均30分钟降至5分钟)
- 🎯 查询准确率达到 **90%以上**
- 📱 支持移动端查询，随时随地访问
- 🔄 文档更新通知及时率 **100%**

**用户体验:**
- 💬 自然语言查询，无需记忆关键词
- 📱 多端访问，Web + 移动端 + 飞书
- 🔍 智能推荐相关文档
- 📊 可视化的查询结果展示

### 成本分析

**开发成本:**
```
人力成本：
- 后端开发工程师 1人 × 2个月 = 40K
- 前端开发工程师 1人 × 1个月 = 20K  
- 产品经理 0.5人 × 2个月 = 15K
- 测试工程师 0.5人 × 1个月 = 8K
小计：83K

技术成本：
- OpenAI API费用：约500元/月
- 服务器费用：约2000元/月
- 其他工具费用：约300元/月
小计：2800元/月
```

**投资回报率(ROI):**
```
时间成本节省：
- 20人团队 × 2小时/周 × 4周/月 × 500元/小时 = 80K/月

效率提升价值：
- 减少重复工作，提升决策速度
- 预估价值：20K/月

总收益：100K/月
投资回收期：约1个月
年化ROI：约1300%
```

---

## 🔄 后续扩展规划

### Phase 4: 智能化增强（未来6个月）

**多模态支持:**
- 📷 图片识别和理解（流程图、原型图）
- 🎥 视频内容分析（产品演示视频）
- 📊 表格数据智能解析

**智能写作助手:**
- 📝 基于历史文档自动生成新需求文档
- 🔄 需求变更影响分析
- 📋 自动生成测试用例

**团队协作增强:**
- 💬 集成钉钉、企业微信
- 📅 与项目管理工具集成（Jira、Trello）
- 🔔 智能提醒和任务分配

### Phase 5: 生态化集成（未来12个月）

**开发工具集成:**
- 🔧 与IDE集成（VS Code插件）
- 📱 移动端原生应用
- 🌐 浏览器插件

**数据洞察:**
- 📈 文档使用情况分析
- 🎯 热门需求识别
- 📊 团队知识图谱构建

---

## ✅ 结论与建议

### 项目可行性评估: ⭐⭐⭐⭐⭐

**技术可行性:** 🟢 **高**
- Eino框架成熟稳定，社区活跃
- RAG技术方案经过验证
- 所需技术栈团队已掌握

**商业价值:** 🟢 **高**
- 直接解决团队痛点
- ROI显著，投资回收期短
- 可扩展到其他业务场景

**实施风险:** 🟡 **低**
- 技术风险可控
- 开发周期合理
- 有降级方案

### 飞书集成特殊优势:

**🎯 无缝体验:**
- 文档源头直接接入，无需导出导入
- 权限体系天然同步，安全可靠
- 用户无需学习新工具，在飞书内直接使用

**🔄 实时性强:**
- Webhook实时监听文档变更
- 增量更新，资源消耗低
- 团队协作信息及时同步

**📱 移动友好:**
- 飞书移动端原生支持
- 随时随地查询文档信息
- 支持语音输入和语音回复

### 下一步行动建议:

1. **飞书应用申请:** 立即创建企业应用并申请权限（3-5天）
2. **API接口测试:** 验证飞书文档API可用性（2天）
3. **Eino环境搭建:** 配置开发和测试环境（3天）
4. **核心功能开发:** 按照实施计划开始开发（4周）

**关键里程碑:**
- **第1周:** 飞书API集成和权限配置完成
- **第2-3周:** 文档同步和向量化功能开发  
- **第4-5周:** 智能问答和实时更新功能
- **第6周:** 飞书机器人和用户界面开发
- **第7-8周:** 测试优化和上线部署

**预计项目启动时间:** 1周后（等待飞书权限审批）
**预计MVP版本上线:** 6周后
**预计完整功能上线:** 8周后

---

*这个方案充分利用了Eino框架的优势，既解决了实际业务痛点，又具备良好的扩展性。建议优先启动！* 🚀