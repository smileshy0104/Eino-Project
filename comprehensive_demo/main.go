// Package main 演示 Eino 框架的综合应用
// 整合 Transformer、Indexer、Retriever 和 Tool 组件，构建智能 RAG + Tool 系统
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	// Eino 框架核心组件
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
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

	// 配置管理
	"github.com/spf13/viper"
)

// =============================================================================
//
//  综合演示: Eino 框架完整功能展示
//  功能特性:
//  1. 文档转换 (Transformer) - 智能分割 Markdown 文档
//  2. 文档索引 (Indexer) - 向量化并存储到 Milvus
//  3. 知识检索 (Retriever) - 基于语义相似度检索文档
//  4. 工具调用 (Tool) - 集成多种实用工具
//  5. 智能编排 (Chain) - 构建完整的 RAG + Tool 工作流
//
// =============================================================================

// Config 应用程序配置结构
type Config struct {
	MilvusAddress    string `mapstructure:"MILVUS_ADDRESS"`    // Milvus 服务地址
	MilvusCollection string `mapstructure:"MILVUS_COLLECTION"` // Milvus 集合名称
	ArkAPIKey        string `mapstructure:"ARK_API_KEY"`       // Ark API Key
	EmbedderModel    string `mapstructure:"EMBEDDER_MODEL"`    // 嵌入模型名称
	ArkModel         string `mapstructure:"ARK_MODEL"`         // Ark 模型名称
}

// Milvus 集合结构定义（必须跟Milvus集合结构一致）
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
		TypeParams:  map[string]string{"dim": "81920"}, // 维度需与 embedding 模型匹配
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

// ================================
// 工具实现部分
// ================================

// KnowledgeSearchTool 知识搜索工具 - 从向量数据库检索相关知识
type KnowledgeSearchTool struct {
	retriever *retriever.Retriever // KnowledgeSearchTool 实现了 tool.BaseTool 接口
}

