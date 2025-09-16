// Package main 演示 Eino 框架中 DocumentLoader 组件的各种用法
// DocumentLoader 是用于从多种数据源加载文档的核心组件
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// MockFileSource 模拟文件数据源
type MockFileSource struct {
	FilePath string
	MetaData map[string]interface{}
}

func (fs *MockFileSource) GetID() string {
	return fs.FilePath
}

func (fs *MockFileSource) GetReader(ctx context.Context) (io.Reader, error) {
	file, err := os.Open(fs.FilePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件 %s: %w", fs.FilePath, err)
	}
	return file, nil
}

func (fs *MockFileSource) GetMetadata() map[string]interface{} {
	if fs.MetaData == nil {
		fs.MetaData = make(map[string]interface{})
	}
	fs.MetaData["source_type"] = "file"
	fs.MetaData["file_path"] = fs.FilePath
	return fs.MetaData
}

// MockURLSource 模拟网络资源数据源
type MockURLSource struct {
	URL      string
	Content  string
	MetaData map[string]interface{}
}

func (us *MockURLSource) GetID() string {
	return us.URL
}

func (us *MockURLSource) GetReader(ctx context.Context) (io.Reader, error) {
	// 模拟网络请求延迟
	time.Sleep(100 * time.Millisecond)
	return strings.NewReader(us.Content), nil
}

func (us *MockURLSource) GetMetadata() map[string]interface{} {
	if us.MetaData == nil {
		us.MetaData = make(map[string]interface{})
	}
	us.MetaData["source_type"] = "url"
	us.MetaData["url"] = us.URL
	us.MetaData["content_length"] = len(us.Content)
	return us.MetaData
}

// MockTextParser 模拟文本解析器
type MockTextParser struct{}

func (p *MockTextParser) Parse(ctx context.Context, reader io.Reader) ([]*schema.Document, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取内容失败: %w", err)
	}

	doc := &schema.Document{
		ID:      fmt.Sprintf("doc_%d", time.Now().UnixNano()),
		Content: string(content),
		MetaData: map[string]interface{}{
			"parser_type":  "text",
			"content_size": len(content),
			"parsed_at":    time.Now().Format(time.RFC3339),
		},
	}

	return []*schema.Document{doc}, nil
}

// MockHTMLParser 模拟HTML解析器
type MockHTMLParser struct{}

func (p *MockHTMLParser) Parse(ctx context.Context, reader io.Reader) ([]*schema.Document, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取HTML内容失败: %w", err)
	}

	htmlContent := string(content)

	// 简单的HTML标签去除（实际项目中会使用专业的HTML解析器）
	cleanContent := strings.ReplaceAll(htmlContent, "<", " <")
	cleanContent = strings.ReplaceAll(cleanContent, ">", "> ")

	// 提取标题
	var title string
	if titleStart := strings.Index(htmlContent, "<title>"); titleStart != -1 {
		titleEnd := strings.Index(htmlContent[titleStart:], "</title>")
		if titleEnd != -1 {
			title = htmlContent[titleStart+7 : titleStart+titleEnd]
		}
	}

	doc := &schema.Document{
		ID:      fmt.Sprintf("html_doc_%d", time.Now().UnixNano()),
		Content: cleanContent,
		MetaData: map[string]interface{}{
			"parser_type":     "html",
			"content_size":    len(content),
			"extracted_title": title,
			"parsed_at":       time.Now().Format(time.RFC3339),
			"original_format": "html",
		},
	}

	return []*schema.Document{doc}, nil
}

// MockExtParser 模拟扩展解析器（根据文件扩展名选择解析器）
type MockExtParser struct {
	parsers        map[string]Parser
	fallbackParser Parser
}

type Parser interface {
	Parse(ctx context.Context, reader io.Reader) ([]*schema.Document, error)
}

func NewMockExtParser() *MockExtParser {
	return &MockExtParser{
		parsers: map[string]Parser{
			".txt":  &MockTextParser{},
			".md":   &MockTextParser{},
			".html": &MockHTMLParser{},
			".htm":  &MockHTMLParser{},
		},
		fallbackParser: &MockTextParser{},
	}
}

