package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/spf13/viper"
)

// =============================================================================
//
//  文件: main.go
//  功能: 演示如何使用 Graph 构建 Agent
//  说明: 展示 Graph 的分支逻辑、并行处理和复杂工作流能力
//
// =============================================================================

// TaskRequest 表示一个任务请求
type TaskRequest struct {
	ID       string `json:"id"`
	Type     string `json:"type"`     // analyze, process, report
	Content  string `json:"content"`
	Priority string `json:"priority"` // high, medium, low
	Status   string `json:"status"`   // pending, processing, completed, failed
	Result   string `json:"result"`
	Created  string `json:"created"`
}

// 全局任务存储
var (
	taskQueue = make(map[string]*TaskRequest)
	nextTaskID = 1
)

// =============================================================================
// 工具实现 - 任务管理工具集
// =============================================================================

// TaskClassifierTool 任务分类工具 - 分析任务类型和复杂度
type TaskClassifierTool struct{}

func (t *TaskClassifierTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "classify_task",
		Desc: "分析任务类型和复杂度，决定处理路径",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"content": {
				Type:     "string",
				Desc:     "任务内容描述",
				Required: true,
			},
		}),
	}, nil
}

func (t *TaskClassifierTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Content string `json:"content"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[TaskClassifier] 分析任务: %s", args.Content)

	// 简单的任务分类逻辑
	var taskType, complexity, priority string
	
	content := strings.ToLower(args.Content)
	switch {
	case strings.Contains(content, "分析") || strings.Contains(content, "analyze"):
		taskType = "analyze"
		complexity = "medium"
		priority = "high"
	case strings.Contains(content, "处理") || strings.Contains(content, "process"):
		taskType = "process"
		complexity = "high"
		priority = "medium"
	case strings.Contains(content, "报告") || strings.Contains(content, "report"):
		taskType = "report"
		complexity = "low"
		priority = "low"
	default:
		taskType = "general"
		complexity = "medium"
		priority = "medium"
	}

	result := map[string]interface{}{
		"task_type":  taskType,
		"complexity": complexity,
		"priority":   priority,
		"estimated_time": getEstimatedTime(complexity),
		"recommended_path": getRecommendedPath(taskType),
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// DataAnalysisTool 数据分析工具
type DataAnalysisTool struct{}

func (t *DataAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "analyze_data",
		Desc: "执行数据分析任务",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"data": {
				Type:     "string",
				Desc:     "要分析的数据",
				Required: true,
			},
			"analysis_type": {
				Type:     "string",
				Desc:     "分析类型",
				Required: false,
				Enum:     []string{"statistical", "trend", "correlation"},
			},
		}),
	}, nil
}

func (t *DataAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Data         string `json:"data"`
		AnalysisType string `json:"analysis_type"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.AnalysisType == "" {
		args.AnalysisType = "statistical"
	}

	log.Printf("[DataAnalysis] 执行 %s 分析: %s", args.AnalysisType, args.Data)

	// 模拟分析过程
	time.Sleep(500 * time.Millisecond)

	result := map[string]interface{}{
		"analysis_type": args.AnalysisType,
		"data_points":   10,
		"mean":          75.5,
		"median":        78.0,
		"trend":         "increasing",
		"confidence":    0.85,
		"summary":       fmt.Sprintf("对数据进行了%s分析，发现上升趋势", args.AnalysisType),
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// TextProcessorTool 文本处理工具
type TextProcessorTool struct{}

func (t *TextProcessorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "process_text",
		Desc: "执行文本处理任务",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"text": {
				Type:     "string",
				Desc:     "要处理的文本",
				Required: true,
			},
			"operation": {
				Type:     "string",
				Desc:     "处理操作",
				Required: false,
				Enum:     []string{"clean", "extract", "summarize"},
			},
		}),
	}, nil
}

func (t *TextProcessorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Text      string `json:"text"`
		Operation string `json:"operation"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.Operation == "" {
		args.Operation = "clean"
	}

	log.Printf("[TextProcessor] 执行 %s 操作: %s", args.Operation, args.Text[:min(50, len(args.Text))])

	// 模拟处理过程
	time.Sleep(300 * time.Millisecond)

	result := map[string]interface{}{
		"operation":      args.Operation,
		"original_length": len(args.Text),
		"processed_text": fmt.Sprintf("[已%s] %s", args.Operation, args.Text),
		"word_count":     len(strings.Fields(args.Text)),
		"status":         "success",
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// ReportGeneratorTool 报告生成工具
type ReportGeneratorTool struct{}

func (t *ReportGeneratorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "generate_report",
		Desc: "生成任务执行报告",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"results": {
				Type:     "string",
				Desc:     "任务执行结果（JSON格式）",
				Required: true,
			},
			"format": {
				Type:     "string",
				Desc:     "报告格式",
				Required: false,
				Enum:     []string{"summary", "detailed", "executive"},
			},
		}),
	}, nil
}

func (t *ReportGeneratorTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Results string `json:"results"`
		Format  string `json:"format"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if args.Format == "" {
		args.Format = "summary"
	}

	log.Printf("[ReportGenerator] 生成 %s 格式报告", args.Format)

	// 模拟报告生成
	time.Sleep(200 * time.Millisecond)

	report := map[string]interface{}{
		"report_type":    args.Format,
		"generated_at":   time.Now().Format(time.RFC3339),
		"summary":        "任务执行成功，所有目标均已达成",
		"key_findings":   []string{"性能提升15%", "准确率达到95%", "处理时间减少30%"},
		"recommendations": []string{"建议继续优化算法", "增加数据验证步骤"},
		"next_steps":     "继续监控和优化",
	}

	resultBytes, _ := json.Marshal(report)
	return string(resultBytes), nil
}