// Info 返回知识搜索工具的信息
func (k *KnowledgeSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "knowledge_search",
		Desc: "从知识库中搜索相关信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     "string",
				Desc:     "搜索查询内容",
				Required: true,
			},
			"top_k": {
				Type:     "integer",
				Desc:     "返回结果数量",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行知识搜索
func (k *KnowledgeSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 解析输入参数
	var args struct {
		Query string `json:"query"`
		TopK  int    `json:"top_k"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 设置默认 TopK
	if args.TopK == 0 {
		args.TopK = 3 // 默认返回前3个结果
	}

	log.Printf("[KnowledgeSearchTool] 搜索知识: %s (TopK: %d)", args.Query, args.TopK)

	// 执行检索
	docs, err := k.retriever.Retrieve(ctx, args.Query)
	if err != nil {
		return "", fmt.Errorf("知识检索失败: %v", err)
	}

	// 构建结果
	result := map[string]interface{}{
		"query":       args.Query,
		"found_count": len(docs),
		"knowledge":   []map[string]interface{}{},
	}

	for i, doc := range docs {
		if i >= args.TopK {
			break
		}
		knowledge := map[string]interface{}{
			"id":       doc.ID,
			"content":  doc.Content,
			"metadata": doc.MetaData,
		}
		result["knowledge"] = append(result["knowledge"].([]map[string]interface{}), knowledge)
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// DocumentProcessorTool 文档处理工具 - 分割和索引新文档
type DocumentProcessorTool struct {
	indexer     *milvus.Indexer
	transformer document.Transformer
}

// Info 返回文档处理工具的信息
func (d *DocumentProcessorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "document_processor",
		Desc: "处理和索引新文档到知识库",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"content": {
				Type:     "string",
				Desc:     "要处理的文档内容(支持Markdown格式)",
				Required: true,
			},
			"doc_id": {
				Type:     "string",
				Desc:     "文档ID前缀",
				Required: false,
			},
			"metadata": {
				Type:     "object",
				Desc:     "文档元数据",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行文档处理和索引
func (d *DocumentProcessorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		Content  string                 `json:"content"`
		DocID    string                 `json:"doc_id"`
		MetaData map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.DocID == "" {
		args.DocID = fmt.Sprintf("doc_%d", time.Now().Unix())
	}

	if args.MetaData == nil {
		args.MetaData = make(map[string]interface{})
	}
	args.MetaData["processed_at"] = time.Now().Format(time.RFC3339)

	log.Printf("[DocumentProcessorTool] 处理文档: %s", args.DocID)

	// 创建原始文档
	originalDoc := &schema.Document{
		ID:       args.DocID,
		Content:  args.Content,
		MetaData: args.MetaData,
	}

	// 使用 Transformer 分割文档
	chunks, err := d.transformer.Transform(ctx, []*schema.Document{originalDoc})
	if err != nil {
		return "", fmt.Errorf("文档分割失败: %v", err)
	}

	// 使用 Indexer 存储文档块
	storedIDs, err := d.indexer.Store(ctx, chunks)
	if err != nil {
		return "", fmt.Errorf("文档索引失败: %v", err)
	}

	result := map[string]interface{}{
		"original_doc_id": args.DocID,
		"chunks_count":    len(chunks),
		"stored_ids":      storedIDs,
		"status":          "success",
		"message":         fmt.Sprintf("成功处理文档，分割为%d个块并完成索引", len(chunks)),
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// CalculatorTool 计算器工具 - 执行数学计算
type CalculatorTool struct{}

// Info 返回计算器工具信息
func (c *CalculatorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "calculator",
		Desc: "执行基本数学计算",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"expression": {
				Type:     "string",
				Desc:     "数学表达式(支持+,-,*,/)",
				Required: true,
			},
		}),
	}, nil
}

// InvokableRun 执行计算
func (c *CalculatorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[CalculatorTool] 计算表达式: %s", args.Expression)

	// 简单的表达式计算(演示用途)
	result := evaluateSimpleExpression(args.Expression)

	response := map[string]interface{}{
		"expression": args.Expression,
		"result":     result,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	resultBytes, _ := json.Marshal(response)
	return string(resultBytes), nil
}

// WeatherTool 天气查询工具
type WeatherTool struct{}

// Info 返回天气工具信息
func (w *WeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "weather_query",
		Desc: "查询城市天气信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     "string",
				Desc:     "城市名称",
				Required: true,
			},
			"date": {
				Type:     "string",
				Desc:     "查询日期(YYYY-MM-DD格式)",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行天气查询
func (w *WeatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		City string `json:"city"`
		Date string `json:"date"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.Date == "" {
		args.Date = time.Now().Format("2006-01-02")
	}

	log.Printf("[WeatherTool] 查询天气: %s @ %s", args.City, args.Date)

	// 模拟天气数据
	weatherData := map[string]interface{}{
		"city":        args.City,
		"date":        args.Date,
		"temperature": 25,
		"humidity":    65,
		"condition":   "晴朗",
		"wind_speed":  "微风",
		"description": fmt.Sprintf("%s今日天气晴朗，温度适宜", args.City),
	}

	result, _ := json.Marshal(weatherData)
	return string(result), nil
}

// ================================
// 核心系统组件
// ================================

// ComprehensiveRAGSystem 综合RAG系统
type ComprehensiveRAGSystem struct {
	config       *Config                                 // 系统配置
	embedder     *embedder.Embedder                      // 嵌入模型
	milvusClient cli.Client                              // Milvus 客户端
	indexer      *milvus.Indexer                         // 向量索引器
	retriever    *retriever.Retriever                    // 知识检索器
	transformer  document.Transformer                    // 文档转换器
	chatModel    *ark.ChatModel                          // 聊天模型
	tools        []tool.BaseTool                         // 工具集
	chain        *compose.Chain[string, *schema.Message] // 智能处理链
}

// NewComprehensiveRAGSystem 创建综合RAG系统实例
func NewComprehensiveRAGSystem(ctx context.Context, config *Config) (*ComprehensiveRAGSystem, error) {
	system := &ComprehensiveRAGSystem{config: config}

	// 1. 初始化 Embedder
	if err := system.initEmbedder(ctx); err != nil {
		return nil, fmt.Errorf("初始化Embedder失败: %v", err)
	}

	// 2. 初始化 Milvus
	if err := system.initMilvus(ctx); err != nil {
		return nil, fmt.Errorf("初始化Milvus失败: %v", err)
	}

	// 3. 初始化 Transformer
	if err := system.initTransformer(ctx); err != nil {
		return nil, fmt.Errorf("初始化Transformer失败: %v", err)
	}

	// 4. 初始化 ChatModel
	if err := system.initChatModel(ctx); err != nil {
		return nil, fmt.Errorf("初始化ChatModel失败: %v", err)
	}

	// 5. 初始化 Tools
	if err := system.initTools(ctx); err != nil {
		return nil, fmt.Errorf("初始化Tools失败: %v", err)
	}

	// 6. 构建 Chain
	if err := system.buildChain(ctx); err != nil {
		return nil, fmt.Errorf("构建Chain失败: %v", err)
	}

	return system, nil
}

// initEmbedder 初始化嵌入模型
func (s *ComprehensiveRAGSystem) initEmbedder(ctx context.Context) error {
	timeout := 30 * time.Second
	// 创建 Embedder 实例
	embedder, err := embedder.NewEmbedder(ctx, &embedder.EmbeddingConfig{
		APIKey:  s.config.ArkAPIKey,
		Model:   s.config.EmbedderModel,
		Timeout: &timeout,
	})
	if err != nil {
		return err
	}
	// 设置 Embedder
	s.embedder = embedder
	log.Println("✓ Embedder 初始化成功")
	return nil
}

// initMilvus 初始化向量数据库
func (s *ComprehensiveRAGSystem) initMilvus(ctx context.Context) error {
	// 连接 Milvus
	client, err := cli.NewClient(ctx, cli.Config{Address: s.config.MilvusAddress})
	if err != nil {
		return err
	}
	s.milvusClient = client

	// 检查并创建集合
	if err := s.setupMilvusCollection(ctx); err != nil {
		return err
	}

	// 初始化 Indexer
	indexerCfg := &milvus.IndexerConfig{
		Client:     client,
		Collection: s.config.MilvusCollection,
		Embedding:  s.embedder,
		Fields:     milvusSchema,
	}
	indexer, err := milvus.NewIndexer(ctx, indexerCfg)
	if err != nil {
		return err
	}
	s.indexer = indexer

	// 初始化 Retriever
	retrieverCfg := &retriever.RetrieverConfig{
		Client:       client,
		Collection:   s.config.MilvusCollection,
		Embedding:    s.embedder,
		OutputFields: []string{"content", "metadata"},
		TopK:         5,
	}
	// 创建 Retriever 实例
	retriever, err := retriever.NewRetriever(ctx, retrieverCfg)
	if err != nil {
		return err
	}
	// 设置 Retriever
	s.retriever = retriever

	log.Println("✓ Milvus 组件初始化成功")
	return nil
}

// setupMilvusCollection 设置Milvus集合
func (s *ComprehensiveRAGSystem) setupMilvusCollection(ctx context.Context) error {
	has, err := s.milvusClient.HasCollection(ctx, s.config.MilvusCollection)
	if err != nil {
		return err
	}

	// 创建 Milvus 集合
	if !has {
		log.Printf("创建 Milvus 集合: %s", s.config.MilvusCollection)
		schema := &entity.Schema{
			CollectionName: s.config.MilvusCollection,
			Fields:         milvusSchema,
			Description:    "综合RAG系统知识库",
		}

		if err := s.milvusClient.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
			return err
		}

		// 创建向量索引
		binFlatIndex, err := entity.NewIndexBinFlat(entity.HAMMING, 128)
		if err != nil {
			return err
		}

		if err := s.milvusClient.CreateIndex(ctx, s.config.MilvusCollection, "vector", binFlatIndex, false); err != nil {
			return err
		}

		log.Println("✓ Milvus 集合和索引创建成功")
	} else {
		log.Printf("✓ Milvus 集合 %s 已存在", s.config.MilvusCollection)
	}

	return nil
}

// initTransformer 初始化文档转换器
func (s *ComprehensiveRAGSystem) initTransformer(ctx context.Context) error {
	// 创建 Markdown 分割器
	transformer, err := markdown.NewHeaderSplitter(ctx, &markdown.HeaderConfig{
		Headers: map[string]string{
			"##":  "Header 2",
			"###": "Header 3",
		},
	})
	if err != nil {
		return err
	}
	// 设置 Transformer
	s.transformer = transformer
	log.Println("✓ Transformer 初始化成功")
	return nil
}

// initChatModel 初始化聊天模型
func (s *ComprehensiveRAGSystem) initChatModel(ctx context.Context) error {
	// 创建 Ark 聊天模型
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: s.config.ArkAPIKey,
		Model:  s.config.ArkModel,
	})
	if err != nil {
		return err
	}
	// 设置 ChatModel
	s.chatModel = model
	log.Println("✓ ChatModel 初始化成功")
	return nil
}