func (p *MockExtParser) Parse(ctx context.Context, reader io.Reader, sourceID string) ([]*schema.Document, error) {
	ext := strings.ToLower(filepath.Ext(sourceID))

	parser, exists := p.parsers[ext]
	if !exists {
		fmt.Printf("  未找到扩展名 %s 对应的解析器，使用默认解析器\n", ext)
		parser = p.fallbackParser
	} else {
		fmt.Printf("  使用 %s 解析器处理文件\n", ext)
	}

	return parser.Parse(ctx, reader)
}

// MockDocumentLoader 模拟文档加载器
type MockDocumentLoader struct {
	parser *MockExtParser
}

func NewMockDocumentLoader() *MockDocumentLoader {
	return &MockDocumentLoader{
		parser: NewMockExtParser(),
	}
}

func (dl *MockDocumentLoader) Load(ctx context.Context, source Source) ([]*schema.Document, error) {
	fmt.Printf("📥 开始加载文档: %s\n", source.GetID())

	reader, err := source.GetReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取数据源读取器失败: %w", err)
	}

	// 如果是文件读取器，确保关闭
	if closer, ok := reader.(io.Closer); ok {
		defer closer.Close()
	}

	docs, err := dl.parser.Parse(ctx, reader, source.GetID())
	if err != nil {
		return nil, fmt.Errorf("文档解析失败: %w", err)
	}

	// 添加数据源元数据到文档中
	sourceMetadata := source.GetMetadata()
	for _, doc := range docs {
		if doc.MetaData == nil {
			doc.MetaData = make(map[string]interface{})
		}
		for key, value := range sourceMetadata {
			doc.MetaData[key] = value
		}
		doc.MetaData["loaded_at"] = time.Now().Format(time.RFC3339)
	}

	fmt.Printf("✅ 成功加载 %d 个文档\n", len(docs))
	return docs, nil
}

// Source 接口定义
type Source interface {
	GetID() string
	GetReader(ctx context.Context) (io.Reader, error)
	GetMetadata() map[string]interface{}
}

func main() {
	fmt.Println("=== Eino DocumentLoader 组件演示 ===")

	ctx := context.Background()

	// 获取命令行参数决定运行哪个示例
	if len(os.Args) > 1 {
		example := os.Args[1]
		switch example {
		case "basic":
			fmt.Println("运行基础文档加载演示...")
			basicLoaderDemo(ctx)
		case "multi":
			fmt.Println("运行多文件加载演示...")
			multiFileLoaderDemo(ctx)
		case "url":
			fmt.Println("运行网络资源加载演示...")
			urlLoaderDemo(ctx)
		case "chain":
			fmt.Println("运行链式集成演示...")
			chainIntegrationDemo(ctx)
		case "advanced":
			fmt.Println("运行高级用法演示...")
			advancedUsageDemo(ctx)
		default:
			fmt.Printf("未知示例: %s\n", example)
			showUsage()
		}
		return
	}

	// 默认运行所有演示
	runAllDemos(ctx)
}

func runAllDemos(ctx context.Context) {
	fmt.Println("📝 演示1: 基础文档加载")
	basicLoaderDemo(ctx)

	//fmt.Println("\n📚 演示2: 多文件批量加载")
	//multiFileLoaderDemo(ctx)
	//
	//fmt.Println("\n🌐 演示3: 网络资源加载")
	//urlLoaderDemo(ctx)
	//
	//fmt.Println("\n🔗 演示4: 链式集成")
	//chainIntegrationDemo(ctx)
	//
	//fmt.Println("\n🚀 演示5: 高级用法")
	//advancedUsageDemo(ctx)
	//
	//fmt.Println("\n✅ 所有 DocumentLoader 演示完成！")
}

