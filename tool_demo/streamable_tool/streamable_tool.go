package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// =============================================================================
//
//  文件: streamable_tool.go
//  功能: 演示如何实现支持流式输出的 StreamableTool
//  说明: 流式工具适用于需要长时间处理或逐步产生结果的场景
//
// =============================================================================

// --- 示例 1: 流式文本生成工具 ---

type StreamTextGeneratorTool struct{}

func (s *StreamTextGeneratorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "stream_text_generator",
		Desc: "逐步生成文本内容，支持流式输出",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"topic": {
				Type:     "string",
				Desc:     "生成文本的主题",
				Required: true,
			},
			"length": {
				Type:     "integer",
				Desc:     "生成文本的段落数量",
				Required: false,
			},
			"delay_ms": {
				Type:     "integer",
				Desc:     "每段之间的延迟毫秒数",
				Required: false,
			},
		}),
	}, nil
}

func (s *StreamTextGeneratorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 对于流式工具，InvokableRun 通常返回完整结果
	reader, err := s.StreamableRun(ctx, argumentsInJSON, opts...)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	
	var result strings.Builder
	for {
		chunk, err := reader.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
		result.WriteString(chunk)
	}
	
	return result.String(), nil
}

func (s *StreamTextGeneratorTool) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (*schema.StreamReader[string], error) {
	var args struct {
		Topic   string `json:"topic"`
		Length  int    `json:"length"`
		DelayMs int    `json:"delay_ms"`
	}
	
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return nil, fmt.Errorf("参数解析失败: %v", err)
	}
	
	// 设置默认值
	if args.Length == 0 {
		args.Length = 3
	}
	if args.DelayMs == 0 {
		args.DelayMs = 500
	}
	
	log.Printf("[StreamTextGenerator] 开始生成关于 '%s' 的文本，共 %d 段", args.Topic, args.Length)
	
	// 创建一个 channel 来传递流式数据
	resultChan := make(chan string, args.Length+1)
	errorChan := make(chan error, 1)
	
	// 在 goroutine 中生成内容
	go func() {
		defer close(resultChan)
		defer close(errorChan)
		
		paragraphs := generateTopicParagraphs(args.Topic, args.Length)
		
		for i, paragraph := range paragraphs {
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			default:
				// 发送段落内容
				resultChan <- fmt.Sprintf("段落 %d: %s\n\n", i+1, paragraph)
				
				// 延迟（除了最后一段）
				if i < len(paragraphs)-1 {
					time.Sleep(time.Duration(args.DelayMs) * time.Millisecond)
				}
			}
		}
		
		// 发送完成消息
		resultChan <- fmt.Sprintf("--- 关于 '%s' 的文本生成完成 ---", args.Topic)
	}()
	
	// 创建 StreamReader
	reader := &schema.StreamReader[string]{
		// 这里需要实现 StreamReader 的接口方法
		// 注意: 实际的 Eino 框架可能有不同的实现方式
	}
	
	// 模拟 StreamReader 的实现（实际使用时需要根据 Eino 框架的具体实现）
	reader = createMockStreamReader(resultChan, errorChan)
	
	return reader, nil
}

// --- 示例 2: 流式数据处理工具 ---

type StreamDataProcessorTool struct{}

func (d *StreamDataProcessorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "stream_data_processor",
		Desc: "逐步处理数据数组，实时返回处理进度",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"data": {
				Type:     "array",
				Desc:     "要处理的数据数组",
				Required: true,
			},
			"operation": {
				Type:     "string",
				Desc:     "处理操作类型",
				Required: true,
				Enum:     []string{"square", "double", "increment"},
			},
			"batch_size": {
				Type:     "integer",
				Desc:     "每批处理的数据量",
				Required: false,
			},
		}),
	}, nil
}

func (d *StreamDataProcessorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	reader, err := d.StreamableRun(ctx, argumentsInJSON, opts...)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	
	var result strings.Builder
	for {
		chunk, err := reader.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
		result.WriteString(chunk)
	}
	
	return result.String(), nil
}

