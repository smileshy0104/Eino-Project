// Package main 演示如何使用 Eino 框架构建基于 Graph 的智能 Agent
// Graph Agent 相比 Chain Agent 具有更强的分支逻辑和并行处理能力
// 可以根据不同的任务类型动态选择处理路径，实现复杂的工作流编排
package main

import (
	"context"      // 上下文管理，用于控制请求生命周期
	"encoding/json" // JSON 编解码，用于工具参数和结果的序列化
	"fmt"          // 格式化输出
	"log"          // 日志记录
	"strings"      // 字符串操作
	"time"         // 时间处理

	"github.com/cloudwego/eino/components/tool" // Eino 工具组件
	"github.com/cloudwego/eino/compose"         // Eino 编排组件，用于构建 Graph
	"github.com/cloudwego/eino/schema"          // Eino 模式定义
	"github.com/cloudwego/eino-ext/components/model/ark" // ARK 模型扩展
	"github.com/spf13/viper"                    // 配置管理
)

// =============================================================================
//
//  文件: main.go
//  功能: 演示如何使用 Graph 构建 Agent
//  说明: 展示 Graph 的分支逻辑、并行处理和复杂工作流能力
//
// =============================================================================

// TaskRequest 表示一个任务请求的完整信息
// 用于在 Graph Agent 中跟踪和管理任务的生命周期
type TaskRequest struct {
	ID       string `json:"id"`       // 任务唯一标识符
	Type     string `json:"type"`     // 任务类型：analyze(分析), process(处理), report(报告)
	Content  string `json:"content"`  // 任务具体内容描述
	Priority string `json:"priority"` // 任务优先级：high(高), medium(中), low(低)
	Status   string `json:"status"`   // 任务状态：pending(待处理), processing(处理中), completed(已完成), failed(失败)
	Result   string `json:"result"`   // 任务执行结果
	Created  string `json:"created"`  // 任务创建时间
}

// 全局任务存储
// 在实际应用中，这些应该存储在数据库或持久化存储中
var (
	taskQueue  = make(map[string]*TaskRequest) // 任务队列，使用 map 存储任务ID到任务的映射
	nextTaskID = 1                             // 下一个任务ID计数器
)

// =============================================================================
// 工具实现 - 任务管理工具集
// =============================================================================

// TaskClassifierTool 任务分类工具 - 智能分析任务类型和复杂度
// 这是 Graph Agent 的核心工具之一，负责：
// 1. 分析输入任务的内容和特征
// 2. 确定任务类型（分析、处理、报告等）
// 3. 评估任务复杂度和优先级
// 4. 为 Graph 路由提供决策依据
type TaskClassifierTool struct{}

// Info 返回工具的元信息，包括名称、描述和参数定义
// 这些信息会被 LLM 用来理解工具的功能和使用方式
func (t *TaskClassifierTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "classify_task",                          // 工具名称，LLM 调用时使用
		Desc: "分析任务类型和复杂度，决定处理路径",              // 工具描述，帮助 LLM 理解功能
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"content": {
				Type:     "string", // 参数类型
				Desc:     "任务内容描述", // 参数描述
				Required: true,    // 是否必需参数
			},
		}),
	}, nil
}

