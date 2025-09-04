package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// ==========================================
// CallOption 接口和实现 (简化版本)
// ==========================================

// CallOption 基础接口
type CallOption interface {
	Type() string
	Value() interface{}
	Apply(ctx context.Context) context.Context
}

// NodeDesignable 节点指定接口
type NodeDesignable interface {
	DesignateNode(nodeName string) CallOption
}

// 通用 ChatModel 选项
type CommonChatModelOption struct {
	Temperature *float64
	MaxTokens   *int
	TopP        *float64
	StopWords   []string
	targetNode  string
}

func (o *CommonChatModelOption) Type() string {
	return "CommonChatModel"
}

func (o *CommonChatModelOption) Value() interface{} {
	return map[string]interface{}{
		"temperature": o.Temperature,
		"max_tokens":  o.MaxTokens,
		"top_p":       o.TopP,
		"stop_words":  o.StopWords,
		"target_node": o.targetNode,
	}
}

func (o *CommonChatModelOption) Apply(ctx context.Context) context.Context {
	if o.Temperature != nil {
		ctx = context.WithValue(ctx, "temperature", *o.Temperature)
	}
	if o.MaxTokens != nil {
		ctx = context.WithValue(ctx, "max_tokens", *o.MaxTokens)
	}
	if o.TopP != nil {
		ctx = context.WithValue(ctx, "top_p", *o.TopP)
	}
	if len(o.StopWords) > 0 {
		ctx = context.WithValue(ctx, "stop_words", o.StopWords)
	}
	if o.targetNode != "" {
		ctx = context.WithValue(ctx, "target_node", o.targetNode)
	}
	return ctx
}

func (o *CommonChatModelOption) DesignateNode(nodeName string) CallOption {
	newOption := *o
	newOption.targetNode = nodeName
	return &newOption
}

// 选项构造函数
func WithTemperature(temp float64) *CommonChatModelOption {
	return &CommonChatModelOption{Temperature: &temp}
}

func WithMaxTokens(tokens int) *CommonChatModelOption {
	return &CommonChatModelOption{MaxTokens: &tokens}
}

func WithTopP(topP float64) *CommonChatModelOption {
	return &CommonChatModelOption{TopP: &topP}
}

func WithStopWords(words []string) *CommonChatModelOption {
	return &CommonChatModelOption{StopWords: words}
}

// 特定实现选项（模拟 Ark 特有选项）
type ArkSpecificOption struct {
	UseCache   *bool
	RetryCount *int
	Timeout    *time.Duration
	targetNode string
}

func (o *ArkSpecificOption) Type() string {
	return "ArkSpecific"
}

func (o *ArkSpecificOption) Value() interface{} {
	return map[string]interface{}{
		"use_cache":   o.UseCache,
		"retry_count": o.RetryCount,
		"timeout":     o.Timeout,
		"target_node": o.targetNode,
	}
}

func (o *ArkSpecificOption) Apply(ctx context.Context) context.Context {
	if o.UseCache != nil {
		ctx = context.WithValue(ctx, "ark_use_cache", *o.UseCache)
	}
	if o.RetryCount != nil {
		ctx = context.WithValue(ctx, "ark_retry_count", *o.RetryCount)
	}
	if o.Timeout != nil {
		ctx = context.WithValue(ctx, "ark_timeout", *o.Timeout)
	}
	if o.targetNode != "" {
		ctx = context.WithValue(ctx, "target_node", o.targetNode)
	}
	return ctx
}

func (o *ArkSpecificOption) DesignateNode(nodeName string) CallOption {
	newOption := *o
	newOption.targetNode = nodeName
	return &newOption
}

func WithArkCache(useCache bool) *ArkSpecificOption {
	return &ArkSpecificOption{UseCache: &useCache}
}

func WithArkRetryCount(count int) *ArkSpecificOption {
	return &ArkSpecificOption{RetryCount: &count}
}

func WithArkTimeout(timeout time.Duration) *ArkSpecificOption {
	return &ArkSpecificOption{Timeout: &timeout}
}