func (d *StreamDataProcessorTool) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (*schema.StreamReader[string], error) {
	var args struct {
		Data      []float64 `json:"data"`
		Operation string    `json:"operation"`
		BatchSize int       `json:"batch_size"`
	}
	
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return nil, fmt.Errorf("参数解析失败: %v", err)
	}
	
	if args.BatchSize == 0 {
		args.BatchSize = 5
	}
	
	log.Printf("[StreamDataProcessor] 处理 %d 个数据点，操作: %s", len(args.Data), args.Operation)
	
	resultChan := make(chan string, len(args.Data)/args.BatchSize+2)
	errorChan := make(chan error, 1)
	
	go func() {
		defer close(resultChan)
		defer close(errorChan)
		
		totalBatches := (len(args.Data) + args.BatchSize - 1) / args.BatchSize
		
		for batchIndex := 0; batchIndex < totalBatches; batchIndex++ {
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			default:
				start := batchIndex * args.BatchSize
				end := start + args.BatchSize
				if end > len(args.Data) {
					end = len(args.Data)
				}
				
				batch := args.Data[start:end]
				processedBatch := make([]float64, len(batch))
				
				// 处理当前批次
				for i, value := range batch {
					switch args.Operation {
					case "square":
						processedBatch[i] = value * value
					case "double":
						processedBatch[i] = value * 2
					case "increment":
						processedBatch[i] = value + 1
					default:
						processedBatch[i] = value
					}
				}
				
				// 发送批次处理结果
				batchResult := map[string]interface{}{
					"batch_index":    batchIndex + 1,
					"total_batches":  totalBatches,
					"processed_data": processedBatch,
					"original_data":  batch,
					"progress":       fmt.Sprintf("%.1f%%", float64(batchIndex+1)/float64(totalBatches)*100),
				}
				
				resultBytes, _ := json.Marshal(batchResult)
				resultChan <- string(resultBytes) + "\n"
				
				// 模拟处理延迟
				time.Sleep(200 * time.Millisecond)
			}
		}
		
		// 发送完成消息
		summary := map[string]interface{}{
			"status":       "completed",
			"total_items":  len(args.Data),
			"operation":    args.Operation,
			"batch_size":   args.BatchSize,
			"total_batches": totalBatches,
		}
		summaryBytes, _ := json.Marshal(summary)
		resultChan <- string(summaryBytes)
	}()
	
	return createMockStreamReader(resultChan, errorChan), nil
}

// --- 示例 3: 流式日志分析工具 ---

type StreamLogAnalyzerTool struct{}

func (l *StreamLogAnalyzerTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "stream_log_analyzer",
		Desc: "实时分析日志内容并流式返回分析结果",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"log_content": {
				Type:     "string",
				Desc:     "要分析的日志内容",
				Required: true,
			},
			"analysis_types": {
				Type:     "array",
				Desc:     "分析类型列表",
				Required: true,
			},
		}),
	}, nil
}

func (l *StreamLogAnalyzerTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	reader, err := l.StreamableRun(ctx, argumentsInJSON, opts...)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	
	var result strings.Builder
	for {
		chunk, err := reader.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
		result.WriteString(chunk)
	}
	
	return result.String(), nil
}

func (l *StreamLogAnalyzerTool) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (*schema.StreamReader[string], error) {
	var args struct {
		LogContent    string   `json:"log_content"`
		AnalysisTypes []string `json:"analysis_types"`
	}
	
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return nil, fmt.Errorf("参数解析失败: %v", err)
	}
	
	log.Printf("[StreamLogAnalyzer] 分析日志，长度: %d", len(args.LogContent))
	
	resultChan := make(chan string, len(args.AnalysisTypes)+2)
	errorChan := make(chan error, 1)
	
	go func() {
		defer close(resultChan)
		defer close(errorChan)
		
		lines := strings.Split(args.LogContent, "\n")
		
		for _, analysisType := range args.AnalysisTypes {
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			default:
				time.Sleep(300 * time.Millisecond) // 模拟分析时间
				
				result := performLogAnalysis(lines, analysisType)
				
				analysisResult := map[string]interface{}{
					"analysis_type": analysisType,
					"result":        result,
					"timestamp":     time.Now().Format(time.RFC3339),
				}
				
				resultBytes, _ := json.Marshal(analysisResult)
				resultChan <- string(resultBytes) + "\n"
			}
		}
		
		// 发送分析完成消息
		resultChan <- `{"status": "analysis_completed", "total_types": ` + fmt.Sprintf("%d", len(args.AnalysisTypes)) + `}`
	}()
	
	return createMockStreamReader(resultChan, errorChan), nil
}

// --- 辅助函数 ---

// generateTopicParagraphs 生成指定主题的段落
func generateTopicParagraphs(topic string, count int) []string {
	templates := []string{
		"%s 是一个非常有趣的主题，它涉及到多个方面的知识和技能。",
		"在 %s 领域，我们可以看到许多创新和发展的机会。",
		"对于 %s 的研究，需要结合理论知识和实践经验。",
		"通过深入了解 %s，我们能够获得更多宝贵的洞察。",
		"%s 的应用范围很广，对多个行业都有重要影响。",
	}
	
	var paragraphs []string
	for i := 0; i < count; i++ {
		template := templates[i%len(templates)]
		paragraph := fmt.Sprintf(template, topic)
		paragraphs = append(paragraphs, paragraph)
	}
	
	return paragraphs
}