// getTestDocumentPath 获取测试文档的绝对路径
func getTestDocumentPath(filename string) string {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		// 如果获取失败，使用相对路径
		return filepath.Join("test_documents", filename)
	}

	// 检查当前目录是否有 test_documents 文件夹
	testDocsPath := filepath.Join(wd, "test_documents", filename)
	if _, err := os.Stat(testDocsPath); err == nil {
		return testDocsPath
	}

	// 如果当前目录没有，检查是否在上级目录运行，需要进入 document_loader_demo 子目录
	demoTestDocsPath := filepath.Join(wd, "document_loader_demo", "test_documents", filename)
	if _, err := os.Stat(demoTestDocsPath); err == nil {
		return demoTestDocsPath
	}

	// 如果都找不到，返回相对路径（可能会失败，但提供更清晰的错误信息）
	return filepath.Join("test_documents", filename)
}

// 1. 基础文档加载演示
func basicLoaderDemo(ctx context.Context) {
	loader := NewMockDocumentLoader()

	// 加载文本文件
	fmt.Println("\n  📄 加载纯文本文件:")
	fileSource := &MockFileSource{
		FilePath: getTestDocumentPath("sample.txt"),
		MetaData: map[string]interface{}{
			"category": "sample",
			"priority": "normal",
		},
	}

	docs, err := loader.Load(ctx, fileSource)
	if err != nil {
		log.Printf("加载失败: %v", err)
		return
	}

	// 显示加载结果
	for _, doc := range docs {
		fmt.Printf("  文档ID: %s\n", doc.ID)
		fmt.Printf("  内容长度: %d 字符\n", len(doc.Content))
		fmt.Printf("  内容预览: %s...\n", truncateString(doc.Content, 100))
		fmt.Printf("  元数据: %v\n", doc.MetaData)
	}
}

// 2. 多文件批量加载演示
func multiFileLoaderDemo(ctx context.Context) {
	loader := NewMockDocumentLoader()

	// 定义多个文件源
	filePaths := []string{
		getTestDocumentPath("sample.txt"),
		getTestDocumentPath("sample.md"),
		getTestDocumentPath("sample.html"),
	}

	var allDocuments []*schema.Document

	fmt.Println("\n  📁 批量加载多个文件:")
	for i, filePath := range filePaths {
		fmt.Printf("\n  处理文件 %d/%d: %s\n", i+1, len(filePaths), filepath.Base(filePath))

		fileSource := &MockFileSource{
			FilePath: filePath,
			MetaData: map[string]interface{}{
				"batch_id":   "batch_001",
				"file_index": i + 1,
			},
		}

		docs, err := loader.Load(ctx, fileSource)
		if err != nil {
			fmt.Printf("  ❌ 文件 %s 加载失败: %v\n", filePath, err)
			continue
		}

		allDocuments = append(allDocuments, docs...)

		// 显示每个文件的处理结果
		for _, doc := range docs {
			fmt.Printf("    - 文档ID: %s, 长度: %d 字符\n", doc.ID, len(doc.Content))
		}
	}

	fmt.Printf("\n  📊 批量加载总结:\n")
	fmt.Printf("    总文件数: %d\n", len(filePaths))
	fmt.Printf("    成功加载的文档数: %d\n", len(allDocuments))

	// 统计不同类型的文档
	typeCount := make(map[string]int)
	for _, doc := range allDocuments {
		if parserType, ok := doc.MetaData["parser_type"].(string); ok {
			typeCount[parserType]++
		}
	}

	fmt.Printf("    文档类型分布:\n")
	for docType, count := range typeCount {
		fmt.Printf("      %s: %d 个\n", docType, count)
	}
}

