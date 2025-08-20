package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// StreamTextGeneratorTool 实现了支持流式输出的文本生成工具
// 该工具演示了如何逐步生成内容并通过流式接口返回结果
type StreamTextGeneratorTool struct{}

// Info 返回工具的基本信息和参数定义
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

// InvokableRun 实现非流式调用接口，返回完整的生成结果
// 该方法内部调用 StreamableRun 并收集所有流式输出
func (s *StreamTextGeneratorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (string, error) {
	// 对于流式工具，InvokableRun 通常返回完整结果
	reader, err := s.StreamableRun(ctx, argumentsInJSON, opts...)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// 收集所有流式输出块并拼接成完整结果
	var result strings.Builder
	for {
		chunk, err := reader.Recv()
		if err != nil {
			if err == io.EOF {
				break // 流结束
			}
			return "", err
		}
		result.WriteString(chunk)
	}

	return result.String(), nil
}

// StreamableRun 实现流式调用接口，返回可以逐步读取结果的 StreamReader
func (s *StreamTextGeneratorTool) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...interface{}) (*schema.StreamReader[string], error) {
	// 定义输入参数结构体
	var args struct {
		Topic   string `json:"topic"`    // 生成文本的主题
		Length  int    `json:"length"`   // 生成段落数量
		DelayMs int    `json:"delay_ms"` // 每段之间的延迟毫秒数
	}

	// 解析 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return nil, fmt.Errorf("参数解析失败: %v", err)
	}

	// 设置默认值
	if args.Length == 0 {
		args.Length = 3 // 默认生成3段
	}
	if args.DelayMs == 0 {
		args.DelayMs = 500 // 默认延迟500毫秒
	}

	log.Printf("[StreamTextGenerator] 开始生成关于 '%s' 的文本，共 %d 段", args.Topic, args.Length)

	// 创建用于流式数据传递的 channels
	// resultChan: 传递生成的文本内容，缓冲大小为段落数+1（包括完成消息）
	// errorChan: 传递错误信息，缓冲大小为1
	resultChan := make(chan string, args.Length+1)
	errorChan := make(chan error, 1)

	// 启动 goroutine 异步生成内容
	go func() {
		// 确保 channels 在函数结束时被关闭
		defer close(resultChan)
		defer close(errorChan)

		// 生成指定主题和数量的段落内容
		paragraphs := generateTopicParagraphs(args.Topic, args.Length)

		for i, paragraph := range paragraphs {
			select {
			case <-ctx.Done():
				// 确保错误被发送后再返回
				select {
				case errorChan <- ctx.Err():
				default:
				}
				return
			case resultChan <- fmt.Sprintf("段落 %d: %s\n\n", i+1, paragraph):
				// 延迟（除了最后一段）
				if i < len(paragraphs)-1 {
					select {
					case <-ctx.Done():
						select {
						case errorChan <- ctx.Err():
						default:
						}
						return
					case <-time.After(time.Duration(args.DelayMs) * time.Millisecond):
					}
				}
			}
		}

		// 发送完成消息
		select {
		case <-ctx.Done():
			select {
			case errorChan <- ctx.Err():
			default:
			}
			return
		case resultChan <- fmt.Sprintf("--- 关于 '%s' 的文本生成完成 ---", args.Topic):
		}
	}()

	// 创建 StreamReader
	// 注意: 实际使用时需要根据 Eino 框架的具体实现来创建 StreamReader
	reader := createMockStreamReader(resultChan, errorChan)

	return reader, nil
}

// --- 示例 2: 流式数据处理工具 ---

// StreamDataProcessorTool 实现了支持流式输出的数据处理工具
// 该工具演示了如何分批处理大量数据并实时返回处理进度
type StreamDataProcessorTool struct{}

// Info 返回数据处理工具的基本信息和参数定义
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

// InvokableRun 实现数据处理工具的非流式调用接口
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
			if err == io.EOF {
				break
			}
			return "", err
		}
		result.WriteString(chunk)
	}

	return result.String(), nil
}

// StreamableRun 实现数据处理工具的流式调用接口，分批处理数据并返回进度
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
				select {
				case errorChan <- ctx.Err():
				default:
				}
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

				select {
				case <-ctx.Done():
					select {
					case errorChan <- ctx.Err():
					default:
					}
					return
				case resultChan <- string(resultBytes) + "\n":
				}

				// 模拟处理延迟，同时检查上下文取消
				select {
				case <-ctx.Done():
					select {
					case errorChan <- ctx.Err():
					default:
					}
					return
				case <-time.After(200 * time.Millisecond):
				}
			}
		}

		// 发送完成消息
		summary := map[string]interface{}{
			"status":        "completed",
			"total_items":   len(args.Data),
			"operation":     args.Operation,
			"batch_size":    args.BatchSize,
			"total_batches": totalBatches,
		}
		summaryBytes, _ := json.Marshal(summary)

		select {
		case <-ctx.Done():
			select {
			case errorChan <- ctx.Err():
			default:
			}
			return
		case resultChan <- string(summaryBytes):
		}
	}()

	return createMockStreamReader(resultChan, errorChan), nil
}

// --- 示例 3: 流式日志分析工具 ---

// StreamLogAnalyzerTool 实现了支持流式输出的日志分析工具
// 该工具演示了如何实时分析日志内容并逐步返回分析结果
type StreamLogAnalyzerTool struct{}

// Info 返回日志分析工具的基本信息和参数定义
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

