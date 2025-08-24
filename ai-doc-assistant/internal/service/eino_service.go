package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	// Eino 框架核心组件
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	// Eino 扩展组件
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	embedder "github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino-ext/components/model/ark"
	retriever "github.com/cloudwego/eino-ext/components/retriever/milvus"

	// Milvus SDK
	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"ai-doc-assistant/internal/config"
	"ai-doc-assistant/internal/model"
	"ai-doc-assistant/internal/repository"
)

// Document 简化的文档结构（用于演示）
type Document struct {
	ID           string
	Title        string
	Content      string
	Author       string
	Department   string
	DocumentType string
	Version      string
	CreatedAt    string
}

// EinoService Eino框架服务
type EinoService struct {
	config       *config.Config
	embedder     *embedder.Embedder
	milvusClient cli.Client
	indexer      *milvus.Indexer
	retriever    *retriever.Retriever
	transformer  document.Transformer
	chatModel    *ark.ChatModel
	tools        []tool.BaseTool
	database     *repository.Database
	initialized  bool
}

// Milvus集合结构定义
var milvusSchema = []*entity.Field{
	{
		Name:        "id",
		DataType:    entity.FieldTypeVarChar,
		TypeParams:  map[string]string{"max_length": "255"},
		PrimaryKey:  true,
		Description: "文档块的唯一标识符",
	},
	{
		Name:        "vector",
		DataType:    entity.FieldTypeBinaryVector,
		TypeParams:  map[string]string{"dim": "81920"}, // 维度需与embedding模型匹配
		Description: "文档内容的向量表示",
	},
	{
		Name:        "content",
		DataType:    entity.FieldTypeVarChar,
		TypeParams:  map[string]string{"max_length": "8192"},
		Description: "原始文本内容",
	},
	{
		Name:        "metadata",
		DataType:    entity.FieldTypeJSON,
		Description: "文档元数据信息",
	},
}

// NewEinoService 创建Eino服务实例
func NewEinoService(cfg *config.Config) (*EinoService, error) {
	service := &EinoService{
		config: cfg,
	}

	ctx := context.Background()

	// 初始化所有组件
	if err := service.initialize(ctx); err != nil {
		return nil, fmt.Errorf("初始化Eino服务失败: %w", err)
	}

	return service, nil
}

// SetDatabase 设置数据库连接
func (s *EinoService) SetDatabase(db *repository.Database) {
	s.database = db
}

// QueryKnowledgeWithHistory 问答并保存历史记录
func (s *EinoService) QueryKnowledgeWithHistory(ctx context.Context, req *model.QueryRequest) (*model.QueryResponse, error) {
	// 执行问答
	response, err := s.QueryKnowledge(ctx, req.Question)
	if err != nil {
		return nil, err
	}

	// 保存历史记录（如果有数据库连接）
	if s.database != nil {
		retrievedDocsJSON, _ := json.Marshal(response.Sources)
		
		history := &model.QueryHistory{
			ID:            response.QueryID,
			UserID:        req.UserID,
			Query:         req.Question,
			Response:      response.Answer,
			ResponseTimeMs: response.ResponseTime,
			RetrievedDocs: string(retrievedDocsJSON),
			CreatedAt:    time.Now(),
		}
		
		// 异步保存，不影响响应速度
		go func() {
			if err := s.database.SaveQueryHistory(history); err != nil {
				log.Printf("保存查询历史失败: %v", err)
			}
		}()
	}

	return response, nil
}

// GetQueryHistory 获取查询历史
func (s *EinoService) GetQueryHistory(userID string, limit int) ([]model.QueryHistory, error) {
	if s.database == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	return s.database.GetQueryHistoryByUserID(userID, limit)
}

// UpdateQueryFeedback 更新查询反馈
func (s *EinoService) UpdateQueryFeedback(queryID string, satisfactionScore int, feedback string) error {
	if s.database == nil {
		return fmt.Errorf("数据库未初始化")
	}
	return s.database.UpdateQueryFeedback(queryID, satisfactionScore, feedback)
}

// initialize 初始化所有Eino组件
func (s *EinoService) initialize(ctx context.Context) error {
	// 1. 初始化Embedder
	if err := s.initEmbedder(ctx); err != nil {
		return fmt.Errorf("初始化Embedder失败: %w", err)
	}

	// 2. 初始化Milvus
	if err := s.initMilvus(ctx); err != nil {
		return fmt.Errorf("初始化Milvus失败: %w", err)
	}

	// 3. 初始化Transformer
	if err := s.initTransformer(ctx); err != nil {
		return fmt.Errorf("初始化Transformer失败: %w", err)
	}

	// 4. 初始化ChatModel
	if err := s.initChatModel(ctx); err != nil {
		return fmt.Errorf("初始化ChatModel失败: %w", err)
	}

	// 5. 初始化Tools
	if err := s.initTools(ctx); err != nil {
		return fmt.Errorf("初始化Tools失败: %w", err)
	}

	s.initialized = true
	log.Println("✅ Eino服务初始化完成")
	return nil
}