// 3. 网络资源加载演示
func urlLoaderDemo(ctx context.Context) {
	loader := NewMockDocumentLoader()

	fmt.Println("\n  🌐 加载网络资源:")

	// 模拟不同的网络资源
	urlSources := []*MockURLSource{
		{
			URL: "https://example.com/article1.html",
			Content: `<!DOCTYPE html>
<html><head><title>示例文章1</title></head>
<body><h1>技术文档</h1><p>这是一篇关于AI技术发展的文章。</p></body></html>`,
			MetaData: map[string]interface{}{
				"domain":       "example.com",
				"fetch_time":   time.Now().Format(time.RFC3339),
				"content_type": "text/html",
			},
		},
		{
			URL: "https://api.example.com/docs/guide.txt",
			Content: `API 使用指南

本指南介绍如何使用我们的API服务。

1. 身份验证
2. 请求格式
3. 响应处理
4. 错误码说明

更多详细信息请参考官方文档。`,
			MetaData: map[string]interface{}{
				"domain":       "api.example.com",
				"fetch_time":   time.Now().Format(time.RFC3339),
				"content_type": "text/plain",
			},
		},
	}

	var allDocs []*schema.Document

	for i, urlSource := range urlSources {
		fmt.Printf("\n  处理 URL %d/%d: %s\n", i+1, len(urlSources), urlSource.URL)

		docs, err := loader.Load(ctx, urlSource)
		if err != nil {
			fmt.Printf("  ❌ URL %s 加载失败: %v\n", urlSource.URL, err)
			continue
		}

		allDocs = append(allDocs, docs...)

		// 显示网络资源的处理结果
		for _, doc := range docs {
			fmt.Printf("    - 文档ID: %s\n", doc.ID)
			fmt.Printf("    - 来源: %s\n", doc.MetaData["url"])
			fmt.Printf("    - 内容长度: %d 字符\n", len(doc.Content))
			fmt.Printf("    - 内容预览: %s...\n", truncateString(doc.Content, 80))
		}
	}

	fmt.Printf("\n  🌐 网络资源加载总结: 成功加载 %d 个文档\n", len(allDocs))
}

// 4. 链式集成演示
func chainIntegrationDemo(ctx context.Context) {
	fmt.Println("\n  🔗 DocumentLoader 链式集成演示")

	// 创建一个模拟的文档处理链
	// Source -> Documents -> Processed Documents

	// 步骤1: 创建文档加载器
	loader := NewMockDocumentLoader()

	// 步骤2: 创建文档后处理器 (模拟)
	processor := &MockDocumentProcessor{}

	// 步骤3: 创建简化的处理链
	pipeline := &DocumentProcessingPipeline{
		loader:    loader,
		processor: processor,
	}

	// 测试数据源
	sources := []Source{
		&MockFileSource{
			FilePath: getTestDocumentPath("sample.txt"),
			MetaData: map[string]interface{}{"category": "text"},
		},
		&MockURLSource{
			URL: "https://example.com/news.html",
			Content: `<html><head><title>新闻</title></head>
<body><h1>今日新闻</h1><p>重要新闻内容...</p></body></html>`,
			MetaData: map[string]interface{}{"category": "news"},
		},
	}

	fmt.Printf("\n  处理 %d 个数据源:\n", len(sources))

	var allResults []ProcessingResult

	for i, source := range sources {
		fmt.Printf("\n  处理数据源 %d: %s\n", i+1, source.GetID())

		result, err := pipeline.Process(ctx, source)
		if err != nil {
			fmt.Printf("    ❌ 处理失败: %v\n", err)
			continue
		}

		allResults = append(allResults, result)

		fmt.Printf("    ✅ 处理成功:\n")
		fmt.Printf("      原始文档数: %d\n", len(result.OriginalDocs))
		fmt.Printf("      处理后文档数: %d\n", len(result.ProcessedDocs))
		fmt.Printf("      处理耗时: %v\n", result.ProcessingTime)
	}

	fmt.Printf("\n  🔗 链式处理总结: 处理了 %d 个数据源\n", len(allResults))
}

// 5. 高级用法演示
func advancedUsageDemo(ctx context.Context) {
	fmt.Println("\n  🚀 高级用法演示")

	// 演示1: 带有自定义元数据的加载
	fmt.Println("\n    📋 自定义元数据处理:")
	advancedMetadataDemo(ctx)

	// 演示2: 错误处理和重试机制
	fmt.Println("\n    🔄 错误处理和重试机制:")
	errorHandlingDemo(ctx)

	// 演示3: 性能监控
	fmt.Println("\n    📊 性能监控演示:")
	performanceMonitoringDemo(ctx)
}