// ==========================================
// 业务数据结构
// ==========================================

type TaskRequest struct {
	Content   string `json:"content"`
	TaskType  string `json:"task_type"`  // creative, analytical, conversational
	UserLevel string `json:"user_level"` // beginner, intermediate, expert
	Urgency   string `json:"urgency"`    // low, medium, high
}

type ProcessedResult struct {
	Content        string        `json:"content"`
	ProcessingTime time.Duration `json:"processing_time"`
	TokensUsed     int           `json:"tokens_used"`
	ConfigUsed     ConfigSummary `json:"config_used"`
}

type ConfigSummary struct {
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	TopP        float64 `json:"top_p"`
	UseCache    bool    `json:"use_cache"`
	NodeName    string  `json:"node_name"`
}

// ==========================================
// 模拟的智能文本处理器
// ==========================================

type IntelligentTextProcessor struct {
	name string
}

func NewIntelligentTextProcessor(name string) *IntelligentTextProcessor {
	return &IntelligentTextProcessor{name: name}
}

func (p *IntelligentTextProcessor) Process(ctx context.Context, req *TaskRequest, options ...CallOption) (*ProcessedResult, error) {
	startTime := time.Now()

	// 应用所有选项到上下文
	processCtx := ctx
	for _, opt := range options {
		processCtx = opt.Apply(processCtx)
	}

	// 从上下文中提取配置
	config := p.extractConfig(processCtx)

	fmt.Printf("🔧 [%s] 使用配置: Temperature=%.1f, MaxTokens=%d, TopP=%.1f, Cache=%v\n",
		p.name, config.Temperature, config.MaxTokens, config.TopP, config.UseCache)

	// 根据配置模拟不同的处理逻辑
	processedContent := p.simulateProcessing(req, config)

	// 模拟 token 使用
	tokensUsed := p.calculateTokenUsage(req.Content, config)

	return &ProcessedResult{
		Content:        processedContent,
		ProcessingTime: time.Since(startTime),
		TokensUsed:     tokensUsed,
		ConfigUsed:     config,
	}, nil
}

func (p *IntelligentTextProcessor) extractConfig(ctx context.Context) ConfigSummary {
	config := ConfigSummary{
		Temperature: 0.7, // 默认值
		MaxTokens:   1000,
		TopP:        0.9,
		UseCache:    false,
		NodeName:    p.name,
	}

	// 从上下文提取配置
	if temp, ok := ctx.Value("temperature").(float64); ok {
		config.Temperature = temp
	}
	if tokens, ok := ctx.Value("max_tokens").(int); ok {
		config.MaxTokens = tokens
	}
	if topP, ok := ctx.Value("top_p").(float64); ok {
		config.TopP = topP
	}
	if cache, ok := ctx.Value("ark_use_cache").(bool); ok {
		config.UseCache = cache
	}
	if node, ok := ctx.Value("target_node").(string); ok && node != "" {
		config.NodeName = node
	}

	return config
}

func (p *IntelligentTextProcessor) simulateProcessing(req *TaskRequest, config ConfigSummary) string {
	baseContent := strings.ToUpper(req.Content)

	// 根据温度调整创意程度
	if config.Temperature > 0.8 {
		baseContent = fmt.Sprintf("✨创意版本: %s (灵感迸发!)", baseContent)
	} else if config.Temperature < 0.3 {
		baseContent = fmt.Sprintf("📊分析版本: %s (精确严谨)", baseContent)
	} else {
		baseContent = fmt.Sprintf("⚖️平衡版本: %s (稳定可靠)", baseContent)
	}

	// 根据任务类型调整
	switch req.TaskType {
	case "creative":
		baseContent += " [创意增强]"
	case "analytical":
		baseContent += " [逻辑分析]"
	case "conversational":
		baseContent += " [对话优化]"
	}

	// 根据缓存策略调整
	if config.UseCache {
		baseContent += " [缓存加速]"
	}

	return baseContent
}