// initEmbedder 初始化嵌入模型
func (s *EinoService) initEmbedder(ctx context.Context) error {
	timeout := 30 * time.Second
	embedder, err := embedder.NewEmbedder(ctx, &embedder.EmbeddingConfig{
		APIKey:  s.config.AI.APIKey,
		Model:   s.config.AI.Models.Embedding,
		Timeout: &timeout,
	})
	if err != nil {
		return err
	}
	s.embedder = embedder
	log.Println("✓ Embedder 初始化成功")
	return nil
}

// initMilvus 初始化Milvus向量数据库
func (s *EinoService) initMilvus(ctx context.Context) error {
	// 连接Milvus
	milvusAddr := fmt.Sprintf("%s:%d", s.config.Database.Milvus.Host, s.config.Database.Milvus.Port)
	client, err := cli.NewClient(ctx, cli.Config{Address: milvusAddr})
	if err != nil {
		return err
	}
	s.milvusClient = client

	// 设置集合
	collectionName := s.config.Database.Milvus.Database + "_documents"
	if err := s.setupMilvusCollection(ctx, collectionName); err != nil {
		return err
	}

	// 初始化Indexer
	indexerCfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: collectionName,
		Embedding:  s.embedder,
		Fields:     milvusSchema,
	}
	indexer, err := milvus.NewIndexer(ctx, indexerCfg)
	if err != nil {
		return err
	}
	s.indexer = indexer

	// 初始化Retriever
	retrieverCfg := &retriever.RetrieverConfig{
		Client:       client,
		Collection:   collectionName,
		Embedding:    s.embedder,
		OutputFields: []string{"content", "metadata"},
		TopK:         5,
	}
	retriever, err := retriever.NewRetriever(ctx, retrieverCfg)
	if err != nil {
		return err
	}
	s.retriever = retriever

	log.Println("✓ Milvus 组件初始化成功")
	return nil
}

// setupMilvusCollection 设置Milvus集合
func (s *EinoService) setupMilvusCollection(ctx context.Context, collectionName string) error {
	has, err := s.milvusClient.HasCollection(ctx, collectionName)
	if err != nil {
		return err
	}

	if !has {
		log.Printf("创建Milvus集合: %s", collectionName)
		schema := &entity.Schema{
			CollectionName: collectionName,
			Fields:         milvusSchema,
			Description:    "AI文档助手知识库",
		}

		if err := s.milvusClient.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
			return err
		}

		// 创建向量索引
		binFlatIndex, err := entity.NewIndexBinFlat(entity.HAMMING, 128)
		if err != nil {
			return err
		}

		if err := s.milvusClient.CreateIndex(ctx, collectionName, "vector", binFlatIndex, false); err != nil {
			return err
		}

		log.Println("✓ Milvus集合和索引创建成功")
	} else {
		log.Printf("✓ Milvus集合 %s 已存在", collectionName)
	}

	return nil
}

// initTransformer 初始化文档转换器
func (s *EinoService) initTransformer(ctx context.Context) error {
	transformer, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"#":   "Header 1",
			"##":  "Header 2", 
			"###": "Header 3",
		},
	})
	if err != nil {
		return err
	}
	s.transformer = transformer
	log.Println("✓ Transformer 初始化成功")
	return nil
}

// initChatModel 初始化聊天模型
func (s *EinoService) initChatModel(ctx context.Context) error {
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: s.config.AI.APIKey,
		Model:  s.config.AI.Models.Chat,
	})
	if err != nil {
		return err
	}
	s.chatModel = model
	log.Println("✓ ChatModel 初始化成功")
	return nil
}

// initTools 初始化工具集
func (s *EinoService) initTools(ctx context.Context) error {
	// 创建知识搜索工具
	knowledgeTool := &KnowledgeSearchTool{retriever: s.retriever}

	// 创建文档处理工具  
	docTool := &DocumentProcessorTool{
		indexer:     s.indexer,
		transformer: s.transformer,
	}

	s.tools = []tool.BaseTool{knowledgeTool, docTool}
	log.Printf("✓ 初始化了 %d 个工具", len(s.tools))
	return nil
}

// ProcessDocument 处理文档
func (s *EinoService) ProcessDocument(ctx context.Context, doc *Document) error {
	if !s.initialized {
		return fmt.Errorf("Eino服务未初始化")
	}

	// 转换为Eino文档格式
	einoDoc := &schema.Document{
		ID:      doc.ID,
		Content: doc.Content,
		MetaData: map[string]interface{}{
			"title":       doc.Title,
			"author":      doc.Author,
			"department":  doc.Department,
			"doc_type":    doc.DocumentType,
			"version":     doc.Version,
			"created_at":  doc.CreatedAt,
		},
	}

	// 使用Transformer分割文档
	chunks, err := s.transformer.Transform(ctx, []*schema.Document{einoDoc})
	if err != nil {
		return fmt.Errorf("文档分割失败: %w", err)
	}

	// 使用Indexer存储文档块
	storedIDs, err := s.indexer.Store(ctx, chunks)
	if err != nil {
		return fmt.Errorf("文档索引失败: %w", err)
	}

	log.Printf("✓ 文档 %s 处理完成，分割为 %d 块，存储ID: %v", doc.ID, len(chunks), storedIDs)
	return nil
}