func advancedMetadataDemo(ctx context.Context) {
	loader := NewMockDocumentLoader()

	// 创建带有丰富元数据的数据源
	source := &MockFileSource{
		FilePath: getTestDocumentPath("sample.md"),
		MetaData: map[string]interface{}{
			"author":        "技术团队",
			"version":       "1.0",
			"tags":          []string{"documentation", "guide", "api"},
			"created_at":    time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"department":    "研发部",
			"priority":      "high",
			"review_status": "approved",
		},
	}

	docs, err := loader.Load(ctx, source)
	if err != nil {
		fmt.Printf("    ❌ 加载失败: %v\n", err)
		return
	}

	// 显示元数据处理结果
	for _, doc := range docs {
		fmt.Printf("    📄 文档: %s\n", doc.ID)
		fmt.Printf("    🏷️  元数据字段数: %d\n", len(doc.MetaData))

		// 显示关键元数据
		if author, ok := doc.MetaData["author"]; ok {
			fmt.Printf("      作者: %v\n", author)
		}
		if tags, ok := doc.MetaData["tags"]; ok {
			fmt.Printf("      标签: %v\n", tags)
		}
		if priority, ok := doc.MetaData["priority"]; ok {
			fmt.Printf("      优先级: %v\n", priority)
		}
	}
}

func errorHandlingDemo(ctx context.Context) {
	loader := NewMockDocumentLoader()

	// 测试不存在的文件
	invalidSource := &MockFileSource{
		FilePath: "./test_documents/nonexistent.txt",
		MetaData: map[string]interface{}{
			"expected_result": "error",
		},
	}

	fmt.Printf("    🔍 测试不存在的文件: %s\n", invalidSource.FilePath)

	docs, err := loader.Load(ctx, invalidSource)
	if err != nil {
		fmt.Printf("    ✅ 预期的错误: %v\n", err)
	} else {
		fmt.Printf("    ❌ 意外成功，加载了 %d 个文档\n", len(docs))
	}

	// 测试空内容处理
	emptySource := &MockURLSource{
		URL:     "https://example.com/empty.txt",
		Content: "",
		MetaData: map[string]interface{}{
			"expected_result": "empty",
		},
	}

	fmt.Printf("    🔍 测试空内容: %s\n", emptySource.URL)

	docs, err = loader.Load(ctx, emptySource)
	if err != nil {
		fmt.Printf("    ❌ 空内容处理失败: %v\n", err)
	} else {
		fmt.Printf("    ✅ 空内容处理成功，生成 %d 个文档\n", len(docs))
		if len(docs) > 0 {
			fmt.Printf("      内容长度: %d 字符\n", len(docs[0].Content))
		}
	}
}

func performanceMonitoringDemo(ctx context.Context) {
	loader := NewMockDocumentLoader()

	// 创建性能监控器
	monitor := &PerformanceMonitor{}

	// 测试不同大小的文档加载性能
	testSizes := []int{100, 1000, 5000, 10000}

	fmt.Printf("    📈 性能测试 (不同文档大小):\n")

	for _, size := range testSizes {
		// 生成指定大小的测试内容
		content := strings.Repeat("测试内容 ", size/4)

		source := &MockURLSource{
			URL:     fmt.Sprintf("https://example.com/test_%d.txt", size),
			Content: content,
			MetaData: map[string]interface{}{
				"test_size": size,
			},
		}

		// 测量加载时间
		startTime := time.Now()

		docs, err := loader.Load(ctx, source)

		duration := time.Since(startTime)

		monitor.RecordLoad(len(docs), len(content), duration, err == nil)

		if err != nil {
			fmt.Printf("      ❌ 大小 %d: 失败 - %v\n", size, err)
		} else {
			fmt.Printf("      ✅ 大小 %d: 成功 - %v (%d docs)\n", size, duration, len(docs))
		}
	}

	// 显示性能统计
	stats := monitor.GetStats()
	fmt.Printf("    📊 性能统计:\n")
	fmt.Printf("      总加载次数: %d\n", stats["total_loads"])
	fmt.Printf("      成功率: %.2f%%\n", stats["success_rate"].(float64)*100)
	fmt.Printf("      平均加载时间: %v\n", stats["average_time"])
	fmt.Printf("      平均文档大小: %.0f 字符\n", stats["average_size"])
}