// InvokableRun 执行任务分类的核心逻辑
// 参数：
//   - ctx: 上下文对象
//   - argumentsInJSON: JSON 格式的工具参数
//   - opts: 可选的工具选项
// 返回：JSON 格式的分类结果和错误信息
func (t *TaskClassifierTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体，用于解析 JSON 参数
	var args struct {
		Content string `json:"content"` // 任务内容
	}

	// 解析 JSON 参数到结构体
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	log.Printf("[TaskClassifier] 分析任务: %s", args.Content)

	// 基于关键词的任务分类逻辑
	// 在实际应用中，这里可以使用更复杂的 NLP 算法或机器学习模型
	var taskType, complexity, priority string
	
	content := strings.ToLower(args.Content) // 转换为小写以便匹配
	switch {
	case strings.Contains(content, "分析") || strings.Contains(content, "analyze"):
		taskType = "analyze"   // 分析类任务
		complexity = "medium"  // 中等复杂度
		priority = "high"     // 高优先级
	case strings.Contains(content, "处理") || strings.Contains(content, "process"):
		taskType = "process"   // 处理类任务
		complexity = "high"    // 高复杂度
		priority = "medium"   // 中等优先级
	case strings.Contains(content, "报告") || strings.Contains(content, "report"):
		taskType = "report"    // 报告类任务
		complexity = "low"     // 低复杂度
		priority = "low"      // 低优先级
	default:
		taskType = "general"   // 通用任务
		complexity = "medium"  // 中等复杂度
		priority = "medium"   // 中等优先级
	}

	// 构建分类结果
	result := map[string]interface{}{
		"task_type":        taskType,                    // 任务类型
		"complexity":       complexity,                 // 复杂度评估
		"priority":         priority,                   // 优先级评估
		"estimated_time":   getEstimatedTime(complexity), // 预估执行时间
		"recommended_path": getRecommendedPath(taskType), // 推荐的处理路径
	}

	// 将结果序列化为 JSON 字符串返回
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// DataAnalysisTool 数据分析工具 - 专门处理数据分析类任务
// 支持多种分析类型：统计分析、趋势分析、相关性分析等
// 在 Graph 中作为分析分支的核心处理节点
type DataAnalysisTool struct{}

// Info 返回数据分析工具的元信息
func (t *DataAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "analyze_data",        // 工具名称
		Desc: "执行数据分析任务",        // 工具描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"data": {
				Type:     "string",      // 数据参数类型
				Desc:     "要分析的数据",    // 数据参数描述
				Required: true,         // 必需参数
			},
			"analysis_type": {
				Type:     "string",                                    // 分析类型参数
				Desc:     "分析类型",                                   // 参数描述
				Required: false,                                   // 可选参数
				Enum:     []string{"statistical", "trend", "correlation"}, // 枚举值：统计、趋势、相关性
			},
		}),
	}, nil
}

// InvokableRun 执行数据分析的具体逻辑
// 根据指定的分析类型对数据进行处理和分析
func (t *DataAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// 定义参数结构体
	var args struct {
		Data         string `json:"data"`          // 要分析的数据
		AnalysisType string `json:"analysis_type"` // 分析类型
	}

	// 解析 JSON 参数
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 设置默认分析类型
	if args.AnalysisType == "" {
		args.AnalysisType = "statistical" // 默认使用统计分析
	}

	log.Printf("[DataAnalysis] 执行 %s 分析: %s", args.AnalysisType, args.Data)

	// 模拟数据分析过程
	// 在实际应用中，这里会调用真实的数据分析算法
	time.Sleep(500 * time.Millisecond)

	// 构建分析结果
	// 这里返回模拟的分析结果，实际应用中会包含真实的计算结果
	result := map[string]interface{}{
		"analysis_type": args.AnalysisType,                                           // 分析类型
		"data_points":   10,                                                        // 数据点数量
		"mean":          75.5,                                                      // 平均值
		"median":        78.0,                                                      // 中位数
		"trend":         "increasing",                                              // 趋势
		"confidence":    0.85,                                                      // 置信度
		"summary":       fmt.Sprintf("对数据进行了%s分析，发现上升趋势", args.AnalysisType), // 分析摘要
	}

	// 返回 JSON 格式的分析结果
	resultBytes, _ := json.Marshal(result)
	return string(resultBytes), nil
}

// TextProcessorTool 文本处理工具 - 专门处理文本相关任务
// 支持多种文本操作：清理、提取、摘要等
// 在 Graph 中作为文本处理分支的核心节点
type TextProcessorTool struct{}