// performLogAnalysis 执行日志分析
func performLogAnalysis(lines []string, analysisType string) interface{} {
	switch analysisType {
	case "error_count":
		count := 0
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "error") {
				count++
			}
		}
		return count
	case "warning_count":
		count := 0
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "warning") {
				count++
			}
		}
		return count
	case "line_count":
		return len(lines)
	case "unique_ips":
		ips := make(map[string]bool)
		for _, line := range lines {
			// 简单的IP提取（实际应该用正则表达式）
			parts := strings.Fields(line)
			if len(parts) > 0 && strings.Contains(parts[0], ".") {
				ips[parts[0]] = true
			}
		}
		return len(ips)
	default:
		return "unknown_analysis_type"
	}
}

// createMockStreamReader 创建模拟的 StreamReader
func createMockStreamReader(resultChan <-chan string, errorChan <-chan error) *schema.StreamReader[string] {
	// 注意：这是一个模拟实现，实际的 Eino 框架可能有不同的 StreamReader 实现
	// 这里只是为了演示流式工具的概念
	
	reader := &MockStreamReader{
		resultChan: resultChan,
		errorChan:  errorChan,
		closed:     false,
	}
	
	// 由于 schema.StreamReader 是具体类型，我们需要用不同的方法
	// 在实际使用中，应该参考 Eino 框架的具体实现
	
	return (*schema.StreamReader[string])(reader)
}

// MockStreamReader 模拟的 StreamReader 实现
type MockStreamReader struct {
	resultChan <-chan string
	errorChan  <-chan error
	closed     bool
}

func (r *MockStreamReader) Recv() (string, error) {
	if r.closed {
		return "", fmt.Errorf("EOF")
	}
	
	select {
	case result, ok := <-r.resultChan:
		if !ok {
			return "", fmt.Errorf("EOF")
		}
		return result, nil
	case err := <-r.errorChan:
		return "", err
	}
}

func (r *MockStreamReader) Close() error {
	r.closed = true
	return nil
}

// --- 演示函数 ---

func demonstrateStreamableTools() {
	ctx := context.Background()
	
	fmt.Println("=== 流式 Tool 实现演示 ===\n")
	
	// 1. 测试流式文本生成工具
	fmt.Println("--- 流式文本生成工具 ---")
	textGenerator := &StreamTextGeneratorTool{}
	
	fmt.Println("开始流式生成文本...")
	textResult, err := textGenerator.InvokableRun(ctx, `{
		"topic": "人工智能",
		"length": 3,
		"delay_ms": 1000
	}`)
	if err != nil {
		log.Printf("文本生成失败: %v", err)
	} else {
		fmt.Printf("生成结果:\n%s\n", textResult)
	}
	
	// 2. 测试流式数据处理工具
	fmt.Println("--- 流式数据处理工具 ---")
	dataProcessor := &StreamDataProcessorTool{}
	
	fmt.Println("开始流式处理数据...")
	dataResult, err := dataProcessor.InvokableRun(ctx, `{
		"data": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
		"operation": "square",
		"batch_size": 3
	}`)
	if err != nil {
		log.Printf("数据处理失败: %v", err)
	} else {
		fmt.Printf("处理结果:\n%s\n", dataResult)
	}
	
	// 3. 测试流式日志分析工具
	fmt.Println("--- 流式日志分析工具 ---")
	logAnalyzer := &StreamLogAnalyzerTool{}
	
	logContent := `2024-08-19 10:00:01 INFO Starting application
2024-08-19 10:00:02 ERROR Failed to connect to database
2024-08-19 10:00:03 WARNING Connection retry attempt 1
2024-08-19 10:00:04 INFO Connection established
2024-08-19 10:00:05 ERROR Authentication failed`
	
	fmt.Println("开始流式分析日志...")
	logResult, err := logAnalyzer.InvokableRun(ctx, fmt.Sprintf(`{
		"log_content": %q,
		"analysis_types": ["error_count", "warning_count", "line_count"]
	}`, logContent))
	if err != nil {
		log.Printf("日志分析失败: %v", err)
	} else {
		fmt.Printf("分析结果:\n%s\n", logResult)
	}
}

func main() {
	demonstrateStreamableTools()
}