func (p *IntelligentTextProcessor) calculateTokenUsage(content string, config ConfigSummary) int {
	baseTokens := len(strings.Fields(content)) * 2 // 简单估算

	// 根据最大 token 限制调整
	if baseTokens > config.MaxTokens {
		return config.MaxTokens
	}

	// 根据温度调整 token 使用（高温度可能产生更多token）
	multiplier := 1.0 + (config.Temperature-0.5)*0.5
	estimatedTokens := int(float64(baseTokens) * multiplier)

	if estimatedTokens > config.MaxTokens {
		return config.MaxTokens
	}

	return estimatedTokens
}

// ==========================================
// 选项配置预设
// ==========================================

// 创意写作预设
func CreativeWritingPreset() []CallOption {
	return []CallOption{
		WithTemperature(0.9),
		WithMaxTokens(2000),
		WithTopP(0.95),
		WithStopWords([]string{"结论", "总结"}),
	}
}

// 技术分析预设
func TechnicalAnalysisPreset() []CallOption {
	return []CallOption{
		WithTemperature(0.2),
		WithMaxTokens(1500),
		WithTopP(0.8),
		WithArkCache(true),
		WithArkRetryCount(3),
	}
}

// 对话交互预设
func ConversationalPreset() []CallOption {
	return []CallOption{
		WithTemperature(0.7),
		WithMaxTokens(1000),
		WithTopP(0.9),
		WithArkTimeout(10 * time.Second),
	}
}

// 动态选项选择器
func SelectOptionsForTask(taskType, userLevel, urgency string) []CallOption {
	var options []CallOption

	// 根据任务类型选择基础配置
	switch taskType {
	case "creative":
		options = append(options, CreativeWritingPreset()...)
	case "analytical":
		options = append(options, TechnicalAnalysisPreset()...)
	case "conversational":
		options = append(options, ConversationalPreset()...)
	default:
		options = append(options, ConversationalPreset()...)
	}

	// 根据用户级别调整
	switch userLevel {
	case "beginner":
		options = append(options, WithMaxTokens(500)) // 简化输出
	case "expert":
		options = append(options, WithMaxTokens(2000)) // 详细输出
	}

	// 根据紧急程度调整
	switch urgency {
	case "high":
		options = append(options,
			WithArkCache(true),
			WithArkTimeout(5*time.Second),
		)
	case "low":
		options = append(options,
			WithArkCache(false), // 不使用缓存，获得最新结果
		)
	}

	return options
}

// ==========================================
// 演示场景
// ==========================================

func demonstrateCallOptions() error {
	fmt.Println("🎯 === Eino CallOption 能力演示 ===\n")

	// 1. 基础选项演示
	fmt.Println("1️⃣  基础选项配置演示")
	if err := demonstrateBasicOptions(); err != nil {
		return err
	}

	// 2. 预设配置演示
	fmt.Println("\n2️⃣  预设配置演示")
	if err := demonstratePresetConfigs(); err != nil {
		return err
	}

	// 3. 动态选项选择演示
	fmt.Println("\n3️⃣  动态选项选择演示")
	if err := demonstrateDynamicOptions(); err != nil {
		return err
	}

	// 4. 节点特定配置演示
	fmt.Println("\n4️⃣  节点特定配置演示")
	if err := demonstrateNodeSpecificOptions(); err != nil {
		return err
	}

	// 5. 选项组合演示
	fmt.Println("\n5️⃣  选项组合演示")
	if err := demonstrateOptionCombination(); err != nil {
		return err
	}

	return nil
}