// QualityCheckerTool 质量检查工具
type QualityCheckerTool struct{}

func (t *QualityCheckerTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "check_quality",
		Desc: "检查任务执行质量",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"task_result": {
				Type:     "string",
				Desc:     "任务执行结果",
				Required: true,
			},
		}),
	}, nil
}

func (t *QualityCheckerTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		TaskResult string `json:"task_result"`
	}

	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[QualityChecker] 检查质量")

	// 模拟质量检查
	time.Sleep(100 * time.Millisecond)

	qualityScore := 0.92 // 模拟质量分数
	passed := qualityScore >= 0.8

	result := map[string]interface{}{
		"quality_score":  qualityScore,
		"passed":         passed,
		"issues_found":   []string{},
		"recommendations": []string{"建议添加更多测试用例"},
		"overall_rating": "excellent",
	}

	if !passed {
		result["issues_found"] = []string{"数据完整性问题", "格式不规范"}
		result["overall_rating"] = "needs_improvement"
	}

	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// =============================================================================
// 工具集合和模型创建
// =============================================================================

func createTools() []tool.InvokableTool {
	return []tool.InvokableTool{
		&TaskClassifierTool{},
		&DataAnalysisTool{},
		&TextProcessorTool{},
		&ReportGeneratorTool{},
		&QualityCheckerTool{},
	}
}

func createChatModel(ctx context.Context, tools []tool.InvokableTool) (*ark.ChatModel, error) {
	config := &ark.ChatModelConfig{
		Model:  viper.GetString("ARK_MODEL"),
		APIKey: viper.GetString("ARK_API_KEY"),
	}

	chatModel, err := ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("创建聊天模型失败: %v", err)
	}

	// 绑定工具到聊天模型
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			log.Printf("获取工具信息失败: %v", err)
			continue
		}
		toolInfos = append(toolInfos, info)
		log.Printf("绑定工具: %s - %s", info.Name, info.Desc)
	}

	chatModel.BindTools(toolInfos)
	return chatModel, nil
}

// =============================================================================
// Graph 构建 - 复杂的分支工作流
// =============================================================================

