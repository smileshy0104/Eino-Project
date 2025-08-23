package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	retriever "github.com/cloudwego/eino-ext/components/retriever/milvus"
)

// KnowledgeSearchTool 知识搜索工具 - 从向量数据库检索相关知识
type KnowledgeSearchTool struct {
	retriever *retriever.Retriever
}

// Info 返回知识搜索工具的信息
func (k *KnowledgeSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "knowledge_search",
		Desc: "从知识库中搜索相关文档信息",
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

	// 设置默认TopK
	if args.TopK == 0 {
		args.TopK = 3
	}

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
			"title": {
				Type:     "string",
				Desc:     "文档标题",
				Required: false,
			},
			"author": {
				Type:     "string",
				Desc:     "文档作者",
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
		Title    string                 `json:"title"`
		Author   string                 `json:"author"`
		MetaData map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 生成默认值
	if args.DocID == "" {
		args.DocID = fmt.Sprintf("doc_%d", time.Now().Unix())
	}
	if args.Title == "" {
		args.Title = "未命名文档"
	}

	// 初始化元数据
	if args.MetaData == nil {
		args.MetaData = make(map[string]interface{})
	}
	args.MetaData["title"] = args.Title
	args.MetaData["author"] = args.Author
	args.MetaData["processed_at"] = time.Now().Format(time.RFC3339)

	// 创建原始文档
	originalDoc := &schema.Document{
		ID:       args.DocID,
		Content:  args.Content,
		MetaData: args.MetaData,
	}

	// 使用Transformer分割文档
	chunks, err := d.transformer.Transform(ctx, []*schema.Document{originalDoc})
	if err != nil {
		return "", fmt.Errorf("文档分割失败: %v", err)
	}

	// 使用Indexer存储文档块
	storedIDs, err := d.indexer.Store(ctx, chunks)
	if err != nil {
		return "", fmt.Errorf("文档索引失败: %v", err)
	}

	result := map[string]interface{}{
		"original_doc_id": args.DocID,
		"title":           args.Title,
		"chunks_count":    len(chunks),
		"stored_ids":      storedIDs,
		"status":          "success",
		"message":         fmt.Sprintf("成功处理文档《%s》，分割为%d个块并完成索引", args.Title, len(chunks)),
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

// WeatherTool 天气查询工具（模拟）
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

// DocumentSummaryTool 文档摘要工具
type DocumentSummaryTool struct {
	retriever *retriever.Retriever
}

// Info 返回文档摘要工具信息
func (ds *DocumentSummaryTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "document_summary",
		Desc: "根据文档ID或关键词生成文档摘要",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"doc_id": {
				Type:     "string",
				Desc:     "文档ID",
				Required: false,
			},
			"keyword": {
				Type:     "string",
				Desc:     "搜索关键词",
				Required: false,
			},
			"max_length": {
				Type:     "integer",
				Desc:     "摘要最大长度",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行文档摘要
func (ds *DocumentSummaryTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	var args struct {
		DocID     string `json:"doc_id"`
		Keyword   string `json:"keyword"`
		MaxLength int    `json:"max_length"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.MaxLength == 0 {
		args.MaxLength = 200
	}

	// 根据输入检索文档
	var searchQuery string
	if args.DocID != "" {
		searchQuery = args.DocID
	} else if args.Keyword != "" {
		searchQuery = args.Keyword
	} else {
		return "", fmt.Errorf("必须提供doc_id或keyword参数")
	}

	docs, err := ds.retriever.Retrieve(ctx, searchQuery)
	if err != nil {
		return "", fmt.Errorf("文档检索失败: %v", err)
	}

	if len(docs) == 0 {
		return `{"summary": "未找到相关文档", "doc_count": 0}`, nil
	}

	// 生成简单摘要（实际应用中可能需要更复杂的摘要算法）
	summary := generateDocumentSummary(docs, args.MaxLength)

	result := map[string]interface{}{
		"summary":     summary,
		"doc_count":   len(docs),
		"search_term": searchQuery,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// 辅助函数

// evaluateSimpleExpression 简单表达式计算
func evaluateSimpleExpression(expr string) float64 {
	// 这里实现简单的表达式计算
	// 实际项目中可以使用更强大的表达式解析库
	var result float64
	if n, err := fmt.Sscanf(expr, "%f + %f", &result, new(float64)); n == 2 && err == nil {
		var a, b float64
		fmt.Sscanf(expr, "%f + %f", &a, &b)
		return a + b
	}
	if n, err := fmt.Sscanf(expr, "%f - %f", &result, new(float64)); n == 2 && err == nil {
		var a, b float64
		fmt.Sscanf(expr, "%f - %f", &a, &b)
		return a - b
	}
	if n, err := fmt.Sscanf(expr, "%f * %f", &result, new(float64)); n == 2 && err == nil {
		var a, b float64
		fmt.Sscanf(expr, "%f * %f", &a, &b)
		return a * b
	}
	if n, err := fmt.Sscanf(expr, "%f / %f", &result, new(float64)); n == 2 && err == nil {
		var a, b float64
		fmt.Sscanf(expr, "%f / %f", &a, &b)
		if b != 0 {
			return a / b
		}
	}
	
	// 如果不是表达式，尝试解析为单个数字
	fmt.Sscanf(expr, "%f", &result)
	return result
}

// generateDocumentSummary 生成文档摘要
func generateDocumentSummary(docs []*schema.Document, maxLength int) string {
	if len(docs) == 0 {
		return "无相关文档内容"
	}

	// 简单的摘要生成：取前几个文档的内容片段
	summary := ""
	for i, doc := range docs {
		if i >= 3 { // 最多使用前3个文档
			break
		}

		content := doc.Content
		if len(content) > maxLength/3 { // 每个文档最多占用总长度的1/3
			content = content[:maxLength/3] + "..."
		}

		if doc.MetaData != nil {
			if title, ok := doc.MetaData["title"].(string); ok && title != "" {
				summary += fmt.Sprintf("《%s》: %s ", title, content)
			} else {
				summary += content + " "
			}
		} else {
			summary += content + " "
		}
	}

	if len(summary) > maxLength {
		summary = summary[:maxLength] + "..."
	}

	return summary
}