// 基础选项演示
func demonstrateBasicOptions() error {
	fmt.Println("   展示基础调用选项的使用...")

	processor := NewIntelligentTextProcessor("basic_processor")

	request := &TaskRequest{
		Content:   "请分析人工智能的发展趋势",
		TaskType:  "analytical",
		UserLevel: "intermediate",
		Urgency:   "medium",
	}

	// 使用基础选项
	options := []CallOption{
		WithTemperature(0.5),
		WithMaxTokens(800),
		WithTopP(0.85),
	}

	fmt.Printf("   🧪 处理任务: %s\n", request.Content)

	result, err := processor.Process(context.Background(), request, options...)
	if err != nil {
		return fmt.Errorf("基础选项演示失败: %w", err)
	}

	fmt.Printf("✅ 处理完成: %s\n", result.Content)
	fmt.Printf("   📊 使用Token: %d, 处理时间: %v\n",
		result.TokensUsed, result.ProcessingTime)

	return nil
}

// 预设配置演示
func demonstratePresetConfigs() error {
	fmt.Println("   展示预设配置的使用...")

	processor := NewIntelligentTextProcessor("preset_processor")

	// 测试不同预设
	testCases := []struct {
		name    string
		request *TaskRequest
		preset  func() []CallOption
	}{
		{
			name: "创意写作",
			request: &TaskRequest{
				Content:  "写一首关于春天的诗",
				TaskType: "creative",
			},
			preset: CreativeWritingPreset,
		},
		{
			name: "技术分析",
			request: &TaskRequest{
				Content:  "解释区块链技术原理",
				TaskType: "analytical",
			},
			preset: TechnicalAnalysisPreset,
		},
		{
			name: "对话交互",
			request: &TaskRequest{
				Content:  "你好，今天天气怎么样",
				TaskType: "conversational",
			},
			preset: ConversationalPreset,
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("   🧪 测试用例 %d: %s\n", i+1, testCase.name)

		options := testCase.preset()
		result, err := processor.Process(context.Background(), testCase.request, options...)
		if err != nil {
			return fmt.Errorf("预设配置演示失败 (%s): %w", testCase.name, err)
		}

		fmt.Printf("      ✅ 结果: %s\n", result.Content)
		fmt.Printf("      📊 配置: T=%.1f, Tokens=%d, 用时=%v\n",
			result.ConfigUsed.Temperature, result.TokensUsed, result.ProcessingTime)
	}

	return nil
}

// 动态选项选择演示
func demonstrateDynamicOptions() error {
	fmt.Println("   展示基于请求内容的动态选项选择...")

	processor := NewIntelligentTextProcessor("dynamic_processor")

	// 测试不同组合的请求
	testRequests := []*TaskRequest{
		{
			Content:   "创作一个科幻故事",
			TaskType:  "creative",
			UserLevel: "expert",
			Urgency:   "low",
		},
		{
			Content:   "分析股市走势",
			TaskType:  "analytical",
			UserLevel: "beginner",
			Urgency:   "high",
		},
		{
			Content:   "日常聊天对话",
			TaskType:  "conversational",
			UserLevel: "intermediate",
			Urgency:   "medium",
		},
	}

	for i, request := range testRequests {
		fmt.Printf("   🧪 动态请求 %d: %s\n", i+1, request.Content)
		fmt.Printf("      类型=%s, 级别=%s, 紧急=%s\n",
			request.TaskType, request.UserLevel, request.Urgency)

		// 动态选择选项
		options := SelectOptionsForTask(request.TaskType, request.UserLevel, request.Urgency)

		result, err := processor.Process(context.Background(), request, options...)
		if err != nil {
			return fmt.Errorf("动态选项演示失败: %w", err)
		}

		fmt.Printf("      ✅ 智能处理: %s\n", result.Content)
		fmt.Printf("      📊 动态配置: T=%.1f, Cache=%v, Tokens=%d\n",
			result.ConfigUsed.Temperature, result.ConfigUsed.UseCache, result.TokensUsed)
	}

	return nil
}

// 节点特定配置演示
func demonstrateNodeSpecificOptions() error {
	fmt.Println("   展示针对特定节点的配置...")

	// 创建多个不同角色的处理器
	researcher := NewIntelligentTextProcessor("researcher")
	creative_director := NewIntelligentTextProcessor("creative_director")
	editor := NewIntelligentTextProcessor("editor")

	request := &TaskRequest{
		Content:  "设计一个创新的移动应用",
		TaskType: "creative",
	}

	// 为不同节点配置不同选项
	researchOptions := []CallOption{
		WithTemperature(0.2).DesignateNode("researcher"),
		WithMaxTokens(1000).DesignateNode("researcher"),
	}

	creativeOptions := []CallOption{
		WithTemperature(0.9).DesignateNode("creative_director"),
		WithTopP(0.95).DesignateNode("creative_director"),
	}

	editOptions := []CallOption{
		WithTemperature(0.5).DesignateNode("editor"),
		WithMaxTokens(800).DesignateNode("editor"),
	}

	// 模拟多节点处理
	processors := []struct {
		processor *IntelligentTextProcessor
		options   []CallOption
		role      string
	}{
		{researcher, researchOptions, "🔬 研究分析"},
		{creative_director, creativeOptions, "🎨 创意设计"},
		{editor, editOptions, "✏️ 编辑优化"},
	}

	for _, p := range processors {
		fmt.Printf("   🧪 %s 节点处理...\n", p.role)

		result, err := p.processor.Process(context.Background(), request, p.options...)
		if err != nil {
			return fmt.Errorf("节点特定配置演示失败 (%s): %w", p.role, err)
		}

		fmt.Printf("      ✅ %s 结果: %s\n", p.role, result.Content)
		fmt.Printf("      📊 节点配置: T=%.1f, Node=%s\n",
			result.ConfigUsed.Temperature, result.ConfigUsed.NodeName)
	}

	return nil
}

// 选项组合演示
func demonstrateOptionCombination() error {
	fmt.Println("   展示选项的灵活组合...")

	processor := NewIntelligentTextProcessor("combo_processor")

	request := &TaskRequest{
		Content:  "综合分析与创意结合的项目方案",
		TaskType: "creative",
	}

	// 组合不同类型的选项
	combinedOptions := []CallOption{
		// 通用选项
		WithTemperature(0.6), // 平衡的温度
		WithMaxTokens(1500),  // 充足的token
		WithTopP(0.9),        // 适中的采样

		// 特定实现选项
		WithArkCache(true),               // 启用缓存
		WithArkRetryCount(2),             // 重试策略
		WithArkTimeout(15 * time.Second), // 超时设置
	}

	fmt.Printf("   🧪 组合选项处理: %s\n", request.Content)

	result, err := processor.Process(context.Background(), request, combinedOptions...)
	if err != nil {
		return fmt.Errorf("选项组合演示失败: %w", err)
	}

	fmt.Printf("✅ 组合处理结果: %s\n", result.Content)
	fmt.Printf("   📊 组合配置效果: T=%.1f, Tokens=%d, Cache=%v, 用时=%v\n",
		result.ConfigUsed.Temperature, result.TokensUsed,
		result.ConfigUsed.UseCache, result.ProcessingTime)

	return nil
}

// 主函数
func main() {
	fmt.Println("🎯 Eino CallOption 能力演示")
	fmt.Println("===============================")

	if err := demonstrateCallOptions(); err != nil {
		log.Fatalf("❌ 演示失败: %v", err)
	}

	fmt.Println("\n🎉 === 演示完成 ===")
	fmt.Println("📚 这个演示展示了:")
	fmt.Println("   • CallOption 基础概念和使用方法")
	fmt.Println("   • 通用选项 vs 特定实现选项")
	fmt.Println("   • 预设配置的模块化设计")
	fmt.Println("   • 动态选项选择策略")
	fmt.Println("   • 节点特定配置的精确控制")
	fmt.Println("   • 选项的灵活组合和复用")

	fmt.Println("\n💡 核心优势:")
	fmt.Println("   🎯 精确控制 - 细粒度的参数调整")
	fmt.Println("   🔧 灵活配置 - 运行时动态调整")
	fmt.Println("   📊 类型安全 - 编译时类型检查")
	fmt.Println("   🧩 可组合性 - 选项的自由组合")
	fmt.Println("   ⚡ 高性能 - 优化的选项处理机制")
}