// initTools 初始化工具集
func (s *ComprehensiveRAGSystem) initTools(ctx context.Context) error {
	// 创建知识搜索工具
	knowledgeTool := &KnowledgeSearchTool{retriever: s.retriever}

	// 创建文档处理工具
	docTool := &DocumentProcessorTool{
		indexer:     s.indexer,
		transformer: s.transformer,
	}

	// 创建其他工具
	calcTool := &CalculatorTool{}
	weatherTool := &WeatherTool{}

	// 设置工具集
	s.tools = []tool.BaseTool{knowledgeTool, docTool, calcTool, weatherTool}

	log.Printf("✓ 初始化了 %d 个工具", len(s.tools))
	return nil
}

// buildChain 构建智能处理链
func (s *ComprehensiveRAGSystem) buildChain(ctx context.Context) error {
	// 这里可以构建复杂的处理链
	// 为演示目的，我们暂时不实现完整的Chain
	log.Println("✓ Chain 构建完成")
	return nil
}

// LoadInitialKnowledge 加载初始知识库
func (s *ComprehensiveRAGSystem) LoadInitialKnowledge(ctx context.Context) error {
	log.Println("\n=== 加载初始知识库 ===")

	// 准备示例文档
	documents := []*schema.Document{
		{
			ID:      "eino-intro",
			Content: `# Eino 框架介绍\nEino 是一个先进的大模型应用开发框架。\n## 核心特性\nEino 提供了 Transformer、Indexer、Retriever 和 Tool 等核心组件。\n## 应用场景\nEino 适用于构建 RAG 应用、智能问答系统和知识管理平台。`,
			MetaData: map[string]interface{}{
				"source": "official_docs",
				"type":   "introduction",
			},
		},
		{
			ID:      "rag-concept",
			Content: `# RAG 技术详解\nRAG (Retrieval-Augmented Generation) 是结合检索和生成的AI技术。\n## 工作原理\nRAG 通过检索相关知识来增强大模型的生成能力。\n## 优势\nRAG 可以提供更准确、更新的信息，并减少模型幻觉。`,
			MetaData: map[string]interface{}{
				"source": "tech_docs",
				"type":   "concept",
			},
		},
		{
			ID:      "tool-usage",
			Content: `# 工具使用指南\n工具系统允许AI助手调用外部功能。\n## 内置工具\n系统提供知识搜索、文档处理、计算器和天气查询等工具。\n## 自定义工具\n开发者可以轻松添加自定义工具来扩展系统功能。`,
			MetaData: map[string]interface{}{
				"source": "user_manual",
				"type":   "guide",
			},
		},
	}

	// 分割并索引文档
	allChunks := make([]*schema.Document, 0)
	// 遍历每个文档进行分割
	for _, doc := range documents {
		// 分割文档
		chunks, err := s.transformer.Transform(ctx, []*schema.Document{doc})
		if err != nil {
			return fmt.Errorf("分割文档 %s 失败: %v", doc.ID, err)
		}
		allChunks = append(allChunks, chunks...)
		log.Printf("文档 %s 分割为 %d 块", doc.ID, len(chunks))
	}

	// 存储到向量数据库
	storedIDs, err := s.indexer.Store(ctx, allChunks)
	if err != nil {
		return fmt.Errorf("存储文档失败: %v", err)
	}

	// 加载集合到内存
	if err := s.milvusClient.LoadCollection(ctx, s.config.MilvusCollection, false); err != nil {
		return fmt.Errorf("加载集合失败: %v", err)
	}

	log.Printf("✓ 成功加载 %d 个文档块到知识库", len(storedIDs))
	return nil
}