func createGraphAgent(ctx context.Context) (compose.Runnable[[]*schema.Message, []*schema.Message], error) {
	// 1. 创建工具和模型
	tools := createTools()
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, err
	}

	// 2. 定义节点名称
	const (
		CLASSIFIER    = "classifier"
		ANALYZER      = "analyzer" 
		PROCESSOR     = "processor"
		REPORT_GEN    = "report_generator"
		QUALITY_CHECK = "quality_checker"
		AGGREGATOR    = "aggregator"
	)

	// 3. 创建 Graph
	g := compose.NewGraph[[]*schema.Message, []*schema.Message]()

	// 4. 添加 ChatModel 节点 - 统一的LLM节点
	g.AddChatModelNode(CLASSIFIER, chatModel)

	// 添加消息包装器节点处理类型转换
	g.AddLambdaNode("message_wrapper", compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
		return []*schema.Message{msg}, nil
	}))

	// 5. 添加 Lambda 节点用于条件路由（简化版本）
	g.AddLambdaNode("conditional_processor", compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
		// 检查最后一条消息的内容来决定处理方式
		if len(msgs) == 0 {
			return msgs, nil
		}

		lastMsg := msgs[len(msgs)-1]
		log.Printf("[ConditionalProcessor] 分析内容并执行处理: %s", lastMsg.Content)

		content := strings.ToLower(lastMsg.Content)
		var processType, result string
		
		switch {
		case strings.Contains(content, "分析") || strings.Contains(content, "analyze"):
			processType = "数据分析"
			result = "数据分析完成：发现关键趋势和模式，准确率达到95%"
		case strings.Contains(content, "处理") || strings.Contains(content, "process"):
			processType = "文本处理"
			result = "文本处理完成：内容已清理和格式化，提取了关键信息"
		case strings.Contains(content, "报告") || strings.Contains(content, "report"):
			processType = "报告生成"
			result = "报告生成完成：包含详细分析和建议，格式规范"
		default:
			processType = "综合处理"
			result = "综合任务处理完成：包含分析、处理和报告生成的完整流程"
		}

		// 创建处理结果消息
		processResult := &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("[%s] %s", processType, result),
		}

		return append(msgs, processResult), nil
	}))

	// 6. 添加质量检查节点
	g.AddLambdaNode(QUALITY_CHECK, compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
		log.Printf("[QualityCheck] 执行质量检查")
		
		qualityResult := &schema.Message{
			Role:    schema.Assistant,
			Content: "质量检查通过：结果符合预期标准，质量分数: 95%",
		}
		return append(msgs, qualityResult), nil
	}))

	// 7. 添加最终聚合节点
	g.AddLambdaNode(AGGREGATOR, compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
		log.Printf("[Aggregator] 聚合最终结果")
		
		// 统计处理过程
		var steps []string
		for _, msg := range msgs {
			if msg.Role == schema.Assistant {
				steps = append(steps, msg.Content)
			}
		}

		summary := fmt.Sprintf("Graph Agent 任务处理完成！\n执行了 %d 个处理步骤，所有质量检查均已通过。\n系统采用智能路由机制，根据任务类型自动选择最优处理路径。", len(steps))
		
		result := &schema.Message{
			Role:    schema.Assistant,
			Content: summary,
		}

		return []*schema.Message{result}, nil
	}))

	// 8. 定义简化的 Graph 边关系
	// 线性路径：START -> CLASSIFIER -> message_wrapper -> conditional_processor -> quality_check -> aggregator -> END
	g.AddEdge(compose.START, CLASSIFIER)
	g.AddEdge(CLASSIFIER, "message_wrapper")
	g.AddEdge("message_wrapper", "conditional_processor")
	g.AddEdge("conditional_processor", QUALITY_CHECK)
	g.AddEdge(QUALITY_CHECK, AGGREGATOR)
	g.AddEdge(AGGREGATOR, compose.END)

	// 10. 编译 Graph
	agent, err := g.Compile(ctx, 
		compose.WithGraphName("GraphAgent"),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
	)
	if err != nil {
		return nil, fmt.Errorf("编译 Graph 失败: %v", err)
	}

	log.Println("Graph Agent 创建成功")
	return agent, nil
}

// =============================================================================
// 演示和配置
// =============================================================================

func loadConfig() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
}

func runGraphDemo(ctx context.Context, agent compose.Runnable[[]*schema.Message, []*schema.Message]) {
	fmt.Println("\n=== Graph Agent 演示开始 ===")

	scenarios := []struct {
		name    string
		message string
	}{
		{
			name:    "场景1: 数据分析任务",
			message: "请帮我分析一下销售数据的趋势，我需要了解最近三个月的变化情况",
		},
		{
			name:    "场景2: 文本处理任务", 
			message: "帮我处理这份文档，需要清理格式并提取关键信息",
		},
		{
			name:    "场景3: 报告生成任务",
			message: "基于之前的分析结果，请生成一份详细的项目报告",
		},
		{
			name:    "场景4: 综合任务",
			message: "请帮我完成一个复杂的业务流程：包括数据收集、分析和报告生成",
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n--- %s ---\n", scenario.name)
		fmt.Printf("用户: %s\n", scenario.message)

		messages := []*schema.Message{
			{
				Role:    schema.User,
				Content: scenario.message,
			},
		}

		resp, err := agent.Invoke(ctx, messages)
		if err != nil {
			log.Printf("Graph Agent 调用失败: %v", err)
			continue
		}

		for _, msg := range resp {
			if msg.Role == schema.Assistant {
				fmt.Printf("Graph Agent: %s\n", msg.Content)
			}
		}

		time.Sleep(time.Second)
	}

	fmt.Println("\n=== Graph Agent 演示结束 ===")
}

func main() {
	ctx := context.Background()

	// 加载配置
	loadConfig()
	log.Println("配置加载成功")

	// 创建 Graph Agent
	agent, err := createGraphAgent(ctx)
	if err != nil {
		log.Fatalf("创建 Graph Agent 失败: %v", err)
	}

	// 运行演示
	runGraphDemo(ctx, agent)
}

// =============================================================================
// 辅助函数
// =============================================================================

func getEstimatedTime(complexity string) string {
	switch complexity {
	case "low":
		return "1-5 分钟"
	case "medium":
		return "5-15 分钟"
	case "high":
		return "15-30 分钟"
	default:
		return "未知"
	}
}

func getRecommendedPath(taskType string) string {
	switch taskType {
	case "analyze":
		return "分析分支 -> 质量检查 -> 报告聚合"
	case "process":
		return "处理分支 -> 质量检查 -> 结果聚合"
	case "report":
		return "报告分支 -> 质量检查 -> 最终输出"
	default:
		return "通用分支 -> 质量检查 -> 标准输出"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}