// 辅助结构和函数

// MockDocumentProcessor 模拟文档后处理器
type MockDocumentProcessor struct{}

func (p *MockDocumentProcessor) Process(ctx context.Context, docs []*schema.Document) ([]*schema.Document, error) {
	var processedDocs []*schema.Document

	for _, doc := range docs {
		// 简单的后处理：添加处理标记、清理内容等
		processedDoc := &schema.Document{
			ID:       doc.ID + "_processed",
			Content:  strings.TrimSpace(doc.Content),
			MetaData: make(map[string]interface{}),
		}

		// 复制原始元数据
		for k, v := range doc.MetaData {
			processedDoc.MetaData[k] = v
		}

		// 添加处理信息
		processedDoc.MetaData["processed"] = true
		processedDoc.MetaData["processed_at"] = time.Now().Format(time.RFC3339)
		processedDoc.MetaData["word_count"] = len(strings.Fields(doc.Content))

		processedDocs = append(processedDocs, processedDoc)
	}

	return processedDocs, nil
}

// DocumentProcessingPipeline 文档处理流水线
type DocumentProcessingPipeline struct {
	loader    *MockDocumentLoader
	processor *MockDocumentProcessor
}

type ProcessingResult struct {
	OriginalDocs   []*schema.Document
	ProcessedDocs  []*schema.Document
	ProcessingTime time.Duration
}

func (p *DocumentProcessingPipeline) Process(ctx context.Context, source Source) (ProcessingResult, error) {
	startTime := time.Now()

	// 步骤1: 加载文档
	originalDocs, err := p.loader.Load(ctx, source)
	if err != nil {
		return ProcessingResult{}, fmt.Errorf("文档加载失败: %w", err)
	}

	// 步骤2: 处理文档
	processedDocs, err := p.processor.Process(ctx, originalDocs)
	if err != nil {
		return ProcessingResult{}, fmt.Errorf("文档处理失败: %w", err)
	}

	processingTime := time.Since(startTime)

	return ProcessingResult{
		OriginalDocs:   originalDocs,
		ProcessedDocs:  processedDocs,
		ProcessingTime: processingTime,
	}, nil
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	totalLoads   int
	successCount int
	totalTime    time.Duration
	totalSize    int
}

func (pm *PerformanceMonitor) RecordLoad(docCount, contentSize int, duration time.Duration, success bool) {
	pm.totalLoads++
	pm.totalTime += duration
	pm.totalSize += contentSize

	if success {
		pm.successCount++
	}
}

func (pm *PerformanceMonitor) GetStats() map[string]interface{} {
	var averageTime time.Duration
	var averageSize float64
	var successRate float64

	if pm.totalLoads > 0 {
		averageTime = pm.totalTime / time.Duration(pm.totalLoads)
		averageSize = float64(pm.totalSize) / float64(pm.totalLoads)
		successRate = float64(pm.successCount) / float64(pm.totalLoads)
	}

	return map[string]interface{}{
		"total_loads":   pm.totalLoads,
		"success_count": pm.successCount,
		"success_rate":  successRate,
		"average_time":  averageTime,
		"average_size":  averageSize,
	}
}

// 辅助函数
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

func showUsage() {
	fmt.Println("用法: go run main.go [example]")
	fmt.Println("示例:")
	fmt.Println("  basic      - 基础文档加载演示")
	fmt.Println("  multi      - 多文件批量加载演示")
	fmt.Println("  url        - 网络资源加载演示")
	fmt.Println("  chain      - 链式集成演示")
	fmt.Println("  advanced   - 高级用法演示")
	fmt.Println("\n不带参数运行所有演示")
}