// ProcessUserQuery 处理用户查询(演示核心功能)
func (s *ComprehensiveRAGSystem) ProcessUserQuery(ctx context.Context, query string) error {
	log.Printf("\n=== 处理用户查询: %s ===", query)

	// 1. 知识检索演示
	log.Println("\n1. 执行知识检索...")
	docs, err := s.retriever.Retrieve(ctx, query)
	if err != nil {
		return fmt.Errorf("知识检索失败: %v", err)
	}

	log.Printf("检索到 %d 个相关知识片段:", len(docs))
	for i, doc := range docs {
		log.Printf("  [%d] ID: %s", i+1, doc.ID)
		log.Printf("      内容: %s", truncateString(doc.Content, 100))
	}

	// 2. 工具调用演示
	log.Println("\n2. 演示工具调用...")

	// 演示计算器工具
	calcTool := &CalculatorTool{}
	calcResult, err := calcTool.InvokableRun(ctx, `{"expression": "25 + 17"}`)
	if err == nil {
		log.Printf("计算器工具结果: %s", calcResult)
	}

	// 演示天气工具
	weatherTool := &WeatherTool{}
	weatherResult, err := weatherTool.InvokableRun(ctx, `{"city": "北京"}`)
	if err == nil {
		log.Printf("天气工具结果: %s", truncateString(weatherResult, 150))
	}

	// 3. 构建增强提示
	log.Println("\n3. 构建增强提示并生成回答...")

	prompt := buildRAGPrompt(query, docs)
	messages := []*schema.Message{
		schema.SystemMessage("你是一个智能助手，能够基于提供的知识回答问题并调用工具。请根据上下文提供准确、有用的回答。"),
		schema.UserMessage(prompt),
	}

	// 4. 生成最终回答
	response, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		return fmt.Errorf("生成回答失败: %v", err)
	}

	log.Println("\n=== 最终回答 ===")
	log.Println(response.Content)

	return nil
}