// Info 返回文本处理工具的元信息
func (t *TextProcessorTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "process_text",      // 工具名称
		Desc: "执行文本处理任务",      // 工具描述
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"text": {
				Type:     "string",    // 文本参数类型
				Desc:     "要处理的文本",  // 文本参数描述
				Required: true,       // 必需参数
			},
			"operation": {
				Type:     "string",                              // 操作类型参数
				Desc:     "处理操作",                             // 参数描述
				Required: false,                             // 可选参数
				Enum:     []string{"clean", "extract", "summarize"}, // 枚举值：清理、提取、摘要
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

// createGraphAgent 创建基于 Graph 的智能 Agent
// Graph Agent 的核心优势：
// 1. 支持复杂的分支逻辑和条件路由
// 2. 可以并行执行多个任务节点
// 3. 具备动态路径选择能力
// 4. 支持循环和回溯机制
// 返回一个可执行的 Graph Agent 实例
func createGraphAgent(ctx context.Context) (compose.Runnable[[]*schema.Message, []*schema.Message], error) {
	// 1. 创建工具集合和聊天模型
	// 工具是 Agent 执行具体任务的能力单元
	tools := createTools()
	chatModel, err := createChatModel(ctx, tools)
	if err != nil {
		return nil, err
	}

	// 2. 定义 Graph 中各个节点的名称常量
	// 使用常量可以避免字符串拼写错误，提高代码可维护性
	const (
		CLASSIFIER    = "classifier"      // 任务分类节点
		ANALYZER      = "analyzer"        // 数据分析节点
		PROCESSOR     = "processor"       // 文本处理节点
		REPORT_GEN    = "report_generator" // 报告生成节点
		QUALITY_CHECK = "quality_checker"  // 质量检查节点
		AGGREGATOR    = "aggregator"      // 结果聚合节点
	)

	// 3. 创建 Graph 实例
	// Graph 是 Eino 中用于构建复杂工作流的核心组件
	// 泛型参数指定输入和输出都是 Message 数组
	g := compose.NewGraph[[]*schema.Message, []*schema.Message]()

	// 4. 添加 ChatModel 节点 - 统一的 LLM 处理节点
	// ChatModel 节点负责与大语言模型交互，执行推理和工具调用
	g.AddChatModelNode(CLASSIFIER, chatModel)

	// 添加消息包装器节点处理类型转换
	// Lambda 节点用于执行自定义逻辑，这里用于消息格式转换
	g.AddLambdaNode("message_wrapper", compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) ([]*schema.Message, error) {
		// 将单个消息包装成消息数组，满足后续节点的输入要求
		return []*schema.Message{msg}, nil
	}))

	// 5. 添加条件处理器节点 - Graph 的核心路由逻辑
	// 这个节点展示了 Graph 相比 Chain 的优势：智能路由和条件分支
	g.AddLambdaNode("conditional_processor", compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
		// 检查消息数组是否为空
		if len(msgs) == 0 {
			return msgs, nil
		}

		// 获取最后一条消息进行内容分析
		lastMsg := msgs[len(msgs)-1]
		log.Printf("[ConditionalProcessor] 分析内容并执行处理: %s", lastMsg.Content)

		// 基于内容关键词进行智能路由决策
		content := strings.ToLower(lastMsg.Content)
		var processType, result string
		
		// 根据任务类型选择不同的处理路径
		// 这里展示了 Graph 的分支逻辑能力
		switch {
		case strings.Contains(content, "分析") || strings.Contains(content, "analyze"):
			processType = "数据分析"                                    // 选择数据分析路径
			result = "数据分析完成：发现关键趋势和模式，准确率达到95%"              // 模拟分析结果
		case strings.Contains(content, "处理") || strings.Contains(content, "process"):
			processType = "文本处理"                                    // 选择文本处理路径
			result = "文本处理完成：内容已清理和格式化，提取了关键信息"              // 模拟处理结果
		case strings.Contains(content, "报告") || strings.Contains(content, "report"):
			processType = "报告生成"                                    // 选择报告生成路径
			result = "报告生成完成：包含详细分析和建议，格式规范"                // 模拟报告结果
		default:
			processType = "综合处理"                                    // 默认综合处理路径
			result = "综合任务处理完成：包含分析、处理和报告生成的完整流程"           // 模拟综合结果
		}

		// 创建处理结果消息
		processResult := &schema.Message{
			Role:    schema.Assistant,                           // 设置为助手角色
			Content: fmt.Sprintf("[%s] %s", processType, result), // 格式化结果内容
		}

		// 将处理结果添加到消息链中
		return append(msgs, processResult), nil
	}))

	// 6. 添加质量检查节点 - 确保输出质量
	// 质量检查是 Graph 工作流中的重要环节，确保每个处理步骤的输出质量
	g.AddLambdaNode(QUALITY_CHECK, compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
		log.Printf("[QualityCheck] 执行质量检查")
		
		// 创建质量检查结果消息
		// 在实际应用中，这里会包含复杂的质量评估逻辑
		qualityResult := &schema.Message{
			Role:    schema.Assistant,                        // 助手角色
			Content: "质量检查通过：结果符合预期标准，质量分数: 95%", // 质量检查结果
		}
		return append(msgs, qualityResult), nil
	}))

	// 7. 添加最终聚合节点 - 汇总所有处理结果
	// 聚合节点负责整合整个 Graph 工作流的执行结果
	g.AddLambdaNode(AGGREGATOR, compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
		log.Printf("[Aggregator] 聚合最终结果")
		
		// 统计整个处理过程中的步骤数量
		var steps []string
		for _, msg := range msgs {
			// 只统计助手角色的消息（处理结果）
			if msg.Role == schema.Assistant {
				steps = append(steps, msg.Content)
			}
		}

		// 生成最终的处理摘要
		summary := fmt.Sprintf("Graph Agent 任务处理完成！\n执行了 %d 个处理步骤，所有质量检查均已通过。\n系统采用智能路由机制，根据任务类型自动选择最优处理路径。", len(steps))
		
		// 创建最终结果消息
		result := &schema.Message{
			Role:    schema.Assistant, // 助手角色
			Content: summary,         // 处理摘要
		}

		// 返回单个结果消息（聚合后的最终输出）
		return []*schema.Message{result}, nil
	}))

	// 8. 定义 Graph 的边关系（节点连接）
	// 边关系定义了数据在 Graph 中的流向，决定了执行顺序
	// 这里采用线性流水线模式，但 Graph 支持更复杂的分支和并行结构
	g.AddEdge(compose.START, CLASSIFIER)                    // 开始 -> 任务分类
	g.AddEdge(CLASSIFIER, "message_wrapper")                // 分类 -> 消息包装
	g.AddEdge("message_wrapper", "conditional_processor")   // 包装 -> 条件处理
	g.AddEdge("conditional_processor", QUALITY_CHECK)       // 处理 -> 质量检查
	g.AddEdge(QUALITY_CHECK, AGGREGATOR)                   // 检查 -> 结果聚合
	g.AddEdge(AGGREGATOR, compose.END)                     // 聚合 -> 结束

	// 9. 编译 Graph 生成可执行的 Agent
	// 编译过程会验证 Graph 的完整性和正确性
	agent, err := g.Compile(ctx, 
		compose.WithGraphName("GraphAgent"),                    // 设置 Graph 名称
		compose.WithNodeTriggerMode(compose.AnyPredecessor),   // 节点触发模式：任意前驱节点完成即可触发
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

// main 函数 - 程序入口点
// 演示 Graph Agent 的完整工作流程
func main() {
	ctx := context.Background()

	// 加载配置文件
	// 配置包含 ARK 模型的 API Key 等敏感信息
	loadConfig()
	log.Println("配置加载成功")

	// 创建 Graph Agent 实例
	// Graph Agent 是本演示的核心，展示复杂工作流能力
	agent, err := createGraphAgent(ctx)
	if err != nil {
		log.Fatalf("创建 Graph Agent 失败: %v", err)
	}

	// 运行演示场景
	// 通过多个场景展示 Graph Agent 的智能路由能力
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