// QueryKnowledge 智能问答
func (s *EinoService) QueryKnowledge(ctx context.Context, question string) (*model.QueryResponse, error) {
	if !s.initialized {
		return nil, fmt.Errorf("Eino服务未初始化")
	}

	startTime := time.Now()

	// 1. 使用Retriever检索相关文档
	docs, err := s.retriever.Retrieve(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("知识检索失败: %w", err)
	}

	log.Printf("检索到 %d 个相关文档块", len(docs))

	// 2. 构建RAG提示
	prompt := s.buildRAGPrompt(question, docs)
	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的文档助手，基于提供的文档内容回答用户问题。请确保答案准确、详细、有条理。"),
		schema.UserMessage(prompt),
	}

	// 3. 使用ChatModel生成回答
	response, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("生成回答失败: %w", err)
	}

	// 4. 构建响应
	responseTime := int(time.Since(startTime).Milliseconds())
	sources := s.convertDocsToSources(docs)
	queryID := uuid.New().String()

	return &model.QueryResponse{
		Answer:       response.Content,
		Sources:      sources,
		ResponseTime: responseTime,
		QueryID:      queryID,
		Confidence:   s.calculateConfidence(docs),
	}, nil
}

// buildRAGPrompt 构建RAG提示
func (s *EinoService) buildRAGPrompt(question string, docs []*schema.Document) string {
	prompt := "请基于以下文档内容回答用户问题：\n\n=== 文档内容 ===\n"

	for i, doc := range docs {
		prompt += fmt.Sprintf("[文档片段 %d]\n", i+1)
		prompt += fmt.Sprintf("内容: %s\n", doc.Content)
		if doc.MetaData != nil {
			if title, ok := doc.MetaData["title"].(string); ok {
				prompt += fmt.Sprintf("标题: %s\n", title)
			}
			if author, ok := doc.MetaData["author"].(string); ok {
				prompt += fmt.Sprintf("作者: %s\n", author)
			}
		}
		prompt += "\n"
	}

	prompt += fmt.Sprintf("=== 用户问题 ===\n%s\n\n", question)
	prompt += "请基于上述文档内容提供准确、详细的回答。如果文档内容不足以回答问题，请明确说明。"

	return prompt
}

// convertDocsToSources 转换文档为源信息
func (s *EinoService) convertDocsToSources(docs []*schema.Document) []model.DocumentSource {
	sources := make([]model.DocumentSource, len(docs))
	
	for i, doc := range docs {
		source := model.DocumentSource{
			DocumentID:   doc.ID,
			ChunkContent: doc.Content,
		}

		if doc.MetaData != nil {
			if title, ok := doc.MetaData["title"].(string); ok {
				source.DocumentTitle = title
			}
			if author, ok := doc.MetaData["author"].(string); ok {
				source.Author = author
			}
			if version, ok := doc.MetaData["version"].(string); ok {
				source.Version = version
			}
		}

		// 这里可以根据实际需要计算相关性分数
		source.Relevance = 1.0 - float64(i)*0.1 // 简单的相关性计算

		sources[i] = source
	}

	return sources
}

// calculateConfidence 计算置信度
func (s *EinoService) calculateConfidence(docs []*schema.Document) float64 {
	if len(docs) == 0 {
		return 0.3
	}

	// 基于检索到的文档数量计算置信度
	baseConfidence := 0.7
	docBonus := float64(len(docs)) * 0.05
	confidence := baseConfidence + docBonus

	if confidence > 0.95 {
		confidence = 0.95
	}

	return confidence
}

// GetRetriever 获取检索器（用于其他服务）
func (s *EinoService) GetRetriever() *retriever.Retriever {
	return s.retriever
}

// GetIndexer 获取索引器（用于其他服务）
func (s *EinoService) GetIndexer() *milvus.Indexer {
	return s.indexer
}

// GetTransformer 获取转换器（用于其他服务）
func (s *EinoService) GetTransformer() document.Transformer {
	return s.transformer
}

// Close 关闭服务
func (s *EinoService) Close() error {
	if s.milvusClient != nil {
		return s.milvusClient.Close()
	}
	return nil
}

// HealthCheck 健康检查
func (s *EinoService) HealthCheck(ctx context.Context) error {
	if !s.initialized {
		return fmt.Errorf("Eino服务未初始化")
	}

	// 简单的Milvus连接测试
	collections, err := s.milvusClient.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("Milvus连接失败: %w", err)
	}

	log.Printf("Milvus健康检查通过，发现 %d 个集合", len(collections))
	return nil
}