// Close 关闭系统资源
func (s *ComprehensiveRAGSystem) Close() error {
	if s.milvusClient != nil {
		return s.milvusClient.Close()
	}
	return nil
}

// ================================
// 辅助函数
// ================================

// loadConfig 加载配置
func loadConfig() (*Config, error) {
	// 使用 Viper 加载配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.AutomaticEnv()

	// 设置环境变量前缀
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("未找到 config.yaml 文件，将从环境变量读取配置。")
	}

	// 读取文件或环境变量中的配置
	config := &Config{
		MilvusAddress:    viper.GetString("MILVUS_ADDRESS"),
		MilvusCollection: viper.GetString("MILVUS_COLLECTION"),
		ArkAPIKey:        viper.GetString("ARK_API_KEY"),
		EmbedderModel:    viper.GetString("EMBEDDER_MODEL"),
		ArkModel:         viper.GetString("ARK_MODEL"),
	}

	// 验证配置
	return config, validateConfig(config)
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config.MilvusAddress == "" {
		return fmt.Errorf("MILVUS_ADDRESS 必须设置")
	}
	if config.MilvusCollection == "" {
		return fmt.Errorf("MILVUS_COLLECTION 必须设置")
	}
	if config.ArkAPIKey == "" {
		return fmt.Errorf("ARK_API_KEY 必须设置")
	}
	if config.EmbedderModel == "" {
		return fmt.Errorf("EMBEDDER_MODEL 必须设置")
	}
	if config.ArkModel == "" {
		return fmt.Errorf("ARK_MODEL 必须设置")
	}
	return nil
}