// InvokableRun 实现日志分析工具的非流式调用接口
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
			if err == io.EOF {
				break
			}
			return "", err
		}
		result.WriteString(chunk)
	}

	return result.String(), nil
}

// StreamableRun 实现日志分析工具的流式调用接口，逐步分析日志并返回结果
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
				select {
				case errorChan <- ctx.Err():
				default:
				}
				return
			default:
				// 模拟分析时间，同时检查上下文取消
				select {
				case <-ctx.Done():
					select {
					case errorChan <- ctx.Err():
					default:
					}
					return
				case <-time.After(300 * time.Millisecond):
				}

				result := performLogAnalysis(lines, analysisType)

				analysisResult := map[string]interface{}{
					"analysis_type": analysisType,
					"result":        result,
					"timestamp":     time.Now().Format(time.RFC3339),
				}

				resultBytes, _ := json.Marshal(analysisResult)

				select {
				case <-ctx.Done():
					select {
					case errorChan <- ctx.Err():
					default:
					}
					return
				case resultChan <- string(resultBytes) + "\n":
				}
			}
		}

		// 发送分析完成消息
		completionMsg := `{"status": "analysis_completed", "total_types": ` + fmt.Sprintf("%d", len(args.AnalysisTypes)) + `}`
		select {
		case <-ctx.Done():
			select {
			case errorChan <- ctx.Err():
			default:
			}
			return
		case resultChan <- completionMsg:
		}
	}()

	return createMockStreamReader(resultChan, errorChan), nil
}

// --- 辅助函数 ---

// generateTopicParagraphs 根据指定主题生成指定数量的段落内容
// 参数:
//
//	topic: 文本主题
//	count: 需要生成的段落数量
//
// 返回:
//
//	[]string: 生成的段落内容数组
func generateTopicParagraphs(topic string, count int) []string {
	// 预定义的段落模板，用于生成不同样式的内容
	templates := []string{
		"%s 是一个非常有趣的主题，它涉及到多个方面的知识和技能。",
		"在 %s 领域，我们可以看到许多创新和发展的机会。",
		"对于 %s 的研究，需要结合理论知识和实践经验。",
		"通过深入了解 %s，我们能够获得更多宝贵的洞察。",
		"%s 的应用范围很广，对多个行业都有重要影响。",
	}

	var paragraphs []string
	// 循环使用模板生成指定数量的段落
	for i := 0; i < count; i++ {
		template := templates[i%len(templates)] // 循环使用模板
		paragraph := fmt.Sprintf(template, topic)
		paragraphs = append(paragraphs, paragraph)
	}

	return paragraphs
}

// performLogAnalysis 执行指定类型的日志分析
// 参数:
//
//	lines: 日志行数组
//	analysisType: 分析类型（error_count, warning_count, line_count, unique_ips）
//
// 返回:
//
//	interface{}: 分析结果（类型根据分析类型而定）
func performLogAnalysis(lines []string, analysisType string) interface{} {
	switch analysisType {
	case "error_count":
		// 统计包含 "error" 的日志行数量
		count := 0
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "error") {
				count++
			}
		}
		return count
	case "warning_count":
		// 统计包含 "warning" 的日志行数量
		count := 0
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "warning") {
				count++
			}
		}
		return count
	case "line_count":
		// 返回总行数
		return len(lines)
	case "unique_ips":
		// 统计唯一IP地址数量（简化实现）
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
	// 使用 schema.Pipe 创建一个 StreamReader 和 StreamWriter 对
	reader, writer := schema.Pipe[string](100) // 缓冲区大小为100
	
	// 启动一个 goroutine 将 channel 数据写入到 StreamWriter 中
	go func() {
		defer writer.Close() // 确保写入器最终被关闭
		
		for {
			select {
			case data, ok := <-resultChan:
				if !ok {
					// resultChan 已关闭，正常结束
					return
				}
				closed := writer.Send(data, nil)
				if closed {
					// 写入器已关闭，退出
					return
				}
			case err := <-errorChan:
				if err != nil {
					// 发生错误，发送错误并退出
					writer.Send("", err)
					return
				}
			}
		}
	}()
	
	return reader
}

// 注意：我们现在使用 Eino 框架提供的 schema.Pipe 来创建真正的 StreamReader，
// 不再需要自定义的 MockStreamReader 实现

// --- 演示函数 ---

// demonstrateStreamableTools 演示各种流式工具的使用方法
func demonstrateStreamableTools() {
	ctx := context.Background()

	fmt.Println("=== 流式 Tool 实现演示 ===")

	// 1. 测试流式文本生成工具
	fmt.Println("--- 流式文本生成工具 ---")
	textGenerator := &StreamTextGeneratorTool{}

	fmt.Println("开始流式生成文本...")
	// 调用流式文本生成工具，生成关于"人工智能"主题的3段文本
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
	// 调用流式数据处理工具，对数据进行平方运算，每批处理3个数据
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

	// 模拟日志内容，包含不同级别的日志信息
	logContent := `2024-08-19 10:00:01 INFO Starting application
2024-08-19 10:00:02 ERROR Failed to connect to database
2024-08-19 10:00:03 WARNING Connection retry attempt 1
2024-08-19 10:00:04 INFO Connection established
2024-08-19 10:00:05 ERROR Authentication failed`

	fmt.Println("开始流式分析日志...")
	// 调用流式日志分析工具，分析错误数量、警告数量和总行数
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

// main 程序入口点，运行流式工具演示
func main() {
	demonstrateStreamableTools()
}