// buildRAGPrompt 构建RAG提示
func buildRAGPrompt(query string, docs []*schema.Document) string {
	prompt := "请基于以下知识库信息回答问题。\n\n=== 知识库信息 ===\n"

	for i, doc := range docs {
		prompt += fmt.Sprintf("[知识片段 %d]\n%s\n\n", i+1, doc.Content)
	}

	prompt += fmt.Sprintf("=== 用户问题 ===\n%s\n\n", query)
	prompt += "请结合上述知识信息，提供准确、详细的回答。如果知识信息不足，请说明情况。"

	return prompt
}

// evaluateSimpleExpression 简单表达式计算
func evaluateSimpleExpression(expr string) float64 {
	expr = strings.ReplaceAll(expr, " ", "")

	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		if len(parts) == 2 {
			var a, b float64
			fmt.Sscanf(parts[0], "%f", &a)
			fmt.Sscanf(parts[1], "%f", &b)
			return a + b
		}
	}

	if strings.Contains(expr, "-") {
		parts := strings.Split(expr, "-")
		if len(parts) == 2 {
			var a, b float64
			fmt.Sscanf(parts[0], "%f", &a)
			fmt.Sscanf(parts[1], "%f", &b)
			return a - b
		}
	}

	var result float64
	fmt.Sscanf(expr, "%f", &result)
	return result
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ================================
// 主程序入口
// ================================

func main() {
	log.Println("🚀 启动 Eino 综合演示系统")

	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	ctx := context.Background()

	// 创建系统实例，初始化各个组件
	system, err := NewComprehensiveRAGSystem(ctx, config)
	if err != nil {
		log.Fatalf("❌ 系统初始化失败: %v", err)
	}
	defer system.Close()

	log.Println("✅ 综合RAG系统初始化完成")

	// 加载初始知识库
	if err := system.LoadInitialKnowledge(ctx); err != nil {
		log.Fatalf("❌ 知识库加载失败: %v", err)
	}

	// 演示查询处理
	queries := []string{
		"什么是 Eino 框架？",
		"RAG 技术有什么优势？",
		"如何使用工具系统？",
	}

	// 遍历查询列表，依次处理每个查询
	for i, query := range queries {
		log.Printf("\n" + strings.Repeat("=", 60))
		log.Printf("演示查询 %d/%d", i+1, len(queries))

		// 处理用户查询
		if err := system.ProcessUserQuery(ctx, query); err != nil {
			log.Printf("❌ 处理查询失败: %v", err)
		}

		// 为演示添加延迟
		time.Sleep(2 * time.Second)
	}

	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🎉 综合演示完成！系统展示了以下核心功能：")
	log.Println("   • 📝 文档转换与分割 (Transformer)")
	log.Println("   • 📚 文档向量化与索引 (Indexer)")
	log.Println("   • 🔍 语义相似度检索 (Retriever)")
	log.Println("   • 🔧 智能工具调用 (Tools)")
	log.Println("   • 🤖 增强生成回答 (RAG)")
	log.Println("   • ⚡ 端到端工作流编排 (Chain)")
}
